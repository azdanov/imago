package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/azdanov/imago/rand"
)

const MinSessionTokenBytes = 32

type Session struct {
	ID     uint `json:"id"`
	UserID uint `json:"user_id"`
	// Token is the actual token that will be sent to the client.
	// Only created once and never stored in the database.
	Token     string `json:"token"`
	TokenHash string `json:"-"`
}

type SessionService struct {
	DB *sql.DB
	// SessionTokenBytes is the number of bytes used to generate a session token.
	// If the value is less than MinSessionTokenBytes, MinSessionTokenBytes will be used.
	SessionTokenBytes int
}

func (s *SessionService) Create(userID uint) (*Session, error) {
	bytesPerToken := s.SessionTokenBytes
	if bytesPerToken < MinSessionTokenBytes {
		bytesPerToken = MinSessionTokenBytes
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	session := &Session{
		UserID: userID,
		Token:  token,
	}

	tokenHash := s.hashToken(token)

	err = s.DB.QueryRow(`
    INSERT INTO sessions (user_id, token_hash)
    VALUES ($1, $2)
    ON CONFLICT (user_id) DO UPDATE SET token_hash = $2
    RETURNING id
  `, session.UserID, tokenHash).Scan(&session.ID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) User(token string) (*User, error) {
	tokenHash := s.hashToken(token)

	user := &User{}
	err := s.DB.QueryRow(`
      SELECT u.id, u.email, u.password_hash
      FROM sessions s
      INNER JOIN users u ON s.user_id = u.id
      WHERE s.token_hash = $1
    `, tokenHash).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	return user, nil
}

func (s *SessionService) Delete(token string) error {
	tokenHash := s.hashToken(token)

	_, err := s.DB.Exec(`
    DELETE FROM sessions WHERE token_hash = $1
  `, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (s *SessionService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
