package service

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/repository"
)

// ── stub repo ─────────────────────────────────────────────────────────────────

type stubStatisticsRepo struct {
	years    []int
	yearsErr error

	curve    []model.AnnualPowerCurvePoint
	curveErr error

	records    []repository.ActivityPowerRecord
	recordsErr error
}

func (s *stubStatisticsRepo) GetAvailablePowerYears(_ context.Context, _ uuid.UUID) ([]int, error) {
	return s.years, s.yearsErr
}
func (s *stubStatisticsRepo) GetAnnualMedianPowerCurve(_ context.Context, _ uuid.UUID, _ int, _ []int) ([]model.AnnualPowerCurvePoint, error) {
	return s.curve, s.curveErr
}
func (s *stubStatisticsRepo) GetAnnualPowerRecords(_ context.Context, _ uuid.UUID, _ int) ([]repository.ActivityPowerRecord, error) {
	return s.records, s.recordsErr
}

// ── helpers ───────────────────────────────────────────────────────────────────

func ptr(f float64) *float64 { return &f }

// standardZones returns the classic 7-zone scheme (min/max as % of FTP).
func standardZones() []model.PowerZone {
	return []model.PowerZone{
		{ZoneNumber: 1, Name: "Active Recovery", MinPercentage: 0, MaxPercentage: ptr(55), Color: "#gray"},
		{ZoneNumber: 2, Name: "Endurance", MinPercentage: 55, MaxPercentage: ptr(75), Color: "#blue"},
		{ZoneNumber: 3, Name: "Tempo", MinPercentage: 75, MaxPercentage: ptr(90), Color: "#green"},
		{ZoneNumber: 4, Name: "Threshold", MinPercentage: 90, MaxPercentage: ptr(105), Color: "#yellow"},
		{ZoneNumber: 5, Name: "VO2Max", MinPercentage: 105, MaxPercentage: ptr(120), Color: "#orange"},
		{ZoneNumber: 6, Name: "Anaerobic", MinPercentage: 120, MaxPercentage: ptr(150), Color: "#red"},
		{ZoneNumber: 7, Name: "Neuromuscular", MinPercentage: 150, MaxPercentage: nil, Color: "#purple"},
	}
}

// makeRecords builds a sequence of ActivityPowerRecord for one activity at a
// constant power, spaced 1 second apart.
func makeRecords(actID uuid.UUID, base time.Time, powerWatts int, count int) []repository.ActivityPowerRecord {
	recs := make([]repository.ActivityPowerRecord, count)
	for i := range recs {
		recs[i] = repository.ActivityPowerRecord{
			ActivityID: actID,
			Timestamp:  base.Add(time.Duration(i) * time.Second),
			Power:      powerWatts,
		}
	}
	return recs
}

const almostEqual = 1e-9

func approxEq(a, b float64) bool { return math.Abs(a-b) < almostEqual }

// ── medianFloat ───────────────────────────────────────────────────────────────

