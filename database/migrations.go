package database

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

func Migrate(db *sql.DB, fs embed.FS, dir string) error {
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
