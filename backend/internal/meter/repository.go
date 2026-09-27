package meter

import (
	"context"

	"learning/internal/anomaly"
	"learning/internal/reading"
)

type ReadingRepository interface {
	AllReadings(context.Context) ([]reading.Reading, error)
	ReadingsByMeter(context.Context, string) ([]reading.Reading, error)
}

type AnomalyRepository interface {
	ListAnomalies(context.Context) ([]anomaly.Anomaly, error)
}