func TestMedianFloat(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"odd count", []float64{3, 1, 2}, 2},
		{"even count", []float64{1, 3, 2, 4}, 2.5},
		{"all same", []float64{7, 7, 7}, 7},
		{"already sorted", []float64{10, 20, 30, 40}, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := medianFloat(tt.values)
			if !approxEq(got, tt.want) {
				t.Errorf("medianFloat(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

// ── computeMedianZoneDistribution ─────────────────────────────────────────────

func TestComputeMedianZoneDistribution_Empty(t *testing.T) {
	result := computeMedianZoneDistribution(nil, 250, standardZones())
	if len(result) != len(standardZones()) {
		t.Fatalf("expected %d zones, got %d", len(standardZones()), len(result))
	}
	for _, z := range result {
		if z.MedianPercentage != 0 {
			t.Errorf("zone %d: expected 0 median, got %v", z.ZoneNumber, z.MedianPercentage)
		}
	}
}

func TestComputeMedianZoneDistribution_SingleActivityAllInOneZone(t *testing.T) {
	// FTP=200, Zone2=55-75% → 110-150 W. Ride at 120 W (60% FTP) for 60 seconds.
	ftp := 200
	actID := uuid.New()
	base := time.Now()
	records := makeRecords(actID, base, 120, 61) // 61 records → 60 intervals

	result := computeMedianZoneDistribution(records, ftp, standardZones())

	if len(result) != len(standardZones()) {
		t.Fatalf("expected %d zones, got %d", len(standardZones()), len(result))
	}

	// Zone 2 (index 1) should be 100%, all others 0.
	for i, z := range result {
		if i == 1 {
			if !approxEq(z.MedianPercentage, 100) {
				t.Errorf("zone 2 MedianPercentage = %.2f, want 100", z.MedianPercentage)
			}
		} else {
			if !approxEq(z.MedianPercentage, 0) {
				t.Errorf("zone %d MedianPercentage = %.2f, want 0", z.ZoneNumber, z.MedianPercentage)
			}
		}
	}
}

func TestComputeMedianZoneDistribution_WattsToZoneMapping(t *testing.T) {
	// FTP=200. Build one activity in zone1 for 30s then zone4 for 30s → 50% each.
	ftp := 200
	actID := uuid.New()
	base := time.Now()

	// Zone1: 0–55% FTP → 0–110W → use 50W
	// Zone4: 90–105% FTP → 180–210W → use 200W
	var records []repository.ActivityPowerRecord
	records = append(records, makeRecords(actID, base, 50, 31)...)
	records = append(records, makeRecords(actID, base.Add(31*time.Second), 200, 31)...)

	result := computeMedianZoneDistribution(records, ftp, standardZones())

	z1 := result[0].MedianPercentage
	z4 := result[3].MedianPercentage

	// The transition record between the two batches falls in Z1 (prev.Power=50W),
	// so Z1 gets one extra interval: 31/(31+30)≈50.8%. Allow 1% tolerance.
	if math.Abs(z1-50) > 1 {
		t.Errorf("zone1 = %.2f%%, want ~50%%", z1)
	}
	if math.Abs(z4-50) > 1 {
		t.Errorf("zone4 = %.2f%%, want ~50%%", z4)
	}
}

func TestComputeMedianZoneDistribution_MedianAcrossActivities(t *testing.T) {
	// Three activities, same zone2 percentages: 20%, 50%, 80% → median=50%.
	ftp := 250
	zones := standardZones()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	buildActivity := func(z1Intervals, z2Intervals int, t0 time.Time) []repository.ActivityPowerRecord {
		id := uuid.New()
		var recs []repository.ActivityPowerRecord
		recs = append(recs, makeRecords(id, t0, 100, z1Intervals+1)...)
		recs = append(recs, makeRecords(id, t0.Add(time.Duration(z1Intervals+1)*time.Second), 160, z2Intervals+1)...)
		return recs
	}

	var records []repository.ActivityPowerRecord
	records = append(records, buildActivity(20, 80, base)...)
	records = append(records, buildActivity(50, 50, base.Add(24*time.Hour))...)
	records = append(records, buildActivity(80, 20, base.Add(48*time.Hour))...)

	result := computeMedianZoneDistribution(records, ftp, zones)

	z2 := result[1].MedianPercentage
	if math.Abs(z2-50) > 0.5 {
		t.Errorf("zone2 median = %.2f%%, want ~50%%", z2)
	}
}

func TestComputeMedianZoneDistribution_GapsLargerThan60sIgnored(t *testing.T) {
	ftp := 200
	actID := uuid.New()
	base := time.Now()
	records := []repository.ActivityPowerRecord{
		{ActivityID: actID, Timestamp: base, Power: 200},
		{ActivityID: actID, Timestamp: base.Add(61 * time.Second), Power: 200},
	}

	result := computeMedianZoneDistribution(records, ftp, standardZones())

	for _, z := range result {
		if z.MedianPercentage != 0 {
			t.Errorf("zone %d: expected 0 (gap ignored), got %.2f", z.ZoneNumber, z.MedianPercentage)
		}
	}
}

func TestComputeMedianZoneDistribution_ZoneWattBoundaries(t *testing.T) {
	ftp := 300
	result := computeMedianZoneDistribution(nil, ftp, standardZones())

	if result[0].MinWatts != 0 {
		t.Errorf("zone1 MinWatts = %d, want 0", result[0].MinWatts)
	}
	if result[0].MaxWatts == nil || *result[0].MaxWatts != 165 {
		t.Errorf("zone1 MaxWatts = %v, want 165", result[0].MaxWatts)
	}
	if result[6].MinWatts != 450 {
		t.Errorf("zone7 MinWatts = %d, want 450", result[6].MinWatts)
	}
	if result[6].MaxWatts != nil {
		t.Errorf("zone7 MaxWatts should be nil (open-ended), got %v", *result[6].MaxWatts)
	}
}

func TestComputeMedianZoneDistribution_ZeroTotalTimeActivitySkipped(t *testing.T) {
	ftp := 200
	zones := standardZones()
	actID := uuid.New()
	records := []repository.ActivityPowerRecord{
		{ActivityID: actID, Timestamp: time.Now(), Power: 300},
	}
	result := computeMedianZoneDistribution(records, ftp, zones)
	for _, z := range result {
		if z.MedianPercentage != 0 {
			t.Errorf("zone %d: expected 0 when single record, got %.2f", z.ZoneNumber, z.MedianPercentage)
		}
	}
}

// ── computeNP ─────────────────────────────────────────────────────────────────

func TestComputeNP_Empty(t *testing.T) {
	if got := computeNP(nil); got != 0 {
		t.Errorf("computeNP(nil) = %v, want 0", got)
	}
}

func TestComputeNP_SingleRecord(t *testing.T) {
	// Need at least 2 records to form any interval.
	recs := makeRecords(uuid.New(), time.Now(), 300, 1)
	if got := computeNP(recs); got != 0 {
		t.Errorf("computeNP(single record) = %v, want 0", got)
	}
}

func TestComputeNP_ConstantPower(t *testing.T) {
	// At constant power the 30s rolling average never changes, so NP == avg power.
	// Use 120 seconds so the full 30s window is reached.
	recs := makeRecords(uuid.New(), time.Now(), 250, 120)
	got := computeNP(recs)
	if math.Abs(got-250) > 0.01 {
		t.Errorf("computeNP(constant 250W) = %.2f, want 250", got)
	}
}

func TestComputeNP_HigherThanAvgForVariableRide(t *testing.T) {
	// NP penalises variability: a ride alternating 100W/400W has same average (250W)
	// as a flat 250W ride, but NP > 250 because high power intervals are weighted ^4.
	base := time.Now()
	actID := uuid.New()
	var recs []repository.ActivityPowerRecord
	for i := 0; i < 120; i++ {
		p := 100
		if i%2 == 0 {
			p = 400
		}
		recs = append(recs, repository.ActivityPowerRecord{
			ActivityID: actID,
			Timestamp:  base.Add(time.Duration(i) * time.Second),
			Power:      p,
		})
	}
	np := computeNP(recs)
	avg := 250.0
	if np <= avg {
		t.Errorf("NP (%.2f) should be > avg (%.2f) for a variable ride", np, avg)
	}
}

func TestComputeNP_HarderRideHasHigherNP(t *testing.T) {
	// Ride A: steady 200W. Ride B: steady 300W. NP(B) > NP(A).
	base := time.Now()
	recA := makeRecords(uuid.New(), base, 200, 120)
	recB := makeRecords(uuid.New(), base, 300, 120)
	npA := computeNP(recA)
	npB := computeNP(recB)
	if npB <= npA {
		t.Errorf("NP(300W ride)=%.2f should exceed NP(200W ride)=%.2f", npB, npA)
	}
}

// ── computeBestActivityZoneDistribution ───────────────────────────────────────

func TestComputeBestActivityZoneDistribution_Empty(t *testing.T) {
	result := computeBestActivityZoneDistribution(nil, 250, standardZones())
	if len(result) != len(standardZones()) {
		t.Fatalf("expected %d zones, got %d", len(standardZones()), len(result))
	}
	for _, z := range result {
		if z.MedianPercentage != 0 {
			t.Errorf("zone %d: expected 0, got %v", z.ZoneNumber, z.MedianPercentage)
		}
	}
}

func TestComputeBestActivityZoneDistribution_SingleActivity(t *testing.T) {
	// One activity, all in zone2. Best is that activity → zone2=100%.
	ftp := 200
	actID := uuid.New()
	records := makeRecords(actID, time.Now(), 120, 61) // 60% FTP → zone2

	result := computeBestActivityZoneDistribution(records, ftp, standardZones())

	for i, z := range result {
		if i == 1 {
			if !approxEq(z.MedianPercentage, 100) {
				t.Errorf("zone2 = %.2f%%, want 100%%", z.MedianPercentage)
			}
		} else {
			if !approxEq(z.MedianPercentage, 0) {
				t.Errorf("zone %d = %.2f%%, want 0%%", z.ZoneNumber, z.MedianPercentage)
			}
		}
	}
}

func TestComputeBestActivityZoneDistribution_PicksHighestNP(t *testing.T) {
	// Act1: 200W steady (lower NP). Act2: 300W steady (higher NP).
	// Best mode should return act2's zone distribution.
	// FTP=250: 200W=80% → zone3 (75–90%); 300W=120% → zone6 (120–150%).
	ftp := 250
	zones := standardZones()
	base := time.Now()

	act1 := uuid.New()
	act2 := uuid.New()
	var records []repository.ActivityPowerRecord
	records = append(records, makeRecords(act1, base, 200, 120)...)
	records = append(records, makeRecords(act2, base.Add(2*time.Hour), 300, 120)...)

	result := computeBestActivityZoneDistribution(records, ftp, zones)

	// act2 is all in zone6 (index 5), so zone6=100%.
	if !approxEq(result[5].MedianPercentage, 100) {
		t.Errorf("zone6 = %.2f%%, want 100%% (from act2)", result[5].MedianPercentage)
	}
	// zone3 (act1's zone) should be 0% because act1 was not picked.
	if !approxEq(result[2].MedianPercentage, 0) {
		t.Errorf("zone3 = %.2f%%, want 0%% (act1 not picked)", result[2].MedianPercentage)
	}
}

func TestComputeBestActivityZoneDistribution_SingleRideNotMultipleZones(t *testing.T) {
	// Unlike old max-per-zone, best mode returns one ride's distribution.
	// Act1 is all zone1, act2 is all zone2. The harder ride (act2, higher NP) wins.
	// So zone1 must be 0% — it does NOT blend both activities.
	ftp := 200
	zones := standardZones()
	base := time.Now()

	act1 := uuid.New()
	act2 := uuid.New()
	var records []repository.ActivityPowerRecord
	records = append(records, makeRecords(act1, base, 50, 120)...)              // Z1, lower NP
	records = append(records, makeRecords(act2, base.Add(2*time.Hour), 120, 120)...) // Z2, higher NP

	result := computeBestActivityZoneDistribution(records, ftp, zones)

	// act2 wins → zone2=100%, zone1=0%
	if !approxEq(result[1].MedianPercentage, 100) {
		t.Errorf("zone2 = %.2f%%, want 100%%", result[1].MedianPercentage)
	}
	if !approxEq(result[0].MedianPercentage, 0) {
		t.Errorf("zone1 = %.2f%%, want 0%% (act1 not picked)", result[0].MedianPercentage)
	}
}

func TestComputeBestActivityZoneDistribution_GapsIgnored(t *testing.T) {
	ftp := 200
	actID := uuid.New()
	base := time.Now()
	records := []repository.ActivityPowerRecord{
		{ActivityID: actID, Timestamp: base, Power: 300},
		{ActivityID: actID, Timestamp: base.Add(61 * time.Second), Power: 300},
	}
	result := computeBestActivityZoneDistribution(records, ftp, standardZones())
	for _, z := range result {
		if z.MedianPercentage != 0 {
			t.Errorf("zone %d: expected 0 (gap ignored), got %.2f", z.ZoneNumber, z.MedianPercentage)
		}
	}
}

func TestComputeBestActivityZoneDistribution_ZoneWattBoundaries(t *testing.T) {
	ftp := 300
	result := computeBestActivityZoneDistribution(nil, ftp, standardZones())

	if result[0].MinWatts != 0 {
		t.Errorf("zone1 MinWatts = %d, want 0", result[0].MinWatts)
	}
	if result[0].MaxWatts == nil || *result[0].MaxWatts != 165 {
		t.Errorf("zone1 MaxWatts = %v, want 165", result[0].MaxWatts)
	}
	if result[6].MaxWatts != nil {
		t.Errorf("zone7 MaxWatts should be nil, got %v", *result[6].MaxWatts)
	}
}

// ── StatisticsService.GetAvailablePowerYears ──────────────────────────────────

func TestGetAvailablePowerYears_Passthrough(t *testing.T) {
	want := []int{2022, 2023, 2024}
	svc := NewStatisticsService(&stubStatisticsRepo{years: want})

	got, err := svc.GetAvailablePowerYears(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestGetAvailablePowerYears_Error(t *testing.T) {
	sentinel := errors.New("db error")
	svc := NewStatisticsService(&stubStatisticsRepo{yearsErr: sentinel})

	_, err := svc.GetAvailablePowerYears(context.Background(), uuid.New())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

// ── StatisticsService.GetAnnualPowerStats ─────────────────────────────────────

func TestGetAnnualPowerStats_NilCurveBecomesEmpty(t *testing.T) {
	svc := NewStatisticsService(&stubStatisticsRepo{curve: nil})

	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 0, nil, "avg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.PowerCurve == nil {
		t.Error("PowerCurve should be non-nil empty slice, got nil")
	}
	if len(stats.PowerCurve) != 0 {
		t.Errorf("expected 0 curve points, got %d", len(stats.PowerCurve))
	}
}

func TestGetAnnualPowerStats_NoFTPSkipsDistribution(t *testing.T) {
	svc := NewStatisticsService(&stubStatisticsRepo{curve: []model.AnnualPowerCurvePoint{{DurationSeconds: 60, MedianPower: 200}}})

	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 0, standardZones(), "avg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats.ZoneDistribution) != 0 {
		t.Errorf("expected empty ZoneDistribution when ftp=0, got %d entries", len(stats.ZoneDistribution))
	}
}

func TestGetAnnualPowerStats_NoZonesSkipsDistribution(t *testing.T) {
	svc := NewStatisticsService(&stubStatisticsRepo{})

	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 250, nil, "avg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats.ZoneDistribution) != 0 {
		t.Errorf("expected empty ZoneDistribution when zones=nil, got %d entries", len(stats.ZoneDistribution))
	}
}

func TestGetAnnualPowerStats_NilDistributionBecomesEmpty(t *testing.T) {
	svc := NewStatisticsService(&stubStatisticsRepo{records: nil})

	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 250, standardZones(), "avg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ZoneDistribution == nil {
		t.Error("ZoneDistribution should be non-nil, got nil")
	}
}

func TestGetAnnualPowerStats_CurveError(t *testing.T) {
	sentinel := errors.New("curve db error")
	svc := NewStatisticsService(&stubStatisticsRepo{curveErr: sentinel})

	_, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 250, standardZones(), "avg")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected curve error, got %v", err)
	}
}

