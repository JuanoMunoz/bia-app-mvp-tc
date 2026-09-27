package anomaly

import (
	"testing"
	"time"

	"learning/internal/event"
	"learning/internal/reading"
)

func TestDetectorClassifiesKnownCases(t *testing.T) {
	baseTime := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		meterID      string
		baseline     float64
		current      float64
		voltage      float64
		currentA     float64
		powerFactor  float64
		changesAt    int
		eventType    string
		wantType     Type
		wantSeverity Severity
	}{
		{name: "M-104 production line", meterID: "M-104", baseline: 50, current: 75, voltage: 220, currentA: 300, powerFactor: 0.92, changesAt: 200, eventType: "OPERATIONAL_CHANGE", wantType: ExplainableAnomaly, wantSeverity: Medium},
		{name: "M-106 scheduled outage", meterID: "M-106", baseline: 50, current: 50, voltage: 220, currentA: 250, powerFactor: 0.93, changesAt: 168, eventType: "SCHEDULED_OUTAGE", wantType: FalsePositive, wantSeverity: Low},
		{name: "M-109 unexplained increase", meterID: "M-109", baseline: 40, current: 90, voltage: 217, currentA: 400, powerFactor: 0.74, changesAt: 200, eventType: "UNKNOWN", wantType: RealAnomaly, wantSeverity: High},
		{name: "M-112 electrical quality", meterID: "M-112", baseline: 30, current: 30, voltage: 240, currentA: 120, powerFactor: 0.58, changesAt: 200, eventType: "DATA_QUALITY", wantType: DataQuality, wantSeverity: High},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := make([]reading.Reading, 14*24)
			for index := range items {
				consumption := test.baseline
				powerFactor := 0.92
				if index >= test.changesAt {
					consumption = test.current
					powerFactor = test.powerFactor
				}
				items[index] = reading.Reading{
					MeterID:        test.meterID,
					Timestamp:      baseTime.Add(time.Duration(index) * time.Hour),
					ConsumptionKWh: consumption,
					VoltageV:       test.voltage,
					CurrentA:       test.currentA,
					PowerFactor:    powerFactor,
					Status:         "OK",
				}
			}
			if test.eventType == "SCHEDULED_OUTAGE" {
				for index := test.changesAt; index < test.changesAt+12; index++ {
					items[index].ConsumptionKWh = 5
				}
			}
			if test.wantType == DataQuality {
				for index := len(items) - 24; index < len(items); index++ {
					items[index].PowerFactor = 0.58
				}
			}
			events := []event.Event{{
				MeterID:     test.meterID,
				Timestamp:   baseTime.Add(time.Duration(test.changesAt) * time.Hour),
				Type:        test.eventType,
				Description: "Known operational event",
			}}

			results := (Detector{}).Detect(items, events)
			if len(results) != 1 {
				t.Fatalf("Detect() returned %d anomalies, want 1", len(results))
			}
			if results[0].Type != test.wantType || results[0].Severity != test.wantSeverity {
				t.Fatalf("Detect() = %s/%s, want %s/%s", results[0].Type, results[0].Severity, test.wantType, test.wantSeverity)
			}
			if results[0].Anomaly == "" {
				t.Fatal("Detect() returned an empty anomaly label")
			}
		})
	}
}

func TestDetectorFindsHourlyProfileShiftAndConsumptionOutlier(t *testing.T) {
	baseTime := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		mutate    func([]reading.Reading)
		wantHours int
		wantSpike int
	}{
		{
			name: "hourly profile shift",
			mutate: func(items []reading.Reading) {
				for index := range items {
					if index >= 7*24 && items[index].Timestamp.Hour() >= 8 && items[index].Timestamp.Hour() < 12 {
						items[index].ConsumptionKWh = 75
					}
				}
			},
			wantHours: 4,
		},
		{
			name: "isolated spike",
			mutate: func(items []reading.Reading) {
				items[len(items)-10].ConsumptionKWh = 100
			},
			wantSpike: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := make([]reading.Reading, 14*24)
			for index := range items {
				items[index] = reading.Reading{
					MeterID:        "M-120",
					Timestamp:      baseTime.Add(time.Duration(index) * time.Hour),
					ConsumptionKWh: 50,
					VoltageV:       220,
					CurrentA:       100,
					PowerFactor:    0.95,
					Status:         "OK",
				}
			}
			test.mutate(items)
			results := (Detector{}).Detect(items, nil)
			if len(results) != 1 {
				t.Fatalf("Detect() returned %d anomalies, want 1", len(results))
			}
			if len(results[0].Evidence.AbnormalHours) != test.wantHours {
				t.Errorf("abnormal hours = %d, want %d", len(results[0].Evidence.AbnormalHours), test.wantHours)
			}
			if results[0].Evidence.OutlierCount != test.wantSpike {
				t.Errorf("outlier count = %d, want %d", results[0].Evidence.OutlierCount, test.wantSpike)
			}
		})
	}
}

