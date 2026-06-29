package repository

import (
	"testing"
	"time"

	"MediaDetox/internal/domain"
)

func TestInMemoryEventRepository_GetEventsByUser(t *testing.T) {
	baseTime := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	start := baseTime.Add(-1 * time.Hour)
	end := baseTime.Add(1 * time.Hour)

	events := []domain.Event{
		{ID: "event-1", UserID: "user-1", AppName: "instagram", EventType: domain.EventTypeOpen, Timestamp: baseTime},
		{ID: "event-2", UserID: "user-1", AppName: "spotify", EventType: domain.EventTypeOpen, Timestamp: baseTime},
		{ID: "event-3", UserID: "user-1", AppName: "instagram", EventType: domain.EventTypeOpen, Timestamp: baseTime.Add(-2 * time.Hour)},
		{ID: "event-4", UserID: "user-2", AppName: "instagram", EventType: domain.EventTypeOpen, Timestamp: baseTime},
	}

	tests := []struct {
		name    string
		userID  string
		filter  domain.EventFilter
		wantIDs []string
	}{
		{
			name:    "GivenEventsForMultipleUsers_WhenGetEventsByUser_ThenOnlyUserEventsAreReturned",
			userID:  "user-1",
			filter:  domain.EventFilter{},
			wantIDs: []string{"event-1", "event-2", "event-3"},
		},
		{
			name:   "GivenAppNameFilter_WhenGetEventsByUser_ThenOnlyMatchingAppEventsAreReturned",
			userID: "user-1",
			filter: domain.EventFilter{
				AppName: "instagram",
			},
			wantIDs: []string{"event-1", "event-3"},
		},
		{
			name:   "GivenDateRangeFilter_WhenGetEventsByUser_ThenOnlyEventsInsideRangeAreReturned",
			userID: "user-1",
			filter: domain.EventFilter{
				StartDate: &start,
				EndDate:   &end,
			},
			wantIDs: []string{"event-1", "event-2"},
		},
		{
			name:   "GivenAppNameAndDateRangeFilters_WhenGetEventsByUser_ThenOnlyMatchingEventsAreReturned",
			userID: "user-1",
			filter: domain.EventFilter{
				StartDate: &start,
				EndDate:   &end,
				AppName:   "instagram",
			},
			wantIDs: []string{"event-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemoryEventRepository()
			for _, event := range events {
				if err := repo.SaveEvent(event); err != nil {
					t.Fatalf("SaveEvent() error = %v", err)
				}
			}

			got, err := repo.GetEventsByUser(tt.userID, tt.filter)
			if err != nil {
				t.Fatalf("GetEventsByUser() error = %v", err)
			}
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("len(events) = %d, want %d", len(got), len(tt.wantIDs))
			}
			for i, event := range got {
				if event.ID != tt.wantIDs[i] {
					t.Fatalf("events[%d].ID = %q, want %q", i, event.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestInMemoryRepositories_ReturnCopies(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "GivenStoredEvents_WhenReturnedSliceIsMutated_ThenRepositoryStateDoesNotChange",
			run: func(t *testing.T) {
				repo := NewInMemoryEventRepository()
				event := domain.Event{ID: "event-1", UserID: "user-1", AppName: "instagram", EventType: domain.EventTypeOpen, Timestamp: time.Now().UTC()}
				if err := repo.SaveEvent(event); err != nil {
					t.Fatalf("SaveEvent() error = %v", err)
				}

				got, err := repo.GetEventsByUser("user-1", domain.EventFilter{})
				if err != nil {
					t.Fatalf("GetEventsByUser() error = %v", err)
				}
				got[0].ID = "changed"

				gotAgain, err := repo.GetEventsByUser("user-1", domain.EventFilter{})
				if err != nil {
					t.Fatalf("GetEventsByUser() error = %v", err)
				}
				if gotAgain[0].ID != "event-1" {
					t.Fatalf("stored event ID = %q, want %q", gotAgain[0].ID, "event-1")
				}
			},
		},
		{
			name: "GivenStoredAlerts_WhenReturnedSliceIsMutated_ThenRepositoryStateDoesNotChange",
			run: func(t *testing.T) {
				repo := NewInMemoryAlertRepository()
				alert := domain.Alert{ID: "alert-1", UserID: "user-1", Type: "checking_loop", Severity: "medium", CreatedAt: time.Now().UTC()}
				if err := repo.SaveAlert(alert); err != nil {
					t.Fatalf("SaveAlert() error = %v", err)
				}

				got, err := repo.GetAlertsByUser("user-1")
				if err != nil {
					t.Fatalf("GetAlertsByUser() error = %v", err)
				}
				got[0].ID = "changed"

				gotAgain, err := repo.GetAlertsByUser("user-1")
				if err != nil {
					t.Fatalf("GetAlertsByUser() error = %v", err)
				}
				if gotAgain[0].ID != "alert-1" {
					t.Fatalf("stored alert ID = %q, want %q", gotAgain[0].ID, "alert-1")
				}
			},
		},
		{
			name: "GivenStoredRules_WhenReturnedSliceIsMutated_ThenRepositoryStateDoesNotChange",
			run: func(t *testing.T) {
				repo := NewInMemoryRuleRepository()
				rule := domain.Rule{ID: "rule-1", Name: "checking-loop", AppName: "instagram", MaxOpens: 10, WindowMinutes: 30, Enabled: true}
				if err := repo.SaveRule(rule); err != nil {
					t.Fatalf("SaveRule() error = %v", err)
				}

				got, err := repo.GetRules()
				if err != nil {
					t.Fatalf("GetRules() error = %v", err)
				}
				got[0].ID = "changed"

				gotAgain, err := repo.GetRules()
				if err != nil {
					t.Fatalf("GetRules() error = %v", err)
				}
				if gotAgain[0].ID != "rule-1" {
					t.Fatalf("stored rule ID = %q, want %q", gotAgain[0].ID, "rule-1")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
