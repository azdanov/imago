package models

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint   `json:"id"`
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
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	u := User{
		Email:        email,
		PasswordHash: string(hash),
	}
	err = us.DB.QueryRow(query, u.Email, u.PasswordHash).Scan(&u.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &u, nil
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	query := `SELECT id, password_hash FROM users WHERE email = $1`

	u := User{
		Email: email,
	}
	err := us.DB.QueryRow(query, email).Scan(&u.ID, &u.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return &u, nil
}
