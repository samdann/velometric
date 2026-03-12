package service

import (
	"testing"
	"time"

	"github.com/velometric/backend/internal/model"
)

// ── generateActivityName ──────────────────────────────────────────────────────

func TestGenerateActivityName(t *testing.T) {
	tests := []struct {
		name     string
		sport    string
		hour     int
		expected string
	}{
		// Time-of-day labels
		{"morning lower boundary (5h)", "cycling", 5, "Morning Ride"},
		{"morning mid (7h)", "cycling", 7, "Morning Ride"},
		{"morning upper boundary (11h)", "cycling", 11, "Morning Ride"},
		{"afternoon lower boundary (12h)", "cycling", 12, "Afternoon Ride"},
		{"afternoon mid (14h)", "cycling", 14, "Afternoon Ride"},
		{"afternoon upper boundary (16h)", "cycling", 16, "Afternoon Ride"},
		{"evening lower boundary (17h)", "cycling", 17, "Evening Ride"},
		{"evening mid (19h)", "cycling", 19, "Evening Ride"},
		{"evening upper boundary (20h)", "cycling", 20, "Evening Ride"},
		{"night lower boundary (21h)", "cycling", 21, "Night Ride"},
		{"night midnight (0h)", "cycling", 0, "Night Ride"},
		{"night upper boundary (4h)", "cycling", 4, "Night Ride"},

		// Sport name mapping
		{"empty sport defaults to Ride", "", 8, "Morning Ride"},
		{"cycling sport defaults to Ride", "cycling", 8, "Morning Ride"},
		{"non-cycling sport used verbatim", "running", 8, "Morning running"},
		{"swim sport used verbatim", "swim", 12, "Afternoon swim"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Date(2024, 1, 1, tt.hour, 0, 0, 0, time.UTC)
			got := generateActivityName(tt.sport, startTime)
			if got != tt.expected {
				t.Errorf("generateActivityName(%q, hour=%d) = %q, want %q", tt.sport, tt.hour, got, tt.expected)
			}
		})
	}
}

// ── classifyHRZone ───────────────────────────────────────────────────────────

// pf is a float64 pointer helper for test zone boundaries.
func pf(v float64) *float64 { return &v }

// standardHRZones returns a 5-zone setup as percentages of maxHR.
//
//	Zone 1:  0–60%
//	Zone 2: 60–70%
//	Zone 3: 70–80%
//	Zone 4: 80–90%
//	Zone 5: 90–∞  (unbounded)
func standardHRZones() []model.HRZone {
	return []model.HRZone{
		{ZoneNumber: 1, MinPercentage: 0, MaxPercentage: pf(60)},
		{ZoneNumber: 2, MinPercentage: 60, MaxPercentage: pf(70)},
		{ZoneNumber: 3, MinPercentage: 70, MaxPercentage: pf(80)},
		{ZoneNumber: 4, MinPercentage: 80, MaxPercentage: pf(90)},
		{ZoneNumber: 5, MinPercentage: 90, MaxPercentage: nil},
	}
}

func TestClassifyHRZone(t *testing.T) {
	zones := standardHRZones()

	tests := []struct {
		name      string
		hrPct     float64
		wantIndex int // 0-based index into zones slice
	}{
		// Zone 1 (0–60%)
		{"well inside zone 1", 50, 0},
		{"zone 1 lower bound (0%)", 0, 0},
		{"just below zone 2 boundary", 59.9, 0},

		// Zone 2 (60–70%)
		{"zone 2 lower bound (60%)", 60, 1},
		{"zone 2 mid", 65, 1},
		{"just below zone 3 boundary", 69.9, 1},

		// Zone 3 (70–80%)
		{"zone 3 lower bound (70%)", 70, 2},
		{"zone 3 mid", 75, 2},

		// Zone 4 (80–90%)
		{"zone 4 lower bound (80%)", 80, 3},
		{"zone 4 mid", 85, 3},
		{"just below zone 5 boundary", 89.9, 3},

		// Zone 5 (90–∞, unbounded)
		{"zone 5 lower bound (90%)", 90, 4},
		{"well above all zones (120%)", 120, 4},
		{"extreme value (200%)", 200, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyHRZone(tt.hrPct, zones)
			if got != tt.wantIndex {
				t.Errorf("classifyHRZone(%.1f) = index %d (zone %d), want index %d (zone %d)",
					tt.hrPct, got, zones[got].ZoneNumber, tt.wantIndex, zones[tt.wantIndex].ZoneNumber)
			}
		})
	}
}

