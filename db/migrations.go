package db

import (
	"fmt"

	"github.com/pressly/goose/v3"
)

const migrationsDir = "migrations"

func Migrate() error {
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if err := goose.Up(DB, migrationsDir); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
