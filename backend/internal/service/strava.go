package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/velometric/backend/internal/config"
	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/repository"
)

// StravaActivitySummary is the summary object returned by the Strava API
type StravaActivitySummary struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	StartDate      time.Time `json:"start_date"`
	StartDateLocal time.Time `json:"start_date_local"`
	Distance       float64   `json:"distance"`
	Private        bool      `json:"private"`
}

const (
	stravaAPIBaseURL = "https://www.strava.com/api/v3"
	matchTimeWindow  = 30   // seconds
	matchDistancePct = 0.01 // 1%
	rateLimitDelay   = 50 * time.Millisecond
	processBatchSize = 50
	stravaPageSize   = 200
)

var ErrJobNotFound = errors.New("strava sync job not found")

// StravaService handles Strava API interactions
type StravaService struct {
	accessToken  string
	stravaRepo   *repository.StravaRepository
	activityRepo *repository.ActivityRepository
	httpClient   *http.Client
}

// NewStravaService creates a new Strava service
func NewStravaService(cfg *config.Config, pool *pgxpool.Pool) *StravaService {
	return &StravaService{
		accessToken:  cfg.StravaAccessToken,
		stravaRepo:   repository.NewStravaRepository(pool),
		activityRepo: repository.NewActivityRepository(pool),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartSync creates a job record, persists it, and launches the background goroutine.
// Returns the job immediately (status PENDING).
// offset: number of activities to skip (Strava returns newest-first, so offset=0 = most recent).
// limit: max activities to fetch (0 = all).
func (s *StravaService) StartSync(ctx context.Context, userID uuid.UUID, offset, limit int) (*model.StravaSyncJob, error) {
	if s.accessToken == "" {
		return nil, fmt.Errorf("STRAVA_ACCESS_TOKEN not configured")
	}

	now := time.Now()
	job := &model.StravaSyncJob{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      model.JobStatusPending,
		LimitCount:  limit,
		OffsetCount: offset,
		StartedAt:   now,
		CreatedAt:   now,
	}

	if err := s.stravaRepo.CreateJob(context.Background(), job); err != nil {
		return nil, fmt.Errorf("failed to create sync job: %w", err)
	}

	log.Printf("[strava-sync][job=%s] job created for user=%s offset=%d limit=%d", job.ID, userID, offset, limit)

	go s.runJob(job)

	return job, nil
}

// GetJob returns the current state of a sync job.
func (s *StravaService) GetJob(ctx context.Context, jobID uuid.UUID) (*model.StravaSyncJob, error) {
	job, err := s.stravaRepo.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJobNotFound, err)
	}
	return job, nil
}

// ReprocessSync skips the fetch phase and re-runs only the process phase.
// Valid for any job that already has strava_activities data (DATA_FETCHED, DATA_PROCESSED, PROCESSING_FAILED).
func (s *StravaService) ReprocessSync(ctx context.Context, jobID uuid.UUID) (*model.StravaSyncJob, error) {
	job, err := s.stravaRepo.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJobNotFound, err)
	}

	switch job.Status {
	case model.JobStatusDataFetched, model.JobStatusDataProcessed, model.JobStatusProcessingFailed:
		// ok
	default:
		return nil, fmt.Errorf("job must have completed fetch phase to reprocess (current: %s)", job.Status)
	}

	log.Printf("[strava-sync][job=%s] reprocess requested (status=%s)", jobID, job.Status)

	go func() {
		bgCtx := context.Background()
		if err := s.runProcessPhase(bgCtx, job); err != nil {
			log.Printf("[strava-sync][job=%s] reprocess failed: %v", job.ID, err)
		}
	}()

	return job, nil
}

