package httpServer

import (
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
)

type meterResponse struct {
	MeterID               string           `json:"meter_id"`
	Provider              string           `json:"provider,omitempty"`
	Region                string           `json:"region,omitempty"`
	RateType              string           `json:"rate_type,omitempty"`
	CurrentConsumptionKWh float64          `json:"current_consumption_kwh"`
	PeriodConsumptionKWh  float64          `json:"period_consumption_kwh"`
	BaselineKWh           float64          `json:"baseline_kwh"`
	ChangePercent         float64          `json:"change_percent"`
	Status                string           `json:"status"`
	Severity              anomaly.Severity `json:"severity"`
	Anomaly               *anomalyResponse `json:"anomaly,omitempty"`
}

type meterConfigRequestDTO struct {
	Provider string `json:"provider,omitempty"`
	Region   string `json:"region,omitempty"`
	RateType string `json:"rate_type,omitempty"`
}

type meterConfigResponseDTO struct {
	MeterID   string    `json:"meter_id"`
	Provider  string    `json:"provider,omitempty"`
	Region    string    `json:"region,omitempty"`
	RateType  string    `json:"rate_type,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type meterDetailResponse struct {
	meterResponse
	VoltageV    float64           `json:"voltage_v"`
	CurrentA    float64           `json:"current_a"`
	PowerFactor float64           `json:"power_factor"`
	History     []readingResponse `json:"history"`
}

type readingResponse struct {
	Timestamp      time.Time `json:"timestamp"`
	ConsumptionKWh float64   `json:"consumption_kwh"`
	VoltageV       float64   `json:"voltage_v"`
	CurrentA       float64   `json:"current_a"`
	PowerFactor    float64   `json:"power_factor"`
	Status         string    `json:"status"`
}

type siteHistoryPointResponse struct {
	Timestamp      time.Time `json:"timestamp"`
	ConsumptionKWh float64   `json:"consumption_kwh"`
}

type evidenceResponse struct {
	BaselineKWh             float64  `json:"baseline_kwh"`
	CurrentKWh              float64  `json:"current_kwh"`
	ChangePercent           float64  `json:"change_percent"`
	VoltageV                float64  `json:"voltage_v"`
	CurrentA                float64  `json:"current_a"`
	PowerFactor             float64  `json:"power_factor"`
	CurrentVariationPercent float64  `json:"current_variation_percent"`
	RelatedEvents           []string `json:"related_events"`
	AbnormalHours           []int    `json:"abnormal_hours"`
	OutlierCount            int      `json:"outlier_count"`
}

type tariffRequestDTO struct {
	Provider    string     `json:"provider"`
	Region      string     `json:"region,omitempty"`
	RateType    string     `json:"rate_type"`
	PricePerKWh float64    `json:"price_per_kwh"`
	Currency    string     `json:"currency,omitempty"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidTo     *time.Time `json:"valid_to,omitempty"`
}

