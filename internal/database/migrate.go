package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migration struct {
	Version  int
	UpSQL    string
	Filename string
}

func RunMigrations(db *sql.DB, logger *slog.Logger) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	migrations, err := loadMigrations(db)
	if err != nil {
		return fmt.Errorf("error loading migrations: %w", err)
	}

	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	if len(migrations) == 0 || currentVersion == migrations[len(migrations)-1].Version {
		logger.Info("Migrations are up to date")
		return nil
	}

	logger.Info("Running migrations", "count", len(migrations))
	return applyMigrations(db, logger, migrations, currentVersion)
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("error getting current version: %w", err)
	}
	return version, nil
}

func loadMigrations(db *sql.DB) ([]Migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	migrations := loadMigrationFiles(entries)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	if err := validateMigrations(db, migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}

func loadMigrationFiles(entries []fs.DirEntry) []Migration {
	var migrations []Migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			version, err := extractVersion(entry.Name())
			if err != nil {
				continue
			}

			content, err := migrationFiles.ReadFile(filepath.Join("migrations", entry.Name()))
			if err != nil {
				continue
			}

			migrations = append(migrations, Migration{
				Version:  version,
				UpSQL:    string(content),
				Filename: entry.Name(),
			})
		}
	}
	return migrations
}

func validateMigrations(db *sql.DB, migrations []Migration) error {
	for i := 0; i < len(migrations); i++ {
		expectedVersion := i + 1
		if migrations[i].Version != expectedVersion {
			return fmt.Errorf("missing migration file for version %d", expectedVersion)
		}
	}

	existingMigrations := make(map[int]bool)
	for _, m := range migrations {
		existingMigrations[m.Version] = true
	}

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("error querying schema_migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("error scanning migration version: %w", err)
		}
		if !existingMigrations[version] {
			return fmt.Errorf("orphaned migration %d", version)
		}
	}

	return nil
}

func applyMigrations(db *sql.DB, logger *slog.Logger, migrations []Migration, currentVersion int) error {
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := applyMigration(db, logger, migration); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyMigration(db *sql.DB, logger *slog.Logger, migration Migration) error {
	logger.Info("running migration", "version", migration.Version, "file", migration.Filename)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range strings.Split(migration.UpSQL, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("error running migration %d: %w", migration.Version, err)
		}
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version); err != nil {
		return fmt.Errorf("error recording migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing migration %d: %w", migration.Version, err)
	}

	logger.Info("migration complete ", "version", migration.Version)
	return nil
}

func extractVersion(filename string) (int, error) {
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return 0, errors.New("invalid migration filename format")
	}

	var version int
	_, err := fmt.Sscanf(parts[0], "%d", &version)
	if err != nil {
		return 0, err
	}

	return version, nil
}
