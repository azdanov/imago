package database

import (
	"database/sql"
	"fmt"

	"github.com/azdanov/imago/config"
)

func NewDB(cnf *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cnf.DB.GetDSN())
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
