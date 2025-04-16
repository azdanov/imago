package models

import "sort"

type Notification struct {
	Type    string
	Message string
}

const (
	NotificationSuccess = "success"
	NotificationError   = "error"
)

// SortNotifications sorts notifications by priority: success > error > others.
func SortNotifications(nl []Notification) []Notification {
	if nl == nil {
		return nil
	}

	sortedNotifications := make([]Notification, len(nl))
	copy(sortedNotifications, nl)

	sort.SliceStable(sortedNotifications, func(i, j int) bool {
		if sortedNotifications[i].Type == NotificationSuccess &&
			sortedNotifications[j].Type != NotificationSuccess {
			return true
		}
		if sortedNotifications[i].Type != NotificationSuccess &&
			sortedNotifications[j].Type == NotificationSuccess {
			return false
		}
		if sortedNotifications[i].Type == NotificationError &&
			sortedNotifications[j].Type != NotificationError &&
			sortedNotifications[j].Type != NotificationSuccess {
			return true
		}
		return false
	})

	return sortedNotifications
}
