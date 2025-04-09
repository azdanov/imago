package controllers

import (
	"net/http"

	"github.com/azdanov/imago/context"
	"github.com/azdanov/imago/models"
)

type NotificationType string

const (
	ErrorNotification   NotificationType = "error"
	SuccessNotification NotificationType = "success"
	InfoNotification    NotificationType = "info"
)

// NotificationMiddleware extracts notification parameters from URL query
// and adds them to the request context
type NotificationMiddleware struct{}

func NewNotificationMiddleware() *NotificationMiddleware {
	return &NotificationMiddleware{}
}

func (m *NotificationMiddleware) ExtractNotifications(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
			ctx = context.AddNotification(ctx, models.Notification{
				Type:    string(ErrorNotification),
				Message: errorMsg,
			})
		}

		if successMsg := r.URL.Query().Get("success"); successMsg != "" {
			ctx = context.AddNotification(ctx, models.Notification{
				Type:    string(SuccessNotification),
				Message: successMsg,
			})
		}

		if infoMsg := r.URL.Query().Get("info"); infoMsg != "" {
			ctx = context.AddNotification(ctx, models.Notification{
				Type:    string(InfoNotification),
				Message: infoMsg,
			})
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
