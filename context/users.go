package context

import (
	"context"

	"github.com/azdanov/imago/models"
)

type key int

const (
	userKey key = iota
	notificationsKey
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	user, ok := ctx.Value(userKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func AddNotification(ctx context.Context, notification models.Notification) context.Context {
	notifications := Notifications(ctx)
	if notifications == nil {
		notifications = []models.Notification{notification}
	} else {
		notifications = append(notifications, notification)
	}
	return WithNotifications(ctx, notifications)
}

func WithNotifications(ctx context.Context, notifications []models.Notification) context.Context {
	return context.WithValue(ctx, notificationsKey, notifications)
}

func Notifications(ctx context.Context) []models.Notification {
	notifications, ok := ctx.Value(notificationsKey).([]models.Notification)
	if !ok {
		return nil
	}
	return notifications
}
