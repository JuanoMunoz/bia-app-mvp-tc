package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	anomalyapp "learning/internal/anomaly/application"
	"learning/internal/event"
	"learning/internal/infrastructure/csvloader"
	"learning/internal/reading"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, connectionString string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	return &Repository{pool: pool}, nil
}

func (repository *Repository) Close() {
	repository.pool.Close()
}

func defaultSeedMeterProfile() (string, string, string) {
	return "bia", "CUND-EAST", "industrial"
}

func (repository *Repository) Ping(ctx context.Context) error {
	return repository.pool.Ping(ctx)
}

func (repository *Repository) Migrate(ctx context.Context, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		statement, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read migration: %w", err)
		}
		if _, err := repository.pool.Exec(ctx, string(statement)); err != nil {
			return fmt.Errorf("apply migration: %w", err)
		}
		return nil
	}
	if !info.IsDir() {
		statement, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration: %w", err)
		}
		if _, err := repository.pool.Exec(ctx, string(statement)); err != nil {
			return fmt.Errorf("apply migration: %w", err)
		}
		return nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		statement, err := os.ReadFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := repository.pool.Exec(ctx, string(statement)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (repository *Repository) Seed(ctx context.Context, dataset csvloader.Dataset) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "TRUNCATE billing, tariffs, anomalies, analyses, readings, events, meters CASCADE"); err != nil {
		return fmt.Errorf("clear existing dataset: %w", err)
	}
	provider, region, rateType := defaultSeedMeterProfile()
	meterIDs := make(map[string]struct{})
	for _, item := range dataset.Readings {
		meterIDs[item.MeterID] = struct{}{}
	}
	for meterID := range meterIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO meters(meter_id, provider, region, rate_type)
			VALUES($1, $2, $3, $4)
			ON CONFLICT (meter_id) DO UPDATE SET provider = EXCLUDED.provider, region = EXCLUDED.region, rate_type = EXCLUDED.rate_type`, meterID, provider, region, rateType); err != nil {
			return fmt.Errorf("seed meter %s: %w", meterID, err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tariffs (tariff_id, provider, region, rate_type, price_per_kwh, currency, valid_from, valid_to, created_at)
		VALUES (1, $1, $2, $3, 690, 'COP', '2026-09-27 00:00:00+00', NULL, '2026-09-27 19:19:00.43602+00')
		ON CONFLICT (tariff_id) DO UPDATE SET provider = EXCLUDED.provider, region = EXCLUDED.region, rate_type = EXCLUDED.rate_type,
		price_per_kwh = EXCLUDED.price_per_kwh, currency = EXCLUDED.currency, valid_from = EXCLUDED.valid_from,
		valid_to = EXCLUDED.valid_to, created_at = EXCLUDED.created_at`, provider, region, rateType); err != nil {
		return fmt.Errorf("seed tariff: %w", err)
	}
	for _, item := range dataset.Readings {
		_, err := tx.Exec(ctx, `INSERT INTO readings(meter_id, recorded_at, consumption_kwh, voltage_v, current_a, power_factor, status)
			VALUES($1, $2, $3, $4, $5, $6, $7)`, item.MeterID, item.Timestamp, item.ConsumptionKWh, item.VoltageV, item.CurrentA, item.PowerFactor, item.Status)
		if err != nil {
			return fmt.Errorf("seed reading for %s: %w", item.MeterID, err)
		}
	}
	for _, item := range dataset.Events {
		_, err := tx.Exec(ctx, `INSERT INTO events(meter_id, occurred_at, event_type, description)
			VALUES($1, $2, $3, $4)`, item.MeterID, item.Timestamp, item.Type, item.Description)
		if err != nil {
			return fmt.Errorf("seed event for %s: %w", item.MeterID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed: %w", err)
	}
	return nil
}

func (repository *Repository) AllReadings(ctx context.Context) ([]reading.Reading, error) {
	return repository.queryReadings(ctx, "SELECT meter_id, recorded_at, consumption_kwh, voltage_v, current_a, power_factor, status FROM readings ORDER BY meter_id, recorded_at")
}

func (repository *Repository) ReadingsByMeter(ctx context.Context, meterID string) ([]reading.Reading, error) {
	return repository.queryReadings(ctx, "SELECT meter_id, recorded_at, consumption_kwh, voltage_v, current_a, power_factor, status FROM readings WHERE meter_id=$1 ORDER BY recorded_at", meterID)
}

func (repository *Repository) ReadingPeriod(ctx context.Context) (time.Time, time.Time, error) {
	var start, end time.Time
	if err := repository.pool.QueryRow(ctx, "SELECT MIN(recorded_at), MAX(recorded_at) FROM readings").Scan(&start, &end); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("query reading period: %w", err)
	}
	return start, end, nil
}

