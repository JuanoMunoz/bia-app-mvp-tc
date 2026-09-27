package meter

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"learning/internal/anomaly"
	"learning/internal/reading"
)

var ErrNotFound = errors.New("meter not found")

type Profile struct {
	Provider string
	Region   string
	RateType string
}

type ProfileRepository interface {
	MeterProfile(context.Context, string) (Profile, error)
}

type Service struct {
	readings  ReadingRepository
	anomalies AnomalyRepository
	profiles  ProfileRepository
}

func NewService(readings ReadingRepository, anomalies AnomalyRepository, profiles ProfileRepository) *Service {
	return &Service{readings: readings, anomalies: anomalies, profiles: profiles}
}

func (service *Service) List(ctx context.Context) ([]Summary, error) {
	allReadings, err := service.readings.AllReadings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load meter readings: %w", err)
	}
	anomalies, err := service.anomalies.ListAnomalies(ctx)
	if err != nil {
		return nil, fmt.Errorf("load meter anomalies: %w", err)
	}
	readingsByMeter := make(map[string][]reading.Reading)
	for _, item := range allReadings {
		readingsByMeter[item.MeterID] = append(readingsByMeter[item.MeterID], item)
	}
	anomalyByMeter := make(map[string]anomaly.Anomaly)
	for _, item := range anomalies {
		anomalyByMeter[item.MeterID] = item
	}
	results := make([]Summary, 0, len(readingsByMeter))
	for meterID, series := range readingsByMeter {
		results = append(results, summarize(meterID, series, anomalyByMeter, service.profileOf(ctx, meterID)))
	}
	return results, nil
}

func (service *Service) Get(ctx context.Context, meterID string) (Detail, error) {
	series, err := service.readings.ReadingsByMeter(ctx, meterID)
	if err != nil {
		return Detail{}, fmt.Errorf("load meter %s readings: %w", meterID, err)
	}
	if len(series) == 0 {
		return Detail{}, ErrNotFound
	}
	anomalies, err := service.anomalies.ListAnomalies(ctx)
	if err != nil {
		return Detail{}, fmt.Errorf("load meter anomalies: %w", err)
	}
	anomalyByMeter := make(map[string]anomaly.Anomaly, len(anomalies))
	for _, item := range anomalies {
		anomalyByMeter[item.MeterID] = item
	}
	summary := summarize(meterID, series, anomalyByMeter, service.profileOf(ctx, meterID))
	latest := series[max(0, len(series)-24):]
	return Detail{
		Summary:     summary,
		Readings:    series,
		VoltageV:    mean(latest, func(item reading.Reading) float64 { return item.VoltageV }),
		CurrentA:    mean(latest, func(item reading.Reading) float64 { return item.CurrentA }),
		PowerFactor: mean(latest, func(item reading.Reading) float64 { return item.PowerFactor }),
	}, nil
}

func (service *Service) SiteHistory(ctx context.Context) ([]SitePoint, error) {
	allReadings, err := service.readings.AllReadings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load site history: %w", err)
	}
	totals := make(map[time.Time]float64)
	for _, item := range allReadings {
		hour := item.Timestamp.Truncate(time.Hour)
		totals[hour] += item.ConsumptionKWh
	}
	points := make([]SitePoint, 0, len(totals))
	for hour, total := range totals {
		points = append(points, SitePoint{Timestamp: hour, ConsumptionKWh: total})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp.Before(points[j].Timestamp) })
	return points, nil
}

func (service *Service) profileOf(ctx context.Context, meterID string) Profile {
	if service.profiles == nil {
		return Profile{}
	}
	profile, err := service.profiles.MeterProfile(ctx, meterID)
	if err != nil {
		return Profile{}
	}
	return profile
}

func summarize(meterID string, series []reading.Reading, anomalies map[string]anomaly.Anomaly, profile Profile) Summary {
	current := series[max(0, len(series)-24):]
	baselineWindow := min(7*24, max(1, len(series)-len(current)))
	baseline := mean(series[:baselineWindow], func(item reading.Reading) float64 { return item.ConsumptionKWh })
	currentMean := mean(current, func(item reading.Reading) float64 { return item.ConsumptionKWh })
	periodTotal := 0.0
	for _, item := range series {
		periodTotal += item.ConsumptionKWh
	}
	result := Summary{
		Meter:                 Meter{MeterID: meterID, Provider: profile.Provider, Region: profile.Region, RateType: profile.RateType},
		CurrentConsumptionKWh: currentMean,
		PeriodConsumptionKWh:  periodTotal,
		BaselineKWh:           baseline,
		ChangePercent:         percentageChange(baseline, currentMean),
		Status:                "NORMAL",
	}
	if item, ok := anomalies[meterID]; ok {
		result.Anomaly = &item
		result.Severity = item.Severity
		switch {
		case item.Severity == anomaly.Critical:
			result.Status = "CRITICAL"
		case item.Type == anomaly.FalsePositive || item.Severity == anomaly.Low:
			result.Status = "NORMAL"
		default:
			result.Status = "ALERT"
		}
	}
	return result
}

func mean(readings []reading.Reading, field func(reading.Reading) float64) float64 {
	if len(readings) == 0 {
		return 0
	}
	total := 0.0
	for _, item := range readings {
		total += field(item)
	}
	return total / float64(len(readings))
}

func percentageChange(baseline, current float64) float64 {
	if baseline == 0 {
		return 0
	}
	return (current/baseline - 1) * 100
}
