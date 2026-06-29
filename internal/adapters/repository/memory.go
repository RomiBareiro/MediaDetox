package repository

import (
	"MediaDetox/internal/domain"
	"sync"
)

type InMemoryEventRepository struct {
	mu    sync.RWMutex
	store map[string][]domain.Event
}

type InMemoryAlertRepository struct {
	mu    sync.RWMutex
	store map[string][]domain.Alert
}

type InMemoryRuleRepository struct {
	mu    sync.RWMutex
	rules []domain.Rule
}

func NewInMemoryEventRepository() *InMemoryEventRepository {
	return &InMemoryEventRepository{
		store: make(map[string][]domain.Event),
	}
}

func NewInMemoryAlertRepository() *InMemoryAlertRepository {
	return &InMemoryAlertRepository{
		store: make(map[string][]domain.Alert),
	}
}

func NewInMemoryRuleRepository() *InMemoryRuleRepository {
	return &InMemoryRuleRepository{
		rules: make([]domain.Rule, 0),
	}
}

func (r *InMemoryEventRepository) SaveEvent(event domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[event.UserID] = append(r.store[event.UserID], event)
	return nil
}

func (r *InMemoryEventRepository) GetEventsByUser(userID string, filter domain.EventFilter) ([]domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	events := append([]domain.Event(nil), r.store[userID]...)
	if filter.AppName != "" {
		filtered := make([]domain.Event, 0, len(events))
		for _, event := range events {
			if event.AppName == filter.AppName {
				filtered = append(filtered, event)
			}
		}
		events = filtered
	}
	if filter.StartDate != nil || filter.EndDate != nil {
		filtered := make([]domain.Event, 0, len(events))
		for _, event := range events {
			if filter.StartDate != nil && event.Timestamp.Before(*filter.StartDate) {
				continue
			}
			if filter.EndDate != nil && event.Timestamp.After(*filter.EndDate) {
				continue
			}
			filtered = append(filtered, event)
		}
		events = filtered
	}
	return events, nil
}

func (r *InMemoryAlertRepository) SaveAlert(alert domain.Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[alert.UserID] = append(r.store[alert.UserID], alert)
	return nil
}

func (r *InMemoryAlertRepository) GetAlertsByUser(userID string) ([]domain.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Alert(nil), r.store[userID]...), nil
}

func (r *InMemoryRuleRepository) SaveRule(rule domain.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
	return nil
}

func (r *InMemoryRuleRepository) GetRules() ([]domain.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Rule(nil), r.rules...), nil
}
