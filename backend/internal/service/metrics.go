package service

import (
	"math"
	"sort"
)

// ComputeNormalizedPower calculates Normalized Power from power data
// NP = 4th root of (average of (30-second rolling average power)^4)
func ComputeNormalizedPower(powers []int) int {
	if len(powers) < 30 {
		return ComputeAverage(powers)
	}

	// Calculate 30-second rolling averages
	rollingAvgs := make([]float64, len(powers)-29)
	var sum float64
	for i := 0; i < 30; i++ {
		sum += float64(powers[i])
	}
	rollingAvgs[0] = sum / 30

	for i := 30; i < len(powers); i++ {
		sum = sum - float64(powers[i-30]) + float64(powers[i])
		rollingAvgs[i-29] = sum / 30
	}

	// Raise to 4th power and average
	var sumFourth float64
	for _, avg := range rollingAvgs {
		sumFourth += math.Pow(avg, 4)
	}
	avgFourth := sumFourth / float64(len(rollingAvgs))

	// Take 4th root (Round, not truncate — avoids off-by-one from float precision)
	return int(math.Round(math.Pow(avgFourth, 0.25)))
}

// ComputeTSS calculates Training Stress Score
// TSS = (duration_seconds * NP * IF) / (FTP * 3600) * 100
func ComputeTSS(durationSeconds int, normalizedPower int, ftp int) float64 {
	if ftp <= 0 {
		return 0
	}
	intensityFactor := float64(normalizedPower) / float64(ftp)
	return (float64(durationSeconds) * float64(normalizedPower) * intensityFactor) / (float64(ftp) * 3600) * 100
}

// ComputeIntensityFactor calculates Intensity Factor (NP/FTP)
func ComputeIntensityFactor(normalizedPower int, ftp int) float64 {
	if ftp <= 0 {
		return 0
	}
	return float64(normalizedPower) / float64(ftp)
}

// ComputeVariabilityIndex calculates VI (NP/AP)
func ComputeVariabilityIndex(normalizedPower int, averagePower int) float64 {
	if averagePower <= 0 {
		return 0
	}
	return float64(normalizedPower) / float64(averagePower)
}

// ComputePowerCurve calculates best power for standard durations
func ComputePowerCurve(powers []int) map[int]int {
	durations := []int{1, 5, 10, 15, 20, 30, 45, 60, 90, 120, 180, 300, 600, 900, 1200, 1800, 2700, 3600, 5400, 7200}

	curve := make(map[int]int)
	n := len(powers)

	for _, duration := range durations {
		if duration > n {
			break
		}

		maxAvg := 0
		var windowSum int
		for i := 0; i < duration; i++ {
			windowSum += powers[i]
		}
		maxAvg = windowSum

		for i := duration; i < n; i++ {
			windowSum = windowSum - powers[i-duration] + powers[i]
			if windowSum > maxAvg {
				maxAvg = windowSum
			}
		}

		curve[duration] = maxAvg / duration
	}

	return curve
}

// PowerCurveRecord holds per-second data for extended power curve computation
type PowerCurveRecord struct {
	Power               int
	HeartRate           int      // 0 if not available
	Cadence             int      // 0 if not available
	Speed               *float64 // m/s, nil if not available
	Gradient            *float64 // %, nil if not available
	LRBalance           *float64 // %, nil if not available
	TorqueEffectiveness *float64 // %, nil if not available
}

// PowerCurveResult holds the result for a single duration
type PowerCurveResult struct {
	BestPower              int
	AvgHeartRate           *int
	AvgSpeed               *float64
	AvgGradient            *float64
	AvgCadence             *int
	AvgLRBalance           *float64
	AvgTorqueEffectiveness *float64
}

