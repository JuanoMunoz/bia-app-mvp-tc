package application

import (
	"context"
	"errors"
	"fmt"

	"learning/internal/anomaly"
)

var ErrNotFound = errors.New("anomaly not found")

type Service struct {
	repository AnomalyRepository
}

func NewService(repository AnomalyRepository) *Service {
	return &Service{repository: repository}
}

func (service *Service) List(ctx context.Context) ([]anomaly.Anomaly, error) {
	results, err := service.repository.ListAnomalies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list anomalies: %w", err)
	}
	return results, nil
}

func (service *Service) Get(ctx context.Context, id string) (anomaly.Anomaly, error) {
	result, err := service.repository.FindAnomaly(ctx, id)
	if err != nil {
		return anomaly.Anomaly{}, fmt.Errorf("find anomaly: %w", err)
	}
	return result, nil
}
