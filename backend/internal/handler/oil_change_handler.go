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

// OilChangeServicer is the interface the OilChangeHandler depends on.
type OilChangeServicer interface {
	CreateOilChangeRecord(ctx context.Context, arg db.CreateOilChangeRecordParams) (db.OilChangeRecord, error)
	GetOilChangeRecordByID(ctx context.Context, id uuid.UUID) (db.OilChangeRecord, error)
	ListOilChangeRecordsByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]db.OilChangeRecord, error)
	GetLatestOilChangeRecord(ctx context.Context, vehicleID uuid.UUID) (db.OilChangeRecord, error)
	DeleteOilChangeRecord(ctx context.Context, id uuid.UUID) error
}

// OilChangeHandler holds the dependencies for oil change record HTTP handlers.
type OilChangeHandler struct {
	svc OilChangeServicer
	log *slog.Logger
}

// NewOilChangeHandler creates a new OilChangeHandler.
func NewOilChangeHandler(svc OilChangeServicer, log *slog.Logger) *OilChangeHandler {
	return &OilChangeHandler{svc: svc, log: log}
}

// mapOilChangeServiceError maps OilChangeService sentinel errors to HTTP responses.
func mapOilChangeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrVehicleNotFound):
		Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrOilChangeRecordNotFound):
		Error(w, http.StatusNotFound, err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}

// oilChangeRecordToResponse converts a db.OilChangeRecord to a dto.OilChangeRecordResponse.
func oilChangeRecordToResponse(r db.OilChangeRecord) dto.OilChangeRecordResponse {
	return dto.OilChangeRecordResponse{
		ID:                 r.ID.String(),
		VehicleID:          r.VehicleID.String(),
		ServiceDate:        r.ServiceDate,
		CurrentMileage:     r.CurrentMileage,
		NextServiceMileage: r.NextServiceMileage,
		NextServiceDate:    r.NextServiceDate,
		OilType:            r.OilType,
		OilFilter:          r.OilFilter,
		NextOilFilter:      r.NextOilFilter,
		CreatedAt:          r.CreatedAt,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// CreateOilChangeRecord godoc
//
//	@Summary      Create an oil change record
//	@Description  Log a new oil change service for a vehicle
//	@Tags         oil-changes
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path      string                          true  "Vehicle UUID"
//	@Param        request    body      dto.CreateOilChangeRecordRequest true  "Oil change data"
//	@Success      201        {object}  dto.OilChangeRecordResponse     "Record created"
//	@Failure      400        {object}  dto.ErrorResponse               "Invalid UUID or malformed JSON"
//	@Failure      401        {object}  dto.ErrorResponse               "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse               "Vehicle not found"
//	@Failure      422        {object}  dto.ErrorResponse               "Validation failed"
//	@Router       /v1/vehicles/{vehicleID}/oil-changes [post]
func (h *OilChangeHandler) CreateOilChangeRecord(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicleID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	var req dto.CreateOilChangeRecordRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	record, err := h.svc.CreateOilChangeRecord(r.Context(), db.CreateOilChangeRecordParams{
		VehicleID:          vehicleID,
		ServiceDate:        req.ServiceDate,
		CurrentMileage:     req.CurrentMileage,
		NextServiceMileage: req.NextServiceMileage,
		NextServiceDate:    req.NextServiceDate,
		OilType:            req.OilType,
		OilFilter:          req.OilFilter,
		NextOilFilter:      req.NextOilFilter,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "create oil change record failed", "error", err)
		mapOilChangeServiceError(w, err)
		return
	}

	JSON(w, http.StatusCreated, oilChangeRecordToResponse(record))
}

// ListOilChangeRecords godoc
//
//	@Summary      List oil change records
//	@Description  Returns all oil change records for a vehicle, ordered by service date descending
//	@Tags         oil-changes
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path      string                        true  "Vehicle UUID"
//	@Success      200        {array}   dto.OilChangeRecordResponse
//	@Failure      400        {object}  dto.ErrorResponse             "Invalid UUID"
//	@Failure      401        {object}  dto.ErrorResponse             "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse             "Vehicle not found"
//	@Router       /v1/vehicles/{vehicleID}/oil-changes [get]
func (h *OilChangeHandler) ListOilChangeRecords(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicleID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	records, err := h.svc.ListOilChangeRecordsByVehicle(r.Context(), vehicleID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "list oil change records failed", "error", err)
		mapOilChangeServiceError(w, err)
		return
	}

	resp := make([]dto.OilChangeRecordResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, oilChangeRecordToResponse(rec))
	}
	JSON(w, http.StatusOK, resp)
}

