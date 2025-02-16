package controllers

import (
	"fmt"
	"net/http"
	"os"
)

type SessionCookie struct{}

const SessionName = "session"

func (c SessionCookie) new(value string) http.Cookie {
	cookie := http.Cookie{
		Name:     SessionName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
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
