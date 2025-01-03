package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type Template struct {
	htmlTpl *template.Template
	Data    any
}

func Parse(fs fs.FS, pattern string) (*Template, error) {
	t, err := template.ParseFS(fs, pattern)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &Template{htmlTpl: t}, nil
}

func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

func (t Template) Execute(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.htmlTpl.Execute(w, data)
	if err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
	}
}
