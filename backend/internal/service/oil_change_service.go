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

// Sentinel errors returned by OilChangeService methods.
var (
	ErrOilChangeRecordNotFound = errors.New("oil change record not found")
)

// OilChangeRepository is the subset of db.Querier used by OilChangeService.
type OilChangeRepository interface {
	CreateOilChangeRecord(ctx context.Context, arg db.CreateOilChangeRecordParams) (db.OilChangeRecord, error)
	GetOilChangeRecordByID(ctx context.Context, id uuid.UUID) (db.OilChangeRecord, error)
	ListOilChangeRecordsByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]db.OilChangeRecord, error)
	GetLatestOilChangeRecord(ctx context.Context, vehicleID uuid.UUID) (db.OilChangeRecord, error)
	DeleteOilChangeRecord(ctx context.Context, id uuid.UUID) error
}

// OilChangeService contains all business logic related to oil change records.
type OilChangeService struct {
	repo        OilChangeRepository
	vehicleRepo VehicleRepository
}

// NewOilChangeService creates a new OilChangeService.
func NewOilChangeService(repo OilChangeRepository, vehicleRepo VehicleRepository) *OilChangeService {
	return &OilChangeService{
		repo:        repo,
		vehicleRepo: vehicleRepo,
	}
}

// CreateOilChangeRecord creates a new oil change record for a given vehicle.
// It verifies the vehicle exists before inserting.
func (s *OilChangeService) CreateOilChangeRecord(ctx context.Context, arg db.CreateOilChangeRecordParams) (db.OilChangeRecord, error) {
	// Ensure the vehicle exists.
	_, err := s.vehicleRepo.GetVehicleByID(ctx, arg.VehicleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.OilChangeRecord{}, ErrVehicleNotFound
		}
		return db.OilChangeRecord{}, fmt.Errorf("service: get vehicle: %w", err)
	}

	record, err := s.repo.CreateOilChangeRecord(ctx, arg)
	if err != nil {
		return db.OilChangeRecord{}, fmt.Errorf("service: create oil change record: %w", err)
	}

	slog.InfoContext(ctx, "oil change record created",
		"record_id", record.ID,
		"vehicle_id", record.VehicleID,
	)
	return record, nil
}

// GetOilChangeRecordByID retrieves a single oil change record by its UUID.
func (s *OilChangeService) GetOilChangeRecordByID(ctx context.Context, id uuid.UUID) (db.OilChangeRecord, error) {
	record, err := s.repo.GetOilChangeRecordByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.OilChangeRecord{}, ErrOilChangeRecordNotFound
		}
		return db.OilChangeRecord{}, fmt.Errorf("service: get oil change record: %w", err)
	}
	return record, nil
}

// ListOilChangeRecordsByVehicle returns all records for the given vehicle,
// ordered by service date descending.
func (s *OilChangeService) ListOilChangeRecordsByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]db.OilChangeRecord, error) {
	// Ensure the vehicle exists.
	_, err := s.vehicleRepo.GetVehicleByID(ctx, vehicleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("service: get vehicle: %w", err)
	}

	records, err := s.repo.ListOilChangeRecordsByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("service: list oil change records: %w", err)
	}
	return records, nil
}

// GetLatestOilChangeRecord retrieves the most recent oil change record for a vehicle.
func (s *OilChangeService) GetLatestOilChangeRecord(ctx context.Context, vehicleID uuid.UUID) (db.OilChangeRecord, error) {
	// Ensure the vehicle exists.
	_, err := s.vehicleRepo.GetVehicleByID(ctx, vehicleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.OilChangeRecord{}, ErrVehicleNotFound
		}
		return db.OilChangeRecord{}, fmt.Errorf("service: get vehicle: %w", err)
	}

	record, err := s.repo.GetLatestOilChangeRecord(ctx, vehicleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.OilChangeRecord{}, ErrOilChangeRecordNotFound
		}
		return db.OilChangeRecord{}, fmt.Errorf("service: get latest oil change record: %w", err)
	}
	return record, nil
}

// DeleteOilChangeRecord removes an oil change record by ID.
func (s *OilChangeService) DeleteOilChangeRecord(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetOilChangeRecordByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOilChangeRecordNotFound
		}
		return fmt.Errorf("service: get oil change record: %w", err)
	}

	if err := s.repo.DeleteOilChangeRecord(ctx, id); err != nil {
		return fmt.Errorf("service: delete oil change record: %w", err)
	}

	slog.InfoContext(ctx, "oil change record deleted", "record_id", id)
	return nil
}
