package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	db "github.com/lam-thinh/customer-oil-change-management/internal/db/sqlc"
	"github.com/lam-thinh/customer-oil-change-management/internal/dto"
	"github.com/lam-thinh/customer-oil-change-management/internal/service"
)

// VehicleServicer is the interface the VehicleHandler depends on.
type VehicleServicer interface {
	CreateVehicle(ctx context.Context, arg db.CreateVehicleParams) (db.Vehicle, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (db.Vehicle, error)
	ListVehicles(ctx context.Context) ([]db.Vehicle, error)
	UpdateVehicle(ctx context.Context, arg db.UpdateVehicleParams) (db.Vehicle, error)
	DeleteVehicle(ctx context.Context, id uuid.UUID) error
}

// VehicleHandler holds the dependencies for vehicle HTTP handlers.
type VehicleHandler struct {
	svc VehicleServicer
	log *slog.Logger
}

// NewVehicleHandler creates a new VehicleHandler.
func NewVehicleHandler(svc VehicleServicer, log *slog.Logger) *VehicleHandler {
	return &VehicleHandler{svc: svc, log: log}
}

// mapVehicleServiceError maps VehicleService sentinel errors to HTTP responses.
func mapVehicleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrVehicleNotFound):
		Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrLicensePlateTaken):
		Error(w, http.StatusConflict, err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}

// vehicleToResponse converts a db.Vehicle to a dto.VehicleResponse.
func vehicleToResponse(v db.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID:           v.ID.String(),
		LicensePlate: v.LicensePlate,
		OwnerName:    v.OwnerName,
		PhoneNumber:  v.PhoneNumber,
		Make:         v.Make,
		Model:        v.Model,
		Year:         v.Year,
		CreatedAt:    v.CreatedAt,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// ListVehicles godoc
//
//	@Summary      List vehicles
//	@Description  Returns all vehicles ordered by creation date descending
//	@Tags         vehicles
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {array}   dto.VehicleResponse
//	@Failure      401  {object}  dto.ErrorResponse  "Not authenticated"
//	@Failure      500  {object}  dto.ErrorResponse  "Internal error"
//	@Router       /v1/vehicles [get]
func (h *VehicleHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.svc.ListVehicles(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "list vehicles failed", "error", err)
		mapVehicleServiceError(w, err)
		return
	}

	resp := make([]dto.VehicleResponse, 0, len(vehicles))
	for _, v := range vehicles {
		resp = append(resp, vehicleToResponse(v))
	}
	JSON(w, http.StatusOK, resp)
}

// CreateVehicle godoc
//
//	@Summary      Create a vehicle
//	@Description  Register a new vehicle
//	@Tags         vehicles
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        request  body      dto.CreateVehicleRequest  true  "Vehicle data"
//	@Success      201      {object}  dto.VehicleResponse       "Vehicle created"
//	@Failure      400      {object}  dto.ErrorResponse         "Malformed JSON"
//	@Failure      401      {object}  dto.ErrorResponse         "Not authenticated"
//	@Failure      409      {object}  dto.ErrorResponse         "License plate already registered"
//	@Failure      422      {object}  dto.ErrorResponse         "Validation failed"
//	@Router       /v1/vehicles [post]
func (h *VehicleHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateVehicleRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	vehicle, err := h.svc.CreateVehicle(r.Context(), db.CreateVehicleParams{
		LicensePlate: req.LicensePlate,
		OwnerName:    req.OwnerName,
		PhoneNumber:  req.PhoneNumber,
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "create vehicle failed", "error", err)
		mapVehicleServiceError(w, err)
		return
	}

	JSON(w, http.StatusCreated, vehicleToResponse(vehicle))
}

// GetVehicle godoc
//
//	@Summary      Get a vehicle
//	@Description  Retrieve a vehicle by its UUID
//	@Tags         vehicles
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path      string              true  "Vehicle UUID"
//	@Success      200        {object}  dto.VehicleResponse "Vehicle found"
//	@Failure      400        {object}  dto.ErrorResponse   "Invalid UUID"
//	@Failure      401        {object}  dto.ErrorResponse   "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse   "Vehicle not found"
//	@Router       /v1/vehicles/{vehicleID} [get]
func (h *VehicleHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "vehicleID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	vehicle, err := h.svc.GetVehicleByID(r.Context(), id)
	if err != nil {
		h.log.ErrorContext(r.Context(), "get vehicle failed", "error", err)
		mapVehicleServiceError(w, err)
		return
	}

	JSON(w, http.StatusOK, vehicleToResponse(vehicle))
}

// UpdateVehicle godoc
//
//	@Summary      Update a vehicle
//	@Description  Update vehicle details by UUID
//	@Tags         vehicles
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path      string                    true  "Vehicle UUID"
//	@Param        request    body      dto.UpdateVehicleRequest  true  "Updated vehicle data"
//	@Success      200        {object}  dto.VehicleResponse       "Vehicle updated"
//	@Failure      400        {object}  dto.ErrorResponse         "Invalid UUID or malformed JSON"
//	@Failure      401        {object}  dto.ErrorResponse         "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse         "Vehicle not found"
//	@Failure      409        {object}  dto.ErrorResponse         "License plate already registered"
//	@Failure      422        {object}  dto.ErrorResponse         "Validation failed"
//	@Router       /v1/vehicles/{vehicleID} [put]
func (h *VehicleHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "vehicleID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	var req dto.UpdateVehicleRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	vehicle, err := h.svc.UpdateVehicle(r.Context(), db.UpdateVehicleParams{
		ID:           id,
		LicensePlate: req.LicensePlate,
		OwnerName:    req.OwnerName,
		PhoneNumber:  req.PhoneNumber,
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "update vehicle failed", "error", err)
		mapVehicleServiceError(w, err)
		return
	}

	JSON(w, http.StatusOK, vehicleToResponse(vehicle))
}

// DeleteVehicle godoc
//
//	@Summary      Delete a vehicle
//	@Description  Delete a vehicle and all associated oil change records by UUID
//	@Tags         vehicles
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path  string  true  "Vehicle UUID"
//	@Success      204        "Vehicle deleted"
//	@Failure      400        {object}  dto.ErrorResponse  "Invalid UUID"
//	@Failure      401        {object}  dto.ErrorResponse  "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse  "Vehicle not found"
//	@Router       /v1/vehicles/{vehicleID} [delete]
func (h *VehicleHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "vehicleID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	if err := h.svc.DeleteVehicle(r.Context(), id); err != nil {
		h.log.ErrorContext(r.Context(), "delete vehicle failed", "error", err)
		mapVehicleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
