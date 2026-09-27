package anomaly

import "time"

type Type string

const (
	RealAnomaly        Type = "REAL_ANOMALY"
	ExplainableAnomaly Type = "EXPLAINABLE_ANOMALY"
	FalsePositive      Type = "FALSE_POSITIVE"
	DataQuality        Type = "DATA_QUALITY"
)

type Severity string

const (
	Low      Severity = "LOW"
	Medium   Severity = "MEDIUM"
	High     Severity = "HIGH"
	Critical Severity = "CRITICAL"
)

type Evidence struct {
	BaselineKWh             float64
	CurrentKWh              float64
	ChangePercent           float64
	VoltageV                float64
	CurrentA                float64
	PowerFactor             float64
	CurrentVariationPercent float64
	RelatedEvents           []string
	AbnormalHours           []int
	OutlierCount            int
}

type Anomaly struct {
	ID                string
	MeterID           string
	Anomaly           string
	Type              Type
	Severity          Severity
	Confidence        float64
	Reason            string
	RecommendedAction string
	DetectedAt        time.Time
	Evidence          Evidence
}
