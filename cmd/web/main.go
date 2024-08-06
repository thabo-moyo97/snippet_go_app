package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

type application struct {
	logger *slog.Logger
}

func main() {
	port := flag.Int("port", 8888, "Port to run the server on")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := application{
		logger: logger,
	}

	logger.Info("starting server on port", slog.Any("port", *port))

	err := http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(*port)), app.routes())

	logger.Error(err.Error())
	os.Exit(1)
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
