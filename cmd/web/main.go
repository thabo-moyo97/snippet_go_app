package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"

	_ "github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"thabomoyo.co.uk/cmd/web/config"
	"thabomoyo.co.uk/cmd/web/routes"
	"thabomoyo.co.uk/internal/database"
	"thabomoyo.co.uk/internal/models"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

func dsn() string {
	return os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?parseTime=true&timeout=5s&readTimeout=5s&writeTimeout=5s&multiStatements=true"
}

func main() {
	port := flag.Int("port", getEnvAsInt("APP_PORT", 8888), "Port to run the server on")
	dsn := dsn()
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(dsn)
	if err != nil {
		logger.Error("DB connection failed: " + err.Error())
		os.Exit(1)
	}
	logger.Info("DB connected")

	// Run migrations
	err = database.RunMigrations(db, logger)
	if err != nil {
		logger.Error("Failed to run migrations: " + err.Error())
		os.Exit(1)
	}

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

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	srv := &http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(*port),
		Handler:        routes.Routes(&app),
		ErrorLog:       slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:      tlsConfig,
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 524288,
	}

	logger.Info("Starting server in " + os.Getenv("APP_MODE") + " mode and debug mode is " + strconv.FormatBool(*debug))
	err = srv.ListenAndServe()
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

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
