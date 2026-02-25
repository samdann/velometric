package fitparser

import "time"

// ParsedActivity represents the complete parsed data from a FIT file
type ParsedActivity struct {
	Name      string
	Sport     string
	StartTime time.Time
	Records   []Record
	Laps      []Lap
	Events    []Event

	// Session-level summary data (pre-computed by device)
	TotalDistance       *float64 // meters
	TotalTimerTime      *float64 // seconds (moving time)
	TotalElapsedTime    *float64 // seconds (total time)
	TotalAscent         *float64 // meters
	TotalDescent        *float64 // meters
	AvgPower            *int     // watts
	MaxPower            *int     // watts
	NormalizedPower     *int     // watts
	AvgHeartRate        *int     // bpm
	MaxHeartRate        *int     // bpm
	AvgCadence          *int     // rpm
	MaxCadence          *int     // rpm
	AvgSpeed            *float64 // m/s
	MaxSpeed            *float64 // m/s
	TotalCalories       *int
	AvgTemperature      *float64 // Celsius
	TrainingStressScore *float64
	IntensityFactor     *float64
}

// Record represents a single data point in the activity
type Record struct {
	Timestamp time.Time

	// Position
	Lat      *float64 // degrees
	Lon      *float64 // degrees
	Altitude *float64 // meters

	// Distance
	Distance *float64 // cumulative meters

	// Power
	Power *int // watts

	// Heart rate
	HeartRate *int // bpm

	// Cadence
	Cadence *int // rpm

	// Speed
	Speed *float64 // m/s

	// Temperature
	Temperature *float64 // Celsius

	// Pedaling dynamics
	LeftRightBalance       *float64 // percentage (50 = balanced)
	LeftTorqueEffectiveness  *float64 // percentage
	RightTorqueEffectiveness *float64 // percentage
	LeftPedalSmoothness      *float64 // percentage
	RightPedalSmoothness     *float64 // percentage
}

// Lap represents a lap in the activity
type Lap struct {
	StartTime time.Time
	Duration  int     // seconds
	Distance  float64 // meters

	// Power
	AvgPower *int
	MaxPower *int

	// Heart rate
	AvgHR *int
	MaxHR *int

	// Cadence
	AvgCadence *int

	// Speed
	AvgSpeed *float64
	MaxSpeed *float64

	// Elevation
	Ascent  *float64
	Descent *float64

	// Trigger
	Trigger string
}

// Event represents an event during the activity
type Event struct {
	Timestamp time.Time
	EventType string
	Data      map[string]interface{}
}
