package anomaly

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"learning/internal/event"
	"learning/internal/reading"
)

type Detector struct{}

const (
	minSeriesLength        = 48
	baselineWindowHours    = 7 * 24
	currentWindowHours     = 24
	hourlyHistoryHours     = 14 * 24
	changeThresholdPct     = 25.0
	qualityMinIssues       = 3
	currentVariationLimit  = 0.25
	abnormalHoursMin       = 4
	outlierMinSeries       = 5
	outlierNeighborRadius  = 2
	outlierDeviationLimit  = 0.5
	outageWindowHours      = 12
	outageDropThreshold    = 0.25
	voltageMinV            = 207.0
	voltageMaxV            = 253.0
	powerFactorMin         = 0.7
	eventCorrelationWindow = 24 * time.Hour
)

func (Detector) Detect(readings []reading.Reading, events []event.Event) []Anomaly {
	byMeter := make(map[string][]reading.Reading)
	for _, item := range readings {
		byMeter[item.MeterID] = append(byMeter[item.MeterID], item)
	}
	eventsByMeter := make(map[string][]event.Event)
	for _, item := range events {
		eventsByMeter[item.MeterID] = append(eventsByMeter[item.MeterID], item)
	}

	meterIDs := make([]string, 0, len(byMeter))
	for meterID := range byMeter {
		meterIDs = append(meterIDs, meterID)
	}
	sort.Strings(meterIDs)

	results := make([]Anomaly, 0)
	for _, meterID := range meterIDs {
		series := byMeter[meterID]
		sort.Slice(series, func(i, j int) bool { return series[i].Timestamp.Before(series[j].Timestamp) })
		if result, ok := detectMeter(series, eventsByMeter[meterID]); ok {
			results = append(results, result)
		}
	}
	return results
}