func TestClassifyHRZone_SingleZone(t *testing.T) {
	// Edge case: only one zone (unbounded)
	zones := []model.HRZone{
		{ZoneNumber: 1, MinPercentage: 0, MaxPercentage: nil},
	}
	for _, hrPct := range []float64{0, 50, 100, 200} {
		got := classifyHRZone(hrPct, zones)
		if got != 0 {
			t.Errorf("classifyHRZone(%.0f) with single zone = %d, want 0", hrPct, got)
		}
	}
}

// ── classifyPowerZone ─────────────────────────────────────────────────────────

// standardPowerZones returns a 7-zone setup (Coggan model) as % of FTP.
//
//	Z1:   0–55%    Active Recovery
//	Z2:  55–75%    Endurance
//	Z3:  75–90%    Tempo
//	Z4:  90–105%   Threshold
//	Z5: 105–120%   VO2 Max
//	Z6: 120–150%   Anaerobic
//	Z7: 150–∞      Neuromuscular
func standardPowerZones() []model.PowerZone {
	return []model.PowerZone{
		{ZoneNumber: 1, MinPercentage: 0, MaxPercentage: pf(55)},
		{ZoneNumber: 2, MinPercentage: 55, MaxPercentage: pf(75)},
		{ZoneNumber: 3, MinPercentage: 75, MaxPercentage: pf(90)},
		{ZoneNumber: 4, MinPercentage: 90, MaxPercentage: pf(105)},
		{ZoneNumber: 5, MinPercentage: 105, MaxPercentage: pf(120)},
		{ZoneNumber: 6, MinPercentage: 120, MaxPercentage: pf(150)},
		{ZoneNumber: 7, MinPercentage: 150, MaxPercentage: nil},
	}
}

func TestClassifyPowerZone(t *testing.T) {
	zones := standardPowerZones()

	tests := []struct {
		name      string
		powerPct  float64
		wantIndex int
	}{
		// Zone 1 (0–55%)
		{"recovery (30%)", 30, 0},
		{"zone 1 lower bound (0%)", 0, 0},
		{"just below zone 2 (54.9%)", 54.9, 0},

		// Zone 2 (55–75%)
		{"zone 2 lower bound (55%)", 55, 1},
		{"endurance (65%)", 65, 1},

		// Zone 3 (75–90%)
		{"tempo lower bound (75%)", 75, 2},
		{"tempo (80%)", 80, 2},

		// Zone 4 (90–105%)
		{"threshold lower bound (90%)", 90, 3},
		{"at FTP (100%)", 100, 3},
		{"just above FTP (104%)", 104, 3},

		// Zone 5 (105–120%)
		{"VO2 max lower bound (105%)", 105, 4},
		{"VO2 max (115%)", 115, 4},

		// Zone 6 (120–150%)
		{"anaerobic lower bound (120%)", 120, 5},
		{"anaerobic (135%)", 135, 5},

		// Zone 7 (150–∞, unbounded)
		{"neuromuscular lower bound (150%)", 150, 6},
		{"sprint (250%)", 250, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyPowerZone(tt.powerPct, zones)
			if got != tt.wantIndex {
				t.Errorf("classifyPowerZone(%.1f) = index %d (zone %d), want index %d (zone %d)",
					tt.powerPct, got, zones[got].ZoneNumber, tt.wantIndex, zones[tt.wantIndex].ZoneNumber)
			}
		})
	}
}

func TestClassifyPowerZone_SingleZone(t *testing.T) {
	zones := []model.PowerZone{
		{ZoneNumber: 1, MinPercentage: 0, MaxPercentage: nil},
	}
	for _, pct := range []float64{0, 100, 300} {
		got := classifyPowerZone(pct, zones)
		if got != 0 {
			t.Errorf("classifyPowerZone(%.0f) with single zone = %d, want 0", pct, got)
		}
	}
}
