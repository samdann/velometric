package service

import (
	"context"
	"sort"

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

// GetAnnualPowerStats returns the median power curve and median zone distribution for a given year.
func (s *StatisticsService) GetAnnualPowerStats(ctx context.Context, userID uuid.UUID, year, ftp int, zones []model.PowerZone) (*model.AnnualPowerStats, error) {
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
		dist = computeMedianZoneDistribution(records, ftp, zones)
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
