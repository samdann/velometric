package service

import (
	"testing"
)

// ── buildHRZones ──────────────────────────────────────────────────────────────
//
// Input: maxHR=200, boundaries=[120, 140, 160, 180]
//
//	Zone 1: Recovery    0%  – 60%  (#64748B)
//	Zone 2: Endurance  60%  – 70%  (#3B82F6)
//	Zone 3: Tempo      70%  – 80%  (#22C55E)
//	Zone 4: Threshold  80%  – 90%  (#EAB308)
//	Zone 5: Anaerobic  90%  – ∞    (#F97316)

func TestBuildHRZones_Count(t *testing.T) {
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	if len(zones) != 5 {
		t.Fatalf("expected 5 zones, got %d", len(zones))
	}
}

func TestBuildHRZones_ZoneNumbers(t *testing.T) {
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	for i, z := range zones {
		if z.ZoneNumber != i+1 {
			t.Errorf("zones[%d].ZoneNumber = %d, want %d", i, z.ZoneNumber, i+1)
		}
	}
}

func TestBuildHRZones_Names(t *testing.T) {
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	want := hrZoneNames
	for i, z := range zones {
		if z.Name != want[i] {
			t.Errorf("zones[%d].Name = %q, want %q", i, z.Name, want[i])
		}
	}
}

func TestBuildHRZones_Colors(t *testing.T) {
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	for i, z := range zones {
		if z.Color != hrZoneColors[i] {
			t.Errorf("zones[%d].Color = %q, want %q", i, z.Color, hrZoneColors[i])
		}
	}
}

func TestBuildHRZones_BoundariesConvertedToPercentages(t *testing.T) {
	// maxHR=200, boundaries=[120, 140, 160, 180] → 60%, 70%, 80%, 90%
	zones := buildHRZones(200, []int{120, 140, 160, 180})

	wantMin := []float64{0, 60, 70, 80, 90}
	wantMax := []float64{60, 70, 80, 90, 0} // 0 = unbounded (nil)

	const epsilon = 1e-9
	for i, z := range zones {
		if diff := z.MinPercentage - wantMin[i]; diff > epsilon || diff < -epsilon {
			t.Errorf("zones[%d].MinPercentage = %.4f, want %.4f", i, z.MinPercentage, wantMin[i])
		}
		if i < 4 {
			if z.MaxPercentage == nil {
				t.Errorf("zones[%d].MaxPercentage is nil, want %.1f", i, wantMax[i])
			} else {
				diff := *z.MaxPercentage - wantMax[i]
				if diff > epsilon || diff < -epsilon {
					t.Errorf("zones[%d].MaxPercentage = %.4f, want %.4f", i, *z.MaxPercentage, wantMax[i])
				}
			}
		}
	}
}

func TestBuildHRZones_LastZoneUnbounded(t *testing.T) {
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	last := zones[len(zones)-1]
	if last.MaxPercentage != nil {
		t.Errorf("last zone MaxPercentage should be nil (unbounded), got %v", last.MaxPercentage)
	}
}

func TestBuildHRZones_ContinuousChain(t *testing.T) {
	// Each zone's Min must equal the previous zone's Max — no gaps or overlaps.
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	for i := 1; i < len(zones); i++ {
		prevMax := *zones[i-1].MaxPercentage
		if zones[i].MinPercentage != prevMax {
			t.Errorf("zone %d Min (%.1f) != zone %d Max (%.1f) — gap in zone chain",
				i+1, zones[i].MinPercentage, i, prevMax)
		}
	}
}

func TestBuildHRZones_FirstZoneStartsAtZero(t *testing.T) {
	zones := buildHRZones(200, []int{120, 140, 160, 180})
	if zones[0].MinPercentage != 0 {
		t.Errorf("first zone MinPercentage = %.1f, want 0", zones[0].MinPercentage)
	}
}

// ── buildPowerZones ───────────────────────────────────────────────────────────
//
// Input: ftp=200, boundaries=[110, 150, 180, 210, 240, 300]  (55%, 75%, 90%, 105%, 120%, 150%)
//
//	Zone 1: Recovery      0%  – 55%
//	Zone 2: Endurance    55%  – 75%
//	Zone 3: Tempo        75%  – 90%
//	Zone 4: Threshold    90%  – 105%
//	Zone 5: VO2 Max     105%  – 120%
//	Zone 6: Anaerobic   120%  – 150%
//	Zone 7: Neuromuscular 150% – ∞

func TestBuildPowerZones_Count(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	if len(zones) != 7 {
		t.Fatalf("expected 7 zones, got %d", len(zones))
	}
}

func TestBuildPowerZones_ZoneNumbers(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	for i, z := range zones {
		if z.ZoneNumber != i+1 {
			t.Errorf("zones[%d].ZoneNumber = %d, want %d", i, z.ZoneNumber, i+1)
		}
	}
}

func TestBuildPowerZones_Names(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	for i, z := range zones {
		if z.Name != powerZoneNames[i] {
			t.Errorf("zones[%d].Name = %q, want %q", i, z.Name, powerZoneNames[i])
		}
	}
}

func TestBuildPowerZones_Colors(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	for i, z := range zones {
		if z.Color != powerZoneColors[i] {
			t.Errorf("zones[%d].Color = %q, want %q", i, z.Color, powerZoneColors[i])
		}
	}
}

func TestBuildPowerZones_BoundariesConvertedToPercentages(t *testing.T) {
	// ftp=200, boundaries=[110, 150, 180, 210, 240, 300]
	// percentages:          55%, 75%, 90%, 105%, 120%, 150%
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})

	wantMin := []float64{0, 55, 75, 90, 105, 120, 150}
	wantMax := []float64{55, 75, 90, 105, 120, 150, 0} // 0 = nil

	const epsilon = 1e-9
	for i, z := range zones {
		if diff := z.MinPercentage - wantMin[i]; diff > epsilon || diff < -epsilon {
			t.Errorf("zones[%d].MinPercentage = %.4f, want %.4f", i, z.MinPercentage, wantMin[i])
		}
		if i < 6 {
			if z.MaxPercentage == nil {
				t.Errorf("zones[%d].MaxPercentage is nil, want %.1f", i, wantMax[i])
			} else {
				diff := *z.MaxPercentage - wantMax[i]
				if diff > epsilon || diff < -epsilon {
					t.Errorf("zones[%d].MaxPercentage = %.4f, want %.4f", i, *z.MaxPercentage, wantMax[i])
				}
			}
		}
	}
}

func TestBuildPowerZones_LastZoneUnbounded(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	last := zones[len(zones)-1]
	if last.MaxPercentage != nil {
		t.Errorf("last zone MaxPercentage should be nil (unbounded), got %v", last.MaxPercentage)
	}
}

func TestBuildPowerZones_ContinuousChain(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	for i := 1; i < len(zones); i++ {
		prevMax := *zones[i-1].MaxPercentage
		if zones[i].MinPercentage != prevMax {
			t.Errorf("zone %d Min (%.1f) != zone %d Max (%.1f)",
				i+1, zones[i].MinPercentage, i, prevMax)
		}
	}
}

func TestBuildPowerZones_FirstZoneStartsAtZero(t *testing.T) {
	zones := buildPowerZones(200, []int{110, 150, 180, 210, 240, 300})
	if zones[0].MinPercentage != 0 {
		t.Errorf("first zone MinPercentage = %.1f, want 0", zones[0].MinPercentage)
	}
}
