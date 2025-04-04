package database

import (
	"database/sql"
	"fmt"

	"github.com/azdanov/imago/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(config config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.DB.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("database: %w", err)
	}

	return db, nil
}
