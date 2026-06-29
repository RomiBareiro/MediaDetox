package usecase

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"MediaDetox/internal/domain"
	"MediaDetox/internal/ports"
)

type UsageService struct {
	events  ports.EventRepository
	alerts  ports.AlertRepository
	rules   ports.RuleRepository
	metrics ports.MetricsRecorder
}

func NewUsageService(events ports.EventRepository, alerts ports.AlertRepository, rules ports.RuleRepository) *UsageService {
	return NewUsageServiceWithMetrics(events, alerts, rules, ports.NoopMetricsRecorder{})
}

func NewUsageServiceWithMetrics(events ports.EventRepository, alerts ports.AlertRepository, rules ports.RuleRepository, metrics ports.MetricsRecorder) *UsageService {
	return &UsageService{
		events:  events,
		alerts:  alerts,
		rules:   rules,
		metrics: metrics,
	}
}

func (s *UsageService) RegisterEvent(event domain.Event) error {
	if err := validateEvent(event); err != nil {
		return err
	}

	if err := s.events.SaveEvent(event); err != nil {
		return err
	}
	s.metrics.RecordEventRegistered(event)

	return s.evaluateRules(event)
}

func validateEvent(event domain.Event) error {
	if event.UserID == "" {
		return errors.New("user_id is required")
	}
	if event.AppName == "" {
		return errors.New("app_name is required")
	}
	if event.EventType != domain.EventTypeOpen && event.EventType != domain.EventTypeClose {
		return errors.New("event_type must be open or close")
	}
	if event.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}

func (s *UsageService) QueryEvents(userID string, filter domain.EventFilter) ([]domain.Event, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}
	return s.events.GetEventsByUser(userID, filter)
}

func (s *UsageService) QueryAlerts(userID string) ([]domain.Alert, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}
	return s.alerts.GetAlertsByUser(userID)
}

func (s *UsageService) CalculateMetrics(userID string) (*domain.Metrics, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	events, err := s.events.GetEventsByUser(userID, domain.EventFilter{})
	if err != nil {
		return nil, err
	}

	opensToday := 0
	alerts, err := s.alerts.GetAlertsByUser(userID)
	if err != nil {
		return nil, err
	}

	countByApp := map[string]int{}
	today := time.Now().UTC()
	for _, event := range events {
		if event.EventType == domain.EventTypeOpen && sameDay(event.Timestamp, today) {
			opensToday++
			countByApp[event.AppName]++
		}
		countByApp[event.AppName]++
	}

	mostUsedApp := ""
	if len(countByApp) > 0 {
		mostUsedApp = mostUsedAppFromCounts(countByApp)
	}

	alertsToday := 0
	for _, alert := range alerts {
		if sameDay(alert.CreatedAt, today) {
			alertsToday++
		}
	}

	return &domain.Metrics{
		OpensToday:  opensToday,
		MostUsedApp: mostUsedApp,
		AlertsToday: alertsToday,
	}, nil
}

func (s *UsageService) CreateRule(rule domain.Rule) (domain.Rule, error) {
	if rule.AppName == "" {
		return domain.Rule{}, errors.New("app_name is required")
	}
	if rule.MaxOpens <= 0 {
		return domain.Rule{}, errors.New("max_opens must be greater than 0")
	}
	if rule.WindowMinutes <= 0 {
		return domain.Rule{}, errors.New("window_minutes must be greater than 0")
	}

	rule.ID = generateID()
	if rule.Name == "" {
		rule.Name = fmt.Sprintf("rule-%s", rule.ID)
	}
	rule.Enabled = true

	if err := s.rules.SaveRule(rule); err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

func (s *UsageService) evaluateRules(event domain.Event) error {
	rules, err := s.rules.GetRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if !rule.Enabled || rule.AppName != event.AppName {
			continue
		}
		if event.EventType != domain.EventTypeOpen {
			continue
		}

		windowStart := event.Timestamp.Add(-time.Duration(rule.WindowMinutes) * time.Minute)
		filtered, err := s.events.GetEventsByUser(event.UserID, domain.EventFilter{
			StartDate: &windowStart,
			EndDate:   &event.Timestamp,
			AppName:   event.AppName,
		})
		if err != nil {
			return err
		}

		opens := 0
		for _, candidate := range filtered {
			if candidate.EventType == domain.EventTypeOpen {
				opens++
			}
		}

		if opens >= rule.MaxOpens {
			alert := domain.Alert{
				ID:        generateID(),
				UserID:    event.UserID,
				Type:      "checking_loop",
				Severity:  "medium",
				Message:   fmt.Sprintf("%s opened %d times in %d minutes", event.AppName, opens, rule.WindowMinutes),
				CreatedAt: time.Now().UTC(),
			}
			if err := s.alerts.SaveAlert(alert); err != nil {
				return err
			}
			s.metrics.RecordAlertGenerated(alert)
		}
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
}

func sameDay(value, reference time.Time) bool {
	y, m, d := value.UTC().Date()
	rY, rM, rD := reference.UTC().Date()
	return y == rY && m == rM && d == rD
}

func mostUsedAppFromCounts(counts map[string]int) string {
	type pair struct {
		app   string
		count int
	}
	list := make([]pair, 0, len(counts))
	for app, count := range counts {
		list = append(list, pair{app: app, count: count})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count == list[j].count {
			return list[i].app < list[j].app
		}
		return list[i].count > list[j].count
	})
	return list[0].app
}