func (repository *Repository) queryReadings(ctx context.Context, query string, args ...any) ([]reading.Reading, error) {
	rows, err := repository.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()
	results := make([]reading.Reading, 0)
	for rows.Next() {
		var item reading.Reading
		if err := rows.Scan(&item.MeterID, &item.Timestamp, &item.ConsumptionKWh, &item.VoltageV, &item.CurrentA, &item.PowerFactor, &item.Status); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate readings: %w", err)
	}
	return results, nil
}

func (repository *Repository) AllEvents(ctx context.Context) ([]event.Event, error) {
	rows, err := repository.pool.Query(ctx, "SELECT meter_id, occurred_at, event_type, description FROM events ORDER BY meter_id, occurred_at")
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()
	results := make([]event.Event, 0)
	for rows.Next() {
		var item event.Event
		if err := rows.Scan(&item.MeterID, &item.Timestamp, &item.Type, &item.Description); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return results, nil
}

func (repository *Repository) SaveAnalysis(ctx context.Context, result ai.Analysis) error {
	anomaliesJSON, err := json.Marshal(result.Anomalies)
	if err != nil {
		return fmt.Errorf("encode analysis anomalies: %w", err)
	}
	var insightJSON []byte
	if result.Insight != nil {
		insightJSON, err = json.Marshal(result.Insight)
		if err != nil {
			return fmt.Errorf("encode analysis insight: %w", err)
		}
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin save analysis: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO analyses(id, status, started_at, completed_at, anomaly_count, anomalies, error, ai_response, ai_error)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(id) DO UPDATE SET status=EXCLUDED.status, completed_at=EXCLUDED.completed_at,
		anomaly_count=EXCLUDED.anomaly_count, anomalies=EXCLUDED.anomalies, error=EXCLUDED.error,
		ai_response=EXCLUDED.ai_response, ai_error=EXCLUDED.ai_error`,
		result.ID, result.Status, result.StartedAt, result.CompletedAt, len(result.Anomalies), anomaliesJSON, result.Error, insightJSON, result.InsightError)
	if err != nil {
		return fmt.Errorf("upsert analysis: %w", err)
	}
	if result.Status == "COMPLETED" {
		if _, err := tx.Exec(ctx, "DELETE FROM anomalies WHERE analysis_id=$1", result.ID); err != nil {
			return fmt.Errorf("replace analysis anomalies: %w", err)
		}
		for _, item := range result.Anomalies {
			evidenceJSON, err := json.Marshal(item.Evidence)
			if err != nil {
				return fmt.Errorf("encode anomaly evidence: %w", err)
			}
			_, err = tx.Exec(ctx, `INSERT INTO anomalies(id, analysis_id, meter_id, type, severity, confidence, reason, recommended_action, detected_at, evidence)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, item.ID, result.ID, item.MeterID, item.Type, item.Severity, item.Confidence, item.Reason, item.RecommendedAction, item.DetectedAt, evidenceJSON)
			if err != nil {
				return fmt.Errorf("save anomaly %s: %w", item.ID, err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit analysis: %w", err)
	}
	return nil
}

func (repository *Repository) FindAnalysis(ctx context.Context, id string) (ai.Analysis, error) {
	var result ai.Analysis
	var anomaliesJSON []byte
	var insightJSON []byte
	err := repository.pool.QueryRow(ctx, "SELECT id, status, started_at, completed_at, anomalies, error, ai_response, ai_error FROM analyses WHERE id=$1", id).
		Scan(&result.ID, &result.Status, &result.StartedAt, &result.CompletedAt, &anomaliesJSON, &result.Error, &insightJSON, &result.InsightError)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ai.Analysis{}, ai.ErrNotFound
		}
		return ai.Analysis{}, err
	}
	if err := json.Unmarshal(anomaliesJSON, &result.Anomalies); err != nil {
		return ai.Analysis{}, fmt.Errorf("decode analysis anomalies: %w", err)
	}
	if len(insightJSON) > 0 {
		result.Insight = &ai.Insight{}
		if err := json.Unmarshal(insightJSON, result.Insight); err != nil {
			return ai.Analysis{}, fmt.Errorf("decode analysis insight: %w", err)
		}
	}
	return result, nil
}

