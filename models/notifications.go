package models

import "sort"

type Notification struct {
	Type    string
	Message string
}

// SortNotifications sorts notifications by priority: success > error > others
func SortNotifications(nl []Notification) []Notification {
	if nl == nil {
		return nil
	}

	sortedNotifications := make([]Notification, len(nl))
	copy(sortedNotifications, nl)

	sort.SliceStable(sortedNotifications, func(i, j int) bool {
		if sortedNotifications[i].Type == "success" && sortedNotifications[j].Type != "success" {
			return true
		}
		if sortedNotifications[i].Type != "success" && sortedNotifications[j].Type == "success" {
			return false
		}
		if sortedNotifications[i].Type == "error" && sortedNotifications[j].Type != "error" && sortedNotifications[j].Type != "success" {
			return true
		}
		return false
	})

	return sortedNotifications
}
