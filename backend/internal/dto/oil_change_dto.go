package dto

import "time"

// ── Oil Change Record Request DTOs ────────────────────────────────────────────

// CreateOilChangeRecordRequest is the request body for POST /vehicles/{vehicleID}/oil-changes.
type CreateOilChangeRecordRequest struct {
	ServiceDate        time.Time  `json:"service_date"          validate:"required"  example:"2026-06-15T00:00:00Z"`
	CurrentMileage     int32      `json:"current_mileage"       validate:"required,min=0" example:"158397"`
	NextServiceMileage *int32     `json:"next_service_mileage,omitempty"           example:"166397"`
	NextServiceDate    *time.Time `json:"next_service_date,omitempty"             example:"2026-12-15T00:00:00Z"`
	OilType            *string    `json:"oil_type,omitempty"                       example:"5W40"`
	OilFilter          *string    `json:"oil_filter,omitempty"                     example:"1002"`
	NextOilFilter      *string    `json:"next_oil_filter,omitempty"                example:"1002"`
}

// ── Oil Change Record Response DTOs ──────────────────────────────────────────

// OilChangeRecordResponse is the representation of a record returned by the API.
type OilChangeRecordResponse struct {
	ID                 string     `json:"id"                           example:"49b586b2-4eda-436c-9b93-33d8306e18f0"`
	VehicleID          string     `json:"vehicle_id"                   example:"b4925809-60f5-45f7-ad6e-1f0ca0922ea9"`
	ServiceDate        time.Time  `json:"service_date"                 example:"2026-06-15T00:00:00Z"`
	CurrentMileage     int32      `json:"current_mileage"              example:"158397"`
	NextServiceMileage *int32     `json:"next_service_mileage,omitempty" example:"166397"`
	NextServiceDate    *time.Time `json:"next_service_date,omitempty"  example:"2026-12-15T00:00:00Z"`
	OilType            *string    `json:"oil_type,omitempty"           example:"5W40"`
	OilFilter          *string    `json:"oil_filter,omitempty"         example:"1002"`
	NextOilFilter      *string    `json:"next_oil_filter,omitempty"    example:"1002"`
	CreatedAt          time.Time  `json:"created_at"                   example:"2026-06-15T07:47:45Z"`
}
