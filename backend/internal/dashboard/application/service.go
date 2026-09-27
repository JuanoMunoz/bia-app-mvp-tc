package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	"learning/internal/meter"
)

type MeterReader interface {
	List(context.Context) ([]meter.Summary, error)
}

type AnomalyReader interface {
	List(context.Context) ([]anomaly.Anomaly, error)
}

type AnalysisReader interface {
	GetLatestAnalysis(context.Context) (ai.Analysis, error)
}

type ReadingPeriodReader interface {
	ReadingPeriod(context.Context) (time.Time, time.Time, error)
}

type Summary struct {
	MeterCount          int
	TotalConsumptionKWh float64
	AnomalyCount        int
	HighPriorityCount   int
	AggregateConfidence float64
	LatestAnalysis      *ai.Analysis
	PeriodStart         time.Time
	PeriodEnd           time.Time
}

type Service struct {
	meters    MeterReader
	anomalies AnomalyReader
	analysis  AnalysisReader
	period    ReadingPeriodReader
}

func NewService(meters MeterReader, anomalies AnomalyReader, analysis AnalysisReader) *Service {
	return &Service{meters: meters, anomalies: anomalies, analysis: analysis}
}

func NewServiceWithPeriod(meters MeterReader, anomalies AnomalyReader, analysis AnalysisReader, period ReadingPeriodReader) *Service {
	return &Service{meters: meters, anomalies: anomalies, analysis: analysis, period: period}
}

func (service *Service) GetSummary(ctx context.Context) (Summary, error) {
	meters, err := service.meters.List(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("load dashboard meters: %w", err)
	}
	anomalies, err := service.anomalies.List(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("load dashboard anomalies: %w", err)
	}
	latest, err := service.analysis.GetLatestAnalysis(ctx)
	if err != nil && !errors.Is(err, ai.ErrNotFound) {
		return Summary{}, fmt.Errorf("load latest analysis: %w", err)
	}

	result := Summary{MeterCount: len(meters), AnomalyCount: len(anomalies)}
	if service.period != nil {
		result.PeriodStart, result.PeriodEnd, err = service.period.ReadingPeriod(ctx)
		if err != nil {
			return Summary{}, fmt.Errorf("load reading period: %w", err)
		}
	}
	if err == nil {
		result.LatestAnalysis = &latest
	}
	for _, item := range meters {
		result.TotalConsumptionKWh += item.PeriodConsumptionKWh
	}
	for _, item := range anomalies {
		result.AggregateConfidence += item.Confidence
		if item.Severity == anomaly.High || item.Severity == anomaly.Critical {
			result.HighPriorityCount++
		}
	}
	if result.AnomalyCount > 0 {
		result.AggregateConfidence /= float64(result.AnomalyCount)
	}
	return result, nil
}
