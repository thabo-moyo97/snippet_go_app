package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"thabomoyo.co.uk/cmd/web/config"
	"thabomoyo.co.uk/cmd/web/routes"
	"thabomoyo.co.uk/internal/models"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

func main() {
	port := flag.Int("port", 8888, "Port to run the server on")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error("DB connection failed: " + err.Error())
		os.Exit(1)
	}
	logger.Info("DB connected")

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := *scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 15 * time.Minute
	sessionManager.Cookie.Secure = true

	app := config.Application{
		Logger:         logger,
		Snippets:       &models.SnippetModel{DB: db},
		Users:          &models.UserModel{DB: db},
		TemplateCache:  templateCache,
		FormDecoder:    form.NewDecoder(),
		SessionManager: &sessionManager,
		DebugMode:      *debug,
	}

	logger.Info("starting server on port", slog.Any("port", *port))

	cert, err := tls.LoadX509KeyPair("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		logger.Error("Failed to load TLS certificate and key: " + err.Error())
		os.Exit(1)
	}

	caCert, err := os.ReadFile("./tls/ca.pem")
	if err != nil {
		logger.Error("Failed to load CA certificate: " + err.Error())
		os.Exit(1)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		logger.Error("Failed to append CA certificate")
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		Certificates:     []tls.Certificate{cert},
		RootCAs:          caCertPool,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(*port),
		Handler:        routes.Routes(&app),
		ErrorLog:       slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:      tlsConfig,
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 524288,
	}

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