type tariffResponseDTO struct {
	TariffID    int        `json:"tariff_id"`
	Provider    string     `json:"provider"`
	Region      string     `json:"region,omitempty"`
	RateType    string     `json:"rate_type"`
	PricePerKWh float64    `json:"price_per_kwh"`
	Currency    string     `json:"currency"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidTo     *time.Time `json:"valid_to,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type billingResponseDTO struct {
	BillingID   int       `json:"billing_id"`
	MeterID     string    `json:"meter_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	TotalKWh    float64   `json:"total_kwh"`
	PricePerKWh float64   `json:"price_per_kwh"`
	TotalCost   float64   `json:"total_cost"`
	Currency    string    `json:"currency"`
	TariffID    *int      `json:"tariff_id,omitempty"`
}

type estimateConsumptionRequestDTO struct {
	MeterID     string    `json:"meter_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

type estimateConsumptionResponseDTO struct {
	MeterID     string    `json:"meter_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	TotalKWh    float64   `json:"total_kwh"`
	PricePerKWh float64   `json:"price_per_kwh"`
	TotalCost   float64   `json:"total_cost"`
	Currency    string    `json:"currency"`
}

type billingSummaryResponseDTO struct {
	TotalCost         float64   `json:"total_cost"`
	PreviousTotalCost float64   `json:"previous_total_cost,omitempty"`
	ChangePercent     float64   `json:"change_percent,omitempty"`
	AverageDailyCost  float64   `json:"average_daily_cost,omitempty"`
	TopMeterID        string    `json:"top_meter_id,omitempty"`
	Currency          string    `json:"currency"`
	PeriodStart       time.Time `json:"period_start,omitempty"`
	PeriodEnd         time.Time `json:"period_end,omitempty"`
}

type costTimePointResponseDTO struct {
	Label     string    `json:"label"`
	Date      time.Time `json:"date,omitempty"`
	TotalCost float64   `json:"total_cost"`
	KWh       float64   `json:"kwh"`
}

type costByMeterRowResponseDTO struct {
	MeterID     string  `json:"meter_id"`
	KWh         float64 `json:"kwh"`
	TariffName  string  `json:"tariff_name,omitempty"`
	DailyCost   float64 `json:"daily_cost"`
	WeeklyCost  float64 `json:"weekly_cost"`
	PeriodCost  float64 `json:"period_cost"`
	Trend       float64 `json:"trend,omitempty"`
	Calculated  bool    `json:"calculated,omitempty"`
	Currency    string  `json:"currency,omitempty"`
}

type billingReportFileResponseDTO struct {
	Content     []byte `json:"-"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type anomalyResponse struct {
	ID                string           `json:"id"`
	MeterID           string           `json:"meter_id"`
	Anomaly           string           `json:"anomaly"`
	Type              anomaly.Type     `json:"type"`
	Severity          anomaly.Severity `json:"severity"`
	Confidence        float64          `json:"confidence"`
	Reason            string           `json:"reason"`
	RecommendedAction string           `json:"recommended_action"`
	DetectedAt        time.Time        `json:"detected_at"`
	Evidence          evidenceResponse `json:"evidence"`
}

type analysisResponse struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	StartedAt    time.Time         `json:"started_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	Anomalies    []anomalyResponse `json:"anomalies"`
	Error        string            `json:"error,omitempty"`
	Insight      *insightResponse  `json:"insight,omitempty"`
	InsightError string            `json:"insight_error,omitempty"`
}

type insightResponse struct {
	Answer             string                      `json:"answer"`
	Explanation        string                      `json:"explanation"`
	SuggestedQuestions []suggestedQuestionResponse `json:"suggested_questions"`
}

type optimizationInsightResponseDTO struct {
	Summary              string    `json:"summary"`
	Recommendations      []string  `json:"recommendations"`
	PotentialSavingsKWh  float64   `json:"potential_savings_kwh"`
	PotentialSavingsCost float64   `json:"potential_savings_cost"`
	Currency             string    `json:"currency"`
	GeneratedAt          time.Time `json:"generated_at"`
}

type suggestedQuestionResponse struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func mapAnalysis(result ai.Analysis) analysisResponse {
	anomalies := make([]anomalyResponse, 0, len(result.Anomalies))
	for _, item := range result.Anomalies {
		anomalies = append(anomalies, mapAnomaly(item))
	}
	response := analysisResponse{ID: result.ID, Status: result.Status, StartedAt: result.StartedAt, CompletedAt: result.CompletedAt, Anomalies: anomalies, Error: result.Error, InsightError: result.InsightError}
	if result.Insight != nil {
		questions := make([]suggestedQuestionResponse, 0, len(result.Insight.SuggestedQuestions))
		for _, item := range result.Insight.SuggestedQuestions {
			questions = append(questions, suggestedQuestionResponse{Question: item.Question, Answer: item.Answer})
		}
		response.Insight = &insightResponse{Answer: result.Insight.Answer, Explanation: result.Insight.Explanation, SuggestedQuestions: questions}
	}
	return response
}
