package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/velometric/backend/internal/model"
)

// svc is a zero-value StravaService; findMatch/findCandidates/mapStravaType
// don't touch the repo or http fields so nil is fine.
var svc = &StravaService{}

// baseTime is an arbitrary fixed instant for test cases.
var baseTime = time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC)

// makeLocalActivity is a helper that returns a *model.Activity with the
// provided start time and distance (meters).
func makeLocalActivity(startTime time.Time, distanceM float64) *model.Activity {
	return &model.Activity{
		ID:        uuid.New(),
		StartTime: startTime,
		Distance:  distanceM,
	}
}

// makeStravaSummary returns a StravaActivitySummary for the given start time
// and distance (meters).
func makeStravaSummary(startLocal time.Time, distanceM float64) StravaActivitySummary {
	return StravaActivitySummary{
		ID:        1,
		Name:      "Morning Ride",
		Type:      "Ride",
		StartDate: startLocal,
		Distance:  distanceM,
	}
}

// ── mapStravaType ─────────────────────────────────────────────────────────────

func TestMapStravaType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Known ride variants
		{"Ride", "Ride"},
		{"VirtualRide", "Ride"},
		{"GravelRide", "Ride"},
		{"MountainBikeRide", "Ride"},
		{"EBikeRide", "Ride"},

		// Known run variants
		{"Run", "Run"},
		{"TrailRun", "Run"},
		{"VirtualRun", "Run"},

		// Other known types
		{"Hike", "Hike"},
		{"Walk", "Walk"},
		{"Swim", "Swim"},
		{"Rowing", "Rowing"},
		{"Workout", "Workout"},
		{"Yoga", "Yoga"},

		// Unknown / empty → "Other"
		{"", "Other"},
		{"Crossfit", "Other"},
		{"WeightTraining", "Other"},
		{"ride", "Other"}, // case-sensitive — lowercase doesn't match
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapStravaType(tt.input)
			if got != tt.expected {
				t.Errorf("mapStravaType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ── findMatch ─────────────────────────────────────────────────────────────────

func TestFindMatch_ExactMatch(t *testing.T) {
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime, 30000)

	got := svc.findMatch(sa, []*model.Activity{la})
	if got == nil {
		t.Fatal("expected a match, got nil")
	}
	if got.ID != la.ID {
		t.Errorf("matched wrong activity: got %v, want %v", got.ID, la.ID)
	}
}

func TestFindMatch_WithinTimeWindow(t *testing.T) {
	// 150 seconds apart is the exact boundary — must still match (condition is >).
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime.Add(150*time.Second), 30000)

	got := svc.findMatch(sa, []*model.Activity{la})
	if got == nil {
		t.Error("150s diff should match (boundary is exclusive)")
	}
}

func TestFindMatch_OutsideTimeWindow(t *testing.T) {
	// 151 seconds apart → no match.
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime.Add(151*time.Second), 30000)

	got := svc.findMatch(sa, []*model.Activity{la})
	if got != nil {
		t.Error("151s diff should not match")
	}
}

func TestFindMatch_WithinDistanceTolerance(t *testing.T) {
	// 1% distance diff is the exact boundary — must still match.
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime, 30000*1.01) // exactly +1%

	got := svc.findMatch(sa, []*model.Activity{la})
	if got == nil {
		t.Error("1% distance diff should match (boundary is exclusive)")
	}
}

func TestFindMatch_OutsideDistanceTolerance(t *testing.T) {
	// >1% distance diff → no match.
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime, 30000*1.02) // 2% diff

	got := svc.findMatch(sa, []*model.Activity{la})
	if got != nil {
		t.Error("2% distance diff should not match")
	}
}

func TestFindMatch_ZeroLocalDistance_Skipped(t *testing.T) {
	// Local activity with Distance=0 must be skipped to avoid divide-by-zero.
	la := makeLocalActivity(baseTime, 0)
	sa := makeStravaSummary(baseTime, 30000)

	got := svc.findMatch(sa, []*model.Activity{la})
	if got != nil {
		t.Error("activity with distance=0 should be skipped")
	}
}