// ComputePowerCurveExtended calculates best power for standard durations
// and also computes avg metrics over the best window.
func ComputePowerCurveExtended(records []PowerCurveRecord) map[int]PowerCurveResult {
	durations := []int{1, 5, 10, 15, 20, 30, 45, 60, 90, 120, 180, 300, 600, 900, 1200, 1800, 2700, 3600, 5400, 7200}

	results := make(map[int]PowerCurveResult)
	n := len(records)

	for _, duration := range durations {
		if duration > n {
			break
		}

		var windowSum int
		for i := 0; i < duration; i++ {
			windowSum += records[i].Power
		}
		maxSum := windowSum
		bestStart := 0

		for i := duration; i < n; i++ {
			windowSum = windowSum - records[i-duration].Power + records[i].Power
			if windowSum > maxSum {
				maxSum = windowSum
				bestStart = i - duration + 1
			}
		}

		result := PowerCurveResult{
			BestPower: maxSum / duration,
		}

		// Compute averages over the best window
		var hrSum, hrCount int
		var cadSum, cadCount int
		var speedSum float64
		var speedCount int
		var gradSum float64
		var gradCount int
		var lrSum float64
		var lrCount int
		var teSum float64
		var teCount int

		for j := bestStart; j < bestStart+duration; j++ {
			r := records[j]
			if r.HeartRate > 0 {
				hrSum += r.HeartRate
				hrCount++
			}
			if r.Cadence > 0 {
				cadSum += r.Cadence
				cadCount++
			}
			if r.Speed != nil {
				speedSum += *r.Speed
				speedCount++
			}
			if r.Gradient != nil {
				gradSum += *r.Gradient
				gradCount++
			}
			if r.LRBalance != nil {
				lrSum += *r.LRBalance
				lrCount++
			}
			if r.TorqueEffectiveness != nil {
				teSum += *r.TorqueEffectiveness
				teCount++
			}
		}

		if hrCount > 0 {
			v := hrSum / hrCount
			result.AvgHeartRate = &v
		}
		if cadCount > 0 {
			v := cadSum / cadCount
			result.AvgCadence = &v
		}
		if speedCount > 0 {
			v := speedSum / float64(speedCount)
			result.AvgSpeed = &v
		}
		if gradCount > 0 {
			v := gradSum / float64(gradCount)
			result.AvgGradient = &v
		}
		if lrCount > 0 {
			v := lrSum / float64(lrCount)
			result.AvgLRBalance = &v
		}
		if teCount > 0 {
			v := teSum / float64(teCount)
			result.AvgTorqueEffectiveness = &v
		}

		results[duration] = result
	}

	return results
}

// ComputeAverage calculates the average of non-zero values
func ComputeAverage(values []int) int {
	if len(values) == 0 {
		return 0
	}
	var sum, count int
	for _, v := range values {
		if v > 0 {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

// ComputeMax returns the maximum value
func ComputeMax(values []int) int {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// ComputeAverageFloat calculates the average of non-zero float values
func ComputeAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	var count int
	for _, v := range values {
		if v > 0 {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// ComputeMaxFloat returns the maximum float value
func ComputeMaxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// ComputeElevationGain calculates total ascent from altitude data
func ComputeElevationGain(altitudes []float64) float64 {
	if len(altitudes) < 2 {
		return 0
	}

	var gain float64
	for i := 1; i < len(altitudes); i++ {
		diff := altitudes[i] - altitudes[i-1]
		if diff > 0 {
			gain += diff
		}
	}
	return gain
}

// ComputeGradient calculates gradient between two points
func ComputeGradient(distanceDelta, altitudeDelta float64) float64 {
	if distanceDelta <= 0 {
		return 0
	}
	return (altitudeDelta / distanceDelta) * 100
}

// ComputeZoneDistribution calculates time in each zone
func ComputeZoneDistribution(values []int, zones []ZoneBoundary) map[int]int {
	distribution := make(map[int]int)
	for _, v := range values {
		for _, z := range zones {
			if v >= z.Min && (z.Max == 0 || v < z.Max) {
				distribution[z.Zone]++
				break
			}
		}
	}
	return distribution
}

// ZoneBoundary defines a power or HR zone
type ZoneBoundary struct {
	Zone int
	Min  int
	Max  int // 0 means no upper limit
}

// DefaultPowerZones returns default power zones based on FTP
func DefaultPowerZones(ftp int) []ZoneBoundary {
	return []ZoneBoundary{
		{1, 0, int(float64(ftp) * 0.55)},
		{2, int(float64(ftp) * 0.55), int(float64(ftp) * 0.75)},
		{3, int(float64(ftp) * 0.75), int(float64(ftp) * 0.90)},
		{4, int(float64(ftp) * 0.90), int(float64(ftp) * 1.05)},
		{5, int(float64(ftp) * 1.05), int(float64(ftp) * 1.20)},
		{6, int(float64(ftp) * 1.20), int(float64(ftp) * 1.50)},
		{7, int(float64(ftp) * 1.50), 0},
	}
}

// DefaultHRZones returns default HR zones based on max HR
func DefaultHRZones(maxHR int) []ZoneBoundary {
	return []ZoneBoundary{
		{1, 0, int(float64(maxHR) * 0.60)},
		{2, int(float64(maxHR) * 0.60), int(float64(maxHR) * 0.70)},
		{3, int(float64(maxHR) * 0.70), int(float64(maxHR) * 0.80)},
		{4, int(float64(maxHR) * 0.80), int(float64(maxHR) * 0.90)},
		{5, int(float64(maxHR) * 0.90), 0},
	}
}

// MedianFilter applies a median filter to smooth data
func MedianFilter(values []int, windowSize int) []int {
	if len(values) == 0 || windowSize <= 1 {
		return values
	}

	result := make([]int, len(values))
	halfWindow := windowSize / 2

	for i := range values {
		start := i - halfWindow
		if start < 0 {
			start = 0
		}
		end := i + halfWindow + 1
		if end > len(values) {
			end = len(values)
		}

		window := make([]int, end-start)
		copy(window, values[start:end])
		sort.Ints(window)
		result[i] = window[len(window)/2]
	}

	return result
}
