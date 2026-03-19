package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/fitparser"
	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/service"
)

// activityServicer is the subset of service.ActivityService used by the handler layer.
// Defining it here (consumer side) lets tests inject fakes without touching service code.
type activityServicer interface {
	ListActivitiesPaginated(ctx context.Context, userID uuid.UUID, page, limit int, f model.ActivityFilter) ([]*model.Activity, int, error)
	GetActivity(ctx context.Context, id uuid.UUID) (*model.Activity, error)
	ProcessFITFile(ctx context.Context, userID uuid.UUID, parsed *fitparser.ParsedActivity, ftp int) (*model.Activity, error)
	GetActivityRecords(ctx context.Context, activityID uuid.UUID) ([]model.ActivityRecord, error)
	GetPowerCurve(ctx context.Context, activityID uuid.UUID) ([]model.PowerCurvePoint, error)
	GetElevationProfile(ctx context.Context, activityID uuid.UUID) ([]model.ElevationPoint, error)
	GetSpeedProfile(ctx context.Context, activityID uuid.UUID) ([]model.SpeedPoint, error)
	GetHRCadenceProfile(ctx context.Context, activityID uuid.UUID) ([]model.HRCadencePoint, error)
	GetRoute(ctx context.Context, activityID uuid.UUID) ([]model.RoutePoint, error)
	GetLaps(ctx context.Context, activityID uuid.UUID) ([]model.ActivityLap, error)
	DeleteActivity(ctx context.Context, id uuid.UUID) error
	ComputeHRZoneDistribution(ctx context.Context, activityID uuid.UUID, maxHR int, zones []model.HRZone) ([]model.HRZoneDistributionPoint, error)
	ComputePowerZoneDistribution(ctx context.Context, activityID uuid.UUID, ftp int, zones []model.PowerZone) ([]model.PowerZoneDistributionPoint, error)
	GetFeed(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.FeedActivity, int, error)
}

// userServicer is the subset of service.UserService used by the handler layer.
type userServicer interface {
	GetProfile(ctx context.Context) (*model.User, error)
	UpdateProfile(ctx context.Context, name, email string, weight *float64) (*model.User, error)
	GetHRZones(ctx context.Context) (int, []model.HRZone, error)
	SaveHRZones(ctx context.Context, maxHR int, boundaries []int) ([]model.HRZone, error)
	GetPowerZones(ctx context.Context) (int, []model.PowerZone, error)
	SavePowerZones(ctx context.Context, ftp int, boundaries []int) ([]model.PowerZone, error)
}

// statisticsServicer is the subset of service.StatisticsService used by the handler layer.
type statisticsServicer interface {
	GetAvailablePowerYears(ctx context.Context, userID uuid.UUID) ([]int, error)
	GetAnnualPowerStats(ctx context.Context, userID uuid.UUID, year, ftp int, zones []model.PowerZone) (*model.AnnualPowerStats, error)
}

// batchImportServicer is the subset of service.BatchImportService used by the handler layer.
type batchImportServicer interface {
	StartImport(ctx context.Context, userID uuid.UUID, ftp int, req service.ImportRequest) (*service.ImportJob, error)
	GetJob(id uuid.UUID) (*service.ImportJob, bool)
}
