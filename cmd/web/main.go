package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // New import
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"thabomoyo.co.uk/internal/models"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

type application struct {
	logger   *slog.Logger
	snippets *models.SnippetModel
}

func main() {
	//commandline terminal flags
	port := flag.Int("port", 8888, "Port to run the server on")
	//commandline terminal dns
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	//structure logging = slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//db connection
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("DB connected")

	app := application{
		logger:   logger,
		snippets: &models.SnippetModel{DB: db},
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			os.Exit(1)
		}
	}(db)

	logger.Info("starting server on port", slog.Any("port", *port))

	err = http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(*port)), app.routes())

	logger.Error(err.Error())
	os.Exit(1)
}

func (app *application) InitializeTemplates(files ...string) (*template.Template, error) {
	baseTemplates := []string{
		"./ui/html/base.gohtml",
		"./ui/html/partials/nav.gohtml",
		"./ui/html/partials/footer.gohtml",
	}
	allFiles := append(baseTemplates, files...)
	return template.ParseFiles(allFiles...)
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
