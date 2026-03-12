package fitparser

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"testing"
)

// activityFixture loads the real FIT file from testdata/.
// All tests that need a valid FIT file use this helper.
func activityFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/activity.fit")
	if err != nil {
		t.Fatalf("failed to read testdata/activity.fit: %v", err)
	}
	return data
}

// ── error cases (synthetic inputs) ───────────────────────────────────────────

func TestParse_EmptyInput(t *testing.T) {
	_, err := Parse(bytes.NewReader([]byte{}))
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestParse_TooSmall(t *testing.T) {
	// FIT header is at least 14 bytes; anything shorter must be rejected.
	_, err := Parse(bytes.NewReader(make([]byte, 10)))
	if err == nil {
		t.Fatal("expected error for input < 14 bytes, got nil")
	}
}

func TestParse_InvalidFITSignature(t *testing.T) {
	// Exactly 14 bytes but bytes[8:12] are not ".FIT".
	data := make([]byte, 14)
	copy(data[8:12], "NOPE")
	_, err := Parse(bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected error for invalid FIT signature, got nil")
	}
}

func TestParse_RandomBytes(t *testing.T) {
	// Arbitrary binary that isn't a ZIP or valid FIT.
	garbage := []byte("this is definitely not a FIT file at all, just garbage bytes!!!!")
	_, err := Parse(bytes.NewReader(garbage))
	if err == nil {
		t.Fatal("expected error for random bytes, got nil")
	}
}

func TestParse_ZipWithNoFITFile(t *testing.T) {
	// ZIP archive containing a non-.fit file — should fail with "no .fit file found".
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, _ := zw.Create("readme.txt")
	f.Write([]byte("hello"))
	zw.Close()

	_, err := Parse(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("expected error for ZIP with no .fit entry, got nil")
	}
}

// ── valid activity FIT file ───────────────────────────────────────────────────

func TestParse_ValidActivity(t *testing.T) {
	data := activityFixture(t)
	result, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil result")
	}

	// Must have a non-zero start time
	if result.StartTime.IsZero() {
		t.Error("StartTime is zero")
	}

	// Sport must be set (defaults to "cycling")
	if result.Sport == "" {
		t.Error("Sport is empty")
	}

	// Must have at least one record
	if len(result.Records) == 0 {
		t.Error("Records slice is empty")
	}

	// Must have at least one lap
	if len(result.Laps) == 0 {
		t.Error("Laps slice is empty")
	}
}

func TestParse_ValidActivity_RecordTimestamps(t *testing.T) {
	data := activityFixture(t)
	result, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Timestamps must be monotonically non-decreasing
	for i := 1; i < len(result.Records); i++ {
		if result.Records[i].Timestamp.Before(result.Records[i-1].Timestamp) {
			t.Errorf("record[%d] timestamp (%v) is before record[%d] (%v)",
				i, result.Records[i].Timestamp, i-1, result.Records[i-1].Timestamp)
		}
	}
}

func TestParse_ValidActivity_NoNilSlices(t *testing.T) {
	// Records, Laps, Events must be initialised slices (not nil) — safe to range over.
	data := activityFixture(t)
	result, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Records == nil {
		t.Error("Records is nil, expected empty slice")
	}
	if result.Laps == nil {
		t.Error("Laps is nil, expected empty slice")
	}
	if result.Events == nil {
		t.Error("Events is nil, expected empty slice")
	}
}

// ── ZIP-wrapped FIT ───────────────────────────────────────────────────────────

func TestParse_ZipWrappedActivity(t *testing.T) {
	fitData := activityFixture(t)

	// Build an in-memory ZIP containing the FIT bytes.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entry, err := zw.Create("activity.fit")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := entry.Write(fitData); err != nil {
		t.Fatalf("failed to write fit data into zip: %v", err)
	}
	zw.Close()

	result, err := Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Parse() on ZIP-wrapped FIT returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil for ZIP-wrapped FIT")
	}
	if result.StartTime.IsZero() {
		t.Error("StartTime is zero after ZIP unwrap")
	}
	if len(result.Records) == 0 {
		t.Error("Records is empty after ZIP unwrap")
	}
}

// ── ErrNotActivity sentinel ───────────────────────────────────────────────────

func TestErrNotActivity_IsDistinct(t *testing.T) {
	// Verify ErrNotActivity is exported and non-nil so callers can use errors.Is.
	if ErrNotActivity == nil {
		t.Fatal("ErrNotActivity should not be nil")
	}
	if !errors.Is(ErrNotActivity, ErrNotActivity) {
		t.Fatal("errors.Is(ErrNotActivity, ErrNotActivity) should be true")
	}
}
