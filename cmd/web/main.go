package main

import (
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"thabomoyo.co.uk/internal/models"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	//commandline terminal flags
	port := flag.Int("port", 8888, "Port to run the server on")
	//move to environment variables
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	//structure logging = slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//db connection
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

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 10 * time.Minute // 10 minutes

	app := application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    form.NewDecoder(),
		sessionManager: sessionManager,
	}

	logger.Info("starting server on port", slog.Any("port", *port))

	srv := &http.Server{
		Addr:     ":" + strconv.Itoa(*port),
		Handler:  app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
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

// Open /
/**
 * The neuteredFileSystem type is a custom implementation of the http.FileSystem
 * interface. It wraps an existing http.FileSystem to remove the ability to
 * navigate directories on the server's file system.
 */
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
