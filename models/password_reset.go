package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/azdanov/imago/rand"
)

const (
	DefaultTokenLifetime = 1 * time.Hour
)

type PasswordReset struct {
	ID     int
	UserID int
	// Token is only created initially and never stored in the database.
	Token     string
	TokenHash string
	CreatedAt time.Time
}

type PasswordResetService struct {
	DB *sql.DB
	// SessionTokenBytes is the number of bytes used to generate a session token.
	// If the value is less than MinSessionTokenBytes, MinSessionTokenBytes will be used.
	BytesPerToken int
	TokenLifetime time.Duration
}

func NewPasswordResetService(db *sql.DB, bytesPerToken int, tokenLifetime time.Duration) *PasswordResetService {
	return &PasswordResetService{
		DB:            db,
		BytesPerToken: bytesPerToken,
		TokenLifetime: tokenLifetime,
	}
}

func (s *PasswordResetService) Generate(email string) (*PasswordReset, error) {
	var userID int
	query := s.DB.QueryRow(`SELECT id FROM users WHERE email = $1;`, email)

	err := query.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	bytesPerToken := s.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = MinSessionTokenBytes
	}

	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	duration := s.TokenLifetime
	if duration == 0 {
		duration = DefaultTokenLifetime
	}

	resetToken := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: s.hash(token),
		CreatedAt: time.Now(),
	}

	query = s.DB.QueryRow(`
		INSERT INTO reset_tokens (user_id, token_hash, created_at)
		VALUES ($1, $2, $3) ON CONFLICT (user_id)
		DO UPDATE SET token_hash = $2, created_at = $3
		RETURNING id;`, resetToken.UserID, resetToken.TokenHash, resetToken.CreatedAt)

	err = query.Scan(&resetToken.ID)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	return &resetToken, nil
}

func (s *PasswordResetService) GetUserByToken(token string) (*User, error) {
	tokenHash := s.hash(token)

	var user User
	var passwordReset PasswordReset

	query := s.DB.QueryRow(`
			SELECT rt.id, rt.created_at, u.id, u.email, u.password_hash
			FROM reset_tokens rt
			JOIN users u ON u.id = rt.user_id
			WHERE rt.token_hash = $1;`, tokenHash)

	err := query.Scan(&passwordReset.ID, &passwordReset.CreatedAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if time.Now().After(passwordReset.CreatedAt.Add(s.TokenLifetime)) {
		return nil, fmt.Errorf("token expired: %v", token)
	}

	err = s.delete(passwordReset.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (prs *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))

	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (prs *PasswordResetService) delete(id int) error {
	_, err := prs.DB.Exec(`DELETE FROM reset_tokens WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