func detectMeter(series []reading.Reading, events []event.Event) (Anomaly, bool) {
	if len(series) < minSeriesLength {
		return Anomaly{}, false
	}
	baselineWindow := baselineWindowHours
	if len(series) < baselineWindow+currentWindowHours {
		baselineWindow = len(series) - currentWindowHours
	}
	baseline := meanConsumption(series[:baselineWindow])
	current := series[len(series)-currentWindowHours:]
	currentMean := meanConsumption(current)
	change := percentChange(baseline, currentMean)
	voltage := meanField(current, func(item reading.Reading) float64 { return item.VoltageV })
	currentA := meanField(current, func(item reading.Reading) float64 { return item.CurrentA })
	powerFactor := meanField(current, func(item reading.Reading) float64 { return item.PowerFactor })

	evidence := Evidence{
		BaselineKWh:             round(baseline),
		CurrentKWh:              round(currentMean),
		ChangePercent:           round(change),
		VoltageV:                round(voltage),
		CurrentA:                round(currentA),
		PowerFactor:             round(powerFactor),
		CurrentVariationPercent: round(coefficientVariation(current) * 100),
		AbnormalHours:           abnormalHours(series),
		OutlierCount:            len(consumptionOutliers(series)),
	}
	qualityReadings := qualityIssue(current)
	currentVariation := coefficientVariation(current)
	if (qualityReadings >= qualityMinIssues || currentVariation >= currentVariationLimit) && math.Abs(change) < changeThresholdPct {
		qualityReason := fmt.Sprintf("El consumo se mantiene cerca del baseline (%+.1f%%), pero %d de las últimas 24 lecturas presentan voltaje, factor de potencia o estado fuera de rango", change, qualityReadings)
		if currentVariation >= currentVariationLimit {
			qualityReason += fmt.Sprintf(" y la corriente tiene una variación relativa de %.1f%%", currentVariation*100)
		}
		return Anomaly{
			ID:                series[0].MeterID + "-quality",
			MeterID:           series[0].MeterID,
			Anomaly:           "Inconsistencia en variables eléctricas",
			Type:              DataQuality,
			Severity:          High,
			Confidence:        0.94,
			Reason:            qualityReason + ".",
			RecommendedAction: "Inspeccionar el medidor, sus transformadores de corriente y la cadena de adquisición antes de tomar decisiones operativas.",
			DetectedAt:        current[len(current)-1].Timestamp,
			Evidence:          evidence,
		}, true
	}

	if math.Abs(change) >= changeThresholdPct {
		changeStart := findChangeStart(series, baseline, baselineWindow)
		related := relatedEvents(events, changeStart, changeStart)
		kind, severity, action := RealAnomaly, High, "Verificar de inmediato el medidor y la carga conectada; comparar con la orden operativa antes de ajustar equipos."
		confidence := 0.91
		if hasEventType(related, "SCHEDULED_OUTAGE") {
			kind, severity, action = FalsePositive, Low, "No escalar: confirmar el cierre de la parada programada y validar el retorno al perfil habitual."
			confidence = 0.88
		} else if hasEventType(related, "OPERATIONAL_CHANGE") {
			kind, severity, action = ExplainableAnomaly, Medium, "Validar que la nueva línea esté operando según lo previsto y actualizar el baseline de consumo."
			confidence = 0.9
		}
		evidence.RelatedEvents = eventDescriptions(related)
		return Anomaly{
			ID:                series[0].MeterID + "-consumption",
			MeterID:           series[0].MeterID,
			Anomaly:           "Cambio de consumo",
			Type:              kind,
			Severity:          severity,
			Confidence:        confidence,
			Reason:            fmt.Sprintf("El consumo medio de las últimas 24 horas es %.1f kWh frente a un baseline de %.1f kWh (%+.1f%%).", currentMean, baseline, change),
			RecommendedAction: action,
			DetectedAt:        current[len(current)-1].Timestamp,
			Evidence:          evidence,
		}, true
	}

	for _, item := range events {
		if item.Type != "SCHEDULED_OUTAGE" {
			continue
		}
		if result, ok := detectScheduledOutage(series, item, evidence); ok {
			return result, true
		}
	}
	if len(evidence.AbnormalHours) >= abnormalHoursMin {
		start := firstAbnormalHour(series)
		related := relatedEvents(events, start, start.Add(23*time.Hour))
		kind, severity, action, confidence := RealAnomaly, High, "Revisar el perfil operativo y comparar las cargas conectadas durante las horas afectadas.", 0.86
		if hasEventType(related, "OPERATIONAL_CHANGE") {
			kind, severity, action, confidence = ExplainableAnomaly, Medium, "Validar el nuevo perfil horario contra el cambio operativo y actualizar el baseline por franja.", 0.9
		}
		evidence.RelatedEvents = eventDescriptions(related)
		return Anomaly{
			ID:                series[0].MeterID + "-hourly",
			MeterID:           series[0].MeterID,
			Anomaly:           "Patrón horario anómalo",
			Type:              kind,
			Severity:          severity,
			Confidence:        confidence,
			Reason:            fmt.Sprintf("El perfil por hora se desvía al menos 25%% del baseline en %d de 24 franjas; el promedio diario reciente solo cambia %+.1f%%.", len(evidence.AbnormalHours), change),
			RecommendedAction: action,
			DetectedAt:        start,
			Evidence:          evidence,
		}, true
	}
	if outlier, ok := firstConsumptionOutlier(series); ok {
		related := relatedEvents(events, outlier.Timestamp, outlier.Timestamp)
		kind, severity, action, confidence := RealAnomaly, High, "Inspeccionar la lectura y el equipo asociado; confirmar si el pico representa una carga transitoria real.", 0.9
		if hasEventType(related, "SCHEDULED_OUTAGE") {
			kind, severity, action, confidence = FalsePositive, Low, "No escalar: confirmar el registro de la parada programada y validar la continuidad de las lecturas.", 0.88
		}
		evidence.RelatedEvents = eventDescriptions(related)
		return Anomaly{
			ID:                series[0].MeterID + "-outlier",
			MeterID:           series[0].MeterID,
			Anomaly:           "Pico puntual de consumo",
			Type:              kind,
			Severity:          severity,
			Confidence:        confidence,
			Reason:            fmt.Sprintf("La lectura de %.1f kWh en %s se desvía más de 50%% de la mediana de las cuatro horas vecinas.", outlier.ConsumptionKWh, outlier.Timestamp.Format(time.RFC3339)),
			RecommendedAction: action,
			DetectedAt:        outlier.Timestamp,
			Evidence:          evidence,
		}, true
	}
	return Anomaly{}, false
}

func abnormalHours(series []reading.Reading) []int {
	if len(series) < hourlyHistoryHours {
		return nil
	}
	baseline := make([][]float64, 24)
	current := make([][]float64, 24)
	for _, item := range series[:baselineWindowHours] {
		hour := item.Timestamp.Hour()
		baseline[hour] = append(baseline[hour], item.ConsumptionKWh)
	}
	for _, item := range series[len(series)-baselineWindowHours:] {
		hour := item.Timestamp.Hour()
		current[hour] = append(current[hour], item.ConsumptionKWh)
	}
	result := make([]int, 0)
	for hour := range baseline {
		baselineMean := meanValues(baseline[hour])
		currentMean := meanValues(current[hour])
		if baselineMean > 0 && len(current[hour]) > 0 && math.Abs(percentChange(baselineMean, currentMean)) >= changeThresholdPct {
			result = append(result, hour)
		}
	}
	return result
}

func firstAbnormalHour(series []reading.Reading) time.Time {
	if len(series) == 0 {
		return time.Time{}
	}
	return series[len(series)-baselineWindowHours].Timestamp
}

type consumptionOutlier struct {
	Timestamp      time.Time
	ConsumptionKWh float64
}

