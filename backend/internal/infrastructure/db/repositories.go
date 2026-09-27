package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	"learning/internal/billing"
	"learning/internal/event"
	"learning/internal/reading"
)

type ReadingRepository struct {
	database *Repository
}

func NewReadingRepository(database *Repository) *ReadingRepository {
	return &ReadingRepository{database: database}
}

func (repository *ReadingRepository) AllReadings(ctx context.Context) ([]reading.Reading, error) {
	return repository.database.AllReadings(ctx)
}

func (repository *ReadingRepository) ReadingsByMeter(ctx context.Context, meterID string) ([]reading.Reading, error) {
	return repository.database.ReadingsByMeter(ctx, meterID)
}

type AnomalyRepository struct {
	database *Repository
}

func NewAnomalyRepository(database *Repository) *AnomalyRepository {
	return &AnomalyRepository{database: database}
}

func (repository *AnomalyRepository) ListAnomalies(ctx context.Context) ([]anomaly.Anomaly, error) {
	return repository.database.ListAnomalies(ctx)
}

func (repository *AnomalyRepository) FindAnomaly(ctx context.Context, id string) (anomaly.Anomaly, error) {
	return repository.database.FindAnomaly(ctx, id)
}

type AnalysisRepository struct {
	database *Repository
}

type InsightCacheRepository struct {
	database *Repository
}

func NewAnalysisRepository(database *Repository) *AnalysisRepository {
	return &AnalysisRepository{database: database}
}

func NewInsightCacheRepository(database *Repository) *InsightCacheRepository {
	return &InsightCacheRepository{database: database}
}

func (repository *AnalysisRepository) AllReadings(ctx context.Context) ([]reading.Reading, error) {
	return repository.database.AllReadings(ctx)
}

func (repository *AnalysisRepository) AllEvents(ctx context.Context) ([]event.Event, error) {
	return repository.database.AllEvents(ctx)
}

func (repository *AnalysisRepository) SaveAnalysis(ctx context.Context, result ai.Analysis) error {
	return repository.database.SaveAnalysis(ctx, result)
}

func (repository *AnalysisRepository) FindAnalysis(ctx context.Context, id string) (ai.Analysis, error) {
	return repository.database.FindAnalysis(ctx, id)
}

func (repository *AnalysisRepository) LatestAnalysis(ctx context.Context) (ai.Analysis, error) {
	return repository.database.LatestAnalysis(ctx)
}

func (repository *InsightCacheRepository) GetRecentInsight(ctx context.Context, meterID string, from, to time.Time) (ai.ConsumptionInsightResult, bool, error) {
	var generatedAt time.Time
	var payload []byte
	query := `SELECT generated_at, payload FROM ai_insights WHERE meter_id=$1 AND from_ts=$2 AND to_ts=$3 AND generated_at >= NOW() - INTERVAL '24 hours' ORDER BY generated_at DESC LIMIT 1`
	if err := repository.database.pool.QueryRow(ctx, query, meterID, from, to).Scan(&generatedAt, &payload); err != nil {
		if err.Error() == "no rows in result set" || strings.Contains(err.Error(), "no rows") {
			return ai.ConsumptionInsightResult{}, false, nil
		}
		return ai.ConsumptionInsightResult{}, false, fmt.Errorf("lookup cached insight: %w", err)
	}
	var result ai.ConsumptionInsightResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return ai.ConsumptionInsightResult{}, false, fmt.Errorf("decode cached insight: %w", err)
	}
	if result.GeneratedAt.IsZero() {
		result.GeneratedAt = generatedAt
	}
	return result, true, nil
}

func (repository *InsightCacheRepository) SaveInsight(ctx context.Context, meterID string, from, to time.Time, insight ai.ConsumptionInsightResult) error {
	payload, err := json.Marshal(insight)
	if err != nil {
		return fmt.Errorf("encode cached insight: %w", err)
	}
	_, err = repository.database.pool.Exec(ctx, `INSERT INTO ai_insights(meter_id, from_ts, to_ts, generated_at, payload)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT(meter_id, from_ts, to_ts) DO UPDATE SET generated_at=EXCLUDED.generated_at, payload=EXCLUDED.payload`, meterID, from, to, time.Now().UTC(), payload)
	if err != nil {
		return fmt.Errorf("save cached insight: %w", err)
	}
	return nil
}

type TariffRepository struct {
	database *Repository
}

func NewTariffRepository(database *Repository) *TariffRepository {
	return &TariffRepository{database: database}
}

func (repository *TariffRepository) Save(ctx context.Context, tariff billing.Tariff) (billing.Tariff, error) {
	var createdAt time.Time
	var tariffID int
	query := `INSERT INTO tariffs(provider, region, rate_type, price_per_kwh, currency, valid_from, valid_to, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING tariff_id, created_at`
	if err := repository.database.pool.QueryRow(ctx, query,
		tariff.Provider,
		tariff.Region,
		tariff.RateType,
		tariff.PricePerKWh,
		tariff.Currency,
		tariff.ValidFrom,
		tariff.ValidTo,
	).Scan(&tariffID, &createdAt); err != nil {
		return billing.Tariff{}, fmt.Errorf("save tariff: %w", err)
	}
	tariff.TariffID = tariffID
	tariff.CreatedAt = createdAt
	return tariff, nil
}

