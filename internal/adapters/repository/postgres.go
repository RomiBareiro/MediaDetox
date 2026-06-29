package repository

import (
	"context"

	"MediaDetox/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveEvent(event domain.Event) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO events (id, user_id, app_name, event_type, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`, event.ID, event.UserID, event.AppName, string(event.EventType), event.Timestamp.UTC())
	return err
}

func (r *PostgresRepository) GetEventsByUser(userID string, filter domain.EventFilter) ([]domain.Event, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, user_id, app_name, event_type, timestamp
		FROM events
		WHERE user_id = $1
		  AND ($2 = '' OR app_name = $2)
		  AND ($3::timestamptz IS NULL OR timestamp >= $3)
		  AND ($4::timestamptz IS NULL OR timestamp <= $4)
		ORDER BY timestamp ASC
	`, userID, filter.AppName, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.ID, &event.UserID, &event.AppName, &event.EventType, &event.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *PostgresRepository) SaveAlert(alert domain.Alert) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO alerts (id, user_id, type, severity, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, alert.ID, alert.UserID, alert.Type, alert.Severity, alert.Message, alert.CreatedAt.UTC())
	return err
}

func (r *PostgresRepository) GetAlertsByUser(userID string) ([]domain.Alert, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, user_id, type, severity, message, created_at
		FROM alerts
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]domain.Alert, 0)
	for rows.Next() {
		var alert domain.Alert
		if err := rows.Scan(&alert.ID, &alert.UserID, &alert.Type, &alert.Severity, &alert.Message, &alert.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

func (r *PostgresRepository) SaveRule(rule domain.Rule) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO rules (id, name, app_name, max_opens, window_minutes, enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, rule.ID, rule.Name, rule.AppName, rule.MaxOpens, rule.WindowMinutes, rule.Enabled)
	return err
}

func (r *PostgresRepository) GetRules() ([]domain.Rule, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, name, app_name, max_opens, window_minutes, enabled
		FROM rules
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]domain.Rule, 0)
	for rows.Next() {
		var rule domain.Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.AppName, &rule.MaxOpens, &rule.WindowMinutes, &rule.Enabled); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
