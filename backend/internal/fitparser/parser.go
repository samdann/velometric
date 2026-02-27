package fitparser

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/tormoder/fit"
)

// Parse reads a FIT file (or ZIP containing FIT) and extracts activity data
func Parse(r io.Reader) (*ParsedActivity, error) {
	// Read entire file into memory
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) < 14 {
		return nil, fmt.Errorf("file too small (%d bytes)", len(data))
	}

	// Check if it's a ZIP file (starts with PK)
	if len(data) >= 2 && data[0] == 'P' && data[1] == 'K' {
		data, err = extractFITFromZIP(data)
		if err != nil {
			return nil, err
		}
	}

	// Verify FIT file signature (bytes 8-11 should be ".FIT")
	if len(data) >= 12 && string(data[8:12]) != ".FIT" {
		return nil, fmt.Errorf("invalid FIT file signature")
	}

	fitFile, err := fit.Decode(bytes.NewReader(data))
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

	// Get activity name, start time, and summary data from session
	if len(activity.Sessions) > 0 {
		session := activity.Sessions[0]
		parsed.StartTime = session.StartTime
		parsed.Sport = session.Sport.String()

		// Extract session-level summary metrics
		if dist := session.GetTotalDistanceScaled(); !math.IsNaN(dist) {
			parsed.TotalDistance = &dist
		}
		if timerTime := session.GetTotalTimerTimeScaled(); !math.IsNaN(timerTime) {
			parsed.TotalTimerTime = &timerTime
		}
		if elapsedTime := session.GetTotalElapsedTimeScaled(); !math.IsNaN(elapsedTime) {
			parsed.TotalElapsedTime = &elapsedTime
		}
		if session.TotalAscent != 0xFFFF {
			ascent := float64(session.TotalAscent)
			parsed.TotalAscent = &ascent
		}
		if session.TotalDescent != 0xFFFF {
			descent := float64(session.TotalDescent)
			parsed.TotalDescent = &descent
		}
		if session.AvgPower != 0xFFFF {
			avgPower := int(session.AvgPower)
			parsed.AvgPower = &avgPower
		}
		if session.MaxPower != 0xFFFF {
			maxPower := int(session.MaxPower)
			parsed.MaxPower = &maxPower
		}
		if session.NormalizedPower != 0xFFFF {
			np := int(session.NormalizedPower)
			parsed.NormalizedPower = &np
		}
		if session.AvgHeartRate != 0xFF {
			avgHR := int(session.AvgHeartRate)
			parsed.AvgHeartRate = &avgHR
		}
		if session.MaxHeartRate != 0xFF {
			maxHR := int(session.MaxHeartRate)
			parsed.MaxHeartRate = &maxHR
		}
		if session.AvgCadence != 0xFF {
			avgCad := int(session.AvgCadence)
			parsed.AvgCadence = &avgCad
		}
		if session.MaxCadence != 0xFF {
			maxCad := int(session.MaxCadence)
			parsed.MaxCadence = &maxCad
		}
		if avgSpeed := session.GetAvgSpeedScaled(); !math.IsNaN(avgSpeed) {
			parsed.AvgSpeed = &avgSpeed
		}
		if maxSpeed := session.GetMaxSpeedScaled(); !math.IsNaN(maxSpeed) {
			parsed.MaxSpeed = &maxSpeed
		}
		if session.TotalCalories != 0xFFFF {
			calories := int(session.TotalCalories)
			parsed.TotalCalories = &calories
		}
		if session.AvgTemperature != 0x7F {
			temp := float64(session.AvgTemperature)
			parsed.AvgTemperature = &temp
		}
		if session.TrainingStressScore != 0xFFFF {
			tss := float64(session.TrainingStressScore) / 10.0 // Scaled by 10 in FIT
			parsed.TrainingStressScore = &tss
		}
		if session.IntensityFactor != 0xFFFF {
			ifactor := float64(session.IntensityFactor) / 1000.0 // Scaled by 1000 in FIT
			parsed.IntensityFactor = &ifactor
		}
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

		// Altitude — try enhanced (uint32) first, fall back to standard (uint16)
		if alt := rec.GetEnhancedAltitudeScaled(); !math.IsNaN(alt) {
			record.Altitude = &alt
		} else if alt := rec.GetAltitudeScaled(); !math.IsNaN(alt) {
			record.Altitude = &alt
		}

		// Distance
		if dist := rec.GetDistanceScaled(); !math.IsNaN(dist) {
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
		if speed := rec.GetSpeedScaled(); !math.IsNaN(speed) {
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
		if avgSpeed := lap.GetAvgSpeedScaled(); !math.IsNaN(avgSpeed) {
			parsedLap.AvgSpeed = &avgSpeed
		}
		if maxSpeed := lap.GetMaxSpeedScaled(); !math.IsNaN(maxSpeed) {
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

// extractFITFromZIP extracts the first .fit file from a ZIP archive
func extractFITFromZIP(data []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP file: %w", err)
	}

	for _, file := range zipReader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".fit") {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open FIT file in ZIP: %w", err)
			}
			defer rc.Close()

			fitData, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read FIT file from ZIP: %w", err)
			}
			return fitData, nil
		}
	}

	return nil, fmt.Errorf("no .fit file found in ZIP archive")
}
