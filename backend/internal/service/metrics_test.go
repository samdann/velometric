package service

import (
	"testing"
)

func TestComputeNormalizedPower(t *testing.T) {
	tests := []struct {
		name     string
		powers   []int
		expected int
	}{
		{
			name:     "empty slice returns 0",
			powers:   []int{},
			expected: 0,
		},
		{
			name:     "fewer than 30 samples falls back to average",
			powers:   []int{100, 200, 300}, // avg of non-zero = 200
			expected: 200,
		},
		{
			name:     "exactly 29 samples falls back to average",
			powers:   makeN(150, 29),
			expected: 150,
		},
		{
			name:     "constant power: NP equals the power value",
			powers:   makeN(200, 60),
			expected: 200,
		},
		{
			// 30s at 300W then 30s at 100W: avg=200W, but NP > 200
			// because high-power windows are weighted heavily by x^4
			name:     "variable power: NP exceeds average",
			powers:   concat(makeN(300, 30), makeN(100, 30)),
			expected: 223, // 4th root of avg(rolling^4) for this input; avg=200, so NP > avg confirms x^4 weighting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeNormalizedPower(tt.powers)
			if got != tt.expected {
				t.Errorf("ComputeNormalizedPower() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestComputeTSS(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		np       int
		ftp      int
		expected float64
	}{
		{
			name:     "zero FTP returns 0",
			duration: 3600,
			np:       200,
			ftp:      0,
			expected: 0,
		},
		{
			name:     "negative FTP returns 0",
			duration: 3600,
			np:       200,
			ftp:      -100,
			expected: 0,
		},
		{
			name:     "zero duration returns 0",
			duration: 0,
			np:       200,
			ftp:      200,
			expected: 0,
		},
		{
			name:     "zero normalized power returns 0",
			duration: 3600,
			np:       0,
			ftp:      200,
			expected: 0,
		},
		{
			name:     "normal case with 1 hour duration, 200NP, 200FTP",
			duration: 3600,
			np:       200,
			ftp:      200,
			expected: 100, // (3600 * 200 * (200/200)) / (200 * 3600) * 100 = 100
		},
		{
			name:     "normal case with 1 hour duration, 300NP, 200FTP",
			duration: 3600,
			np:       300,
			ftp:      200,
			expected: 225, // (3600 * 300 * (300/200)) / (200 * 3600) * 100 = 225
		},
		{
			name:     "normal case with 30 min duration, 200NP, 200FTP",
			duration: 1800,
			np:       200,
			ftp:      200,
			expected: 50, // (1800 * 200 * (200/200)) / (200 * 3600) * 100 = 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeTSS(tt.duration, tt.np, tt.ftp)
			if got != tt.expected {
				t.Errorf("ComputeTSS() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeIntensityFactor(t *testing.T) {
	tests := []struct {
		name     string
		np       int
		ftp      int
		expected float64
	}{
		{"zero FTP returns 0", 200, 0, 0},
		{"negative FTP returns 0", 200, -100, 0},
		{"NP equals FTP returns 1.0", 200, 200, 1.0},
		{"NP at 75% FTP returns 0.75", 150, 200, 0.75},
		{"NP above FTP", 300, 200, 1.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeIntensityFactor(tt.np, tt.ftp)
			if got != tt.expected {
				t.Errorf("ComputeIntensityFactor() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeVariabilityIndex(t *testing.T) {
	tests := []struct {
		name     string
		np       int
		ap       int
		expected float64
	}{
		{"zero average power returns 0", 200, 0, 0},
		{"negative average power returns 0", 200, -1, 0},
		{"NP equals AP returns 1.0", 200, 200, 1.0},
		{"NP greater than AP", 220, 200, 1.1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeVariabilityIndex(tt.np, tt.ap)
			if got != tt.expected {
				t.Errorf("ComputeVariabilityIndex() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeAverage(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		expected int
	}{
		{"empty slice returns 0", []int{}, 0},
		{"all zeros returns 0", []int{0, 0, 0}, 0},
		{"ignores zero values", []int{0, 100, 200}, 150},
		{"uniform values", []int{100, 100, 100}, 100},
		{"mixed values", []int{50, 150, 200, 0}, 133}, // (50+150+200)/3 = 133
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeAverage(tt.values)
			if got != tt.expected {
				t.Errorf("ComputeAverage() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestComputeMax(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		expected int
	}{
		{"empty slice returns 0", []int{}, 0},
		{"single element", []int{42}, 42},
		{"max at end", []int{1, 2, 3, 100}, 100},
		{"max at start", []int{100, 3, 2, 1}, 100},
		{"all same", []int{5, 5, 5}, 5},
		{"includes negative", []int{-10, -5, -1}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeMax(tt.values)
			if got != tt.expected {
				t.Errorf("ComputeMax() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestComputeAverageFloat(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty slice returns 0", []float64{}, 0},
		{"all zeros returns 0", []float64{0, 0, 0}, 0},
		{"ignores zero values", []float64{0, 10.0, 20.0}, 15.0},
		{"uniform values", []float64{5.5, 5.5, 5.5}, 5.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeAverageFloat(tt.values)
			if got != tt.expected {
				t.Errorf("ComputeAverageFloat() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeMaxFloat(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty slice returns 0", []float64{}, 0},
		{"single element", []float64{3.14}, 3.14},
		{"max in middle", []float64{1.0, 9.9, 2.0}, 9.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeMaxFloat(tt.values)
			if got != tt.expected {
				t.Errorf("ComputeMaxFloat() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeElevationGain(t *testing.T) {
	tests := []struct {
		name      string
		altitudes []float64
		expected  float64
	}{
		{"empty slice returns 0", []float64{}, 0},
		{"single element returns 0", []float64{100}, 0},
		{"flat returns 0", []float64{100, 100, 100}, 0},
		{"only descending returns 0", []float64{300, 200, 100}, 0},
		{"simple climb", []float64{100, 110, 120}, 20},
		{"climb and descent counts only ascent", []float64{100, 150, 120, 170}, 100}, // +50 +0 +50 = 100
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeElevationGain(tt.altitudes)
			if got != tt.expected {
				t.Errorf("ComputeElevationGain() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeGradient(t *testing.T) {
	tests := []struct {
		name          string
		distanceDelta float64
		altitudeDelta float64
		expected      float64
	}{
		{"zero distance returns 0", 0, 10, 0},
		{"negative distance returns 0", -5, 10, 0},
		{"flat", 100, 0, 0},
		{"10% climb: 10m rise over 100m", 100, 10, 10},
		{"descent is negative", 100, -5, -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeGradient(tt.distanceDelta, tt.altitudeDelta)
			if got != tt.expected {
				t.Errorf("ComputeGradient() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestComputeZoneDistribution(t *testing.T) {
	zones := []ZoneBoundary{
		{Zone: 1, Min: 0, Max: 100},
		{Zone: 2, Min: 100, Max: 200},
		{Zone: 3, Min: 200, Max: 0}, // no upper limit
	}
	tests := []struct {
		name     string
		values   []int
		expected map[int]int
	}{
		{"empty values returns empty map", []int{}, map[int]int{}},
		{"all in zone 1", []int{50, 60, 70}, map[int]int{1: 3}},
		{"spread across zones", []int{50, 150, 250}, map[int]int{1: 1, 2: 1, 3: 1}},
		{"boundary value goes to higher zone", []int{100}, map[int]int{2: 1}},
		{"unbounded zone 3 includes high values", []int{999}, map[int]int{3: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeZoneDistribution(tt.values, zones)
			if len(got) != len(tt.expected) {
				t.Errorf("ComputeZoneDistribution() len=%d, want %d: %v", len(got), len(tt.expected), got)
				return
			}
			for zone, count := range tt.expected {
				if got[zone] != count {
					t.Errorf("zone %d: got %d, want %d", zone, got[zone], count)
				}
			}
		})
	}
}

func TestComputePowerCurve(t *testing.T) {
	t.Run("empty returns empty map", func(t *testing.T) {
		got := ComputePowerCurve([]int{})
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})
	t.Run("constant power: best avg equals the power", func(t *testing.T) {
		powers := makeN(200, 120)
		got := ComputePowerCurve(powers)
		if got[60] != 200 {
			t.Errorf("60s best power = %d, want 200", got[60])
		}
		if got[120] != 200 {
			t.Errorf("120s best power = %d, want 200", got[120])
		}
	})
	t.Run("duration longer than data is excluded", func(t *testing.T) {
		powers := makeN(200, 10)
		got := ComputePowerCurve(powers)
		if _, ok := got[60]; ok {
			t.Errorf("60s should not appear when data has only 10 samples")
		}
	})
	t.Run("best window is selected, not first window", func(t *testing.T) {
		// 5s at 100W then 5s at 300W: best 5s avg should be 300
		powers := concat(makeN(100, 5), makeN(300, 5))
		got := ComputePowerCurve(powers)
		if got[5] != 300 {
			t.Errorf("5s best power = %d, want 300", got[5])
		}
	})
}

func TestComputePowerCurveExtended(t *testing.T) {
	t.Run("empty returns empty map", func(t *testing.T) {
		got := ComputePowerCurveExtended([]PowerCurveRecord{})
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})

	t.Run("constant power and HR: best power and avg HR match", func(t *testing.T) {
		records := make([]PowerCurveRecord, 60)
		for i := range records {
			records[i] = PowerCurveRecord{Power: 200, HeartRate: 150}
		}
		got := ComputePowerCurveExtended(records)
		r, ok := got[60]
		if !ok {
			t.Fatal("60s entry missing")
		}
		if r.BestPower != 200 {
			t.Errorf("BestPower = %d, want 200", r.BestPower)
		}
		if r.AvgHeartRate == nil || *r.AvgHeartRate != 150 {
			t.Errorf("AvgHeartRate = %v, want 150", r.AvgHeartRate)
		}
	})

	t.Run("nil optional fields when not present", func(t *testing.T) {
		records := make([]PowerCurveRecord, 5)
		for i := range records {
			records[i] = PowerCurveRecord{Power: 100}
		}
		got := ComputePowerCurveExtended(records)
		r, ok := got[5]
		if !ok {
			t.Fatal("5s entry missing")
		}
		if r.AvgHeartRate != nil {
			t.Errorf("AvgHeartRate should be nil when no HR data")
		}
		if r.AvgSpeed != nil {
			t.Errorf("AvgSpeed should be nil when no speed data")
		}
	})
}

func TestMedianFilter(t *testing.T) {
	tests := []struct {
		name       string
		values     []int
		windowSize int
		expected   []int
	}{
		{"empty slice", []int{}, 3, []int{}},
		{"window size 1 returns same", []int{5, 3, 8}, 1, []int{5, 3, 8}},
		{"removes spike in middle", []int{10, 10, 100, 10, 10}, 3, []int{10, 10, 10, 10, 10}},
		{"single element", []int{42}, 3, []int{42}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MedianFilter(tt.values, tt.windowSize)
			if len(got) != len(tt.expected) {
				t.Errorf("MedianFilter() len=%d, want %d", len(got), len(tt.expected))
				return
			}
			for i, v := range tt.expected {
				if got[i] != v {
					t.Errorf("MedianFilter()[%d] = %d, want %d", i, got[i], v)
				}
			}
		})
	}
}

func TestDefaultPowerZones(t *testing.T) {
	zones := DefaultPowerZones(200)
	if len(zones) != 7 {
		t.Fatalf("expected 7 zones, got %d", len(zones))
	}
	// Zone numbering
	for i, z := range zones {
		if z.Zone != i+1 {
			t.Errorf("zones[%d].Zone = %d, want %d", i, z.Zone, i+1)
		}
	}
	// Last zone has no upper limit
	if zones[6].Max != 0 {
		t.Errorf("last zone Max = %d, want 0 (unbounded)", zones[6].Max)
	}
	// Zones are ordered: each Min equals previous Max
	for i := 1; i < len(zones); i++ {
		if zones[i].Min != zones[i-1].Max {
			t.Errorf("zone %d Min (%d) != zone %d Max (%d)", i+1, zones[i].Min, i, zones[i-1].Max)
		}
	}
}

func TestDefaultHRZones(t *testing.T) {
	zones := DefaultHRZones(200)
	if len(zones) != 5 {
		t.Fatalf("expected 5 zones, got %d", len(zones))
	}
	for i, z := range zones {
		if z.Zone != i+1 {
			t.Errorf("zones[%d].Zone = %d, want %d", i, z.Zone, i+1)
		}
	}
	if zones[4].Max != 0 {
		t.Errorf("last zone Max = %d, want 0 (unbounded)", zones[4].Max)
	}
	for i := 1; i < len(zones); i++ {
		if zones[i].Min != zones[i-1].Max {
			t.Errorf("zone %d Min (%d) != zone %d Max (%d)", i+1, zones[i].Min, i, zones[i-1].Max)
		}
	}
}

// makeN returns a slice of n identical values.
func makeN(val, n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = val
	}
	return s
}

// concat joins two slices.
func concat(a, b []int) []int {
	result := make([]int, len(a)+len(b))
	copy(result, a)
	copy(result[len(a):], b)
	return result
}