func (repository *TariffRepository) List(ctx context.Context, provider string, region string, active bool) ([]billing.Tariff, error) {
	query := `SELECT tariff_id, provider, region, rate_type, price_per_kwh, currency, valid_from, valid_to, created_at FROM tariffs WHERE 1=1`
	args := []any{}
	if provider != "" {
		query += " AND provider=$" + fmt.Sprint(len(args)+1)
		args = append(args, provider)
	}
	if region != "" {
		query += " AND region=$" + fmt.Sprint(len(args)+1)
		args = append(args, region)
	}
	if active {
		query += " AND valid_from <= NOW() AND (valid_to IS NULL OR valid_to >= NOW())"
	}
	query += " ORDER BY valid_from DESC, tariff_id DESC"
	rows, err := repository.database.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tariffs: %w", err)
	}
	defer rows.Close()
	results := make([]billing.Tariff, 0)
	for rows.Next() {
		var item billing.Tariff
		var validTo *time.Time
		if err := rows.Scan(&item.TariffID, &item.Provider, &item.Region, &item.RateType, &item.PricePerKWh, &item.Currency, &item.ValidFrom, &validTo, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tariff: %w", err)
		}
		item.ValidTo = validTo
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tariffs: %w", err)
	}
	return results, nil
}

func (repository *TariffRepository) GetByID(ctx context.Context, tariffID int) (billing.Tariff, error) {
	query := `SELECT tariff_id, provider, region, rate_type, price_per_kwh, currency, valid_from, valid_to, created_at
		FROM tariffs WHERE tariff_id=$1 LIMIT 1`
	var item billing.Tariff
	var validTo *time.Time
	if err := repository.database.pool.QueryRow(ctx, query, tariffID).Scan(&item.TariffID, &item.Provider, &item.Region, &item.RateType, &item.PricePerKWh, &item.Currency, &item.ValidFrom, &validTo, &item.CreatedAt); err != nil {
		return billing.Tariff{}, fmt.Errorf("get tariff by id: %w", err)
	}
	item.ValidTo = validTo
	return item, nil
}

func (repository *TariffRepository) GetActiveForMeter(ctx context.Context, provider string, region string, rateType string) (billing.Tariff, error) {
	query := `SELECT tariff_id, provider, region, rate_type, price_per_kwh, currency, valid_from, valid_to, created_at
		FROM tariffs WHERE LOWER(provider)=LOWER($1) AND LOWER(region)=LOWER($2) AND LOWER(rate_type)=LOWER($3) AND valid_from <= NOW() AND (valid_to IS NULL OR valid_to >= NOW())
		ORDER BY valid_from DESC LIMIT 1`
	var item billing.Tariff
	var validTo *time.Time
	if err := repository.database.pool.QueryRow(ctx, query, provider, region, rateType).Scan(&item.TariffID, &item.Provider, &item.Region, &item.RateType, &item.PricePerKWh, &item.Currency, &item.ValidFrom, &validTo, &item.CreatedAt); err != nil {
		return billing.Tariff{}, fmt.Errorf("get active tariff: %w", err)
	}
	item.ValidTo = validTo
	return item, nil
}

type BillingRepository struct {
	database *Repository
}

func NewBillingRepository(database *Repository) *BillingRepository {
	return &BillingRepository{database: database}
}

func (repository *BillingRepository) Save(ctx context.Context, item billing.Billing) (billing.Billing, error) {
	var billingID int
	var createdAt time.Time
	query := `INSERT INTO billing(meter_id, tariff_id, period_start, period_end, total_kwh, price_per_kwh, total_cost, currency, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING billing_id, created_at`
	if err := repository.database.pool.QueryRow(ctx, query,
		item.MeterID,
		item.TariffID,
		item.PeriodStart,
		item.PeriodEnd,
		item.TotalKWh,
		item.PricePerKWh,
		item.TotalCost,
		item.Currency,
	).Scan(&billingID, &createdAt); err != nil {
		return billing.Billing{}, fmt.Errorf("save billing: %w", err)
	}
	item.BillingID = billingID
	item.CreatedAt = createdAt
	return item, nil
}

func (repository *BillingRepository) ListByMeter(ctx context.Context, meterID string) ([]billing.Billing, error) {
	rows, err := repository.database.pool.Query(ctx, `SELECT billing_id, meter_id, tariff_id, period_start, period_end, total_kwh, price_per_kwh, total_cost, currency, created_at
		FROM billing WHERE meter_id=$1 ORDER BY period_end DESC`, meterID)
	if err != nil {
		return nil, fmt.Errorf("list meter billing: %w", err)
	}
	defer rows.Close()
	results := make([]billing.Billing, 0)
	for rows.Next() {
		var item billing.Billing
		var tariffID *int
		if err := rows.Scan(&item.BillingID, &item.MeterID, &tariffID, &item.PeriodStart, &item.PeriodEnd, &item.TotalKWh, &item.PricePerKWh, &item.TotalCost, &item.Currency, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan billing: %w", err)
		}
		item.TariffID = tariffID
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate billing: %w", err)
	}
	return results, nil
}

