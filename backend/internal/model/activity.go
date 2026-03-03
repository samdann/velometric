package model

import (
	"time"

	"github.com/google/uuid"
)

// Activity represents a cycling activity
type Activity struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Sport     string    `json:"sport"`
	StartTime time.Time `json:"startTime"`

	// Basic metrics
	Duration      int     `json:"duration"`      // seconds
	Distance      float64 `json:"distance"`      // meters
	ElevationGain float64 `json:"elevationGain"` // meters

	// Power metrics
	AvgPower         *int     `json:"avgPower,omitempty"`
	MaxPower         *int     `json:"maxPower,omitempty"`
	NormalizedPower  *int     `json:"normalizedPower,omitempty"`
	TSS              *float64 `json:"tss,omitempty"`
	IntensityFactor  *float64 `json:"intensityFactor,omitempty"`
	VariabilityIndex *float64 `json:"variabilityIndex,omitempty"`

	// Heart rate metrics
	AvgHR *int `json:"avgHeartRate,omitempty"`
	MaxHR *int `json:"maxHeartRate,omitempty"`

	// Cadence metrics
	AvgCadence *int `json:"avgCadence,omitempty"`
	MaxCadence *int `json:"maxCadence,omitempty"`

	// Speed metrics
	AvgSpeed *float64 `json:"avgSpeed,omitempty"`
	MaxSpeed *float64 `json:"maxSpeed,omitempty"`

	// Other
	Calories       *int     `json:"calories,omitempty"`
	AvgTemperature *float64 `json:"avgTemperature,omitempty"`
	FitFileURL     *string  `json:"fitFileUrl,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ActivityRecord represents a single time-series data point
type ActivityRecord struct {
	ActivityID uuid.UUID `json:"activityId"`
	Timestamp  time.Time `json:"timestamp"`

	// Position
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
	Altitude *float64 `json:"altitude,omitempty"`

	// Distance
	Distance *float64 `json:"distance,omitempty"`

	// Metrics
	Power     *int     `json:"power,omitempty"`
	HeartRate *int     `json:"heartRate,omitempty"`
	Cadence   *int     `json:"cadence,omitempty"`
	Speed     *float64 `json:"speed,omitempty"`

	// Temperature
	Temperature *float64 `json:"temperature,omitempty"`

	// Pedaling dynamics
	LeftRightBalance         *float64 `json:"leftRightBalance,omitempty"`
	LeftTorqueEffectiveness  *float64 `json:"leftTorqueEffectiveness,omitempty"`
	RightTorqueEffectiveness *float64 `json:"rightTorqueEffectiveness,omitempty"`
	LeftPedalSmoothness      *float64 `json:"leftPedalSmoothness,omitempty"`
	RightPedalSmoothness     *float64 `json:"rightPedalSmoothness,omitempty"`

	// Gradient
	Gradient *float64 `json:"gradient,omitempty"`
}

// ActivityLap represents a lap in an activity
type ActivityLap struct {
	ID         uuid.UUID `json:"id"`
	ActivityID uuid.UUID `json:"activityId"`
	LapNumber  int       `json:"lapNumber"`
	StartTime  time.Time `json:"startTime"`

	Duration int     `json:"duration"`
	Distance float64 `json:"distance"`

	AvgPower        *int `json:"avgPower,omitempty"`
	MaxPower        *int `json:"maxPower,omitempty"`
	NormalizedPower *int `json:"normalizedPower,omitempty"`

	AvgHR *int `json:"avgHeartRate,omitempty"`
	MaxHR *int `json:"maxHeartRate,omitempty"`

	AvgCadence *int `json:"avgCadence,omitempty"`

	AvgSpeed *float64 `json:"avgSpeed,omitempty"`
	MaxSpeed *float64 `json:"maxSpeed,omitempty"`

	Ascent  *float64 `json:"ascent,omitempty"`
	Descent *float64 `json:"descent,omitempty"`

	Trigger *string `json:"trigger,omitempty"`
}

// PowerCurvePoint represents a best effort at a specific duration
type PowerCurvePoint struct {
	ActivityID      uuid.UUID `json:"activityId"`
	DurationSeconds int       `json:"durationSeconds"`
	BestPower       int       `json:"bestPower"`
	AvgHeartRate    *int      `json:"avgHeartRate,omitempty"`
}

// ElevationPoint represents a distance/altitude pair for the elevation profile
type ElevationPoint struct {
	Distance float64 `json:"distance"` // km
	Altitude float64 `json:"altitude"` // meters
}

// SpeedPoint represents a distance/speed/power point for the speed profile
type SpeedPoint struct {
	Distance float64  `json:"distance"`        // km
	Speed    float64  `json:"speed"`           // km/h
	Power    *float64 `json:"power,omitempty"` // watts
}

// HRCadencePoint represents a distance/heart-rate/cadence point
type HRCadencePoint struct {
	Distance  float64 `json:"distance"`           // km
	HeartRate *int    `json:"heartRate,omitempty"` // bpm
	Cadence   *int    `json:"cadence,omitempty"`  // rpm
}

// RoutePoint represents a GPS coordinate point for the route map
type RoutePoint struct {
	Lat      float64  `json:"lat"`
	Lon      float64  `json:"lon"`
	Distance *float64 `json:"distance,omitempty"` // km
}

// ActivityEvent represents an event during an activity
type ActivityEvent struct {
	ID         uuid.UUID              `json:"id"`
	ActivityID uuid.UUID              `json:"activityId"`
	Timestamp  time.Time              `json:"timestamp"`
	EventType  string                 `json:"eventType"`
	Data       map[string]interface{} `json:"data,omitempty"`
}
