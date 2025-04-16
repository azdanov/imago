package models

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type UserService struct {
	DB *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		DB: db,
	}
}

func (us *UserService) Create(email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	u := User{
		Email:        email,
		PasswordHash: string(hash),
	}

	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	err = us.DB.QueryRow(query, u.Email, u.PasswordHash).Scan(&u.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create: %w", err)
	}

	return &u, nil
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	u := User{
		Email: email,
	}

	query := `SELECT id, password_hash FROM users WHERE email = $1`
	err := us.DB.QueryRow(query, email).Scan(&u.ID, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return &u, nil
}

func (us *UserService) UpdatePassword(userID int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err = us.DB.Exec(query, hash, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