func (repository *BillingRepository) ListByPeriod(ctx context.Context, from, to time.Time) ([]billing.Billing, error) {
	query := `SELECT billing_id, meter_id, tariff_id, period_start, period_end, total_kwh, price_per_kwh, total_cost, currency, created_at
		FROM billing WHERE 1=1`
	args := []any{}
	if !from.IsZero() {
		query += " AND period_end >= $" + fmt.Sprint(len(args)+1)
		args = append(args, from)
	}
	if !to.IsZero() {
		query += " AND period_start <= $" + fmt.Sprint(len(args)+1)
		args = append(args, to)
	}
	query += " ORDER BY period_end DESC"
	rows, err := repository.database.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list billing by period: %w", err)
	}
	defer rows.Close()
	results := make([]billing.Billing, 0)
	for rows.Next() {
		var item billing.Billing
		var tariffID *int
		if err := rows.Scan(&item.BillingID, &item.MeterID, &tariffID, &item.PeriodStart, &item.PeriodEnd, &item.TotalKWh, &item.PricePerKWh, &item.TotalCost, &item.Currency, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan billing period: %w", err)
		}
		item.TariffID = tariffID
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate billing period: %w", err)
	}
	return results, nil
}

func (repository *BillingRepository) SummaryByPeriod(ctx context.Context, from, to time.Time) (billing.BillingSummary, error) {
	query := `SELECT COALESCE(SUM(total_cost), 0), COALESCE(currency, 'COP') FROM billing WHERE 1=1`
	args := []any{}
	if !from.IsZero() {
		query += " AND period_end >= $" + fmt.Sprint(len(args)+1)
		args = append(args, from)
	}
	if !to.IsZero() {
		query += " AND period_start <= $" + fmt.Sprint(len(args)+1)
		args = append(args, to)
	}
	query += " GROUP BY currency ORDER BY currency LIMIT 1"
	rows, err := repository.database.pool.Query(ctx, query, args...)
	if err != nil {
		return billing.BillingSummary{}, fmt.Errorf("summary billing period: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return billing.BillingSummary{TotalCost: 0, Currency: "COP"}, nil
	}
	var totalCost float64
	var currency string
	if err := rows.Scan(&totalCost, &currency); err != nil {
		return billing.BillingSummary{}, fmt.Errorf("scan summary billing period: %w", err)
	}
	if err := rows.Err(); err != nil {
		return billing.BillingSummary{}, fmt.Errorf("iterate summary billing period: %w", err)
	}
	return billing.BillingSummary{TotalCost: totalCost, Currency: currency}, nil
}

type MeterProfileRepository struct {
	database *Repository
}

func NewMeterProfileRepository(database *Repository) *MeterProfileRepository {
	return &MeterProfileRepository{database: database}
}

func (repository *MeterProfileRepository) GetMeterProfile(ctx context.Context, meterID string) (billing.MeterProfile, error) {
	var profile billing.MeterProfile
	if err := repository.database.pool.QueryRow(ctx, `SELECT meter_id, COALESCE(provider, ''), COALESCE(region, ''), COALESCE(rate_type, 'industrial') FROM meters WHERE meter_id=$1`, meterID).Scan(&profile.MeterID, &profile.Provider, &profile.Region, &profile.RateType); err != nil {
		return billing.MeterProfile{}, fmt.Errorf("load meter profile: %w", err)
	}
	return profile, nil
}

func (repository *MeterProfileRepository) GetProviderAndRegion(ctx context.Context, meterID string) (string, string, error) {
	profile, err := repository.GetMeterProfile(ctx, meterID)
	if err != nil {
		return "", "", err
	}
	return profile.Provider, profile.Region, nil
}

func (repository *MeterProfileRepository) UpdateMeterProfile(ctx context.Context, meterID string, provider, region, rateType string) error {
	_, err := repository.database.pool.Exec(ctx, `UPDATE meters SET provider=$1, region=$2, rate_type=$3 WHERE meter_id=$4`, provider, region, rateType, meterID)
	if err != nil {
		return fmt.Errorf("update meter profile: %w", err)
	}
	return nil
}

type ConsumptionEstimator struct {
	database *Repository
}

func NewConsumptionEstimator(database *Repository) *ConsumptionEstimator {
	return &ConsumptionEstimator{database: database}
}

func (repository *ConsumptionEstimator) SumConsumption(ctx context.Context, meterID string, periodStart, periodEnd time.Time) (float64, error) {
	var total float64
	if err := repository.database.pool.QueryRow(ctx, `SELECT COALESCE(SUM(consumption_kwh), 0) FROM readings WHERE meter_id=$1 AND recorded_at >= $2 AND recorded_at <= $3`, meterID, periodStart, periodEnd).Scan(&total); err != nil {
		return 0, fmt.Errorf("sum consumption: %w", err)
	}
	return total, nil
}
