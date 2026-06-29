package ports

import (
	"time"

	"MediaDetox/internal/domain"
)

type MetricsRecorder interface {
	RecordEventRegistered(event domain.Event)
	RecordAlertGenerated(alert domain.Alert)
	RecordRepositoryOperation(repository, operation string, duration time.Duration, err error)
}

type NoopMetricsRecorder struct{}

func (NoopMetricsRecorder) RecordEventRegistered(domain.Event) {}

func (NoopMetricsRecorder) RecordAlertGenerated(domain.Alert) {}

func (NoopMetricsRecorder) RecordRepositoryOperation(string, string, time.Duration, error) {}
