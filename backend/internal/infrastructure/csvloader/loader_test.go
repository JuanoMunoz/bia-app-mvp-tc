package csvloader

import (
	"path/filepath"
	"testing"

	"learning/internal/anomaly"
)

func TestSuppliedDatasetMatchesRequiredAnomalyCases(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "data")
	dataset, err := Load(filepath.Join(root, "readings.csv"), filepath.Join(root, "events.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(dataset.Readings) != 4032 {
		t.Fatalf("loaded %d readings, want 4032", len(dataset.Readings))
	}
	if len(dataset.Events) != 4 {
		t.Fatalf("loaded %d events, want 4", len(dataset.Events))
	}

	results := (anomaly.Detector{}).Detect(dataset.Readings, dataset.Events)
	byMeter := make(map[string]anomaly.Anomaly, len(results))
	for _, result := range results {
		byMeter[result.MeterID] = result
	}
	wants := map[string]struct {
		kind     anomaly.Type
		severity anomaly.Severity
	}{
		"M-104": {kind: anomaly.ExplainableAnomaly, severity: anomaly.Medium},
		"M-106": {kind: anomaly.FalsePositive, severity: anomaly.Low},
		"M-109": {kind: anomaly.RealAnomaly, severity: anomaly.High},
		"M-112": {kind: anomaly.DataQuality, severity: anomaly.High},
	}
	for meterID, want := range wants {
		got, ok := byMeter[meterID]
		if !ok {
			t.Errorf("no anomaly detected for %s", meterID)
			continue
		}
		if got.Type != want.kind || got.Severity != want.severity {
			t.Errorf("%s = %s/%s, want %s/%s", meterID, got.Type, got.Severity, want.kind, want.severity)
		}
		if meterID == "M-104" && len(got.Evidence.AbnormalHours) < 4 {
			t.Errorf("M-104 has %d abnormal hours, want at least 4", len(got.Evidence.AbnormalHours))
		}
		if meterID == "M-106" && got.Evidence.OutlierCount == 0 {
			t.Error("M-106 scheduled outage should include point-outlier evidence")
		}
		if meterID == "M-112" && got.Evidence.CurrentVariationPercent < 25 {
			t.Errorf("M-112 current variation = %.1f%%, want at least 25%%", got.Evidence.CurrentVariationPercent)
		}
	}
}
