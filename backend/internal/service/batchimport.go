package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/fitparser"
	"github.com/velometric/backend/internal/repository"
)

// ── Job state ────────────────────────────────────────────────────────────────

type JobStatus string

const (
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
)

type ImportError struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

type ImportReport struct {
	TotalFiles       int            `json:"totalFiles"`
	Success          int            `json:"success"`
	Duplicates       int            `json:"duplicates"`
	SkippedType      int            `json:"skippedType"`
	SkippedDateRange int            `json:"skippedDateRange"`
	SkippedNotAct    int            `json:"skippedNotActivity"`
	Errors           int            `json:"errors"`
	ErrorsByMessage  map[string]int `json:"errorsByMessage"`
	ErrorDetails     []ImportError  `json:"errorDetails"`
	DuplicateFiles   []string       `json:"duplicateFiles"`
	SportBreakdown   map[string]int `json:"sportBreakdown"`
}

type ImportProgress struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
}

type ImportJob struct {
	ID          uuid.UUID      `json:"id"`
	Status      JobStatus      `json:"status"`
	StartedAt   time.Time      `json:"startedAt"`
	CompletedAt *time.Time     `json:"completedAt,omitempty"`
	Progress    ImportProgress `json:"progress"`
	Report      *ImportReport  `json:"report,omitempty"`
	mu          sync.RWMutex
}

func (j *ImportJob) snapshot() ImportJob {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return ImportJob{
		ID:          j.ID,
		Status:      j.Status,
		StartedAt:   j.StartedAt,
		CompletedAt: j.CompletedAt,
		Progress:    j.Progress,
		Report:      j.Report,
	}
}

// ── Request ───────────────────────────────────────────────────────────────────

// ImportRequest configures a batch import run.
type ImportRequest struct {
	// Types filters by sport (case-insensitive). Empty = all activity types.
	// Valid values: "cycling", "running", "swimming", etc.
	Types []string `json:"types"`
	// From / To filter by activity start time (inclusive). Nil = no bound.
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type BatchImportService struct {
	activityRepo *repository.ActivityRepository
	fitDir       string

	mu   sync.RWMutex
	jobs map[uuid.UUID]*ImportJob
}

func NewBatchImportService(activityRepo *repository.ActivityRepository, fitDir string) *BatchImportService {
	return &BatchImportService{
		activityRepo: activityRepo,
		fitDir:       fitDir,
		jobs:         make(map[uuid.UUID]*ImportJob),
	}
}

// StartImport enqueues an async import job and returns immediately.
func (s *BatchImportService) StartImport(ctx context.Context, userID uuid.UUID, ftp int, req ImportRequest) (*ImportJob, error) {
	// List .fit files upfront so we can report total count immediately.
	entries, err := os.ReadDir(s.fitDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read FIT directory %q: %w", s.fitDir, err)
	}
	var fitFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".fit") {
			fitFiles = append(fitFiles, filepath.Join(s.fitDir, e.Name()))
		}
	}

	job := &ImportJob{
		ID:        uuid.New(),
		Status:    JobRunning,
		StartedAt: time.Now(),
		Progress:  ImportProgress{Total: len(fitFiles)},
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	// Run in background; use a detached context so the job survives request cancellation.
	go s.run(context.Background(), job, userID, ftp, req, fitFiles)

	return job, nil
}

// GetJob returns a point-in-time snapshot of a job.
func (s *BatchImportService) GetJob(id uuid.UUID) (*ImportJob, bool) {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	snap := job.snapshot()
	return &snap, true
}

// ── Runner ────────────────────────────────────────────────────────────────────

