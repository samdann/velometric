package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/velometric/backend/internal/database"
	"github.com/velometric/backend/internal/fitparser"
	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/repository"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

// fakeActivityService implements activityServicer with configurable return values.
type fakeActivityService struct {
	activity     *model.Activity
	activities   []*model.Activity
	total        int
	records      []model.ActivityRecord
	powerCurve   []model.PowerCurvePoint
	elevation    []model.ElevationPoint
	speed        []model.SpeedPoint
	hrCadence    []model.HRCadencePoint
	route        []model.RoutePoint
	laps         []model.ActivityLap
	feed         []model.FeedActivity
	hrDist       []model.HRZoneDistributionPoint
	powerDist    []model.PowerZoneDistributionPoint
	err          error
	deleteErr    error
	processFITFn func() (*model.Activity, error)
}

func (f *fakeActivityService) GetActivity(_ context.Context, _ uuid.UUID) (*model.Activity, error) {
	return f.activity, f.err
}
func (f *fakeActivityService) ListActivitiesPaginated(_ context.Context, _ uuid.UUID, _, _ int, _ model.ActivityFilter) ([]*model.Activity, int, error) {
	return f.activities, f.total, f.err
}
func (f *fakeActivityService) ProcessFITFile(_ context.Context, _ uuid.UUID, _ *fitparser.ParsedActivity, _ int) (*model.Activity, error) {
	if f.processFITFn != nil {
		return f.processFITFn()
	}
	return f.activity, f.err
}
func (f *fakeActivityService) GetActivityRecords(_ context.Context, _ uuid.UUID) ([]model.ActivityRecord, error) {
	return f.records, f.err
}
func (f *fakeActivityService) GetPowerCurve(_ context.Context, _ uuid.UUID) ([]model.PowerCurvePoint, error) {
	return f.powerCurve, f.err
}
func (f *fakeActivityService) GetElevationProfile(_ context.Context, _ uuid.UUID) ([]model.ElevationPoint, error) {
	return f.elevation, f.err
}
func (f *fakeActivityService) GetSpeedProfile(_ context.Context, _ uuid.UUID) ([]model.SpeedPoint, error) {
	return f.speed, f.err
}
func (f *fakeActivityService) GetHRCadenceProfile(_ context.Context, _ uuid.UUID) ([]model.HRCadencePoint, error) {
	return f.hrCadence, f.err
}
func (f *fakeActivityService) GetRoute(_ context.Context, _ uuid.UUID) ([]model.RoutePoint, error) {
	return f.route, f.err
}
func (f *fakeActivityService) GetLaps(_ context.Context, _ uuid.UUID) ([]model.ActivityLap, error) {
	return f.laps, f.err
}
func (f *fakeActivityService) DeleteActivity(_ context.Context, _ uuid.UUID) error {
	return f.deleteErr
}
func (f *fakeActivityService) ComputeHRZoneDistribution(_ context.Context, _ uuid.UUID, _ int, _ []model.HRZone) ([]model.HRZoneDistributionPoint, error) {
	return f.hrDist, f.err
}
func (f *fakeActivityService) ComputePowerZoneDistribution(_ context.Context, _ uuid.UUID, _ int, _ []model.PowerZone) ([]model.PowerZoneDistributionPoint, error) {
	return f.powerDist, f.err
}
func (f *fakeActivityService) GetFeed(_ context.Context, _ uuid.UUID, _, _ int) ([]model.FeedActivity, int, error) {
	return f.feed, f.total, f.err
}
func (f *fakeActivityService) GetDistinctSports(_ context.Context, _ uuid.UUID) ([]string, error) {
	return []string{}, f.err
}

// fakeUserService implements userServicer with configurable return values.
type fakeUserService struct {
	user       *model.User
	hrZones    []model.HRZone
	maxHR      int
	powerZones []model.PowerZone
	ftp        int
	err        error
}