// GetLatestOilChangeRecord godoc
//
//	@Summary      Get latest oil change record
//	@Description  Retrieve the most recent oil change record for a vehicle
//	@Tags         oil-changes
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path      string                      true  "Vehicle UUID"
//	@Success      200        {object}  dto.OilChangeRecordResponse "Latest record"
//	@Failure      400        {object}  dto.ErrorResponse           "Invalid UUID"
//	@Failure      401        {object}  dto.ErrorResponse           "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse           "Vehicle or record not found"
//	@Router       /v1/vehicles/{vehicleID}/oil-changes/latest [get]
func (h *OilChangeHandler) GetLatestOilChangeRecord(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicleID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	record, err := h.svc.GetLatestOilChangeRecord(r.Context(), vehicleID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "get latest oil change record failed", "error", err)
		mapOilChangeServiceError(w, err)
		return
	}

	JSON(w, http.StatusOK, oilChangeRecordToResponse(record))
}

// GetOilChangeRecord godoc
//
//	@Summary      Get an oil change record
//	@Description  Retrieve a specific oil change record by its UUID
//	@Tags         oil-changes
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path      string                      true  "Vehicle UUID"
//	@Param        recordID   path      string                      true  "Record UUID"
//	@Success      200        {object}  dto.OilChangeRecordResponse "Record found"
//	@Failure      400        {object}  dto.ErrorResponse           "Invalid UUID"
//	@Failure      401        {object}  dto.ErrorResponse           "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse           "Record not found"
//	@Router       /v1/vehicles/{vehicleID}/oil-changes/{recordID} [get]
func (h *OilChangeHandler) GetOilChangeRecord(w http.ResponseWriter, r *http.Request) {
	recordID, err := uuid.Parse(chi.URLParam(r, "recordID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	record, err := h.svc.GetOilChangeRecordByID(r.Context(), recordID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "get oil change record failed", "error", err)
		mapOilChangeServiceError(w, err)
		return
	}

	JSON(w, http.StatusOK, oilChangeRecordToResponse(record))
}

// DeleteOilChangeRecord godoc
//
//	@Summary      Delete an oil change record
//	@Description  Remove a specific oil change record by its UUID
//	@Tags         oil-changes
//	@Produce      json
//	@Security     BearerAuth
//	@Param        vehicleID  path  string  true  "Vehicle UUID"
//	@Param        recordID   path  string  true  "Record UUID"
//	@Success      204        "Record deleted"
//	@Failure      400        {object}  dto.ErrorResponse  "Invalid UUID"
//	@Failure      401        {object}  dto.ErrorResponse  "Not authenticated"
//	@Failure      404        {object}  dto.ErrorResponse  "Record not found"
//	@Router       /v1/vehicles/{vehicleID}/oil-changes/{recordID} [delete]
func (h *OilChangeHandler) DeleteOilChangeRecord(w http.ResponseWriter, r *http.Request) {
	recordID, err := uuid.Parse(chi.URLParam(r, "recordID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	if err := h.svc.DeleteOilChangeRecord(r.Context(), recordID); err != nil {
		h.log.ErrorContext(r.Context(), "delete oil change record failed", "error", err)
		mapOilChangeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
