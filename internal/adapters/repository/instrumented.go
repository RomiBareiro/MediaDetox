package repository

import (
	"time"

	"MediaDetox/internal/domain"
	"MediaDetox/internal/ports"
)

type InstrumentedEventRepository struct {
	next    ports.EventRepository
	metrics ports.MetricsRecorder
}

func NewInstrumentedEventRepository(next ports.EventRepository, metrics ports.MetricsRecorder) *InstrumentedEventRepository {
	return &InstrumentedEventRepository{next: next, metrics: metrics}
}

func (r *InstrumentedEventRepository) SaveEvent(event domain.Event) error {
	start := time.Now()
	err := r.next.SaveEvent(event)
	r.metrics.RecordRepositoryOperation("events", "save", time.Since(start), err)
	return err
}

func (r *InstrumentedEventRepository) GetEventsByUser(userID string, filter domain.EventFilter) ([]domain.Event, error) {
	start := time.Now()
	events, err := r.next.GetEventsByUser(userID, filter)
	r.metrics.RecordRepositoryOperation("events", "get_by_user", time.Since(start), err)
	return events, err
}

type InstrumentedAlertRepository struct {
	next    ports.AlertRepository
	metrics ports.MetricsRecorder
}

func NewInstrumentedAlertRepository(next ports.AlertRepository, metrics ports.MetricsRecorder) *InstrumentedAlertRepository {
	return &InstrumentedAlertRepository{next: next, metrics: metrics}
}

func (r *InstrumentedAlertRepository) SaveAlert(alert domain.Alert) error {
	start := time.Now()
	err := r.next.SaveAlert(alert)
	r.metrics.RecordRepositoryOperation("alerts", "save", time.Since(start), err)
	return err
}

func (r *InstrumentedAlertRepository) GetAlertsByUser(userID string) ([]domain.Alert, error) {
	start := time.Now()
	alerts, err := r.next.GetAlertsByUser(userID)
	r.metrics.RecordRepositoryOperation("alerts", "get_by_user", time.Since(start), err)
	return alerts, err
}

type InstrumentedRuleRepository struct {
	next    ports.RuleRepository
	metrics ports.MetricsRecorder
}

func NewInstrumentedRuleRepository(next ports.RuleRepository, metrics ports.MetricsRecorder) *InstrumentedRuleRepository {
	return &InstrumentedRuleRepository{next: next, metrics: metrics}
}

func (r *InstrumentedRuleRepository) SaveRule(rule domain.Rule) error {
	start := time.Now()
	err := r.next.SaveRule(rule)
	r.metrics.RecordRepositoryOperation("rules", "save", time.Since(start), err)
	return err
}

func (r *InstrumentedRuleRepository) GetRules() ([]domain.Rule, error) {
	start := time.Now()
	rules, err := r.next.GetRules()
	r.metrics.RecordRepositoryOperation("rules", "get_all", time.Since(start), err)
	return rules, err
}