func (f *fakeUserService) GetProfile(_ context.Context) (*model.User, error) {
	return f.user, f.err
}
func (f *fakeUserService) UpdateProfile(_ context.Context, _, _ string, _ *float64) (*model.User, error) {
	return f.user, f.err
}
func (f *fakeUserService) GetHRZones(_ context.Context) (int, []model.HRZone, error) {
	return f.maxHR, f.hrZones, f.err
}
func (f *fakeUserService) SaveHRZones(_ context.Context, _ int, _ []int) ([]model.HRZone, error) {
	return f.hrZones, f.err
}
func (f *fakeUserService) GetPowerZones(_ context.Context) (int, []model.PowerZone, error) {
	return f.ftp, f.powerZones, f.err
}
func (f *fakeUserService) SavePowerZones(_ context.Context, _ int, _ []int) ([]model.PowerZone, error) {
	return f.powerZones, f.err
}

// ── test helpers ──────────────────────────────────────────────────────────────

var fixedUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// newTestHandler builds a Handler with fake services and no real DB.
// h.db is set to a non-nil sentinel so HasDB() returns true.
func newTestHandler(act activityServicer, usr userServicer) *Handler {
	return &Handler{
		db:              &database.DB{}, // non-nil → HasDB() == true; Pool is nil (never touched by fakes)
		activityService: act,
		userService:     usr,
		resolveUserID: func(_ context.Context) (uuid.UUID, error) {
			return fixedUserID, nil
		},
	}
}

// withIDParam injects a chi URL param "id" into req's context.
func withIDParam(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// executeWithID sends a GET to handler h with the given UUID as chi "id" param.
func executeWithID(h http.HandlerFunc, id uuid.UUID) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/"+id.String(), nil)
	req = withIDParam(req, id.String())
	h(rr, req)
	return rr
}

// decodeJSON is a test helper that decodes JSON from a recorder into v.
func decodeJSON(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(rr.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode JSON response: %v (body: %s)", err, rr.Body.String())
	}
}

