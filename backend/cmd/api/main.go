package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"learning/internal/ai"
	aiapp "learning/internal/ai/application"
	anomalyapp "learning/internal/anomaly/application"
	dashboardapp "learning/internal/dashboard/application"
	"learning/internal/infrastructure/config"
	"learning/internal/infrastructure/db"
	geminiadapter "learning/internal/infrastructure/gemini"
	"learning/internal/infrastructure/httpServer"
	"learning/internal/infrastructure/report"
	"learning/internal/meter"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

type meterProfileAdapter struct {
	repository *db.MeterProfileRepository
}

func (adapter meterProfileAdapter) MeterProfile(ctx context.Context, meterID string) (meter.Profile, error) {
	profile, err := adapter.repository.GetMeterProfile(ctx, meterID)
	if err != nil {
		return meter.Profile{}, err
	}
	return meter.Profile{Provider: profile.Provider, Region: profile.Region, RateType: profile.RateType}, nil
}

func run() error {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repository, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer repository.Close()

	if err := repository.Migrate(ctx, cfg.MigrationsSQL); err != nil {
		return err
	}

	readingRepository := db.NewReadingRepository(repository)
	anomalyRepository := db.NewAnomalyRepository(repository)
	analysisRepository := db.NewAnalysisRepository(repository)
	insightCacheRepository := db.NewInsightCacheRepository(repository)
	tariffRepository := db.NewTariffRepository(repository)
	billingRepository := db.NewBillingRepository(repository)
	meterProfileRepository := db.NewMeterProfileRepository(repository)
	consumptionEstimator := db.NewConsumptionEstimator(repository)
	var insightAnalyzer ai.InsightAnalyzer
	var aiRecommendationPort ai.AIRecommendationPort

	if strings.TrimSpace(cfg.GeminiAPIKey) != "" {
		client := geminiadapter.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel)
		insightAnalyzer = client
		aiRecommendationPort = client
	}
	meters := meter.NewService(readingRepository, anomalyRepository, meterProfileAdapter{repository: meterProfileRepository})
	anomalies := anomalyapp.NewService(anomalyRepository)
	analysis := ai.NewService(analysisRepository, analysisRepository, insightAnalyzer)
	dashboard := dashboardapp.NewServiceWithPeriod(meters, anomalies, analysis, repository)
	billingService := httpServer.NewBillingService(
		tariffRepository,
		billingRepository,
		meterProfileRepository,
		consumptionEstimator,
		report.NewGenerator(),
	)
	optimizationUseCase := aiapp.NewGetConsumptionOptimizationUseCase(
		billingRepository,
		tariffRepository,
		meterProfileRepository,
		readingRepository,
		anomalyRepository,
		insightCacheRepository,
		aiRecommendationPort,
	)
	billingService.SetOptimizationInsightUseCase(optimizationUseCase)
	router, err := httpServer.NewRouterWithDashboardAndHealth(meters, anomalies, analysis, dashboard, repository, meterProfileRepository, logger, cfg.AllowedOrigins, billingService)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 10 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return err
		}
		return nil
	}
}
