package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// Attempt reverse geocoding from first GPS record (best-effort, non-fatal)
	var location *string
	for _, rec := range parsed.Records {
		if rec.Lat != nil && rec.Lon != nil {
			if loc, err := reverseGeocode(ctx, *rec.Lat, *rec.Lon); err == nil && loc != "" {
				location = &loc
			}
			break
		}
	}

	// Create activity model
	activity := &model.Activity{
		UserID:     userID,
		Name:       name,
		Sport:      parsed.Sport,
		StartTime:  parsed.StartTime,
		Duration:   duration,
		Distance:   distance,
		DeviceName: parsed.DeviceName,
		Location:   location,
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
			Trigger:    &lap.Trigger,
		}
	}

	if err := s.repo.InsertLaps(ctx, laps); err != nil {
		return nil, fmt.Errorf("failed to insert laps: %w", err)
	}

	// Compute and save power curve — use model records so gradient is available
	var pcRecords []PowerCurveRecord
	for _, rec := range records {
		if rec.Power == nil {
			continue
		}
		pcr := PowerCurveRecord{
			Power:               *rec.Power,
			Speed:               rec.Speed,
			Gradient:            rec.Gradient,
			LRBalance:           rec.LeftRightBalance,
			TorqueEffectiveness: rec.LeftTorqueEffectiveness,
		}
		if rec.HeartRate != nil {
			pcr.HeartRate = *rec.HeartRate
		}
		if rec.Cadence != nil {
			pcr.Cadence = *rec.Cadence
		}
		pcRecords = append(pcRecords, pcr)
	}

	if len(pcRecords) > 0 {
		powerCurveMap := ComputePowerCurveExtended(pcRecords)
		powerCurve := make([]model.PowerCurvePoint, 0, len(powerCurveMap))
		for duration, result := range powerCurveMap {
			powerCurve = append(powerCurve, model.PowerCurvePoint{
				ActivityID:             activityID,
				DurationSeconds:        duration,
				BestPower:              result.BestPower,
				AvgHeartRate:           result.AvgHeartRate,
				AvgSpeed:               result.AvgSpeed,
				AvgGradient:            result.AvgGradient,
				AvgCadence:             result.AvgCadence,
				AvgLRBalance:           result.AvgLRBalance,
				AvgTorqueEffectiveness: result.AvgTorqueEffectiveness,
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

// ListActivitiesPaginated retrieves a page of activities for a user
func (s *ActivityService) ListActivitiesPaginated(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.Activity, int, error) {
	return s.repo.ListByUserIDPaginated(ctx, userID, page, limit)
}

// GetActivityRecords retrieves all records for an activity
func (s *ActivityService) GetActivityRecords(ctx context.Context, activityID uuid.UUID) ([]model.ActivityRecord, error) {
	return s.repo.GetRecords(ctx, activityID)
}

// GetPowerCurve retrieves the power curve for an activity
func (s *ActivityService) GetPowerCurve(ctx context.Context, activityID uuid.UUID) ([]model.PowerCurvePoint, error) {
	return s.repo.GetPowerCurve(ctx, activityID)
}

// GetElevationProfile retrieves a smoothed, downsampled elevation profile
func (s *ActivityService) GetElevationProfile(ctx context.Context, activityID uuid.UUID) ([]model.ElevationPoint, error) {
	points, err := s.repo.GetElevationProfile(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return points, nil
	}

	// Smooth with a moving average (window=7) to remove corrupt spikes
	const smoothWindow = 7
	smoothed := make([]model.ElevationPoint, len(points))
	for i, p := range points {
		start := i - smoothWindow/2
		end := i + smoothWindow/2 + 1
		if start < 0 {
			start = 0
		}
		if end > len(points) {
			end = len(points)
		}
		var sum float64
		for j := start; j < end; j++ {
			sum += points[j].Altitude
		}
		smoothed[i] = model.ElevationPoint{
			Distance:    p.Distance,
			Altitude:    sum / float64(end-start),
			Temperature: p.Temperature,
		}
	}

	// Downsample to ~400 points
	const target = 400
	if len(smoothed) <= target {
		return smoothed, nil
	}
	step := float64(len(smoothed)) / float64(target)
	result := make([]model.ElevationPoint, 0, target)
	for i := 0; i < target; i++ {
		result = append(result, smoothed[int(float64(i)*step)])
	}
	result = append(result, smoothed[len(smoothed)-1])
	return result, nil
}

// GetSpeedProfile retrieves a smoothed, downsampled speed profile
func (s *ActivityService) GetSpeedProfile(ctx context.Context, activityID uuid.UUID) ([]model.SpeedPoint, error) {
	points, err := s.repo.GetSpeedProfile(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return points, nil
	}

	// Smooth with a moving average (window=7) to remove spikes
	const smoothWindow = 7
	smoothed := make([]model.SpeedPoint, len(points))
	for i, p := range points {
		start := i - smoothWindow/2
		end := i + smoothWindow/2 + 1
		if start < 0 {
			start = 0
		}
		if end > len(points) {
			end = len(points)
		}
		var speedSum, powerSum float64
		powerCount := 0
		for j := start; j < end; j++ {
			speedSum += points[j].Speed
			if points[j].Power != nil {
				powerSum += *points[j].Power
				powerCount++
			}
		}
		sp := model.SpeedPoint{
			Distance: p.Distance,
			Speed:    speedSum / float64(end-start),
		}
		if powerCount > 0 {
			avg := powerSum / float64(powerCount)
			sp.Power = &avg
		}
		smoothed[i] = sp
	}

	// Downsample to ~400 points
	const target = 400
	if len(smoothed) <= target {
		return smoothed, nil
	}
	step := float64(len(smoothed)) / float64(target)
	result := make([]model.SpeedPoint, 0, target)
	for i := 0; i < target; i++ {
		result = append(result, smoothed[int(float64(i)*step)])
	}
	result = append(result, smoothed[len(smoothed)-1])
	return result, nil
}

// GetHRCadenceProfile retrieves a smoothed, downsampled HR+cadence profile
func (s *ActivityService) GetHRCadenceProfile(ctx context.Context, activityID uuid.UUID) ([]model.HRCadencePoint, error) {
	points, err := s.repo.GetHRCadenceProfile(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return points, nil
	}

	const smoothWindow = 7
	smoothed := make([]model.HRCadencePoint, len(points))
	for i, p := range points {
		start := i - smoothWindow/2
		end := i + smoothWindow/2 + 1
		if start < 0 {
			start = 0
		}
		if end > len(points) {
			end = len(points)
		}
		var hrSum, cadSum int
		hrCount, cadCount := 0, 0
		for j := start; j < end; j++ {
			if points[j].HeartRate != nil {
				hrSum += *points[j].HeartRate
				hrCount++
			}
			if points[j].Cadence != nil {
				cadSum += *points[j].Cadence
				cadCount++
			}
		}
		sp := model.HRCadencePoint{Distance: p.Distance}
		if hrCount > 0 {
			avg := hrSum / hrCount
			sp.HeartRate = &avg
		}
		if cadCount > 0 {
			avg := cadSum / cadCount
			sp.Cadence = &avg
		}
		smoothed[i] = sp
	}

	const target = 400
	if len(smoothed) <= target {
		return smoothed, nil
	}
	step := float64(len(smoothed)) / float64(target)
	result := make([]model.HRCadencePoint, 0, target)
	for i := 0; i < target; i++ {
		result = append(result, smoothed[int(float64(i)*step)])
	}
	result = append(result, smoothed[len(smoothed)-1])
	return result, nil
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


// GetRoute retrieves a downsampled GPS route for an activity
func (s *ActivityService) GetRoute(ctx context.Context, activityID uuid.UUID) ([]model.RoutePoint, error) {
	points, err := s.repo.GetRoute(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return points, nil
	}

	const target = 1000
	if len(points) <= target {
		return points, nil
	}
	step := float64(len(points)) / float64(target)
	result := make([]model.RoutePoint, 0, target)
	for i := 0; i < target; i++ {
		result = append(result, points[int(float64(i)*step)])
	}
	result = append(result, points[len(points)-1])
	return result, nil
}

// ComputeHRZoneDistribution computes time spent in each HR zone for an activity.
// maxHR is the user's maximum heart rate; zones are sorted by zone_number ascending.
func (s *ActivityService) ComputeHRZoneDistribution(ctx context.Context, activityID uuid.UUID, maxHR int, zones []model.HRZone) ([]model.HRZoneDistributionPoint, error) {
	pts, err := s.repo.GetHRTimeSeries(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if len(pts) < 2 || maxHR <= 0 || len(zones) == 0 {
		return nil, nil
	}

	// Accumulate seconds per zone index
	zoneSecs := make([]float64, len(zones))
	totalSecs := 0.0

	for i := 1; i < len(pts); i++ {
		dt := pts[i].Timestamp.Sub(pts[i-1].Timestamp).Seconds()
		// Ignore gaps >60 s (pauses)
		if dt <= 0 || dt > 60 {
			continue
		}
		hrPct := float64(pts[i-1].HeartRate) / float64(maxHR) * 100.0
		zoneIdx := classifyHRZone(hrPct, zones)
		if zoneIdx >= 0 {
			zoneSecs[zoneIdx] += dt
			totalSecs += dt
		}
	}

	result := make([]model.HRZoneDistributionPoint, len(zones))
	for i, z := range zones {
		minBPM := int(z.MinPercentage / 100.0 * float64(maxHR))
		var maxBPM *int
		if z.MaxPercentage != nil {
			v := int(*z.MaxPercentage / 100.0 * float64(maxHR))
			maxBPM = &v
		}
		pct := 0.0
		if totalSecs > 0 {
			pct = zoneSecs[i] / totalSecs * 100.0
		}
		result[i] = model.HRZoneDistributionPoint{
			ZoneNumber: z.ZoneNumber,
			Name:       z.Name,
			Color:      z.Color,
			MinBPM:     minBPM,
			MaxBPM:     maxBPM,
			Seconds:    zoneSecs[i],
			Percentage: pct,
		}
	}
	return result, nil
}

func classifyHRZone(hrPct float64, zones []model.HRZone) int {
	for i, z := range zones {
		if hrPct < z.MinPercentage {
			continue
		}
		if z.MaxPercentage == nil || hrPct < *z.MaxPercentage {
			return i
		}
	}
	// If above all zones, assign to last zone
	return len(zones) - 1
}

// ComputePowerZoneDistribution computes time spent in each power zone for an activity.
// ftp is the user's FTP in watts; zones are sorted by zone_number ascending.
func (s *ActivityService) ComputePowerZoneDistribution(ctx context.Context, activityID uuid.UUID, ftp int, zones []model.PowerZone) ([]model.PowerZoneDistributionPoint, error) {
	pts, err := s.repo.GetPowerTimeSeries(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if len(pts) < 2 || ftp <= 0 || len(zones) == 0 {
		return nil, nil
	}

	zoneSecs := make([]float64, len(zones))
	totalSecs := 0.0

	for i := 1; i < len(pts); i++ {
		dt := pts[i].Timestamp.Sub(pts[i-1].Timestamp).Seconds()
		if dt <= 0 || dt > 60 {
			continue
		}
		powerPct := float64(pts[i-1].Power) / float64(ftp) * 100.0
		zoneIdx := classifyPowerZone(powerPct, zones)
		if zoneIdx >= 0 {
			zoneSecs[zoneIdx] += dt
			totalSecs += dt
		}
	}

	result := make([]model.PowerZoneDistributionPoint, len(zones))
	for i, z := range zones {
		minWatts := int(z.MinPercentage / 100.0 * float64(ftp))
		var maxWatts *int
		if z.MaxPercentage != nil {
			v := int(*z.MaxPercentage / 100.0 * float64(ftp))
			maxWatts = &v
		}
		pct := 0.0
		if totalSecs > 0 {
			pct = zoneSecs[i] / totalSecs * 100.0
		}
		result[i] = model.PowerZoneDistributionPoint{
			ZoneNumber: z.ZoneNumber,
			Name:       z.Name,
			Color:      z.Color,
			MinWatts:   minWatts,
			MaxWatts:   maxWatts,
			Seconds:    zoneSecs[i],
			Percentage: pct,
		}
	}
	return result, nil
}

func classifyPowerZone(powerPct float64, zones []model.PowerZone) int {
	for i, z := range zones {
		if powerPct < z.MinPercentage {
			continue
		}
		if z.MaxPercentage == nil || powerPct < *z.MaxPercentage {
			return i
		}
	}
	return len(zones) - 1
}

func (s *ActivityService) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteActivity(ctx, id)
}

// GetFeed returns a paginated feed of activities with embedded mini-routes.
func (s *ActivityService) GetFeed(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.FeedActivity, int, error) {
	return s.repo.GetFeedActivities(ctx, userID, page, limit)
}

// reverseGeocode resolves a GPS coordinate to a human-readable location string
// using the Nominatim OpenStreetMap API. Returns an empty string on failure.
func reverseGeocode(ctx context.Context, lat, lon float64) (string, error) {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%.6f&lon=%.6f", lat, lon)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Velometric/1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Address struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			County  string `json:"county"`
			Country string `json:"country"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	city := result.Address.City
	if city == "" {
		city = result.Address.Town
	}
	if city == "" {
		city = result.Address.Village
	}
	if city == "" {
		city = result.Address.County
	}
	if city != "" && result.Address.Country != "" {
		return city + ", " + result.Address.Country, nil
	}
	return city, nil
}