func TestDetectorIgnoresDistantEvents(t *testing.T) {
	baseTime := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	items := make([]reading.Reading, 14*24)
	for index := range items {
		consumption := 40.0
		powerFactor := 0.95
		currentA := 150.0
		if index >= 200 {
			consumption = 90
			powerFactor = 0.74
			currentA = 400
		}
		items[index] = reading.Reading{
			MeterID:        "M-109",
			Timestamp:      baseTime.Add(time.Duration(index) * time.Hour),
			ConsumptionKWh: consumption,
			VoltageV:       217,
			CurrentA:       currentA,
			PowerFactor:    powerFactor,
			Status:         "OK",
		}
	}
	events := []event.Event{{
		MeterID:     "M-109",
		Timestamp:   baseTime.Add(20 * time.Hour),
		Type:        "OPERATIONAL_CHANGE",
		Description: "Old unrelated operational event",
	}}

	results := (Detector{}).Detect(items, events)
	if len(results) != 1 {
		t.Fatalf("Detect() returned %d anomalies, want 1", len(results))
	}
	if results[0].Type != RealAnomaly || results[0].Severity != High {
		t.Fatalf("Detect() = %s/%s, want %s/%s", results[0].Type, results[0].Severity, RealAnomaly, High)
	}
	if len(results[0].Evidence.RelatedEvents) != 0 {
		t.Fatalf("RelatedEvents = %v, want empty", results[0].Evidence.RelatedEvents)
	}
}

func TestDetectorIgnoresShortSeries(t *testing.T) {
	baseTime := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	items := make([]reading.Reading, 20)
	for index := range items {
		items[index] = reading.Reading{
			MeterID:        "M-99",
			Timestamp:      baseTime.Add(time.Duration(index) * time.Hour),
			ConsumptionKWh: 50,
			VoltageV:       220,
			CurrentA:       100,
			PowerFactor:    0.95,
			Status:         "OK",
		}
	}

	if results := (Detector{}).Detect(items, nil); len(results) != 0 {
		t.Fatalf("Detect() on short series returned %d anomalies, want 0", len(results))
	}
}

func TestDetectorDetectsScheduledOutageAsFalsePositive(t *testing.T) {
	baseTime := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	items := make([]reading.Reading, 14*24)
	for index := range items {
		items[index] = reading.Reading{
			MeterID:        "M-200",
			Timestamp:      baseTime.Add(time.Duration(index) * time.Hour),
			ConsumptionKWh: 50,
			VoltageV:       220,
			CurrentA:       120,
			PowerFactor:    0.94,
			Status:         "OK",
		}
	}
	for index := 200; index < 212; index++ {
		items[index].ConsumptionKWh = 5
	}
	events := []event.Event{{
		MeterID:     "M-200",
		Timestamp:   baseTime.Add(200 * time.Hour),
		Type:        "SCHEDULED_OUTAGE",
		Description: "Parada programada de equipo",
	}}

	results := (Detector{}).Detect(items, events)
	if len(results) != 1 {
		t.Fatalf("Detect() returned %d anomalies, want 1", len(results))
	}
	if results[0].Type != FalsePositive || results[0].Severity != Low {
		t.Fatalf("Detect() = %s/%s, want %s/%s", results[0].Type, results[0].Severity, FalsePositive, Low)
	}
}
