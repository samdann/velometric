package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/repository"
)

// tableDurations are the standard power curve durations shown in the UI.
var tableDurations = []int{5, 15, 30, 60, 300, 600, 1200, 1800, 2700, 3600}

type statisticsRepoer interface {
	GetAvailablePowerYears(ctx context.Context, userID uuid.UUID) ([]int, error)
	GetAnnualMedianPowerCurve(ctx context.Context, userID uuid.UUID, year int, durations []int) ([]model.AnnualPowerCurvePoint, error)
	GetAnnualPowerRecords(ctx context.Context, userID uuid.UUID, year int) ([]repository.ActivityPowerRecord, error)
}

// StatisticsService computes annual power statistics.
type StatisticsService struct {
	repo statisticsRepoer
}

func NewStatisticsService(repo statisticsRepoer) *StatisticsService {
	return &StatisticsService{repo: repo}
}

func (s *StatisticsService) GetAvailablePowerYears(ctx context.Context, userID uuid.UUID) ([]int, error) {
	return s.repo.GetAvailablePowerYears(ctx, userID)
}

// GetAnnualPowerStats returns the power curve and zone distribution for a given year.
// mode must be "avg" (median across activities) or "best" (max per zone independently).
func (s *StatisticsService) GetAnnualPowerStats(ctx context.Context, userID uuid.UUID, year, ftp int, zones []model.PowerZone, mode string) (*model.AnnualPowerStats, error) {
	curve, err := s.repo.GetAnnualMedianPowerCurve(ctx, userID, year, tableDurations)
	if err != nil {
		return nil, err
	}
	if curve == nil {
		curve = make([]model.AnnualPowerCurvePoint, 0)
	}

	var dist []model.AnnualZoneDistributionPoint
	if ftp > 0 && len(zones) > 0 {
		records, err := s.repo.GetAnnualPowerRecords(ctx, userID, year)
		if err != nil {
			return nil, err
		}
		if mode == "best" {
			dist = computeBestActivityZoneDistribution(records, ftp, zones)
		} else {
			dist = computeMedianZoneDistribution(records, ftp, zones)
		}
	}
	if dist == nil {
		dist = make([]model.AnnualZoneDistributionPoint, 0)
	}

	return &model.AnnualPowerStats{
		PowerCurve:       curve,
		ZoneDistribution: dist,
	}, nil
}

// computeMedianZoneDistribution groups records by activity, computes % per zone per activity, then takes
// the median across all activities for each zone.
func computeMedianZoneDistribution(records []repository.ActivityPowerRecord, ftp int, zones []model.PowerZone) []model.AnnualZoneDistributionPoint {
	type actState struct {
		secs  []float64
		total float64
		prev  *repository.ActivityPowerRecord
	}

	actMap := make(map[uuid.UUID]*actState)

	for i := range records {
		r := &records[i]
		act, ok := actMap[r.ActivityID]
		if !ok {
			act = &actState{secs: make([]float64, len(zones))}
			actMap[r.ActivityID] = act
		}
		if act.prev != nil {
			dt := r.Timestamp.Sub(act.prev.Timestamp).Seconds()
			if dt > 0 && dt <= 60 {
				powerPct := float64(act.prev.Power) / float64(ftp) * 100.0
				zoneIdx := classifyPowerZone(powerPct, zones)
				if zoneIdx >= 0 {
					act.secs[zoneIdx] += dt
					act.total += dt
				}
			}
		}
		prev := *r
		act.prev = &prev
	}

	// Per zone, collect percentages across activities
	zonePcts := make([][]float64, len(zones))
	for i := range zones {
		zonePcts[i] = make([]float64, 0, len(actMap))
	}
	for _, act := range actMap {
		if act.total == 0 {
			continue
		}
		for i := range zones {
			pct := act.secs[i] / act.total * 100.0
			zonePcts[i] = append(zonePcts[i], pct)
		}
	}

	result := make([]model.AnnualZoneDistributionPoint, len(zones))
	for i, z := range zones {
		minWatts := int(z.MinPercentage / 100.0 * float64(ftp))
		var maxWatts *int
		if z.MaxPercentage != nil {
			v := int(*z.MaxPercentage / 100.0 * float64(ftp))
			maxWatts = &v
		}
		result[i] = model.AnnualZoneDistributionPoint{
			ZoneNumber:       z.ZoneNumber,
			Name:             z.Name,
			Color:            z.Color,
			MinWatts:         minWatts,
			MaxWatts:         maxWatts,
			MedianPercentage: medianFloat(zonePcts[i]),
		}
	}
	return result
}

// computeNP computes Normalized Power for a time-ordered slice of records from
// a single activity. NP = ⁴√(mean of (30 s rolling average power)⁴).
func computeNP(records []repository.ActivityPowerRecord) float64 {
	if len(records) < 2 {
		return 0
	}
	sum4 := 0.0
	n := 0
	for i := range records {
		cutoff := records[i].Timestamp.Add(-30 * time.Second)
		sum, count := 0.0, 0
		for j := i; j >= 0; j-- {
			if records[j].Timestamp.Before(cutoff) {
				break
			}
			sum += float64(records[j].Power)
			count++
		}
		if count > 0 {
			avg := sum / float64(count)
			sum4 += avg * avg * avg * avg
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return math.Pow(sum4/float64(n), 0.25)
}

// computeBestActivityZoneDistribution picks the activity with the highest Normalized
// Power and returns its zone distribution.
func computeBestActivityZoneDistribution(records []repository.ActivityPowerRecord, ftp int, zones []model.PowerZone) []model.AnnualZoneDistributionPoint {
	// Group records by activity, preserving timestamp order (DB returns activity_id, timestamp).
	grouped := make(map[uuid.UUID][]repository.ActivityPowerRecord)
	var order []uuid.UUID
	for i := range records {
		id := records[i].ActivityID
		if _, ok := grouped[id]; !ok {
			order = append(order, id)
		}
		grouped[id] = append(grouped[id], records[i])
	}

	// Find the activity with the highest NP.
	var bestID uuid.UUID
	bestNP := -1.0
	for _, id := range order {
		if np := computeNP(grouped[id]); np > bestNP {
			bestNP = np
			bestID = id
		}
	}

	// Build the result template (watt boundaries are the same regardless of which ride).
	result := make([]model.AnnualZoneDistributionPoint, len(zones))
	for i, z := range zones {
		minWatts := int(z.MinPercentage / 100.0 * float64(ftp))
		var maxWatts *int
		if z.MaxPercentage != nil {
			v := int(*z.MaxPercentage / 100.0 * float64(ftp))
			maxWatts = &v
		}
		result[i] = model.AnnualZoneDistributionPoint{
			ZoneNumber: z.ZoneNumber,
			Name:       z.Name,
			Color:      z.Color,
			MinWatts:   minWatts,
			MaxWatts:   maxWatts,
		}
	}

	if bestNP <= 0 {
		return result
	}

	// Compute zone percentages for the best activity's records.
	secs := make([]float64, len(zones))
	total := 0.0
	best := grouped[bestID]
	for i := 1; i < len(best); i++ {
		dt := best[i].Timestamp.Sub(best[i-1].Timestamp).Seconds()
		if dt > 0 && dt <= 60 {
			powerPct := float64(best[i-1].Power) / float64(ftp) * 100.0
			if zoneIdx := classifyPowerZone(powerPct, zones); zoneIdx >= 0 {
				secs[zoneIdx] += dt
				total += dt
			}
		}
	}
	if total > 0 {
		for i := range zones {
			result[i].MedianPercentage = secs[i] / total * 100.0
		}
	}
	return result
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}
