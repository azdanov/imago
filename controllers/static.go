package controllers

import (
	"html/template"
	"net/http"

	"github.com/azdanov/imago/views"
)

func StaticHandler(tmpl *views.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func FAQ(tmpl *views.Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   template.HTML
	}{
		{
			Question: "What is Imago?",
			Answer:   "Imago is a web application that allows you to upload images and share them with others.",
		},
		{
			Question: "How do I upload an image?",
			Answer:   "You can upload an image by clicking on the upload button and selecting an image from your computer.",
		},
		{
			Question: "How do I share an image?",
			Answer:   "You can share an image by clicking on the share button and copying the link to the image.",
		},
		{
			Question: "How do I contact support?",
			Answer:   `You can contact support by sending an email to <a href="mailto:contact@example.com">contact@example.com</a>.`,
		},
		{
			Question: "Where is your office located?",
			Answer:   "Our office is located at 123 Main Street, Anytown, USA.",
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, questions)
	}
}
