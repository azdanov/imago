package db

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		DB.Close()
		return err
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
