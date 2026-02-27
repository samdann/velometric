package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/fitparser"
	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/repository"
)

type ActivityService struct {
	repo *repository.ActivityRepository
}

func NewActivityService(repo *repository.ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

// ProcessFITFile parses a FIT file, computes metrics, and stores the activity
func (s *ActivityService) ProcessFITFile(ctx context.Context, userID uuid.UUID, parsed *fitparser.ParsedActivity, ftp int) (*model.Activity, error) {
	// Extract power, HR, and other data arrays for computation
	var powers []int
	var heartRates []int
	var cadences []int
	var speeds []float64
	var altitudes []float64
	var temperatures []float64

	for _, rec := range parsed.Records {
		if rec.Power != nil {
			powers = append(powers, *rec.Power)
		}
		if rec.HeartRate != nil {
			heartRates = append(heartRates, *rec.HeartRate)
		}
		if rec.Cadence != nil {
			cadences = append(cadences, *rec.Cadence)
		}
		if rec.Speed != nil {
			speeds = append(speeds, *rec.Speed)
		}
		if rec.Altitude != nil {
			altitudes = append(altitudes, *rec.Altitude)
		}
		if rec.Temperature != nil {
			temperatures = append(temperatures, *rec.Temperature)
		}
	}

	// Use session-level values when available, otherwise compute from records
	var duration int
	if parsed.TotalTimerTime != nil {
		duration = int(*parsed.TotalTimerTime)
	} else if len(parsed.Records) >= 2 {
		duration = int(parsed.Records[len(parsed.Records)-1].Timestamp.Sub(parsed.Records[0].Timestamp).Seconds())
	}

	// Get distance from session or last record
	var distance float64
	if parsed.TotalDistance != nil {
		distance = *parsed.TotalDistance
	} else {
		for i := len(parsed.Records) - 1; i >= 0; i-- {
			if parsed.Records[i].Distance != nil {
				distance = *parsed.Records[i].Distance
				break
			}
		}
	}

	// Use session-level metrics when available, otherwise compute
	var avgPower, maxPower, normalizedPower int
	if parsed.AvgPower != nil {
		avgPower = *parsed.AvgPower
	} else {
		avgPower = ComputeAverage(powers)
	}
	if parsed.MaxPower != nil {
		maxPower = *parsed.MaxPower
	} else {
		maxPower = ComputeMax(powers)
	}
	if parsed.NormalizedPower != nil {
		normalizedPower = *parsed.NormalizedPower
	} else {
		normalizedPower = ComputeNormalizedPower(powers)
	}

	var avgHR, maxHR int
	if parsed.AvgHeartRate != nil {
		avgHR = *parsed.AvgHeartRate
	} else {
		avgHR = ComputeAverage(heartRates)
	}
	if parsed.MaxHeartRate != nil {
		maxHR = *parsed.MaxHeartRate
	} else {
		maxHR = ComputeMax(heartRates)
	}

	var avgCadence, maxCadence int
	if parsed.AvgCadence != nil {
		avgCadence = *parsed.AvgCadence
	} else {
		avgCadence = ComputeAverage(cadences)
	}
	if parsed.MaxCadence != nil {
		maxCadence = *parsed.MaxCadence
	} else {
		maxCadence = ComputeMax(cadences)
	}

	var avgSpeed, maxSpeed float64
	if parsed.AvgSpeed != nil {
		avgSpeed = *parsed.AvgSpeed
	} else {
		avgSpeed = ComputeAverageFloat(speeds)
	}
	if parsed.MaxSpeed != nil {
		maxSpeed = *parsed.MaxSpeed
	} else {
		maxSpeed = ComputeMaxFloat(speeds)
	}

	var elevationGain float64
	if parsed.TotalAscent != nil {
		elevationGain = *parsed.TotalAscent
	} else {
		elevationGain = ComputeElevationGain(altitudes)
	}

	var avgTemp float64
	if parsed.AvgTemperature != nil {
		avgTemp = *parsed.AvgTemperature
	} else {
		avgTemp = ComputeAverageFloat(temperatures)
	}

	// Use session-level derived metrics when available, otherwise compute
	var tss, intensityFactor, variabilityIndex float64
	if parsed.TrainingStressScore != nil {
		tss = *parsed.TrainingStressScore
	} else if ftp > 0 && normalizedPower > 0 {
		tss = ComputeTSS(duration, normalizedPower, ftp)
	}
	if parsed.IntensityFactor != nil {
		intensityFactor = *parsed.IntensityFactor
	} else if ftp > 0 && normalizedPower > 0 {
		intensityFactor = ComputeIntensityFactor(normalizedPower, ftp)
	}
	if avgPower > 0 && normalizedPower > 0 {
		variabilityIndex = ComputeVariabilityIndex(normalizedPower, avgPower)
	}

	// Debug logging for session values
	log.Printf("ProcessFITFile: Session values - Elevation: %.0f, AvgPower: %d, NP: %d, IF: %.3f, TSS: %.1f",
		elevationGain, avgPower, normalizedPower, intensityFactor, tss)
	log.Printf("ProcessFITFile: Session sources - TotalAscent: %v, AvgPower: %v, NP: %v, IF: %v",
		parsed.TotalAscent != nil, parsed.AvgPower != nil, parsed.NormalizedPower != nil, parsed.IntensityFactor != nil)

	// Generate activity name if not provided
	name := parsed.Name
	if name == "" {
		name = generateActivityName(parsed.Sport, parsed.StartTime)
	}

	// Create activity model
	activity := &model.Activity{
		UserID:    userID,
		Name:      name,
		Sport:     parsed.Sport,
		StartTime: parsed.StartTime,
		Duration:  duration,
		Distance:  distance,
		ElevationGain: elevationGain,
	}

	// Set optional fields
	if avgPower > 0 {
		activity.AvgPower = &avgPower
	}
	if maxPower > 0 {
		activity.MaxPower = &maxPower
	}
	if normalizedPower > 0 {
		activity.NormalizedPower = &normalizedPower
	}
	if tss > 0 {
		activity.TSS = &tss
	}
	if intensityFactor > 0 {
		activity.IntensityFactor = &intensityFactor
	}
	if variabilityIndex > 0 {
		activity.VariabilityIndex = &variabilityIndex
	}
	if avgHR > 0 {
		activity.AvgHR = &avgHR
	}
	if maxHR > 0 {
		activity.MaxHR = &maxHR
	}
	if avgCadence > 0 {
		activity.AvgCadence = &avgCadence
	}
	if maxCadence > 0 {
		activity.MaxCadence = &maxCadence
	}
	if avgSpeed > 0 {
		activity.AvgSpeed = &avgSpeed
	}
	if maxSpeed > 0 {
		activity.MaxSpeed = &maxSpeed
	}
	if avgTemp != 0 {
		activity.AvgTemperature = &avgTemp
	}

	// Save activity
	activityID, err := s.repo.Create(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}
	activity.ID = activityID

	// Convert and save records
	records := make([]model.ActivityRecord, len(parsed.Records))
	for i, rec := range parsed.Records {
		records[i] = model.ActivityRecord{
			ActivityID:               activityID,
			Timestamp:                rec.Timestamp,
			Lat:                      rec.Lat,
			Lon:                      rec.Lon,
			Altitude:                 rec.Altitude,
			Distance:                 rec.Distance,
			Power:                    rec.Power,
			HeartRate:                rec.HeartRate,
			Cadence:                  rec.Cadence,
			Speed:                    rec.Speed,
			Temperature:              rec.Temperature,
			LeftRightBalance:         rec.LeftRightBalance,
			LeftTorqueEffectiveness:  rec.LeftTorqueEffectiveness,
			RightTorqueEffectiveness: rec.RightTorqueEffectiveness,
			LeftPedalSmoothness:      rec.LeftPedalSmoothness,
			RightPedalSmoothness:     rec.RightPedalSmoothness,
		}

		// Compute gradient if we have altitude and distance
		if i > 0 && rec.Altitude != nil && rec.Distance != nil {
			prevRec := parsed.Records[i-1]
			if prevRec.Altitude != nil && prevRec.Distance != nil {
				distDelta := *rec.Distance - *prevRec.Distance
				altDelta := *rec.Altitude - *prevRec.Altitude
				if distDelta > 0 {
					gradient := ComputeGradient(distDelta, altDelta)
					records[i].Gradient = &gradient
				}
			}
		}
	}

	if err := s.repo.InsertRecords(ctx, records); err != nil {
		return nil, fmt.Errorf("failed to insert records: %w", err)
	}

	// Convert and save laps
	laps := make([]model.ActivityLap, len(parsed.Laps))
	for i, lap := range parsed.Laps {
		laps[i] = model.ActivityLap{
			ActivityID: activityID,
			LapNumber:  i + 1,
			StartTime:  lap.StartTime,
			Duration:   lap.Duration,
			Distance:   lap.Distance,
			AvgPower:   lap.AvgPower,
			MaxPower:   lap.MaxPower,
			AvgHR:      lap.AvgHR,
			MaxHR:      lap.MaxHR,
			AvgCadence: lap.AvgCadence,
			AvgSpeed:   lap.AvgSpeed,
			MaxSpeed:   lap.MaxSpeed,
			Ascent:     lap.Ascent,
			Descent:    lap.Descent,
			Trigger:    lap.Trigger,
		}
	}

	if err := s.repo.InsertLaps(ctx, laps); err != nil {
		return nil, fmt.Errorf("failed to insert laps: %w", err)
	}

	// Compute and save power curve with heart rate and elevation data
	var pcRecords []PowerCurveRecord
	for _, rec := range parsed.Records {
		if rec.Power == nil {
			continue
		}
		pcr := PowerCurveRecord{Power: *rec.Power}
		if rec.HeartRate != nil {
			pcr.HeartRate = *rec.HeartRate
		}
		pcRecords = append(pcRecords, pcr)
	}

	if len(pcRecords) > 0 {
		powerCurveMap := ComputePowerCurveExtended(pcRecords)
		powerCurve := make([]model.PowerCurvePoint, 0, len(powerCurveMap))
		for duration, result := range powerCurveMap {
			powerCurve = append(powerCurve, model.PowerCurvePoint{
				ActivityID:      activityID,
				DurationSeconds: duration,
				BestPower:       result.BestPower,
				AvgHeartRate:    result.AvgHeartRate,
			})
		}

		if err := s.repo.InsertPowerCurve(ctx, powerCurve); err != nil {
			return nil, fmt.Errorf("failed to insert power curve: %w", err)
		}
	}

	// Convert and save events
	events := make([]model.ActivityEvent, len(parsed.Events))
	for i, evt := range parsed.Events {
		events[i] = model.ActivityEvent{
			ActivityID: activityID,
			Timestamp:  evt.Timestamp,
			EventType:  evt.EventType,
			Data:       evt.Data,
		}
	}

	if err := s.repo.InsertEvents(ctx, events); err != nil {
		return nil, fmt.Errorf("failed to insert events: %w", err)
	}

	return activity, nil
}

// GetActivity retrieves an activity by ID
func (s *ActivityService) GetActivity(ctx context.Context, id uuid.UUID) (*model.Activity, error) {
	return s.repo.GetByID(ctx, id)
}

// ListActivities retrieves all activities for a user
func (s *ActivityService) ListActivities(ctx context.Context, userID uuid.UUID) ([]*model.Activity, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// GetActivityRecords retrieves all records for an activity
func (s *ActivityService) GetActivityRecords(ctx context.Context, activityID uuid.UUID) ([]model.ActivityRecord, error) {
	return s.repo.GetRecords(ctx, activityID)
}

// GetPowerCurve retrieves the power curve for an activity
func (s *ActivityService) GetPowerCurve(ctx context.Context, activityID uuid.UUID) ([]model.PowerCurvePoint, error) {
	return s.repo.GetPowerCurve(ctx, activityID)
}

// GetLaps retrieves all laps for an activity
func (s *ActivityService) GetLaps(ctx context.Context, activityID uuid.UUID) ([]model.ActivityLap, error) {
	return s.repo.GetLaps(ctx, activityID)
}

func generateActivityName(sport string, startTime time.Time) string {
	timeOfDay := "Morning"
	hour := startTime.Hour()
	switch {
	case hour >= 12 && hour < 17:
		timeOfDay = "Afternoon"
	case hour >= 17 && hour < 21:
		timeOfDay = "Evening"
	case hour >= 21 || hour < 5:
		timeOfDay = "Night"
	}

	sportName := "Ride"
	if sport != "" && sport != "cycling" {
		sportName = sport
	}

	return fmt.Sprintf("%s %s", timeOfDay, sportName)
}
