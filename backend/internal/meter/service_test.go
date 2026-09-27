package meter

import (
	"context"
	"testing"
	"time"

	"learning/internal/anomaly"
	"learning/internal/reading"
)

func TestSiteHistoryAggregatesHours(t *testing.T) {
	base := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	items := []reading.Reading{
		{MeterID: "M-101", Timestamp: base, ConsumptionKWh: 10},
		{MeterID: "M-102", Timestamp: base.Add(30 * time.Minute), ConsumptionKWh: 20},
		{MeterID: "M-101", Timestamp: base.Add(time.Hour), ConsumptionKWh: 5},
	}
	service := NewService(fakeServiceReadings{items: items}, fakeServiceAnomalies{}, fakeServiceProfiles{})
	points, err := service.SiteHistory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 {
		t.Fatalf("len(points) = %d, want 2", len(points))
	}
	if points[0].ConsumptionKWh != 30 || points[1].ConsumptionKWh != 5 {
		t.Fatalf("points = %+v, want [30 5]", points)
	}
	if points[0].Timestamp.After(points[1].Timestamp) {
		t.Fatalf("points not ordered: %+v", points)
	}
}

func TestListEnrichesMetersWithProfile(t *testing.T) {
	items := make([]reading.Reading, 48)
	for index := range items {
		items[index] = reading.Reading{MeterID: "M-101", ConsumptionKWh: 20}
	}
	service := NewService(fakeServiceReadings{items: items}, fakeServiceAnomalies{}, fakeServiceProfiles{profile: Profile{Provider: "EPM", Region: "Medellín", RateType: "industrial"}})
	results, err := service.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Provider != "EPM" || results[0].Region != "Medellín" || results[0].RateType != "industrial" {
		t.Fatalf("profile = %+v, want EPM/Medellín/industrial", results[0].Meter)
	}
}

type fakeServiceReadings struct {
	items []reading.Reading
}

func (fake fakeServiceReadings) AllReadings(context.Context) ([]reading.Reading, error) {
	return fake.items, nil
}

func (fake fakeServiceReadings) ReadingsByMeter(_ context.Context, meterID string) ([]reading.Reading, error) {
	results := make([]reading.Reading, 0)
	for _, item := range fake.items {
		if item.MeterID == meterID {
			results = append(results, item)
		}
	}
	return results, nil
}

type fakeServiceAnomalies struct{}

func (fakeServiceAnomalies) ListAnomalies(context.Context) ([]anomaly.Anomaly, error) {
	return nil, nil
}

type fakeServiceProfiles struct {
	profile Profile
}

func (fake fakeServiceProfiles) MeterProfile(context.Context, string) (Profile, error) {
	return fake.profile, nil
}

func TestSummaryDoesNotEscalateLowFalsePositive(t *testing.T) {
	series := make([]reading.Reading, 192)
	for index := range series {
		series[index] = reading.Reading{ConsumptionKWh: 50, Timestamp: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(index) * time.Hour)}
	}
	falsePositive := anomaly.Anomaly{MeterID: "M-106", Type: anomaly.FalsePositive, Severity: anomaly.Low}
	result := summarize("M-106", series, map[string]anomaly.Anomaly{"M-106": falsePositive}, Profile{})
	if result.Status != "NORMAL" {
		t.Fatalf("false-positive status = %s for %s/%s, want NORMAL", result.Status, result.Anomaly.Type, result.Anomaly.Severity)
	}

	highPriority := anomaly.Anomaly{MeterID: "M-109", Type: anomaly.RealAnomaly, Severity: anomaly.High}
	result = summarize("M-109", series, map[string]anomaly.Anomaly{"M-109": highPriority}, Profile{})
	if result.Status != "ALERT" {
		t.Fatalf("high-severity status = %s, want ALERT", result.Status)
	}
}

func TestSummaryEscalatesCriticalAndExplainableAnomalies(t *testing.T) {
	series := make([]reading.Reading, 192)
	for index := range series {
		series[index] = reading.Reading{ConsumptionKWh: 50, Timestamp: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(index) * time.Hour)}
	}

	critical := anomaly.Anomaly{MeterID: "M-300", Type: anomaly.RealAnomaly, Severity: anomaly.Critical}
	result := summarize("M-300", series, map[string]anomaly.Anomaly{"M-300": critical}, Profile{})
	if result.Status != "CRITICAL" {
		t.Fatalf("critical status = %s, want CRITICAL", result.Status)
	}

	explainable := anomaly.Anomaly{MeterID: "M-301", Type: anomaly.ExplainableAnomaly, Severity: anomaly.Medium}
	result = summarize("M-301", series, map[string]anomaly.Anomaly{"M-301": explainable}, Profile{})
	if result.Status != "ALERT" {
		t.Fatalf("explainable anomaly status = %s, want ALERT", result.Status)
	}
}