// RetrySync re-launches a failed job from the appropriate phase.
func (s *StravaService) RetrySync(ctx context.Context, jobID uuid.UUID) (*model.StravaSyncJob, error) {
	job, err := s.stravaRepo.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJobNotFound, err)
	}

	if job.Status != model.JobStatusFetchingFailed && job.Status != model.JobStatusProcessingFailed {
		return nil, fmt.Errorf("job is not in a failed state (current: %s)", job.Status)
	}

	log.Printf("[strava-sync][job=%s] retrying from status=%s", jobID, job.Status)

	go s.runJob(job)

	return job, nil
}

// runJob executes fetch and process phases sequentially in a background goroutine.
func (s *StravaService) runJob(job *model.StravaSyncJob) {
	ctx := context.Background()

	// If retrying a processing failure, skip re-fetch.
	if job.Status != model.JobStatusProcessingFailed {
		if err := s.runFetchPhase(ctx, job); err != nil {
			log.Printf("[strava-sync][job=%s] fetch phase failed: %v", job.ID, err)
			return
		}
	}

	if err := s.runProcessPhase(ctx, job); err != nil {
		log.Printf("[strava-sync][job=%s] process phase failed: %v", job.ID, err)
	}
}

// runFetchPhase fetches activities from Strava and upserts them to strava_activities.
func (s *StravaService) runFetchPhase(ctx context.Context, job *model.StravaSyncJob) error {
	if err := s.stravaRepo.SetJobFetching(ctx, job.ID); err != nil {
		return fmt.Errorf("failed to set FETCHING: %w", err)
	}
	log.Printf("[strava-sync][job=%s] fetch phase started", job.ID)

	stravaActivities, err := s.fetchActivities(ctx, job.OffsetCount, job.LimitCount)
	if err != nil {
		msg := fmt.Sprintf("failed to fetch from Strava: %v", err)
		_ = s.stravaRepo.SetJobFetchFailed(ctx, job.ID, msg)
		return fmt.Errorf("%s", msg)
	}

	log.Printf("[strava-sync][job=%s] fetched %d activities from Strava (offset=%d limit=%d)",
		job.ID, len(stravaActivities), job.OffsetCount, job.LimitCount)

	for _, sa := range stravaActivities {
		stravaModel := &model.StravaActivity{
			ID:           uuid.New(),
			UserID:       job.UserID,
			StravaID:     sa.ID,
			Title:        &sa.Name,
			ActivityType: &sa.Type,
			StartTime:    sa.StartDate, // UTC — matches how FIT files store start_time
			Distance:     &sa.Distance,
			IsPrivate:    sa.Private,
			IsFlagged:    false,
			SyncedAt:     time.Now(),
		}
		if data, err := json.Marshal(sa); err == nil {
			json.Unmarshal(data, &stravaModel.RawJSON)
		}

		if err := s.stravaRepo.Upsert(ctx, stravaModel, job.ID); err != nil {
			log.Printf("[strava-sync][job=%s] upsert failed for strava_id=%d: %v", job.ID, sa.ID, err)
		}

		time.Sleep(rateLimitDelay)
	}

	if err := s.stravaRepo.SetJobDataFetched(ctx, job.ID, len(stravaActivities)); err != nil {
		return fmt.Errorf("failed to set DATA_FETCHED: %w", err)
	}
	log.Printf("[strava-sync][job=%s] fetch phase complete, %d activities cached", job.ID, len(stravaActivities))
	return nil
}

