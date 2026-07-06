package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/lam-thinh/customer-oil-change-management/internal/db/sqlc"
)

// Sentinel errors returned by VehicleService methods.
var (
	ErrVehicleNotFound    = errors.New("vehicle not found")
	ErrLicensePlateTaken  = errors.New("license plate already registered")
)

// VehicleRepository is the subset of db.Querier used by VehicleService.
type VehicleRepository interface {
	CreateVehicle(ctx context.Context, arg db.CreateVehicleParams) (db.Vehicle, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (db.Vehicle, error)
	GetVehicleByLicensePlate(ctx context.Context, licensePlate string) (db.Vehicle, error)
	ListVehicles(ctx context.Context) ([]db.Vehicle, error)
	UpdateVehicle(ctx context.Context, arg db.UpdateVehicleParams) (db.Vehicle, error)
	DeleteVehicle(ctx context.Context, id uuid.UUID) error
}

// VehicleService contains all business logic related to vehicles.
type VehicleService struct {
	repo VehicleRepository
	log  *slog.Logger
}

// NewVehicleService creates a new VehicleService.
func NewVehicleService(repo VehicleRepository, log *slog.Logger) *VehicleService {
	return &VehicleService{
		repo: repo,
		log:  log,
	}
}

// CreateVehicle validates license plate uniqueness and creates a new vehicle.
func (s *VehicleService) CreateVehicle(ctx context.Context, arg db.CreateVehicleParams) (db.Vehicle, error) {
	// Check for duplicate license plate.
	_, err := s.repo.GetVehicleByLicensePlate(ctx, arg.LicensePlate)
	if err == nil {
		return db.Vehicle{}, ErrLicensePlateTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.Vehicle{}, fmt.Errorf("service: check license plate: %w", err)
	}

	vehicle, err := s.repo.CreateVehicle(ctx, arg)
	if err != nil {
		return db.Vehicle{}, fmt.Errorf("service: create vehicle: %w", err)
	}

	s.log.InfoContext(ctx, "vehicle created", "vehicle_id", vehicle.ID)
	return vehicle, nil
}

// GetVehicleByID retrieves a single vehicle by its UUID.
func (s *VehicleService) GetVehicleByID(ctx context.Context, id uuid.UUID) (db.Vehicle, error) {
	vehicle, err := s.repo.GetVehicleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Vehicle{}, ErrVehicleNotFound
		}
		return db.Vehicle{}, fmt.Errorf("service: get vehicle: %w", err)
	}
	return vehicle, nil
}

// ListVehicles returns all vehicles ordered by creation date descending.
func (s *VehicleService) ListVehicles(ctx context.Context) ([]db.Vehicle, error) {
	vehicles, err := s.repo.ListVehicles(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list vehicles: %w", err)
	}
	return vehicles, nil
}

// UpdateVehicle updates fields on an existing vehicle, verifying the license
// plate is not already taken by a different vehicle.
func (s *VehicleService) UpdateVehicle(ctx context.Context, arg db.UpdateVehicleParams) (db.Vehicle, error) {
	// Ensure the vehicle exists.
	existing, err := s.repo.GetVehicleByID(ctx, arg.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Vehicle{}, ErrVehicleNotFound
		}
		return db.Vehicle{}, fmt.Errorf("service: get vehicle: %w", err)
	}

	// Check license plate uniqueness only when it changes.
	if arg.LicensePlate != existing.LicensePlate {
		conflict, err := s.repo.GetVehicleByLicensePlate(ctx, arg.LicensePlate)
		if err == nil && conflict.ID != arg.ID {
			return db.Vehicle{}, ErrLicensePlateTaken
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return db.Vehicle{}, fmt.Errorf("service: check license plate: %w", err)
		}
	}

	vehicle, err := s.repo.UpdateVehicle(ctx, arg)
	if err != nil {
		return db.Vehicle{}, fmt.Errorf("service: update vehicle: %w", err)
	}

	s.log.InfoContext(ctx, "vehicle updated", "vehicle_id", vehicle.ID)
	return vehicle, nil
}

// DeleteVehicle removes a vehicle by ID.
func (s *VehicleService) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetVehicleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVehicleNotFound
		}
		return fmt.Errorf("service: get vehicle: %w", err)
	}

	if err := s.repo.DeleteVehicle(ctx, id); err != nil {
		return fmt.Errorf("service: delete vehicle: %w", err)
	}

	s.log.InfoContext(ctx, "vehicle deleted", "vehicle_id", id)
	return nil
}