func TestGetAnnualPowerStats_RecordsError(t *testing.T) {
	sentinel := errors.New("records db error")
	svc := NewStatisticsService(&stubStatisticsRepo{recordsErr: sentinel})

	_, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 250, standardZones(), "avg")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected records error, got %v", err)
	}
}

func TestGetAnnualPowerStats_ReturnsCurvePoints(t *testing.T) {
	curve := []model.AnnualPowerCurvePoint{
		{DurationSeconds: 5, MedianPower: 400},
		{DurationSeconds: 60, MedianPower: 300},
	}
	svc := NewStatisticsService(&stubStatisticsRepo{curve: curve})

	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 0, nil, "avg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats.PowerCurve) != 2 {
		t.Fatalf("expected 2 curve points, got %d", len(stats.PowerCurve))
	}
	if stats.PowerCurve[0].MedianPower != 400 || stats.PowerCurve[1].MedianPower != 300 {
		t.Errorf("unexpected curve values: %+v", stats.PowerCurve)
	}
}

// ── GetAnnualPowerStats mode routing ─────────────────────────────────────────

func TestGetAnnualPowerStats_ModeBestPicksHardestRide(t *testing.T) {
	// Act1: 200W (zone3 with FTP=250). Act2: 300W (zone6). Act2 has higher NP → picked.
	ftp := 250
	zones := standardZones()
	base := time.Now()

	act1 := uuid.New()
	act2 := uuid.New()
	var records []repository.ActivityPowerRecord
	records = append(records, makeRecords(act1, base, 200, 120)...)
	records = append(records, makeRecords(act2, base.Add(2*time.Hour), 300, 120)...)

	svc := NewStatisticsService(&stubStatisticsRepo{records: records})
	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, ftp, zones, "best")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Act2 (zone6, index 5) should be 100%.
	if !approxEq(stats.ZoneDistribution[5].MedianPercentage, 100) {
		t.Errorf("zone6 best = %.2f%%, want 100%%", stats.ZoneDistribution[5].MedianPercentage)
	}
	// Zone3 (act1) must be 0% — only one ride is returned.
	if !approxEq(stats.ZoneDistribution[2].MedianPercentage, 0) {
		t.Errorf("zone3 best = %.2f%%, want 0%% (act1 not picked)", stats.ZoneDistribution[2].MedianPercentage)
	}
}

