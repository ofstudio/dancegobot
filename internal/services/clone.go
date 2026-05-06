package services

import "github.com/ofstudio/dancegobot/internal/models"

func cloneProfilePtr(profile *models.Profile) *models.Profile {
	if profile == nil {
		return nil
	}
	cloned := *profile
	return &cloned
}

func cloneDancer(dancer models.Dancer) models.Dancer {
	cloned := dancer
	cloned.Profile = cloneProfilePtr(dancer.Profile)
	return cloned
}

func cloneCouple(couple models.Couple) models.Couple {
	cloned := couple
	if couple.Dancers != nil {
		cloned.Dancers = make([]models.Dancer, len(couple.Dancers))
		for i, dancer := range couple.Dancers {
			cloned.Dancers[i] = cloneDancer(dancer)
		}
	}
	return cloned
}

func cloneEvent(event *models.Event) *models.Event {
	if event == nil {
		return nil
	}
	cloned := *event
	if event.Post != nil {
		post := *event.Post
		if event.Post.Chat != nil {
			chat := *event.Post.Chat
			post.Chat = &chat
		}
		cloned.Post = &post
	}
	if event.Couples != nil {
		cloned.Couples = make([]models.Couple, len(event.Couples))
		for i, couple := range event.Couples {
			cloned.Couples[i] = cloneCouple(couple)
		}
	}
	if event.Singles != nil {
		cloned.Singles = make([]models.Dancer, len(event.Singles))
		for i, dancer := range event.Singles {
			cloned.Singles[i] = cloneDancer(dancer)
		}
	}
	return &cloned
}

func cloneHistoryItems(items []*models.HistoryItem) []*models.HistoryItem {
	if items == nil {
		return nil
	}
	cloned := make([]*models.HistoryItem, len(items))
	for i, item := range items {
		if item == nil {
			continue
		}
		itemClone := *item
		itemClone.Initiator = cloneProfilePtr(item.Initiator)
		if item.EventID != nil {
			eventID := *item.EventID
			itemClone.EventID = &eventID
		}
		itemClone.Details = cloneHistoryDetails(item.Details)
		cloned[i] = &itemClone
	}
	return cloned
}

func cloneHistoryDetails(details any) any {
	switch v := details.(type) {
	case models.Dancer:
		return cloneDancer(v)
	case *models.Dancer:
		if v == nil {
			return (*models.Dancer)(nil)
		}
		cloned := cloneDancer(*v)
		return &cloned
	case models.Couple:
		return cloneCouple(v)
	case *models.Couple:
		if v == nil {
			return (*models.Couple)(nil)
		}
		cloned := cloneCouple(*v)
		return &cloned
	case models.Event:
		return *cloneEvent(&v)
	case *models.Event:
		return cloneEvent(v)
	default:
		return details
	}
}

func cloneNotifications(notifications []*models.Notification) []*models.Notification {
	if notifications == nil {
		return nil
	}
	cloned := make([]*models.Notification, len(notifications))
	for i, notification := range notifications {
		if notification == nil {
			continue
		}
		notificationClone := *notification
		notificationClone.Recipient = cloneProfilePtr(notification.Recipient)
		notificationClone.Payload = cloneNotificationPayload(notification.Payload)
		cloned[i] = &notificationClone
	}
	return cloned
}

func cloneNotificationPayload(payload models.NotificationPayload) models.NotificationPayload {
	return models.NotificationPayload{
		Event:      cloneEvent(payload.Event),
		Partner:    cloneDancerPtr(payload.Partner),
		NewPartner: cloneDancerPtr(payload.NewPartner),
		WaitList:   payload.WaitList,
	}
}

func cloneDancerPtr(dancer *models.Dancer) *models.Dancer {
	if dancer == nil {
		return nil
	}
	cloned := cloneDancer(*dancer)
	return &cloned
}
