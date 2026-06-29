package usecase

import (
	"errors"
	"testing"
	"time"

	"MediaDetox/internal/adapters/repository"
	"MediaDetox/internal/domain"
)

func TestUsageService_RegisterEvent(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name               string
		event              domain.Event
		rule               *domain.Rule
		wantErr            string
		wantEvents         int
		wantAlerts         int
		wantRecordedEvents int
		wantRecordedAlerts int
	}{
		{
			name: "GivenValidOpenEvent_WhenRegisterEvent_ThenEventIsSavedAndMetricRecorded",
			event: domain.Event{
				ID:        "event-1",
				UserID:    "user-1",
				AppName:   "instagram",
				EventType: domain.EventTypeOpen,
				Timestamp: now,
			},
			wantEvents:         1,
			wantRecordedEvents: 1,
		},
		{
			name: "GivenMatchingRule_WhenRegisterEventExceedsMaxOpens_ThenAlertIsGeneratedAndMetricRecorded",
			event: domain.Event{
				ID:        "event-2",
				UserID:    "user-1",
				AppName:   "instagram",
				EventType: domain.EventTypeOpen,
				Timestamp: now,
			},
			rule: &domain.Rule{
				ID:            "rule-1",
				Name:          "checking-loop",
				AppName:       "instagram",
				MaxOpens:      1,
				WindowMinutes: 10,
				Enabled:       true,
			},
			wantEvents:         1,
			wantAlerts:         1,
			wantRecordedEvents: 1,
			wantRecordedAlerts: 1,
		},
		{
			name: "GivenMissingUserID_WhenRegisterEvent_ThenValidationErrorIsReturned",
			event: domain.Event{
				ID:        "event-3",
				AppName:   "instagram",
				EventType: domain.EventTypeOpen,
				Timestamp: now,
			},
			wantErr: "user_id is required",
		},
		{
			name: "GivenInvalidEventType_WhenRegisterEvent_ThenValidationErrorIsReturned",
			event: domain.Event{
				ID:        "event-4",
				UserID:    "user-1",
				AppName:   "instagram",
				EventType: domain.EventType("invalid"),
				Timestamp: now,
			},
			wantErr: "event_type must be open or close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := repository.NewInMemoryEventRepository()
			alertRepo := repository.NewInMemoryAlertRepository()
			ruleRepo := repository.NewInMemoryRuleRepository()
			metrics := &spyMetricsRecorder{}
			service := NewUsageServiceWithMetrics(eventRepo, alertRepo, ruleRepo, metrics)

			if tt.rule != nil {
				if err := ruleRepo.SaveRule(*tt.rule); err != nil {
					t.Fatalf("SaveRule() error = %v", err)
				}
			}

			err := service.RegisterEvent(tt.event)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("RegisterEvent() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("RegisterEvent() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RegisterEvent() error = %v", err)
			}

			events, err := eventRepo.GetEventsByUser(tt.event.UserID, domain.EventFilter{})
			if err != nil {
				t.Fatalf("GetEventsByUser() error = %v", err)
			}
			if len(events) != tt.wantEvents {
				t.Fatalf("len(events) = %d, want %d", len(events), tt.wantEvents)
			}

			alerts, err := alertRepo.GetAlertsByUser(tt.event.UserID)
			if err != nil {
				t.Fatalf("GetAlertsByUser() error = %v", err)
			}
			if len(alerts) != tt.wantAlerts {
				t.Fatalf("len(alerts) = %d, want %d", len(alerts), tt.wantAlerts)
			}
			if metrics.eventsRegistered != tt.wantRecordedEvents {
				t.Fatalf("eventsRegistered = %d, want %d", metrics.eventsRegistered, tt.wantRecordedEvents)
			}
			if metrics.alertsGenerated != tt.wantRecordedAlerts {
				t.Fatalf("alertsGenerated = %d, want %d", metrics.alertsGenerated, tt.wantRecordedAlerts)
			}
		})
	}
}

func TestUsageService_CalculateMetrics(t *testing.T) {
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name    string
		events  []domain.Event
		alerts  []domain.Alert
		userID  string
		want    *domain.Metrics
		wantErr string
	}{
		{
			name:    "GivenEmptyUserID_WhenCalculateMetrics_ThenValidationErrorIsReturned",
			userID:  "",
			wantErr: "user id is required",
		},
		{
			name:   "GivenEventsAndAlerts_WhenCalculateMetrics_ThenUsageSummaryIsReturned",
			userID: "user-1",
			events: []domain.Event{
				{ID: "event-1", UserID: "user-1", AppName: "instagram", EventType: domain.EventTypeOpen, Timestamp: now},
				{ID: "event-2", UserID: "user-1", AppName: "instagram", EventType: domain.EventTypeClose, Timestamp: now},
				{ID: "event-3", UserID: "user-1", AppName: "spotify", EventType: domain.EventTypeOpen, Timestamp: yesterday},
			},
			alerts: []domain.Alert{
				{ID: "alert-1", UserID: "user-1", Type: "checking_loop", Severity: "medium", CreatedAt: now},
				{ID: "alert-2", UserID: "user-1", Type: "checking_loop", Severity: "medium", CreatedAt: yesterday},
			},
			want: &domain.Metrics{
				OpensToday:  1,
				MostUsedApp: "instagram",
				AlertsToday: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := repository.NewInMemoryEventRepository()
			alertRepo := repository.NewInMemoryAlertRepository()
			ruleRepo := repository.NewInMemoryRuleRepository()
			service := NewUsageService(eventRepo, alertRepo, ruleRepo)

			for _, event := range tt.events {
				if err := eventRepo.SaveEvent(event); err != nil {
					t.Fatalf("SaveEvent() error = %v", err)
				}
			}
			for _, alert := range tt.alerts {
				if err := alertRepo.SaveAlert(alert); err != nil {
					t.Fatalf("SaveAlert() error = %v", err)
				}
			}

			got, err := service.CalculateMetrics(tt.userID)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("CalculateMetrics() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("CalculateMetrics() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CalculateMetrics() error = %v", err)
			}
			if *got != *tt.want {
				t.Fatalf("CalculateMetrics() = %+v, want %+v", *got, *tt.want)
			}
		})
	}
}

