package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"learning/internal/billing"
)

type fakeTariffRepo struct {
	stored []billing.Tariff
}

func (repository *fakeTariffRepo) Save(ctx context.Context, tariff billing.Tariff) (billing.Tariff, error) {
	repository.stored = append(repository.stored, tariff)
	return tariff, nil
}

func (repository *fakeTariffRepo) List(ctx context.Context, provider, region string, active bool) ([]billing.Tariff, error) {
	return repository.stored, nil
}

func (repository *fakeTariffRepo) GetByID(ctx context.Context, tariffID int) (billing.Tariff, error) {
	for _, tariff := range repository.stored {
		if tariff.TariffID == tariffID {
			return tariff, nil
		}
	}
	return billing.Tariff{}, errors.New("tariff not found")
}

func (repository *fakeTariffRepo) GetActiveForMeter(ctx context.Context, provider, region, rateType string) (billing.Tariff, error) {
	for _, tariff := range repository.stored {
		if tariff.Provider == provider && tariff.Region == region && tariff.RateType == rateType {
			return tariff, nil
		}
	}
	return billing.Tariff{}, errors.New("tariff not found")
}

type fakeMeterProfileRepo struct {
	provider string
	region   string
	rateType string
	err      error
}

func (repo fakeMeterProfileRepo) GetMeterProfile(ctx context.Context, meterID string) (billing.MeterProfile, error) {
	return billing.MeterProfile{MeterID: meterID, Provider: repo.provider, Region: repo.region, RateType: repo.rateType}, repo.err
}

func (repo fakeMeterProfileRepo) GetProviderAndRegion(ctx context.Context, meterID string) (string, string, error) {
	return repo.provider, repo.region, repo.err
}

func (repo fakeMeterProfileRepo) UpdateMeterProfile(ctx context.Context, meterID string, provider, region, rateType string) error {
	return nil
}

type fakeBillingRepo struct {
	stored []billing.Billing
}

func (repository *fakeBillingRepo) Save(ctx context.Context, item billing.Billing) (billing.Billing, error) {
	repository.stored = append(repository.stored, item)
	return item, nil
}

func (repository *fakeBillingRepo) ListByMeter(ctx context.Context, meterID string) ([]billing.Billing, error) {
	return repository.stored, nil
}

func (repository *fakeBillingRepo) ListByPeriod(ctx context.Context, from, to time.Time) ([]billing.Billing, error) {
	return repository.stored, nil
}

func (repository *fakeBillingRepo) SummaryByPeriod(ctx context.Context, from, to time.Time) (billing.BillingSummary, error) {
	return billing.BillingSummary{TotalCost: 1250.5, Currency: "COP"}, nil
}

type fakeConsumption struct{}

func (fakeConsumption) SumConsumption(ctx context.Context, meterID string, start, end time.Time) (float64, error) {
	return 125.5, nil
}

