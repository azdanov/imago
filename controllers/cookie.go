package controllers

import (
	"fmt"
	"net/http"
)

type SessionCookie struct {
	Secure bool
}

func NewSessionCookie(secure bool) *SessionCookie {
	return &SessionCookie{
		Secure: secure,
	}
}

const SessionName = "session"

func (c SessionCookie) new(value string) http.Cookie {
	cookie := http.Cookie{
		Name:     SessionName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
	}
	return cookie
}

func (c SessionCookie) Set(w http.ResponseWriter, token string) {
	cookie := c.new(token)
	http.SetCookie(w, &cookie)
}

func (c SessionCookie) Get(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionName)
	if err != nil {
		return "", fmt.Errorf("get cookie: %w", err)
	}
	return cookie.Value, nil
}

func (c SessionCookie) Clear(w http.ResponseWriter) {
	cookie := c.new("")
	cookie.MaxAge = -1
	http.SetCookie(w, &cookie)
}
