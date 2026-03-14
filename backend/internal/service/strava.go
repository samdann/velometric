package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
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
func (s *StravaService) StartSync(ctx context.Context, userID uuid.UUID, limit int) (*model.StravaSyncJob, error) {
	if s.accessToken == "" {
		return nil, fmt.Errorf("STRAVA_ACCESS_TOKEN not configured")
	}

	now := time.Now()
	job := &model.StravaSyncJob{
		ID:         uuid.New(),
		UserID:     userID,
		Status:     model.JobStatusPending,
		LimitCount: limit,
		StartedAt:  now,
		CreatedAt:  now,
	}

	if err := s.stravaRepo.CreateJob(context.Background(), job); err != nil {
		return nil, fmt.Errorf("failed to create sync job: %w", err)
	}

	log.Printf("[strava-sync][job=%s] job created for user=%s limit=%d", job.ID, userID, limit)

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

	stravaActivities, err := s.fetchActivities(ctx)
	if err != nil {
		msg := fmt.Sprintf("failed to fetch from Strava: %v", err)
		_ = s.stravaRepo.SetJobFetchFailed(ctx, job.ID, msg)
		return fmt.Errorf("%s", msg)
	}

	if job.LimitCount > 0 && len(stravaActivities) > job.LimitCount {
		stravaActivities = stravaActivities[:job.LimitCount]
	}

	log.Printf("[strava-sync][job=%s] fetched %d activities from Strava", job.ID, len(stravaActivities))

	for _, sa := range stravaActivities {
		stravaModel := &model.StravaActivity{
			ID:           uuid.New(),
			UserID:       job.UserID,
			StravaID:     sa.ID,
			Title:        &sa.Name,
			ActivityType: &sa.Type,
			StartTime:    sa.StartDateLocal,
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

// runProcessPhase reads cached strava_activities for this job and matches them to local activities.
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

	// Build time bounds for local activity lookup
	earliest := cached[0].StartTime
	latest := cached[0].StartTime
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
		earliest.Add(-timeWindow),
		latest.Add(timeWindow))
	if err != nil {
		msg := fmt.Sprintf("failed to fetch local activities: %v", err)
		_ = s.stravaRepo.SetJobProcessFailed(ctx, job.ID, msg)
		return fmt.Errorf("%s", msg)
	}

	updatedCount := 0
	createdCount := 0

	for _, sa := range cached {
		summary := StravaActivitySummary{
			ID:             sa.StravaID,
			StartDateLocal: sa.StartTime,
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
			if err := s.activityRepo.UpdateActivity(ctx, match.ID, summary.Name, sport); err == nil {
				updatedCount++
				log.Printf("[strava-sync][job=%s] matched strava_id=%d → local=%s", job.ID, sa.StravaID, match.ID)
			}
		} else {
			candidates := s.findCandidates(summary, localActivities)
			if len(candidates) > 0 {
				log.Printf("[strava-sync][job=%s] candidate found for strava_id=%d (%d options)", job.ID, sa.StravaID, len(candidates))
			} else {
				createdCount++
			}
		}
	}

	if err := s.stravaRepo.SetJobDataProcessed(ctx, job.ID, updatedCount, createdCount); err != nil {
		return fmt.Errorf("failed to set DATA_PROCESSED: %w", err)
	}
	log.Printf("[strava-sync][job=%s] process phase complete: updated=%d created=%d", job.ID, updatedCount, createdCount)
	return nil
}

// fetchActivities retrieves all activities from Strava (paginated)
func (s *StravaService) fetchActivities(ctx context.Context) ([]StravaActivitySummary, error) {
	var allActivities []StravaActivitySummary
	page := 1
	perPage := 200

	for {
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/athlete/activities?page=%d&per_page=%d",
			stravaAPIBaseURL, page, perPage), nil)
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
		page++

		if page > 10 { // 2000 activities max
			break
		}
	}

	return allActivities, nil
}

// findMatch looks for a local activity matching the Strava activity
func (s *StravaService) findMatch(sa StravaActivitySummary, localActivities []*model.Activity) *model.Activity {
	for _, la := range localActivities {
		timeDiff := math.Abs(la.StartTime.Sub(sa.StartDateLocal).Seconds())
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
		timeDiff := math.Abs(la.StartTime.Sub(sa.StartDateLocal).Seconds())
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