func TestEstimateMeterConsumptionCostUseCaseStoresBillingRecord(t *testing.T) {
	ctx := context.Background()
	tariffRepo := &fakeTariffRepo{stored: []billing.Tariff{{
		TariffID:    1,
		Provider:    "EPM",
		Region:      "Medellín",
		RateType:    "industrial",
		PricePerKWh: 225.5,
		Currency:    "COP",
		ValidFrom:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}}}
	billingRepo := &fakeBillingRepo{}
	useCase := NewEstimateMeterConsumptionCostUseCase(
		billingRepo,
		fakeMeterProfileRepo{provider: "EPM", region: "Medellín", rateType: "industrial"},
		tariffRepo,
		fakeConsumption{},
	)

	result, err := useCase.Execute(ctx, "M-101", time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 28, 23, 59, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if result.TotalKWh != 125.5 {
		t.Fatalf("total_kwh = %v, want 125.5", result.TotalKWh)
	}
	if result.TotalCost != 28300.25 {
		t.Fatalf("total_cost = %v, want 28300.25", result.TotalCost)
	}
	if len(billingRepo.stored) != 1 {
		t.Fatalf("stored billing records = %d, want 1", len(billingRepo.stored))
	}
}

func TestGetCostTimeSeriesUseCaseGroupsWeeklyLabels(t *testing.T) {
	ctx := context.Background()
	repo := &fakeBillingRepo{stored: []billing.Billing{
		{MeterID: "M-1", PeriodStart: time.Date(2025, 2, 3, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(2025, 2, 9, 23, 59, 0, 0, time.UTC), TotalKWh: 10, TotalCost: 100, Currency: "COP"},
		{MeterID: "M-1", PeriodStart: time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(2025, 2, 11, 23, 59, 0, 0, time.UTC), TotalKWh: 15, TotalCost: 150, Currency: "COP"},
		{MeterID: "M-1", PeriodStart: time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(2025, 2, 16, 23, 59, 0, 0, time.UTC), TotalKWh: 20, TotalCost: 200, Currency: "COP"},
	}}
	useCase := NewGetCostTimeSeriesUseCase(repo)

	result, err := useCase.Execute(ctx, time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC), "weekly")
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	for _, point := range result {
		if point.Label == "" {
			t.Fatalf("point label is empty: %+v", point)
		}
		if point.TotalCost <= 0 {
			t.Fatalf("point total_cost = %v, want > 0", point.TotalCost)
		}
	}
	for i := 1; i < len(result); i++ {
		if result[i-1].Date.After(result[i].Date) {
			t.Fatalf("result not ordered by date: %+v", result)
		}
	}
}

func TestGetCostByMeterUseCaseProjectsDailyWeeklyPeriod(t *testing.T) {
	ctx := context.Background()
	periodStart := time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 9, 26, 23, 59, 0, 0, time.UTC)
	periodDays := 30
	totalCost := 12795222.0

	tests := []struct {
		name    string
		stored  []billing.Billing
		meterID string
		daily   float64
		weekly  float64
		period  float64
	}{
		{
			name: "single billing record spreads total across granularity",
			stored: []billing.Billing{
				{MeterID: "M-104", PeriodStart: periodStart, PeriodEnd: periodEnd, TotalKWh: 18543.8, TotalCost: totalCost, Currency: "COP"},
			},
			meterID: "M-104",
			daily:   totalCost / float64(periodDays),
			weekly:  totalCost / float64(periodDays) * 7,
			period:  totalCost,
		},
		{
			name: "multiple records accumulate before projecting",
			stored: []billing.Billing{
				{MeterID: "M-104", PeriodStart: periodStart, PeriodEnd: periodEnd, TotalKWh: 9000, TotalCost: 6000000, Currency: "COP"},
				{MeterID: "M-104", PeriodStart: periodStart, PeriodEnd: periodEnd, TotalKWh: 9543.8, TotalCost: 6795222, Currency: "COP"},
			},
			meterID: "M-104",
			daily:   totalCost / float64(periodDays),
			weekly:  totalCost / float64(periodDays) * 7,
			period:  totalCost,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := NewGetCostByMeterUseCase(&fakeBillingRepo{stored: test.stored}, &fakeTariffRepo{})
			results, err := useCase.Execute(ctx, periodStart, periodEnd)
			if err != nil {
				t.Fatalf("execute returned error: %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("len(results) = %d, want 1", len(results))
			}
			row := results[0]
			if row.MeterID != test.meterID {
				t.Fatalf("meter_id = %q, want %q", row.MeterID, test.meterID)
			}
			if row.DailyCost != test.daily {
				t.Fatalf("daily_cost = %v, want %v", row.DailyCost, test.daily)
			}
			if row.WeeklyCost != test.weekly {
				t.Fatalf("weekly_cost = %v, want %v", row.WeeklyCost, test.weekly)
			}
			if row.PeriodCost != test.period {
				t.Fatalf("period_cost = %v, want %v", row.PeriodCost, test.period)
			}
			if row.DailyCost == row.WeeklyCost || row.DailyCost == row.PeriodCost {
				t.Fatalf("granularities must differ: daily=%v weekly=%v period=%v", row.DailyCost, row.WeeklyCost, row.PeriodCost)
			}
		})
	}
}
func TestGetActiveTariffForMeterUseCaseRequiresRateType(t *testing.T) {
	ctx := context.Background()
	tariffRepo := &fakeTariffRepo{stored: []billing.Tariff{
		{TariffID: 1, Provider: "EPM", Region: "Medellín", RateType: "industrial", PricePerKWh: 215, Currency: "COP", ValidFrom: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{TariffID: 2, Provider: "EPM", Region: "Medellín", RateType: "residential", PricePerKWh: 300, Currency: "COP", ValidFrom: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}}
	useCase := NewGetActiveTariffForMeterUseCase(tariffRepo, fakeMeterProfileRepo{provider: "EPM", region: "Medellín", rateType: "industrial"})

	result, err := useCase.Execute(ctx, "M-101")
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if result.RateType != "industrial" {
		t.Fatalf("rate_type = %q, want industrial", result.RateType)
	}
}

func TestGetActiveTariffForMeterUseCaseFailsWhenMeterNotConfigured(t *testing.T) {
	ctx := context.Background()
	tariffRepo := &fakeTariffRepo{stored: []billing.Tariff{{TariffID: 1, Provider: "EPM", Region: "Medellín", RateType: "industrial", PricePerKWh: 215, Currency: "COP", ValidFrom: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}}}
	useCase := NewGetActiveTariffForMeterUseCase(tariffRepo, fakeMeterProfileRepo{provider: "", region: "", rateType: "", err: billing.ErrMeterNotConfigured})

	_, err := useCase.Execute(ctx, "M-101")
	if !errors.Is(err, billing.ErrMeterNotConfigured) {
		t.Fatalf("err = %v, want ErrMeterNotConfigured", err)
	}
}