// runProcessPhase reads cached strava_activities for this job, splits into batches of
// processBatchSize, processes each batch in parallel, then aggregates results.
func (s *StravaService) runProcessPhase(ctx context.Context, job *model.StravaSyncJob) error {
	if err := s.stravaRepo.SetJobProcessing(ctx, job.ID); err != nil {
		return fmt.Errorf("failed to set PROCESSING: %w", err)
	}
	log.Printf("[strava-sync][job=%s] process phase started", job.ID)

	cached, err := s.stravaRepo.GetStravaActivitiesByJob(ctx, job.ID)
	if err != nil {
		msg := fmt.Sprintf("failed to load cached strava activities: %v", err)
		_ = s.stravaRepo.SetJobProcessFailed(ctx, job.ID, msg)
		return fmt.Errorf("%s", msg)
	}

	if len(cached) == 0 {
		_ = s.stravaRepo.SetJobDataProcessed(ctx, job.ID, 0, 0)
		log.Printf("[strava-sync][job=%s] no cached activities to process", job.ID)
		return nil
	}

	// Fetch local activities once for the full time range — shared across all batches (read-only).
	earliest, latest := cached[0].StartTime, cached[0].StartTime
	for _, a := range cached {
		if a.StartTime.Before(earliest) {
			earliest = a.StartTime
		}
		if a.StartTime.After(latest) {
			latest = a.StartTime
		}
	}
	timeWindow := time.Duration(matchTimeWindow*2) * time.Second
	localActivities, err := s.activityRepo.FindByTimeRange(ctx, job.UserID,
		earliest.Add(-timeWindow), latest.Add(timeWindow))
	if err != nil {
		msg := fmt.Sprintf("failed to fetch local activities: %v", err)
		_ = s.stravaRepo.SetJobProcessFailed(ctx, job.ID, msg)
		return fmt.Errorf("%s", msg)
	}

	// Split cached into batches of processBatchSize and process in parallel.
	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		totalUpdated int
		totalCreated int
		firstErr     error
	)

	for i := 0; i < len(cached); i += processBatchSize {
		end := i + processBatchSize
		if end > len(cached) {
			end = len(cached)
		}
		batch := cached[i:end]
		batchNum := i/processBatchSize + 1

		wg.Add(1)
		go func(b []*model.StravaActivity, n int) {
			defer wg.Done()
			updated, created, err := s.processBatch(ctx, job, b, localActivities, n)
			mu.Lock()
			totalUpdated += updated
			totalCreated += created
			if err != nil && firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}(batch, batchNum)
	}

	wg.Wait()

	if firstErr != nil {
		_ = s.stravaRepo.SetJobProcessFailed(ctx, job.ID, firstErr.Error())
		return firstErr
	}

	if err := s.stravaRepo.SetJobDataProcessed(ctx, job.ID, totalUpdated, totalCreated); err != nil {
		return fmt.Errorf("failed to set DATA_PROCESSED: %w", err)
	}
	log.Printf("[strava-sync][job=%s] process phase complete: updated=%d created=%d", job.ID, totalUpdated, totalCreated)
	return nil
}

// processBatch matches a slice of cached Strava activities against local activities.
func (s *StravaService) processBatch(ctx context.Context, job *model.StravaSyncJob, batch []*model.StravaActivity, localActivities []*model.Activity, batchNum int) (updated, created int, err error) {
	log.Printf("[strava-sync][job=%s] batch=%d processing %d activities", job.ID, batchNum, len(batch))
	for _, sa := range batch {
		summary := StravaActivitySummary{
			ID:        sa.StravaID,
			StartDate: sa.StartTime, // stored as UTC
		}
		if sa.Distance != nil {
			summary.Distance = *sa.Distance
		}
		if sa.ActivityType != nil {
			summary.Type = *sa.ActivityType
		}
		if sa.Title != nil {
			summary.Name = *sa.Title
		}

		match := s.findMatch(summary, localActivities)
		if match != nil {
			sport := mapStravaType(summary.Type)
			if e := s.activityRepo.UpdateActivity(ctx, match.ID, summary.Name, sport); e == nil {
				updated++
				log.Printf("[strava-sync][job=%s] batch=%d matched strava_id=%d → local=%s", job.ID, batchNum, sa.StravaID, match.ID)
			}
		} else {
			candidates := s.findCandidates(summary, localActivities)
			if len(candidates) > 0 {
				log.Printf("[strava-sync][job=%s] batch=%d candidate for strava_id=%d (%d options)", job.ID, batchNum, sa.StravaID, len(candidates))
			} else {
				created++
			}
		}
	}
	return updated, created, nil
}

