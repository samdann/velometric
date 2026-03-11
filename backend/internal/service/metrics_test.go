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
