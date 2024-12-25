package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	_ "github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"thabomoyo.co.uk/cmd/web/routes"
	"thabomoyo.co.uk/internal/database"
	"thabomoyo.co.uk/internal/services"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

func dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s&readTimeout=5s&writeTimeout=5s&multiStatements=true&collation=utf8mb4_unicode_ci",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))
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

	//TODO:
	err = database.RunMigrations(db.DB, logger)
	if err != nil {
		logger.Error("Failed to run migrations: " + err.Error())
		os.Exit(1)
	}

	defer db.Close()

	services := services.NewServices(
		db,
		logger,
		*debug,
	)

	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	rh := routes.NewRouteHandler(services)

	srv := &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(*port),
		Handler:      rh.Routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	logger.Info("Starting server",
		slog.String("mode", os.Getenv("APP_MODE")),
		slog.Bool("debug", *debug),
		slog.Int("port", *port))

	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	if certFile == "" || keyFile == "" {
		logger.Error("TLS_CERT and TLS_KEY environment variables must be set")
		os.Exit(1)
	}

	// Start server
	logger.Info("starting server", "addr", srv.Addr)

	// Graceful shutdown setup
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	err = srv.ListenAndServeTLS(certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	err = <-shutdownError
	if err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("stopped server", "addr", srv.Addr)
}

func openDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.tmpl")
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
