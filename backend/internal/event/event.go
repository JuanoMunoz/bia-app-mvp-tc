package event

import "time"

type Event struct {
	MeterID     string
	Timestamp   time.Time
	Type        string
	Description string
}
