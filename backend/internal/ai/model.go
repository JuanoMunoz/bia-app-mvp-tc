package ai

import (
	"learning/internal/anomaly"
	"learning/internal/event"
	"learning/internal/reading"
	"time"
)

type Analysis struct {
	ID           string
	Status       string
	StartedAt    time.Time
	CompletedAt  *time.Time
	Anomalies    []anomaly.Anomaly
	Error        string
	Insight      *Insight
	InsightError string
}

type Insight struct {
	Answer             string
	Explanation        string
	SuggestedQuestions []SuggestedQuestion
}

type SuggestedQuestion struct {
	Question string
	Answer   string
}

type AnalysisContext struct {
	ReadingCount int
	MeterCount   int
	Readings     []reading.Reading
	Events       []event.Event
	Anomalies    []anomaly.Anomaly
}

type ConsumptionInsightInput struct {
	MeterID               string
	Provider              string
	Region                string
	From                  time.Time
	To                    time.Time
	Readings              []reading.Reading
	Events                []event.Event
	Anomalies             []anomaly.Anomaly
	CurrentConsumptionKWh float64
	BaselineKWh           float64
	AveragePowerFactor    float64
	EstimatedCost         float64
	Currency              string
}

type ConsumptionInsightResult struct {
	Summary              string
	Recommendations      []string
	PotentialSavingsKWh  float64
	PotentialSavingsCost float64
	Currency             string
	GeneratedAt          time.Time
}

func (result ConsumptionInsightResult) HasRecommendations() bool {
	return len(result.Recommendations) > 0
}