func consumptionOutliers(series []reading.Reading) []consumptionOutlier {
	if len(series) < outlierMinSeries {
		return nil
	}
	start := max(0, len(series)-baselineWindowHours)
	result := make([]consumptionOutlier, 0)
	for index := max(start, outlierNeighborRadius); index < len(series)-outlierNeighborRadius; index++ {
		neighbors := []float64{
			series[index-2].ConsumptionKWh,
			series[index-1].ConsumptionKWh,
			series[index+1].ConsumptionKWh,
			series[index+2].ConsumptionKWh,
		}
		sort.Float64s(neighbors)
		median := (neighbors[1] + neighbors[2]) / 2
		if median > 0 && math.Abs(series[index].ConsumptionKWh-median)/median > outlierDeviationLimit {
			result = append(result, consumptionOutlier{Timestamp: series[index].Timestamp, ConsumptionKWh: series[index].ConsumptionKWh})
		}
	}
	return result
}

func firstConsumptionOutlier(series []reading.Reading) (consumptionOutlier, bool) {
	outliers := consumptionOutliers(series)
	if len(outliers) == 0 {
		return consumptionOutlier{}, false
	}
	return outliers[0], true
}

func detectScheduledOutage(series []reading.Reading, scheduled event.Event, evidence Evidence) (Anomaly, bool) {
	start := sort.Search(len(series), func(index int) bool { return !series[index].Timestamp.Before(scheduled.Timestamp) })
	if start == 0 || start+outageWindowHours > len(series) {
		return Anomaly{}, false
	}
	before := meanConsumption(series[max(0, start-outageWindowHours):start])
	during := meanConsumption(series[start : start+outageWindowHours])
	if before == 0 || (before-during)/before < outageDropThreshold {
		return Anomaly{}, false
	}
	evidence.RelatedEvents = []string{scheduled.Description}
	evidence.ChangePercent = round((during/before - 1) * 100)
	return Anomaly{
		ID:                series[0].MeterID + "-outage",
		MeterID:           series[0].MeterID,
		Anomaly:           "Cambio de consumo durante parada programada",
		Type:              FalsePositive,
		Severity:          Low,
		Confidence:        0.88,
		Reason:            fmt.Sprintf("El consumo cayó %.1f%% durante una parada programada de 12 horas; el descenso coincide temporalmente con el evento registrado.", -evidence.ChangePercent),
		RecommendedAction: "No escalar: confirmar el cierre de la parada programada y validar el retorno al perfil habitual.",
		DetectedAt:        scheduled.Timestamp,
		Evidence:          evidence,
	}, true
}

func qualityIssue(readings []reading.Reading) int {
	if len(readings) == 0 {
		return 0
	}
	issueCount := 0
	for _, item := range readings {
		if item.PowerFactor < powerFactorMin || item.VoltageV < voltageMinV || item.VoltageV > voltageMaxV || !strings.EqualFold(item.Status, "OK") {
			issueCount++
		}
	}
	return issueCount
}

func relatedEvents(events []event.Event, start, end time.Time) []event.Event {
	result := make([]event.Event, 0)
	for _, item := range events {
		if !item.Timestamp.Before(start.Add(-eventCorrelationWindow)) && !item.Timestamp.After(end.Add(eventCorrelationWindow)) {
			result = append(result, item)
		}
	}
	return result
}

func findChangeStart(series []reading.Reading, baseline float64, baselineWindow int) time.Time {
	for start := baselineWindow; start+currentWindowHours <= len(series); start++ {
		windowMean := meanConsumption(series[start : start+currentWindowHours])
		if math.Abs(percentChange(baseline, windowMean)) >= changeThresholdPct {
			return series[start].Timestamp
		}
	}
	return series[len(series)-currentWindowHours].Timestamp
}

func eventDescriptions(events []event.Event) []string {
	result := make([]string, 0, len(events))
	for _, item := range events {
		result = append(result, item.Description)
	}
	return result
}

func hasEventType(events []event.Event, kind string) bool {
	for _, item := range events {
		if item.Type == kind {
			return true
		}
	}
	return false
}

func meanConsumption(readings []reading.Reading) float64 {
	return meanField(readings, func(item reading.Reading) float64 { return item.ConsumptionKWh })
}

func meanField(readings []reading.Reading, field func(reading.Reading) float64) float64 {
	if len(readings) == 0 {
		return 0
	}
	total := 0.0
	for _, item := range readings {
		total += field(item)
	}
	return total / float64(len(readings))
}

func meanValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func coefficientVariation(readings []reading.Reading) float64 {
	average := meanField(readings, func(item reading.Reading) float64 { return item.CurrentA })
	if average == 0 || len(readings) == 0 {
		return 0
	}
	variance := 0.0
	for _, item := range readings {
		difference := item.CurrentA - average
		variance += difference * difference
	}
	return math.Sqrt(variance/float64(len(readings))) / average
}

func percentChange(baseline, current float64) float64 {
	if baseline == 0 {
		return 0
	}
	return (current/baseline - 1) * 100
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}
