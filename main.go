package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	httpadapter "MediaDetox/internal/adapters/http"
	metricsadapter "MediaDetox/internal/adapters/metrics"
	"MediaDetox/internal/adapters/repository"
	"MediaDetox/internal/ports"
	"MediaDetox/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

var logger = slog.New(
	slog.NewJSONHandler(os.Stdout, nil),
)

func main() {
	metrics := metricsadapter.NewPrometheusMetrics()

	var eventRepo ports.EventRepository = repository.NewInMemoryEventRepository()
	var alertRepo ports.AlertRepository = repository.NewInMemoryAlertRepository()
	var ruleRepo ports.RuleRepository = repository.NewInMemoryRuleRepository()

	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		db, err := pgxpool.New(context.Background(), databaseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		if err := db.Ping(context.Background()); err != nil {
			log.Fatal(err)
		}

		postgresRepo := repository.NewPostgresRepository(db)
		eventRepo = postgresRepo
		alertRepo = postgresRepo
		ruleRepo = postgresRepo
		logger.Info("using postgres repository")
	} else {
		logger.Info("using in-memory repository")
	}

	eventRepo = repository.NewInstrumentedEventRepository(eventRepo, metrics)
	alertRepo = repository.NewInstrumentedAlertRepository(alertRepo, metrics)
	ruleRepo = repository.NewInstrumentedRuleRepository(ruleRepo, metrics)

	service := usecase.NewUsageServiceWithMetrics(eventRepo, alertRepo, ruleRepo, metrics)
	handler := httpadapter.NewHandler(service, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	mux.Handle("/metrics", metrics.Handler())

	logger.Info("MediaDetox server running", "url", "http://localhost:8080")
	if err := http.ListenAndServe(":8080", metrics.Middleware(mux)); err != nil {
		log.Fatal(err)
	}
}
