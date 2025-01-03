package views

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Template struct {
	htmlTpl *template.Template
	Data    any
}

func Parse(filename string) (*Template, error) {
	p := filepath.Join("templates", filename+".tmpl.html")
	t, err := template.ParseFiles(p)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", filename, err)
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
