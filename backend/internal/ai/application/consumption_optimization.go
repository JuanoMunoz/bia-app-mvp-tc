package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	"learning/internal/billing"
	"learning/internal/reading"
)

var ErrOptimizationInsightUnavailable = errors.New("optimization insight unavailable")

type MeterReadingsRepository interface {
	ReadingsByMeter(context.Context, string) ([]reading.Reading, error)
}

type MeterAnomaliesRepository interface {
	ListAnomalies(context.Context) ([]anomaly.Anomaly, error)
}

type MeterProfileRepository interface {
	GetMeterProfile(context.Context, string) (billing.MeterProfile, error)
	GetProviderAndRegion(context.Context, string) (string, string, error)
}

type TariffRepository interface {
	GetActiveForMeter(context.Context, string, string, string) (billing.Tariff, error)
}

type BillingRepository interface {
	SummaryByPeriod(context.Context, time.Time, time.Time) (billing.BillingSummary, error)
}

type GetConsumptionOptimizationUseCase struct {
	billingRepository      BillingRepository
	tariffRepository       TariffRepository
	meterProfileRepository MeterProfileRepository
	readingsRepository     MeterReadingsRepository
	anomaliesRepository    MeterAnomaliesRepository
	cacheRepository        ai.InsightCacheRepository
	recommendationPort     ai.AIRecommendationPort
}

func NewGetConsumptionOptimizationUseCase(
	billingRepository BillingRepository,
	tariffRepository TariffRepository,
	meterProfileRepository MeterProfileRepository,
	readingsRepository MeterReadingsRepository,
	anomaliesRepository MeterAnomaliesRepository,
	cacheRepository ai.InsightCacheRepository,
	recommendationPort ai.AIRecommendationPort,
) *GetConsumptionOptimizationUseCase {
	return &GetConsumptionOptimizationUseCase{
		billingRepository:      billingRepository,
		tariffRepository:       tariffRepository,
		meterProfileRepository: meterProfileRepository,
		readingsRepository:     readingsRepository,
		anomaliesRepository:    anomaliesRepository,
		cacheRepository:        cacheRepository,
		recommendationPort:     recommendationPort,
	}
}

func (useCase *GetConsumptionOptimizationUseCase) Execute(ctx context.Context, meterID string, from, to time.Time) (ai.ConsumptionInsightResult, error) {
	meterID = strings.TrimSpace(meterID)
	if meterID == "" {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: meter_id is required", ErrOptimizationInsightUnavailable)
	}
	if to.Before(from) || to.Equal(from) {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: from must be before to", ErrOptimizationInsightUnavailable)
	}
	if useCase.cacheRepository != nil {
		if cached, found, err := useCase.cacheRepository.GetRecentInsight(ctx, meterID, from, to); err == nil && found {
			return cached, nil
		}
	}
	profile, err := useCase.meterProfileRepository.GetMeterProfile(ctx, meterID)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: resolve meter profile: %v", ErrOptimizationInsightUnavailable, err)
	}
	provider, region, rateType := profile.Provider, profile.Region, profile.RateType
	if strings.TrimSpace(provider) == "" || strings.TrimSpace(region) == "" || strings.TrimSpace(rateType) == "" {
		rateType = "industrial"
	}
	if useCase.tariffRepository != nil {
		if _, err := useCase.tariffRepository.GetActiveForMeter(ctx, provider, region, rateType); err != nil {
			return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: resolve tariff for meter: %v", ErrOptimizationInsightUnavailable, err)
		}
	}
	readings, err := useCase.readingsRepository.ReadingsByMeter(ctx, meterID)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: load meter readings: %v", ErrOptimizationInsightUnavailable, err)
	}
	filtered := filterReadings(readings, from, to)
	if len(filtered) == 0 {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: no readings available for the selected period", ErrOptimizationInsightUnavailable)
	}
	var meterAnomalies []anomaly.Anomaly
	if useCase.anomaliesRepository != nil {
		allAnomalies, err := useCase.anomaliesRepository.ListAnomalies(ctx)
		if err != nil {
			return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: load meter anomalies: %v", ErrOptimizationInsightUnavailable, err)
		}
		for _, item := range allAnomalies {
			if item.MeterID == meterID {
				meterAnomalies = append(meterAnomalies, item)
			}
		}
	}
	input := ai.ConsumptionInsightInput{
		MeterID:               meterID,
		Provider:              provider,
		Region:                region,
		From:                  from,
		To:                    to,
		Readings:              filtered,
		Events:                nil,
		Anomalies:             meterAnomalies,
		CurrentConsumptionKWh: currentKWh(filtered),
		BaselineKWh:           baselineKWh(filtered),
		AveragePowerFactor:    averagePowerFactor(filtered),
		EstimatedCost:         estimateTotalCost(useCase.billingRepository, ctx, from, to),
		Currency:              "COP",
	}
	if useCase.recommendationPort == nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: recommendation engine not configured", ErrOptimizationInsightUnavailable)
	}
	result, err := useCase.recommendationPort.GenerateConsumptionInsight(ctx, input)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: %v", ErrOptimizationInsightUnavailable, err)
	}
	if result.GeneratedAt.IsZero() {
		result.GeneratedAt = time.Now().UTC()
	}
	if useCase.cacheRepository != nil {
		if saveErr := useCase.cacheRepository.SaveInsight(ctx, meterID, from, to, result); saveErr != nil {
			return ai.ConsumptionInsightResult{}, fmt.Errorf("%w: save optimization insight: %v", ErrOptimizationInsightUnavailable, saveErr)
		}
	}
	return result, nil
}

func filterReadings(readings []reading.Reading, from, to time.Time) []reading.Reading {
	filtered := make([]reading.Reading, 0, len(readings))
	for _, item := range readings {
		if item.Timestamp.Before(from) || item.Timestamp.After(to) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func currentKWh(readings []reading.Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	var total float64
	for _, item := range readings {
		total += item.ConsumptionKWh
	}
	return total
}

func baselineKWh(readings []reading.Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	var total float64
	for _, item := range readings {
		total += item.ConsumptionKWh
	}
	return total / float64(len(readings))
}

func averagePowerFactor(readings []reading.Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	var total float64
	for _, item := range readings {
		total += item.PowerFactor
	}
	return total / float64(len(readings))
}

func estimateTotalCost(repository BillingRepository, ctx context.Context, from, to time.Time) float64 {
	if repository == nil {
		return 0
	}
	summary, err := repository.SummaryByPeriod(ctx, from, to)
	if err != nil || summary.TotalCost == 0 {
		return 0
	}
	return summary.TotalCost
}
