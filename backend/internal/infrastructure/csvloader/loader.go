package csvloader

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"learning/internal/event"
	"learning/internal/reading"
)

type Dataset struct {
	Readings []reading.Reading
	Events   []event.Event
}

func Load(readingsPath, eventsPath string) (Dataset, error) {
	readings, err := loadReadings(readingsPath)
	if err != nil {
		return Dataset{}, fmt.Errorf("load readings: %w", err)
	}
	events, err := loadEvents(eventsPath)
	if err != nil {
		return Dataset{}, fmt.Errorf("load events: %w", err)
	}
	return Dataset{Readings: readings, Events: events}, nil
}

func loadReadings(path string) ([]reading.Reading, error) {
	rows, err := readRows(path)
	if err != nil {
		return nil, err
	}
	columns, err := columnIndexes(rows[0], "meter_id", "timestamp", "consumption_kwh", "voltage_v", "current_a", "power_factor", "status")
	if err != nil {
		return nil, err
	}
	results := make([]reading.Reading, 0, len(rows)-1)
	for rowIndex, row := range rows[1:] {
		consumption, err := parseNumber(row[columns["consumption_kwh"]])
		if err != nil {
			return nil, fmt.Errorf("row %d consumption_kwh: %w", rowIndex+2, err)
		}
		voltage, err := parseNumber(row[columns["voltage_v"]])
		if err != nil {
			return nil, fmt.Errorf("row %d voltage_v: %w", rowIndex+2, err)
		}
		current, err := parseNumber(row[columns["current_a"]])
		if err != nil {
			return nil, fmt.Errorf("row %d current_a: %w", rowIndex+2, err)
		}
		powerFactor, err := parseNumber(row[columns["power_factor"]])
		if err != nil {
			return nil, fmt.Errorf("row %d power_factor: %w", rowIndex+2, err)
		}
		timestamp, err := parseTimestamp(row[columns["timestamp"]])
		if err != nil {
			return nil, fmt.Errorf("row %d timestamp: %w", rowIndex+2, err)
		}
		results = append(results, reading.Reading{
			MeterID:        row[columns["meter_id"]],
			Timestamp:      timestamp,
			ConsumptionKWh: consumption,
			VoltageV:       voltage,
			CurrentA:       current,
			PowerFactor:    powerFactor,
			Status:         row[columns["status"]],
		})
	}
	return results, nil
}

func loadEvents(path string) ([]event.Event, error) {
	rows, err := readRows(path)
	if err != nil {
		return nil, err
	}
	columns, err := columnIndexes(rows[0], "meter_id", "event_timestamp", "event_type", "description")
	if err != nil {
		return nil, err
	}
	results := make([]event.Event, 0, len(rows)-1)
	for rowIndex, row := range rows[1:] {
		timestamp, err := parseTimestamp(row[columns["event_timestamp"]])
		if err != nil {
			return nil, fmt.Errorf("row %d event_timestamp: %w", rowIndex+2, err)
		}
		results = append(results, event.Event{
			MeterID:     row[columns["meter_id"]],
			Timestamp:   timestamp,
			Type:        row[columns["event_type"]],
			Description: row[columns["description"]],
		})
	}
	return results, nil
}

func readRows(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("CSV has no header")
	}
	return rows, nil
}

func columnIndexes(header []string, names ...string) (map[string]int, error) {
	indexes := make(map[string]int, len(header))
	for index, name := range header {
		indexes[name] = index
	}
	for _, name := range names {
		if _, ok := indexes[name]; !ok {
			return nil, fmt.Errorf("missing CSV column %q", name)
		}
	}
	return indexes, nil
}

func parseNumber(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func parseTimestamp(value string) (time.Time, error) {
	for _, layout := range []string{time.DateTime, "2006-01-02 15:04"} {
		parsed, err := time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp %q", value)
}
