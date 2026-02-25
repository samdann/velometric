package fitparser

import (
	"fmt"
	"io"
	"math"

	"github.com/tormoder/fit"
)

// Parse reads a FIT file and extracts activity data
func Parse(r io.Reader) (*ParsedActivity, error) {
	fitFile, err := fit.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode FIT file: %w", err)
	}

	activity, err := fitFile.Activity()
	if err != nil {
		return nil, fmt.Errorf("failed to get activity from FIT file: %w", err)
	}

	parsed := &ParsedActivity{
		Sport:   "cycling",
		Records: make([]Record, 0, len(activity.Records)),
		Laps:    make([]Lap, 0, len(activity.Laps)),
		Events:  make([]Event, 0, len(activity.Events)),
	}

	// Get activity name and start time from session
	if len(activity.Sessions) > 0 {
		session := activity.Sessions[0]
		parsed.StartTime = session.StartTime
		parsed.Sport = session.Sport.String()
	}

	// Parse records
	for _, rec := range activity.Records {
		record := Record{
			Timestamp: rec.Timestamp,
		}

		// Position
		if !rec.PositionLat.Invalid() && !rec.PositionLong.Invalid() {
			lat := rec.PositionLat.Degrees()
			lon := rec.PositionLong.Degrees()
			record.Lat = &lat
			record.Lon = &lon
		}

		// Altitude
		if rec.GetAltitudeScaled() != math.MaxFloat64 {
			alt := rec.GetAltitudeScaled()
			record.Altitude = &alt
		}

		// Distance
		if rec.GetDistanceScaled() != math.MaxFloat64 {
			dist := rec.GetDistanceScaled()
			record.Distance = &dist
		}

		// Power
		if rec.Power != 0xFFFF {
			power := int(rec.Power)
			record.Power = &power
		}

		// Heart rate
		if rec.HeartRate != 0xFF {
			hr := int(rec.HeartRate)
			record.HeartRate = &hr
		}

		// Cadence
		if rec.Cadence != 0xFF {
			cad := int(rec.Cadence)
			record.Cadence = &cad
		}

		// Speed
		if rec.GetSpeedScaled() != math.MaxFloat64 {
			speed := rec.GetSpeedScaled()
			record.Speed = &speed
		}

		// Temperature
		if rec.Temperature != 0x7F {
			temp := float64(rec.Temperature)
			record.Temperature = &temp
		}

		// Left/right balance
		if rec.LeftRightBalance != 0xFF {
			balance := float64(rec.LeftRightBalance & 0x7F)
			record.LeftRightBalance = &balance
		}

		// Torque effectiveness
		if rec.LeftTorqueEffectiveness != 0xFF {
			lte := float64(rec.LeftTorqueEffectiveness) / 2.0
			record.LeftTorqueEffectiveness = &lte
		}
		if rec.RightTorqueEffectiveness != 0xFF {
			rte := float64(rec.RightTorqueEffectiveness) / 2.0
			record.RightTorqueEffectiveness = &rte
		}

		// Pedal smoothness
		if rec.LeftPedalSmoothness != 0xFF {
			lps := float64(rec.LeftPedalSmoothness) / 2.0
			record.LeftPedalSmoothness = &lps
		}
		if rec.RightPedalSmoothness != 0xFF {
			rps := float64(rec.RightPedalSmoothness) / 2.0
			record.RightPedalSmoothness = &rps
		}

		parsed.Records = append(parsed.Records, record)
	}

	// Parse laps
	for _, lap := range activity.Laps {
		parsedLap := Lap{
			StartTime: lap.StartTime,
			Duration:  int(lap.GetTotalElapsedTimeScaled()),
			Distance:  lap.GetTotalDistanceScaled(),
			Trigger:   lap.LapTrigger.String(),
		}

		if lap.AvgPower != 0xFFFF {
			avgPower := int(lap.AvgPower)
			parsedLap.AvgPower = &avgPower
		}
		if lap.MaxPower != 0xFFFF {
			maxPower := int(lap.MaxPower)
			parsedLap.MaxPower = &maxPower
		}
		if lap.AvgHeartRate != 0xFF {
			avgHR := int(lap.AvgHeartRate)
			parsedLap.AvgHR = &avgHR
		}
		if lap.MaxHeartRate != 0xFF {
			maxHR := int(lap.MaxHeartRate)
			parsedLap.MaxHR = &maxHR
		}
		if lap.AvgCadence != 0xFF {
			avgCad := int(lap.AvgCadence)
			parsedLap.AvgCadence = &avgCad
		}
		if lap.GetAvgSpeedScaled() != math.MaxFloat64 {
			avgSpeed := lap.GetAvgSpeedScaled()
			parsedLap.AvgSpeed = &avgSpeed
		}
		if lap.GetMaxSpeedScaled() != math.MaxFloat64 {
			maxSpeed := lap.GetMaxSpeedScaled()
			parsedLap.MaxSpeed = &maxSpeed
		}
		if lap.TotalAscent != 0xFFFF {
			ascent := float64(lap.TotalAscent)
			parsedLap.Ascent = &ascent
		}
		if lap.TotalDescent != 0xFFFF {
			descent := float64(lap.TotalDescent)
			parsedLap.Descent = &descent
		}

		parsed.Laps = append(parsed.Laps, parsedLap)
	}

	// Parse events
	for _, evt := range activity.Events {
		event := Event{
			Timestamp: evt.Timestamp,
			EventType: evt.Event.String(),
			Data:      make(map[string]interface{}),
		}
		event.Data["event_type"] = evt.EventType.String()

		parsed.Events = append(parsed.Events, event)
	}

	return parsed, nil
}
