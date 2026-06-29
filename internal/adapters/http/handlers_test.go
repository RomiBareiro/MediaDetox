package httpadapter

import (
	"net/http/httptest"
	"testing"
)

func TestParseEventFilter(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		wantAppName string
		wantStart   bool
		wantEnd     bool
		wantErr     string
	}{
		{
			name:        "GivenAppNameQuery_WhenParseEventFilter_ThenAppNameFilterIsReturned",
			target:      "/users/user-1/events?app_name=instagram",
			wantAppName: "instagram",
		},
		{
			name:      "GivenStartAndEndDateQuery_WhenParseEventFilter_ThenDateRangeFilterIsReturned",
			target:    "/users/user-1/events?start_date=2026-06-01&end_date=2026-06-29",
			wantStart: true,
			wantEnd:   true,
		},
		{
			name:    "GivenInvalidStartDateQuery_WhenParseEventFilter_ThenValidationErrorIsReturned",
			target:  "/users/user-1/events?start_date=06-01-2026",
			wantErr: "start_date must be YYYY-MM-DD",
		},
		{
			name:    "GivenInvalidEndDateQuery_WhenParseEventFilter_ThenValidationErrorIsReturned",
			target:  "/users/user-1/events?end_date=06-29-2026",
			wantErr: "end_date must be YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", tt.target, nil)

			got, err := parseEventFilter(request)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseEventFilter() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("parseEventFilter() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseEventFilter() error = %v", err)
			}
			if got.AppName != tt.wantAppName {
				t.Fatalf("AppName = %q, want %q", got.AppName, tt.wantAppName)
			}
			if (got.StartDate != nil) != tt.wantStart {
				t.Fatalf("StartDate present = %t, want %t", got.StartDate != nil, tt.wantStart)
			}
			if (got.EndDate != nil) != tt.wantEnd {
				t.Fatalf("EndDate present = %t, want %t", got.EndDate != nil, tt.wantEnd)
			}
		})
	}
}