func (repository *Repository) ListAnomalies(ctx context.Context) ([]anomaly.Anomaly, error) {
	rows, err := repository.pool.Query(ctx, `SELECT id, meter_id, type, severity, confidence, reason, recommended_action, detected_at, evidence
		FROM anomalies WHERE analysis_id=(SELECT id FROM analyses WHERE status='COMPLETED' ORDER BY completed_at DESC LIMIT 1)
		ORDER BY CASE severity WHEN 'CRITICAL' THEN 0 WHEN 'HIGH' THEN 1 WHEN 'MEDIUM' THEN 2 ELSE 3 END, meter_id`)
	if err != nil {
		return nil, fmt.Errorf("query anomalies: %w", err)
	}
	defer rows.Close()
	results := make([]anomaly.Anomaly, 0)
	for rows.Next() {
		var item anomaly.Anomaly
		var evidenceJSON []byte
		if err := rows.Scan(&item.ID, &item.MeterID, &item.Type, &item.Severity, &item.Confidence, &item.Reason, &item.RecommendedAction, &item.DetectedAt, &evidenceJSON); err != nil {
			return nil, fmt.Errorf("scan anomaly: %w", err)
		}
		if err := json.Unmarshal(evidenceJSON, &item.Evidence); err != nil {
			return nil, fmt.Errorf("decode anomaly evidence: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate anomalies: %w", err)
	}
	return results, nil
}

func (repository *Repository) FindAnomaly(ctx context.Context, id string) (anomaly.Anomaly, error) {
	var item anomaly.Anomaly
	var evidenceJSON []byte
	err := repository.pool.QueryRow(ctx, `SELECT id, meter_id, type, severity, confidence, reason, recommended_action, detected_at, evidence
		FROM anomalies WHERE id=$1`, id).Scan(&item.ID, &item.MeterID, &item.Type, &item.Severity, &item.Confidence, &item.Reason, &item.RecommendedAction, &item.DetectedAt, &evidenceJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return anomaly.Anomaly{}, anomalyapp.ErrNotFound
		}
		return anomaly.Anomaly{}, err
	}
	if err := json.Unmarshal(evidenceJSON, &item.Evidence); err != nil {
		return anomaly.Anomaly{}, fmt.Errorf("decode anomaly evidence: %w", err)
	}
	return item, nil
}

func (repository *Repository) LatestAnalysis(ctx context.Context) (ai.Analysis, error) {
	var id string
	err := repository.pool.QueryRow(ctx, "SELECT id FROM analyses ORDER BY started_at DESC LIMIT 1").Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ai.Analysis{}, ai.ErrNotFound
		}
		return ai.Analysis{}, err
	}
	return repository.FindAnalysis(ctx, id)
}

func (repository *Repository) MeterIDs(ctx context.Context) ([]string, error) {
	rows, err := repository.pool.Query(ctx, "SELECT meter_id FROM meters ORDER BY meter_id")
	if err != nil {
		return nil, fmt.Errorf("query meters: %w", err)
	}
	defer rows.Close()
	results := make([]string, 0)
	for rows.Next() {
		var meterID string
		if err := rows.Scan(&meterID); err != nil {
			return nil, fmt.Errorf("scan meter id: %w", err)
		}
		results = append(results, meterID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meters: %w", err)
	}
	return results, nil
}
