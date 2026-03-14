package service

import (
	"context"
	"encoding/json"
	"fmt"
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

// SyncResult contains the results of a Strava sync
type SyncResult struct {
	UpdatedCount int                          `json:"updatedCount"`
	CreatedCount int                          `json:"createdCount"`
	Candidates   []model.StravaMatchCandidate `json:"candidates"`
	Error        string                       `json:"error,omitempty"`
}

// FetchAndSync fetches activities from Strava and syncs them to local DB.
// If limit > 0, only the first N activities are processed (useful for testing).
func (s *StravaService) FetchAndSync(ctx context.Context, userID uuid.UUID, limit int) (*SyncResult, error) {
	if s.accessToken == "" {
		return nil, fmt.Errorf("STRAVA_ACCESS_TOKEN not configured")
	}

	result := &SyncResult{}

	// Fetch activities from Strava
	stravaActivities, err := s.fetchActivities(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("failed to fetch from Strava: %v", err)
		return result, err
	}

	if len(stravaActivities) == 0 {
		return result, nil
	}

	if limit > 0 && len(stravaActivities) > limit {
		stravaActivities = stravaActivities[:limit]
	}

	// Find time bounds
	earliest := stravaActivities[0].StartDate
	latest := stravaActivities[len(stravaActivities)-1].StartDate
	timeWindow := time.Duration(matchTimeWindow*2) * time.Second

	localActivities, err := s.activityRepo.FindByTimeRange(ctx, userID,
		earliest.Add(-timeWindow),
		latest.Add(timeWindow))
	if err != nil {
		result.Error = fmt.Sprintf("failed to fetch local activities: %v", err)
		return result, err
	}

	// Process each Strava activity
	for _, sa := range stravaActivities {
		// Upsert to cache
		stravaModel := &model.StravaActivity{
			ID:           uuid.New(),
			UserID:       userID,
			StravaID:     sa.ID,
			Title:        &sa.Name,
			ActivityType: &sa.Type,
			StartTime:    sa.StartDateLocal,
			Distance:     &sa.Distance,
			IsPrivate:    sa.Private,
			IsFlagged:    false, // Strava doesn't expose this, user can update manually
			RawJSON:      nil,
			SyncedAt:     time.Now(),
		}

		// Store raw JSON
		if data, err := json.Marshal(sa); err == nil {
			json.Unmarshal(data, &stravaModel.RawJSON)
		}

		if err := s.stravaRepo.Upsert(ctx, stravaModel); err != nil {
			continue // log but continue
		}

		// Try to match with local activity
		match := s.findMatch(sa, localActivities)
		if match != nil {
			// Update existing activity
			sport := mapStravaType(sa.Type)
			if err := s.activityRepo.UpdateActivity(ctx, match.ID, sa.Name, sport); err == nil {
				result.UpdatedCount++
			}
		} else {
			// Find potential candidates
			candidates := s.findCandidates(sa, localActivities)
			if len(candidates) > 0 {
				result.Candidates = append(result.Candidates, model.StravaMatchCandidate{
					StravaActivity:  stravaModel,
					LocalActivity:   nil,
					TimeDiffSecs:    0,
					DistanceDiffPct: 0,
				})
			} else {
				// No match, could create new (not implemented yet - requires .fit file)
				result.CreatedCount++
			}
		}

		// Rate limiting
		time.Sleep(rateLimitDelay)
	}

	return result, nil
}

// fetchActivities retrieves all activities from Strava (paginated)
func (s *StravaService) fetchActivities(ctx context.Context) ([]StravaActivitySummary, error) {
	var allActivities []StravaActivitySummary
	page := 1
	perPage := 200

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/athlete/activities?page=%d&per_page=%d",
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

		// Safety limit
		if page > 10 { // 2000 activities max
			break
		}
	}

	return allActivities, nil
}

// findMatch looks for a local activity matching the Strava activity
func (s *StravaService) findMatch(sa StravaActivitySummary, localActivities []*model.Activity) *model.Activity {
	for _, la := range localActivities {
		// Time diff
		timeDiff := math.Abs(la.StartTime.Sub(sa.StartDateLocal).Seconds())
		if timeDiff > float64(matchTimeWindow) {
			continue
		}

		// Distance diff
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
		// Time diff (looser window for candidates)
		timeDiff := math.Abs(la.StartTime.Sub(sa.StartDateLocal).Seconds())
		if timeDiff > float64(matchTimeWindow*10) { // 5 minutes
			continue
		}

		// Distance diff (looser window)
		if la.Distance == 0 {
			continue
		}
		distDiff := math.Abs(la.Distance-sa.Distance) / la.Distance
		if distDiff > matchDistancePct*10 { // 10%
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

// IsConfigured returns whether Strava is configured
func IsStravaConfigured(cfg *config.Config) bool {
	return cfg.StravaAccessToken != ""
}
