package ports

import "MediaDetox/internal/domain"

type EventRepository interface {
	SaveEvent(domain.Event) error
	GetEventsByUser(userID string, filter domain.EventFilter) ([]domain.Event, error)
}

type AlertRepository interface {
	SaveAlert(domain.Alert) error
	GetAlertsByUser(userID string) ([]domain.Alert, error)
}

type RuleRepository interface {
	SaveRule(domain.Rule) error
	GetRules() ([]domain.Rule, error)
}
