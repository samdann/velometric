package service

import (
	"context"
	"fmt"

	"github.com/velometric/backend/internal/model"
	"github.com/velometric/backend/internal/repository"
)

// Zone names and colors aligned with CSS design tokens.
var hrZoneNames = []string{"Recovery", "Endurance", "Tempo", "Threshold", "Anaerobic"}
var hrZoneColors = []string{"#64748B", "#3B82F6", "#22C55E", "#EAB308", "#F97316"}

var powerZoneNames = []string{"Recovery", "Endurance", "Tempo", "Threshold", "VO2 Max", "Anaerobic", "Neuromuscular"}
var powerZoneColors = []string{"#64748B", "#3B82F6", "#22C55E", "#EAB308", "#F97316", "#EF4444", "#DC2626"}

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(ctx context.Context) (*model.User, error) {
	return s.repo.GetFirst(ctx)
}

func (s *UserService) UpdateProfile(ctx context.Context, name, email string, weight *float64) (*model.User, error) {
	u, err := s.repo.GetFirst(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateProfile(ctx, u.ID, name, email, weight); err != nil {
		return nil, err
	}
	return s.repo.GetFirst(ctx)
}

func (s *UserService) GetHRZones(ctx context.Context) (int, []model.HRZone, error) {
	u, err := s.repo.GetFirst(ctx)
	if err != nil {
		return 0, nil, err
	}
	zones, err := s.repo.GetHRZones(ctx, u.ID)
	if err != nil {
		return 0, nil, err
	}
	maxHR := 0
	if u.MaxHR != nil {
		maxHR = *u.MaxHR
	}
	return maxHR, zones, nil
}

// SaveHRZones accepts 4 boundary bpm values and saves 5 HR zones as percentages of maxHR.
// Also updates the user's max_hr field.
func (s *UserService) SaveHRZones(ctx context.Context, maxHR int, boundaries []int) ([]model.HRZone, error) {
	if len(boundaries) != 4 {
		return nil, fmt.Errorf("expected 4 boundaries, got %d", len(boundaries))
	}
	if maxHR <= 0 {
		return nil, fmt.Errorf("max_hr must be positive")
	}
	u, err := s.repo.GetFirst(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateMaxHR(ctx, u.ID, maxHR); err != nil {
		return nil, err
	}
	zones := make([]model.HRZone, 5)
	prevPct := 0.0
	for i := 0; i < 5; i++ {
		var maxPct *float64
		if i < 4 {
			pct := float64(boundaries[i]) / float64(maxHR) * 100
			maxPct = &pct
		}
		zones[i] = model.HRZone{
			ZoneNumber:    i + 1,
			Name:          hrZoneNames[i],
			MinPercentage: prevPct,
			MaxPercentage: maxPct,
			Color:         hrZoneColors[i],
		}
		if maxPct != nil {
			prevPct = *maxPct
		}
	}
	if err := s.repo.UpsertHRZones(ctx, u.ID, zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (s *UserService) GetPowerZones(ctx context.Context) (int, []model.PowerZone, error) {
	u, err := s.repo.GetFirst(ctx)
	if err != nil {
		return 0, nil, err
	}
	zones, err := s.repo.GetPowerZones(ctx, u.ID)
	if err != nil {
		return 0, nil, err
	}
	ftp := 0
	if u.FTP != nil {
		ftp = *u.FTP
	}
	return ftp, zones, nil
}

// SavePowerZones accepts 6 boundary watt values and saves 7 power zones as percentages of FTP.
// Also updates the user's ftp field.
func (s *UserService) SavePowerZones(ctx context.Context, ftp int, boundaries []int) ([]model.PowerZone, error) {
	if len(boundaries) != 6 {
		return nil, fmt.Errorf("expected 6 boundaries, got %d", len(boundaries))
	}
	if ftp <= 0 {
		return nil, fmt.Errorf("ftp must be positive")
	}
	u, err := s.repo.GetFirst(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateFTP(ctx, u.ID, ftp); err != nil {
		return nil, err
	}
	zones := make([]model.PowerZone, 7)
	prevPct := 0.0
	for i := 0; i < 7; i++ {
		var maxPct *float64
		if i < 6 {
			pct := float64(boundaries[i]) / float64(ftp) * 100
			maxPct = &pct
		}
		zones[i] = model.PowerZone{
			ZoneNumber:    i + 1,
			Name:          powerZoneNames[i],
			MinPercentage: prevPct,
			MaxPercentage: maxPct,
			Color:         powerZoneColors[i],
		}
		if maxPct != nil {
			prevPct = *maxPct
		}
	}
	if err := s.repo.UpsertPowerZones(ctx, u.ID, zones); err != nil {
		return nil, err
	}
	return zones, nil
}
