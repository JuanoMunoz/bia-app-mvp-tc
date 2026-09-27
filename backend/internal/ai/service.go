package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"learning/internal/anomaly"
	"learning/internal/event"
	"learning/internal/reading"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("analysis not found")

type Service struct {
	source   DatasetRepository
	store    AnalysisRepository
	insight  InsightAnalyzer
	detector anomaly.Detector
}

func NewService(source DatasetRepository, store AnalysisRepository, insight InsightAnalyzer) *Service {
	return &Service{source: source, store: store, insight: insight, detector: anomaly.Detector{}}
}

func (service *Service) Analyze(ctx context.Context) (Analysis, error) {
	result := Analysis{ID: uuid.NewString(), Status: "RUNNING", StartedAt: time.Now().UTC(), Anomalies: []anomaly.Anomaly{}}
	if err := service.store.SaveAnalysis(ctx, result); err != nil {
		return Analysis{}, fmt.Errorf("save running analysis: %w", err)
	}

	readings, err := service.source.AllReadings(ctx)
	var events []event.Event
	if err == nil {
		events, err = service.source.AllEvents(ctx)
		if err == nil {
			result.Anomalies = service.detector.Detect(readings, events)
		}
	}
	if err != nil {
		result.Status = "FAILED"
		result.Error = "No se pudo completar el análisis de datos."
		if saveErr := service.store.SaveAnalysis(ctx, result); saveErr != nil {
			return Analysis{}, fmt.Errorf("save failed analysis: %w", saveErr)
		}
		return result, fmt.Errorf("analyze dataset: %w", err)
	}

	for index := range result.Anomalies {
		result.Anomalies[index].ID = uuid.NewString()
	}
	if service.insight != nil {
		insight, insightErr := service.insight.Analyze(ctx, AnalysisContext{
			ReadingCount: len(readings),
			MeterCount:   countMeters(readings),
			Readings:     readings,
			Events:       events,
			Anomalies:    result.Anomalies,
		})
		if insightErr != nil {
			result.InsightError = "No fue posible generar la explicación del análisis."
		} else {
			result.Insight = &insight
		}
	}
	completedAt := time.Now().UTC()
	result.Status = "COMPLETED"
	result.CompletedAt = &completedAt
	if err := service.store.SaveAnalysis(ctx, result); err != nil {
		return Analysis{}, fmt.Errorf("save completed analysis: %w", err)
	}
	return result, nil
}

func countMeters(readings []reading.Reading) int {
	meters := make(map[string]struct{})
	for _, item := range readings {
		meters[item.MeterID] = struct{}{}
	}
	return len(meters)
}

func (service *Service) GetLatestAnalysis(ctx context.Context) (Analysis, error) {
	result, err := service.store.LatestAnalysis(ctx)
	if err != nil {
		return Analysis{}, fmt.Errorf("find latest analysis: %w", err)
	}
	return result, nil
}

func (service *Service) GetAnalysis(ctx context.Context, id string) (Analysis, error) {
	result, err := service.store.FindAnalysis(ctx, id)
	if err != nil {
		return Analysis{}, fmt.Errorf("find analysis: %w", err)
	}
	return result, nil
}