func TestUsageService_CreateRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    domain.Rule
		wantErr string
	}{
		{
			name: "GivenValidRule_WhenCreateRule_ThenRuleIsEnabledAndSaved",
			rule: domain.Rule{
				AppName:       "instagram",
				MaxOpens:      10,
				WindowMinutes: 30,
			},
		},
		{
			name: "GivenMissingAppName_WhenCreateRule_ThenValidationErrorIsReturned",
			rule: domain.Rule{
				MaxOpens:      10,
				WindowMinutes: 30,
			},
			wantErr: "app_name is required",
		},
		{
			name: "GivenInvalidMaxOpens_WhenCreateRule_ThenValidationErrorIsReturned",
			rule: domain.Rule{
				AppName:       "instagram",
				MaxOpens:      0,
				WindowMinutes: 30,
			},
			wantErr: "max_opens must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := repository.NewInMemoryEventRepository()
			alertRepo := repository.NewInMemoryAlertRepository()
			ruleRepo := repository.NewInMemoryRuleRepository()
			service := NewUsageService(eventRepo, alertRepo, ruleRepo)

			got, err := service.CreateRule(tt.rule)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("CreateRule() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("CreateRule() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateRule() error = %v", err)
			}
			if got.ID == "" {
				t.Fatal("CreateRule() ID is empty")
			}
			if got.Name == "" {
				t.Fatal("CreateRule() Name is empty")
			}
			if !got.Enabled {
				t.Fatal("CreateRule() Enabled = false, want true")
			}

			rules, err := ruleRepo.GetRules()
			if err != nil {
				t.Fatalf("GetRules() error = %v", err)
			}
			if len(rules) != 1 {
				t.Fatalf("len(rules) = %d, want 1", len(rules))
			}
		})
	}
}

func TestUsageService_QueryEvents(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr string
	}{
		{
			name:   "GivenUserID_WhenQueryEvents_ThenRepositoryResultIsReturned",
			userID: "user-1",
		},
		{
			name:    "GivenEmptyUserID_WhenQueryEvents_ThenValidationErrorIsReturned",
			userID:  "",
			wantErr: "user id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUsageService(
				repository.NewInMemoryEventRepository(),
				repository.NewInMemoryAlertRepository(),
				repository.NewInMemoryRuleRepository(),
			)

			_, err := service.QueryEvents(tt.userID, domain.EventFilter{})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("QueryEvents() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("QueryEvents() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("QueryEvents() error = %v", err)
			}
		})
	}
}

func TestUsageService_RegisterEventRepositoryErrors(t *testing.T) {
	tests := []struct {
		name    string
		service *UsageService
		wantErr string
	}{
		{
			name: "GivenEventRepositoryFails_WhenRegisterEvent_ThenErrorIsReturned",
			service: NewUsageService(
				&failingEventRepository{saveErr: errors.New("save failed")},
				repository.NewInMemoryAlertRepository(),
				repository.NewInMemoryRuleRepository(),
			),
			wantErr: "save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.RegisterEvent(domain.Event{
				ID:        "event-1",
				UserID:    "user-1",
				AppName:   "instagram",
				EventType: domain.EventTypeOpen,
				Timestamp: time.Now().UTC(),
			})
			if err == nil {
				t.Fatalf("RegisterEvent() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("RegisterEvent() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

type spyMetricsRecorder struct {
	eventsRegistered int
	alertsGenerated  int
}

func (r *spyMetricsRecorder) RecordEventRegistered(domain.Event) {
	r.eventsRegistered++
}

func (r *spyMetricsRecorder) RecordAlertGenerated(domain.Alert) {
	r.alertsGenerated++
}

func (r *spyMetricsRecorder) RecordRepositoryOperation(string, string, time.Duration, error) {}

type failingEventRepository struct {
	saveErr error
}

func (r *failingEventRepository) SaveEvent(domain.Event) error {
	return r.saveErr
}

func (r *failingEventRepository) GetEventsByUser(string, domain.EventFilter) ([]domain.Event, error) {
	return nil, nil
}
