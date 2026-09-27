package application

import (
	"context"

	"learning/internal/anomaly"
)

type AnomalyRepository interface {
	ListAnomalies(context.Context) ([]anomaly.Anomaly, error)
	FindAnomaly(context.Context, string) (anomaly.Anomaly, error)
}
