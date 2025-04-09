package controllers

import (
	"net/http"
	"net/url"
)

// RedirectWithNotification redirects to the given path with notification parameters
func RedirectWithNotification(
	w http.ResponseWriter,
	r *http.Request,
	path string,
	notificationType NotificationType,
	message string,
	params map[string]string,
) {
	values := url.Values{}

	switch notificationType {
	case ErrorNotification:
		values.Set("error", message)
	case SuccessNotification:
		values.Set("success", message)
	case InfoNotification:
		values.Set("info", message)
	}

	// Add any additional parameters
	for k, v := range params {
		values.Set(k, v)
	}

	http.Redirect(w, r, path+"?"+values.Encode(), http.StatusSeeOther)
}
