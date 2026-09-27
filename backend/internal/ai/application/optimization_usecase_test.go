package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	"learning/internal/billing"
	"learning/internal/reading"
)

type fakeRecommendationPort struct {
	called bool
	input  ai.ConsumptionInsightInput
	err    error
	result ai.ConsumptionInsightResult
}

func (port *fakeRecommendationPort) GenerateConsumptionInsight(ctx context.Context, data ai.ConsumptionInsightInput) (ai.ConsumptionInsightResult, error) {
	port.called = true
	port.input = data
	return port.result, port.err
}

type fakeMeterReadingsRepo struct {
	series []reading.Reading
}

func (repo *fakeMeterReadingsRepo) ReadingsByMeter(ctx context.Context, meterID string) ([]reading.Reading, error) {
	return repo.series, nil
}

type fakeAnomalyRepo struct {
	items []anomaly.Anomaly
}

func (repo *fakeAnomalyRepo) ListAnomalies(ctx context.Context) ([]anomaly.Anomaly, error) {
	return repo.items, nil
}

type fakeInsightCacheRepo struct {
	cached ai.ConsumptionInsightResult
	found  bool
	called bool
}

func (repo *fakeInsightCacheRepo) GetRecentInsight(ctx context.Context, meterID string, from, to time.Time) (ai.ConsumptionInsightResult, bool, error) {
	repo.called = true
	return repo.cached, repo.found, nil
}

func (repo *fakeInsightCacheRepo) SaveInsight(ctx context.Context, meterID string, from, to time.Time, insight ai.ConsumptionInsightResult) error {
	return nil
}

type fakeTariffRepo struct{}

func (repo *fakeTariffRepo) Save(ctx context.Context, tariff billing.Tariff) (billing.Tariff, error) {
	return tariff, nil
}
func (repo *fakeTariffRepo) List(ctx context.Context, provider, region string, active bool) ([]billing.Tariff, error) {
	return nil, nil
}
func (repo *fakeTariffRepo) GetByID(ctx context.Context, tariffID int) (billing.Tariff, error) {
	return billing.Tariff{}, nil
}
func (repo *fakeTariffRepo) GetActiveForMeter(ctx context.Context, provider, region, rateType string) (billing.Tariff, error) {
	return billing.Tariff{TariffID: 7, Provider: provider, Region: region, PricePerKWh: 150, Currency: "COP", RateType: rateType}, nil
}

type fakeMeterProfileRepo struct{}

func (repo *fakeMeterProfileRepo) GetMeterProfile(ctx context.Context, meterID string) (billing.MeterProfile, error) {
	return billing.MeterProfile{MeterID: meterID, Provider: "EPM", Region: "Medellín", RateType: "industrial"}, nil
}

func (repo *fakeMeterProfileRepo) GetProviderAndRegion(ctx context.Context, meterID string) (string, string, error) {
	return "EPM", "Medellín", nil
}

type fakeBillingRepo struct{}

func (repo *fakeBillingRepo) Save(ctx context.Context, billing billing.Billing) (billing.Billing, error) {
	return billing, nil
}
func (repo *fakeBillingRepo) ListByMeter(ctx context.Context, meterID string) ([]billing.Billing, error) {
	return []billing.Billing{{MeterID: meterID, TotalKWh: 410, TotalCost: 42000, Currency: "COP", PeriodStart: time.Now().AddDate(0, 0, -30), PeriodEnd: time.Now()}}, nil
}
func (repo *fakeBillingRepo) ListByPeriod(ctx context.Context, from, to time.Time) ([]billing.Billing, error) {
	return nil, nil
}
func (repo *fakeBillingRepo) SummaryByPeriod(ctx context.Context, from, to time.Time) (billing.BillingSummary, error) {
	return billing.BillingSummary{TotalCost: 42000, Currency: "COP"}, nil
}

func TestGetConsumptionOptimizationUseCaseBuildsInputAndCallsAI(t *testing.T) {
	ctx := context.Background()
	base := time.Now().UTC().AddDate(0, 0, -30)
	series := []reading.Reading{
		{MeterID: "M-101", Timestamp: base.Add(2 * time.Hour), ConsumptionKWh: 12.5},
		{MeterID: "M-101", Timestamp: base.Add(3 * time.Hour), ConsumptionKWh: 14.0},
	}
	port := &fakeRecommendationPort{result: ai.ConsumptionInsightResult{Summary: "Resumen", Recommendations: []string{"Ajusta arranques"}, GeneratedAt: time.Now().UTC()}}
	uc := NewGetConsumptionOptimizationUseCase(
		&fakeBillingRepo{},
		&fakeTariffRepo{},
		&fakeMeterProfileRepo{},
		&fakeMeterReadingsRepo{series: series},
		&fakeAnomalyRepo{},
		&fakeInsightCacheRepo{},
		port,
	)

	result, err := uc.Execute(ctx, "M-101", base, time.Now().UTC())
	if err != nil {
		t.Fatalf("execute returned unexpected error: %v", err)
	}
	if !port.called {
		t.Fatal("expected recommendation port to be called")
	}
	if port.input.MeterID != "M-101" {
		t.Fatalf("meter id mismatch: %s", port.input.MeterID)
	}
	if port.input.Provider != "EPM" {
		t.Fatalf("provider mismatch: %s", port.input.Provider)
	}
	if len(result.Recommendations) == 0 {
		t.Fatal("recommendations should be returned")
	}
}

func TestGetConsumptionOptimizationUseCaseUsesCacheWhenAvailable(t *testing.T) {
	ctx := context.Background()
	cache := &fakeInsightCacheRepo{found: true, cached: ai.ConsumptionInsightResult{Summary: "cacheado", Recommendations: []string{"Acción cacheada"}, GeneratedAt: time.Now().UTC()}}
	port := &fakeRecommendationPort{}
	uc := NewGetConsumptionOptimizationUseCase(
		&fakeBillingRepo{},
		&fakeTariffRepo{},
		&fakeMeterProfileRepo{},
		&fakeMeterReadingsRepo{series: []reading.Reading{{MeterID: "M-101", Timestamp: time.Now().UTC(), ConsumptionKWh: 5}}},
		&fakeAnomalyRepo{},
		cache,
		port,
	)

	_, err := uc.Execute(ctx, "M-101", time.Now().UTC().AddDate(0, 0, -7), time.Now().UTC())
	if err != nil {
		t.Fatalf("execute returned unexpected error: %v", err)
	}
	if port.called {
		t.Fatal("expected recommendation port not to be called when cache is valid")
	}
	if !cache.called {
		t.Fatal("expected cache lookup to be called")
	}
}

func TestGetConsumptionOptimizationUseCasePropagatesAIError(t *testing.T) {
	ctx := context.Background()
	port := &fakeRecommendationPort{err: errors.New("gemini rate limited")}
	uc := NewGetConsumptionOptimizationUseCase(
		&fakeBillingRepo{},
		&fakeTariffRepo{},
		&fakeMeterProfileRepo{},
		&fakeMeterReadingsRepo{series: []reading.Reading{{MeterID: "M-101", Timestamp: time.Now().UTC().Add(-time.Hour), ConsumptionKWh: 8}}},
		&fakeAnomalyRepo{},
		&fakeInsightCacheRepo{},
		port,
	)

	_, err := uc.Execute(ctx, "M-101", time.Now().UTC().AddDate(0, 0, -7), time.Now().UTC())
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() == "gemini rate limited" {
		t.Fatal("expected domain error, not raw infrastructure error")
	}
}
