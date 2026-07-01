package httpadapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"MediaDetox/internal/domain"
	"MediaDetox/internal/usecase"
	"log/slog"
)

const (
	ErrInvalidJSON      = "invalid json"
	ErrMethodNotAllowed = "method not allowed"
	ErrNotFound         = "not found"
)
const (
	UsersEvents  = "events"
	UsersMetrics = "metrics"
	UsersAlerts  = "alerts"
)

type Handler struct {
	service *usecase.UsageService
	logger  *slog.Logger
}

func NewHandler(service *usecase.UsageService, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/events", h.handleEvents)
	mux.HandleFunc("/rules", h.handleRules)
	mux.HandleFunc("/users/", h.handleUserRoutes)
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.replyError(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var payload registerEventRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.replyError(w, http.StatusBadRequest, ErrInvalidJSON)
		return
	}

	timestamp, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, "invalid timestamp format")
		return
	}

	event := domain.Event{
		ID:        generateID(),
		UserID:    payload.UserID,
		AppName:   payload.AppName,
		EventType: domain.EventType(payload.EventType),
		Timestamp: timestamp.UTC(),
	}

	if err := h.service.RegisterEvent(event); err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.replyJSON(w, http.StatusCreated, event)
}

func (h *Handler) handleRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.replyError(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var payload createRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.replyError(w, http.StatusBadRequest, ErrInvalidJSON)
		return
	}

	rule, err := h.service.CreateRule(domain.Rule{
		AppName:       payload.AppName,
		MaxOpens:      payload.MaxOpens,
		WindowMinutes: payload.WindowMinutes,
	})
	if err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.replyJSON(w, http.StatusCreated, rule)
}

func (h *Handler) handleUserRoutes(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/users/") {
		h.replyError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "users" {
		h.replyError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	userID := parts[1]
	switch parts[2] {
	case UsersEvents:
		if r.Method != http.MethodGet {
			h.replyError(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}
		h.handleGetEvents(w, r, userID)
	case UsersAlerts:
		if r.Method != http.MethodGet {
			h.replyError(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}
		h.handleGetAlerts(w, userID)
	case UsersMetrics:
		if r.Method != http.MethodGet {
			h.replyError(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}
		h.handleGetMetrics(w, userID)
	default:
		h.replyError(w, http.StatusNotFound, ErrNotFound)
	}
}

func (h *Handler) handleGetEvents(w http.ResponseWriter, r *http.Request, userID string) {
	filter, err := parseEventFilter(r)
	if err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	events, err := h.service.QueryEvents(userID, filter)
	if err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.replyJSON(w, http.StatusOK, events)
}

func (h *Handler) handleGetAlerts(w http.ResponseWriter, userID string) {
	alerts, err := h.service.QueryAlerts(userID)
	if err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.replyJSON(w, http.StatusOK, alerts)
}

func (h *Handler) handleGetMetrics(w http.ResponseWriter, userID string) {
	metrics, err := h.service.CalculateMetrics(userID)
	if err != nil {
		h.replyError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.replyJSON(w, http.StatusOK, metrics)
}

func parseEventFilter(r *http.Request) (domain.EventFilter, error) {
	var filter domain.EventFilter

	if start := r.URL.Query().Get("start_date"); start != "" {
		parsed, err := time.Parse("2006-01-02", start)
		if err != nil {
			return filter, errors.New("start_date must be YYYY-MM-DD")
		}
		startTime := parsed.UTC()
		filter.StartDate = &startTime
	}

	if end := r.URL.Query().Get("end_date"); end != "" {
		parsed, err := time.Parse("2006-01-02", end)
		if err != nil {
			return filter, errors.New("end_date must be YYYY-MM-DD")
		}
		endTime := parsed.UTC().Add(24*time.Hour - time.Nanosecond)
		filter.EndDate = &endTime
	}

	filter.AppName = r.URL.Query().Get("app_name")
	return filter, nil
}

func (h *Handler) replyJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) replyError(w http.ResponseWriter, status int, message string) {
	h.logger.Warn("http error", "status", status, "message", message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Code:    status,
		Message: message,
	})
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
}