// buildMultipartFIT creates a multipart/form-data body with a "file" field containing data.
// Returns the body buffer and the Content-Type header value.
func buildMultipartFIT(t *testing.T, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "activity.fit")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("failed to write fit data: %v", err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

// ── GetActivity ───────────────────────────────────────────────────────────────

func TestGetActivity_Success(t *testing.T) {
	id := uuid.New()
	act := &model.Activity{ID: id, Name: "Morning Ride", Sport: "cycling"}
	h := newTestHandler(&fakeActivityService{activity: act}, &fakeUserService{})

	rr := executeWithID(h.GetActivity, id)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got model.Activity
	decodeJSON(t, rr, &got)
	if got.ID != id {
		t.Errorf("activity ID = %v, want %v", got.ID, id)
	}
	if got.Name != "Morning Ride" {
		t.Errorf("activity Name = %q, want %q", got.Name, "Morning Ride")
	}
}

func TestGetActivity_NotFound(t *testing.T) {
	h := newTestHandler(&fakeActivityService{activity: nil, err: nil}, &fakeUserService{})

	rr := executeWithID(h.GetActivity, uuid.New())

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetActivity_ServiceError(t *testing.T) {
	h := newTestHandler(&fakeActivityService{err: errors.New("db down")}, &fakeUserService{})

	rr := executeWithID(h.GetActivity, uuid.New())

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetActivity_InvalidID(t *testing.T) {
	h := newTestHandler(&fakeActivityService{}, &fakeUserService{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil)
	req = withIDParam(req, "not-a-uuid")
	h.GetActivity(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetActivity_ResponseIsJSON(t *testing.T) {
	act := &model.Activity{ID: uuid.New()}
	h := newTestHandler(&fakeActivityService{activity: act}, &fakeUserService{})

	rr := executeWithID(h.GetActivity, act.ID)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

// ── DeleteActivity ────────────────────────────────────────────────────────────

func TestDeleteActivity_Success(t *testing.T) {
	h := newTestHandler(&fakeActivityService{deleteErr: nil}, &fakeUserService{})

	rr := httptest.NewRecorder()
	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/"+id.String(), nil)
	req = withIDParam(req, id.String())
	h.DeleteActivity(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestDeleteActivity_NotFound(t *testing.T) {
	h := newTestHandler(&fakeActivityService{deleteErr: repository.ErrActivityNotFound}, &fakeUserService{})

	rr := httptest.NewRecorder()
	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/"+id.String(), nil)
	req = withIDParam(req, id.String())
	h.DeleteActivity(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestDeleteActivity_InvalidID(t *testing.T) {
	h := newTestHandler(&fakeActivityService{}, &fakeUserService{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/bad", nil)
	req = withIDParam(req, "bad")
	h.DeleteActivity(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ── GetLaps ───────────────────────────────────────────────────────────────────

func TestGetLaps_EmptyArray(t *testing.T) {
	h := newTestHandler(&fakeActivityService{laps: []model.ActivityLap{}}, &fakeUserService{})

	rr := executeWithID(h.GetLaps, uuid.New())

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var got []model.ActivityLap
	decodeJSON(t, rr, &got)
	if len(got) != 0 {
		t.Errorf("expected empty laps array, got %d", len(got))
	}
}

func TestGetLaps_ReturnsList(t *testing.T) {
	laps := []model.ActivityLap{
		{LapNumber: 1, Duration: 300},
		{LapNumber: 2, Duration: 350},
	}
	h := newTestHandler(&fakeActivityService{laps: laps}, &fakeUserService{})

	rr := executeWithID(h.GetLaps, uuid.New())

	var got []model.ActivityLap
	decodeJSON(t, rr, &got)
	if len(got) != 2 {
		t.Errorf("expected 2 laps, got %d", len(got))
	}
}

// ── ListActivities ────────────────────────────────────────────────────────────

func TestListActivities_ReturnsPagedResponse(t *testing.T) {
	activities := []*model.Activity{
		{ID: uuid.New(), Name: "Ride 1"},
		{ID: uuid.New(), Name: "Ride 2"},
	}
	h := newTestHandler(&fakeActivityService{activities: activities, total: 2}, &fakeUserService{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/activities?page=1&limit=25", nil)
	h.ListActivities(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusOK, rr.Body.String())
	}
	var resp PaginatedActivitiesResponse
	decodeJSON(t, rr, &resp)
	if resp.Total != 2 {
		t.Errorf("Total = %d, want 2", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("Page = %d, want 1", resp.Page)
	}
}

func TestListActivities_InvalidLimitFallsBackToDefault(t *testing.T) {
	// limit=999 is not in the allowed set (10, 25, 50) — must fall back to default 25.
	h := newTestHandler(&fakeActivityService{}, &fakeUserService{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/activities?limit=999", nil)
	h.ListActivities(rr, req)

	var resp PaginatedActivitiesResponse
	decodeJSON(t, rr, &resp)
	if resp.Limit != 25 {
		t.Errorf("Limit = %d, want 25 (invalid limit should fall back to default)", resp.Limit)
	}
}

func TestListActivities_ValidLimitsAccepted(t *testing.T) {
	h := newTestHandler(&fakeActivityService{}, &fakeUserService{})

	for _, tc := range []struct{ query string; want int }{
		{"limit=10", 10},
		{"limit=25", 25},
		{"limit=50", 50},
	} {
		t.Run(tc.query, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/activities?"+tc.query, nil)
			h.ListActivities(rr, req)
			var resp PaginatedActivitiesResponse
			decodeJSON(t, rr, &resp)
			if resp.Limit != tc.want {
				t.Errorf("Limit = %d, want %d", resp.Limit, tc.want)
			}
		})
	}
}

// ── GetElevationProfile ───────────────────────────────────────────────────────

func TestGetElevationProfile_EmptyResult(t *testing.T) {
	h := newTestHandler(&fakeActivityService{elevation: []model.ElevationPoint{}}, &fakeUserService{})

	rr := executeWithID(h.GetElevationProfile, uuid.New())

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGetElevationProfile_ServiceError(t *testing.T) {
	h := newTestHandler(&fakeActivityService{err: errors.New("db error")}, &fakeUserService{})

	rr := executeWithID(h.GetElevationProfile, uuid.New())

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// ── GetPowerCurve ─────────────────────────────────────────────────────────────

func TestGetPowerCurve_AttachesWattsPerKg(t *testing.T) {
	weight := 80.0
	curve := []model.PowerCurvePoint{
		{DurationSeconds: 5, BestPower: 800},
		{DurationSeconds: 60, BestPower: 400},
	}
	h := newTestHandler(
		&fakeActivityService{powerCurve: curve},
		&fakeUserService{user: &model.User{Weight: &weight}},
	)

	rr := executeWithID(h.GetPowerCurve, uuid.New())

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var got []model.PowerCurvePoint
	decodeJSON(t, rr, &got)
	if len(got) != 2 {
		t.Fatalf("expected 2 points, got %d", len(got))
	}
	if got[0].WattsPerKg == nil {
		t.Fatal("WattsPerKg is nil, want a value")
	}
	want := 800.0 / 80.0 // 10.0
	if *got[0].WattsPerKg != want {
		t.Errorf("WattsPerKg = %.2f, want %.2f", *got[0].WattsPerKg, want)
	}
}

func TestGetPowerCurve_NoWattsPerKgWhenWeightUnset(t *testing.T) {
	curve := []model.PowerCurvePoint{{DurationSeconds: 5, BestPower: 800}}
	h := newTestHandler(
		&fakeActivityService{powerCurve: curve},
		&fakeUserService{user: &model.User{}}, // Weight is nil
	)

	rr := executeWithID(h.GetPowerCurve, uuid.New())

	var got []model.PowerCurvePoint
	decodeJSON(t, rr, &got)
	if got[0].WattsPerKg != nil {
		t.Errorf("WattsPerKg should be nil when user weight is not set, got %v", *got[0].WattsPerKg)
	}
}

// ── HasDB guard ───────────────────────────────────────────────────────────────

func TestGetActivity_NoDBReturns503(t *testing.T) {
	h := &Handler{db: nil}

	rr := executeWithID(h.GetActivity, uuid.New())

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
}

func TestListActivities_NoDBReturnsEmptyList(t *testing.T) {
	// ListActivities returns an empty paginated response (not 503) when DB is unavailable.
	h := &Handler{db: nil}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/activities", nil)
	h.ListActivities(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp PaginatedActivitiesResponse
	decodeJSON(t, rr, &resp)
	if resp.Total != 0 {
		t.Errorf("Total = %d, want 0", resp.Total)
	}
}

// ── CreateActivity ────────────────────────────────────────────────────────────

func TestCreateActivity_DuplicateReturns409(t *testing.T) {
	fake := &fakeActivityService{
		processFITFn: func() (*model.Activity, error) {
			return nil, repository.ErrDuplicateActivity
		},
	}
	h := newTestHandler(fake, &fakeUserService{})

	fitData, err := os.ReadFile("../fitparser/testdata/activity.fit")
	if err != nil {
		t.Fatalf("could not read fixture: %v", err)
	}

	body, contentType := buildMultipartFIT(t, fitData)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/activities", body)
	req.Header.Set("Content-Type", contentType)
	h.CreateActivity(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d (body: %s)", rr.Code, http.StatusConflict, rr.Body.String())
	}
}

func TestCreateActivity_NoFileReturns400(t *testing.T) {
	h := newTestHandler(&fakeActivityService{}, &fakeUserService{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/activities", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	h.CreateActivity(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateActivity_Success(t *testing.T) {
	id := uuid.New()
	fake := &fakeActivityService{
		processFITFn: func() (*model.Activity, error) {
			return &model.Activity{ID: id, Name: "Morning Ride"}, nil
		},
	}
	h := newTestHandler(fake, &fakeUserService{})

	fitData, err := os.ReadFile("../fitparser/testdata/activity.fit")
	if err != nil {
		t.Fatalf("could not read fixture: %v", err)
	}

	body, contentType := buildMultipartFIT(t, fitData)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/activities", body)
	req.Header.Set("Content-Type", contentType)
	h.CreateActivity(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d (body: %s)", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var resp UploadResponse
	decodeJSON(t, rr, &resp)
	if resp.ID != id.String() {
		t.Errorf("response ID = %q, want %q", resp.ID, id.String())
	}
}
