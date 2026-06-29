package domain

import "time"

type EventType string

const (
	EventTypeOpen  EventType = "open"
	EventTypeClose EventType = "close"
)

type Event struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	AppName   string    `json:"app_name"`
	EventType EventType `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

type Alert struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Rule struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	AppName       string `json:"app_name"`
	MaxOpens      int    `json:"max_opens"`
	WindowMinutes int    `json:"window_minutes"`
	Enabled       bool   `json:"enabled"`
}

type Metrics struct {
	OpensToday  int    `json:"opens_today"`
	MostUsedApp string `json:"most_used_app"`
	AlertsToday int    `json:"alerts_today"`
}

type EventFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	AppName   string
}
