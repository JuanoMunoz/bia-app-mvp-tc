package ai

import (
	"context"
	"time"

	"learning/internal/event"
	"learning/internal/reading"
)

type DatasetRepository interface {
	AllReadings(context.Context) ([]reading.Reading, error)
	AllEvents(context.Context) ([]event.Event, error)
}

type AnalysisRepository interface {
	SaveAnalysis(context.Context, Analysis) error
	FindAnalysis(context.Context, string) (Analysis, error)
	LatestAnalysis(context.Context) (Analysis, error)
}

type InsightAnalyzer interface {
	Analyze(context.Context, AnalysisContext) (Insight, error)
}

type AIRecommendationPort interface {
	GenerateConsumptionInsight(context.Context, ConsumptionInsightInput) (ConsumptionInsightResult, error)
}

type InsightCacheRepository interface {
	GetRecentInsight(ctx context.Context, meterID string, from time.Time, to time.Time) (ConsumptionInsightResult, bool, error)
	SaveInsight(ctx context.Context, meterID string, from time.Time, to time.Time, insight ConsumptionInsightResult) error
}