func TestFindMatch_EmptyLocals(t *testing.T) {
	sa := makeStravaSummary(baseTime, 30000)
	got := svc.findMatch(sa, []*model.Activity{})
	if got != nil {
		t.Error("empty locals should return nil")
	}
}

func TestFindMatch_MultipleLocals_PicksCorrectOne(t *testing.T) {
	// First local is a mismatch (wrong distance); second is a good match.
	noMatch := makeLocalActivity(baseTime, 999) // completely different distance
	goodMatch := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime, 30000)

	got := svc.findMatch(sa, []*model.Activity{noMatch, goodMatch})
	if got == nil {
		t.Fatal("expected a match, got nil")
	}
	if got.ID != goodMatch.ID {
		t.Errorf("matched wrong activity: got %v, want %v", got.ID, goodMatch.ID)
	}
}

// ── findCandidates ────────────────────────────────────────────────────────────
//
// Candidate window: ≤300 s time diff  AND  ≤10% distance diff.

func TestFindCandidates_EmptyLocals(t *testing.T) {
	sa := makeStravaSummary(baseTime, 30000)
	got := svc.findCandidates(sa, []*model.Activity{})
	if len(got) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(got))
	}
}

func TestFindCandidates_WithinCandidateWindow(t *testing.T) {
	// 5 minutes (300 s) time diff, 10% distance diff → candidate.
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime.Add(300*time.Second), 30000*1.10)

	got := svc.findCandidates(sa, []*model.Activity{la})
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	if got[0].LocalActivity.ID != la.ID {
		t.Errorf("candidate points to wrong activity")
	}
}

func TestFindCandidates_OutsideTimeWindow(t *testing.T) {
	// 1501 seconds → outside candidate window (matchTimeWindow * 10 = 1500s).
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime.Add(1501*time.Second), 30000)

	got := svc.findCandidates(sa, []*model.Activity{la})
	if len(got) != 0 {
		t.Errorf("expected 0 candidates for 1501s diff, got %d", len(got))
	}
}

func TestFindCandidates_OutsideDistanceWindow(t *testing.T) {
	// 11% distance diff → outside candidate window.
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime, 30000*1.11)

	got := svc.findCandidates(sa, []*model.Activity{la})
	if len(got) != 0 {
		t.Errorf("expected 0 candidates for 11%% distance diff, got %d", len(got))
	}
}

func TestFindCandidates_ZeroLocalDistance_Skipped(t *testing.T) {
	la := makeLocalActivity(baseTime, 0)
	sa := makeStravaSummary(baseTime, 30000)

	got := svc.findCandidates(sa, []*model.Activity{la})
	if len(got) != 0 {
		t.Error("activity with distance=0 should be skipped in candidate search")
	}
}

func TestFindCandidates_MultipleCandidates(t *testing.T) {
	// Two locals both within candidate window — both should appear.
	la1 := makeLocalActivity(baseTime.Add(1*time.Minute), 29500)
	la2 := makeLocalActivity(baseTime.Add(2*time.Minute), 30500)
	sa := makeStravaSummary(baseTime, 30000)

	got := svc.findCandidates(sa, []*model.Activity{la1, la2})
	if len(got) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(got))
	}
}

func TestFindCandidates_PopulatesTimeDiffAndDistanceDiff(t *testing.T) {
	//this is a comment for testing
	timeDiff := 60 * time.Second
	// 5% distance difference
	la := makeLocalActivity(baseTime, 30000)
	sa := makeStravaSummary(baseTime.Add(timeDiff), 30000*1.05)

	got := svc.findCandidates(sa, []*model.Activity{la})
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	if got[0].TimeDiffSecs != int64(timeDiff.Seconds()) {
		t.Errorf("TimeDiffSecs = %d, want %d", got[0].TimeDiffSecs, int64(timeDiff.Seconds()))
	}
	// Distance diff should be approximately 0.05
	const wantDistDiff = 0.05
	const epsilon = 0.001
	if d := got[0].DistanceDiffPct; d < wantDistDiff-epsilon || d > wantDistDiff+epsilon {
		t.Errorf("DistanceDiffPct = %.4f, want ~%.2f", d, wantDistDiff)
	}
}