// fetchActivities retrieves activities from Strava applying offset and limit.
// Strava returns activities newest-first, so offset=0 / limit=10 = 10 most recent.
// offset and limit are applied over the full ordered result set, not Strava pages.
func (s *StravaService) fetchActivities(ctx context.Context, offset, limit int) ([]StravaActivitySummary, error) {
	var allActivities []StravaActivitySummary
	page := 1
	need := offset + limit // stop fetching once we have this many (0 = fetch all)

	for {
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/athlete/activities?page=%d&per_page=%d",
			stravaAPIBaseURL, page, stravaPageSize), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Strava API returned status %d", resp.StatusCode)
		}

		var activities []StravaActivitySummary
		if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
			return nil, err
		}

		if len(activities) == 0 {
			break
		}

		allActivities = append(allActivities, activities...)

		// Stop early if we have enough to cover offset+limit.
		if need > 0 && len(allActivities) >= need {
			break
		}

		if len(activities) < stravaPageSize {
			break // last page
		}

		page++
		if page > 10 { // hard cap: 2000 activities
			break
		}
	}

	// Apply offset.
	if offset >= len(allActivities) {
		return nil, nil
	}
	allActivities = allActivities[offset:]

	// Apply limit.
	if limit > 0 && len(allActivities) > limit {
		allActivities = allActivities[:limit]
	}

	return allActivities, nil
}

// findMatch looks for a local activity matching the Strava activity
func (s *StravaService) findMatch(sa StravaActivitySummary, localActivities []*model.Activity) *model.Activity {
	for _, la := range localActivities {
		timeDiff := math.Abs(la.StartTime.Sub(sa.StartDate).Seconds())
		if timeDiff > float64(matchTimeWindow) {
			continue
		}
		if la.Distance == 0 {
			continue
		}
		distDiff := math.Abs(la.Distance-sa.Distance) / la.Distance
		if distDiff > matchDistancePct {
			continue
		}
		return la
	}
	return nil
}

// findCandidates finds potential matches that are outside strict matching criteria
func (s *StravaService) findCandidates(sa StravaActivitySummary, localActivities []*model.Activity) []model.StravaMatchCandidate {
	var candidates []model.StravaMatchCandidate

	for _, la := range localActivities {
		timeDiff := math.Abs(la.StartTime.Sub(sa.StartDate).Seconds())
		if timeDiff > float64(matchTimeWindow*10) {
			continue
		}
		if la.Distance == 0 {
			continue
		}
		distDiff := math.Abs(la.Distance-sa.Distance) / la.Distance
		if distDiff > matchDistancePct*10 {
			continue
		}
		candidates = append(candidates, model.StravaMatchCandidate{
			StravaActivity:  nil,
			LocalActivity:   la,
			TimeDiffSecs:    int64(timeDiff),
			DistanceDiffPct: distDiff,
		})
	}

	return candidates
}

// mapStravaType maps Strava activity types to Velometric sport types
func mapStravaType(stravaType string) string {
	mapping := map[string]string{
		"Run":              "Run",
		"TrailRun":         "Run",
		"VirtualRun":       "Run",
		"Ride":             "Ride",
		"VirtualRide":      "Ride",
		"GravelRide":       "Ride",
		"MountainBikeRide": "Ride",
		"EBikeRide":        "Ride",
		"Hike":             "Hike",
		"Walk":             "Walk",
		"Swim":             "Swim",
		"Rowing":           "Rowing",
		"Workout":          "Workout",
		"Yoga":             "Yoga",
	}
	if v, ok := mapping[stravaType]; ok {
		return v
	}
	return "Other"
}

// HasToken checks if Strava token is configured
func (s *StravaService) HasToken() bool {
	return s.accessToken != ""
}

// IsStravaConfigured returns whether Strava is configured
func IsStravaConfigured(cfg *config.Config) bool {
	return cfg.StravaAccessToken != ""
}