func (s *BatchImportService) run(ctx context.Context, job *ImportJob, userID uuid.UUID, ftp int, req ImportRequest, files []string) {
	report := &ImportReport{
		TotalFiles:      len(files),
		ErrorsByMessage: make(map[string]int),
		SportBreakdown:  make(map[string]int),
	}

	// Normalise type filter to lowercase set for O(1) lookup.
	typeFilter := make(map[string]bool, len(req.Types))
	for _, t := range req.Types {
		typeFilter[strings.ToLower(t)] = true
	}

	for i, path := range files {
		// Update progress counter.
		job.mu.Lock()
		job.Progress.Processed = i + 1
		job.mu.Unlock()

		fname := filepath.Base(path)
		result := s.importOne(ctx, userID, ftp, req, typeFilter, path)

		switch result.outcome {
		case outcomeSuccess:
			report.Success++
			report.SportBreakdown[result.sport]++
		case outcomeDuplicate:
			report.Duplicates++
			report.DuplicateFiles = append(report.DuplicateFiles, fname)
		case outcomeSkippedType:
			report.SkippedType++
		case outcomeSkippedDate:
			report.SkippedDateRange++
		case outcomeSkippedNotActivity:
			report.SkippedNotAct++
		case outcomeError:
			report.Errors++
			report.ErrorsByMessage[result.errMsg]++
			report.ErrorDetails = append(report.ErrorDetails, ImportError{
				File:  fname,
				Error: result.errMsg,
			})
			log.Printf("batch-import error [%s]: %s", fname, result.errMsg)
		}
	}

	now := time.Now()
	job.mu.Lock()
	job.Status = JobCompleted
	job.CompletedAt = &now
	job.Report = report
	job.mu.Unlock()

	log.Printf("batch-import done — success=%d dupes=%d errors=%d skipped(type=%d date=%d notAct=%d)",
		report.Success, report.Duplicates, report.Errors,
		report.SkippedType, report.SkippedDateRange, report.SkippedNotAct)
}

// ── Per-file import ───────────────────────────────────────────────────────────

type outcome int

const (
	outcomeSuccess outcome = iota
	outcomeDuplicate
	outcomeSkippedType
	outcomeSkippedDate
	outcomeSkippedNotActivity
	outcomeError
)

type importResult struct {
	outcome outcome
	sport   string
	errMsg  string
}

func (s *BatchImportService) importOne(
	ctx context.Context,
	userID uuid.UUID,
	ftp int,
	req ImportRequest,
	typeFilter map[string]bool,
	path string,
) importResult {
	f, err := os.Open(path)
	if err != nil {
		return importResult{outcome: outcomeError, errMsg: fmt.Sprintf("open: %v", err)}
	}
	defer f.Close()

	parsed, err := fitparser.Parse(f)
	if err != nil {
		if errors.Is(err, fitparser.ErrNotActivity) {
			return importResult{outcome: outcomeSkippedNotActivity}
		}
		// The tormoder/fit library rejects newer Garmin file types (41, 44, 58…)
		// and files whose first message isn't file_id — all are non-activity files.
		msg := err.Error()
		if strings.Contains(msg, "unknown file type") ||
			strings.Contains(msg, "was not for file_id") {
			return importResult{outcome: outcomeSkippedNotActivity}
		}
		return importResult{outcome: outcomeError, errMsg: fmt.Sprintf("parse: %v", err)}
	}

	sport := strings.ToLower(parsed.Sport)

	// Type filter
	if len(typeFilter) > 0 && !typeFilter[sport] {
		return importResult{outcome: outcomeSkippedType}
	}

	// Date range filter
	if req.From != nil && parsed.StartTime.Before(*req.From) {
		return importResult{outcome: outcomeSkippedDate}
	}
	if req.To != nil && parsed.StartTime.After(*req.To) {
		return importResult{outcome: outcomeSkippedDate}
	}

	activityService := NewActivityService(s.activityRepo)
	_, err = activityService.ProcessFITFile(ctx, userID, parsed, ftp)
	if err != nil {
		if err == repository.ErrDuplicateActivity {
			return importResult{outcome: outcomeDuplicate, sport: sport}
		}
		return importResult{outcome: outcomeError, errMsg: err.Error()}
	}

	return importResult{outcome: outcomeSuccess, sport: sport}
}