func TestGetAnnualPowerStats_ModeAvgUsesMedian(t *testing.T) {
	// Three activities with zone2 at 20%, 50%, 80% → median=50%.
	ftp := 250
	zones := standardZones()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	buildActivity := func(z1Intervals, z2Intervals int, t0 time.Time) []repository.ActivityPowerRecord {
		id := uuid.New()
		var recs []repository.ActivityPowerRecord
		recs = append(recs, makeRecords(id, t0, 100, z1Intervals+1)...)
		recs = append(recs, makeRecords(id, t0.Add(time.Duration(z1Intervals+1)*time.Second), 160, z2Intervals+1)...)
		return recs
	}

	var records []repository.ActivityPowerRecord
	records = append(records, buildActivity(20, 80, base)...)
	records = append(records, buildActivity(50, 50, base.Add(24*time.Hour))...)
	records = append(records, buildActivity(80, 20, base.Add(48*time.Hour))...)

	svc := NewStatisticsService(&stubStatisticsRepo{records: records})
	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, ftp, zones, "avg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	z2 := stats.ZoneDistribution[1].MedianPercentage
	if math.Abs(z2-50) > 0.5 {
		t.Errorf("zone2 avg = %.2f%%, want ~50%%", z2)
	}
}

func TestGetAnnualPowerStats_UnknownModeFallsBackToAvg(t *testing.T) {
	svc := NewStatisticsService(&stubStatisticsRepo{})
	stats, err := svc.GetAnnualPowerStats(context.Background(), uuid.New(), 2024, 0, nil, "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
}
