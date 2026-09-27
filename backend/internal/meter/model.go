package meter

import (
	"time"

	"learning/internal/anomaly"
	"learning/internal/reading"
)

type Meter struct {
	MeterID  string
	Provider string
	Region   string
	RateType string
}

type Summary struct {
	Meter
	CurrentConsumptionKWh float64
	PeriodConsumptionKWh  float64
	BaselineKWh           float64
	ChangePercent         float64
	Status                string
	Severity              anomaly.Severity
	Anomaly               *anomaly.Anomaly
}

type Detail struct {
	Summary
	Readings    []reading.Reading
	VoltageV    float64
	CurrentA    float64
	PowerFactor float64
}

type SitePoint struct {
	Timestamp      time.Time
	ConsumptionKWh float64
}
