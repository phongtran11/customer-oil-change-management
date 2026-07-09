package dto

import "time"

// ── Vehicle Request DTOs ──────────────────────────────────────────────────────

// CreateVehicleRequest is the request body for POST /vehicles.
type CreateVehicleRequest struct {
	LicensePlate string  `json:"license_plate" validate:"required,max=50"  example:"72A429.14"`
	OwnerName    string  `json:"owner_name"    validate:"required,max=255" example:"Nguyễn Văn A"`
	PhoneNumber  string  `json:"phone_number"  validate:"required,max=20"  example:"0981811837"`
	Make         *string `json:"make,omitempty"                            example:"Toyota"`
	Model        *string `json:"model,omitempty"                           example:"Vios"`
	Year         *int32  `json:"year,omitempty"                            example:"2022"`
}

// UpdateVehicleRequest is the request body for PUT /vehicles/{vehicleID}.
type UpdateVehicleRequest struct {
	LicensePlate string  `json:"license_plate" validate:"required,max=50"  example:"72A429.14"`
	OwnerName    string  `json:"owner_name"    validate:"required,max=255" example:"Nguyễn Văn A"`
	PhoneNumber  string  `json:"phone_number"  validate:"required,max=20"  example:"0981811837"`
	Make         *string `json:"make,omitempty"                            example:"Toyota"`
	Model        *string `json:"model,omitempty"                           example:"Vios"`
	Year         *int32  `json:"year,omitempty"                            example:"2022"`
}

// ── Vehicle Response DTOs ─────────────────────────────────────────────────────

// VehicleResponse is the representation of a vehicle returned by the API.
type VehicleResponse struct {
	ID           string    `json:"id"            example:"1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a"`
	LicensePlate string    `json:"license_plate" example:"72A429.14"`
	OwnerName    string    `json:"owner_name"    example:"Nguyễn Văn A"`
	PhoneNumber  string    `json:"phone_number"  example:"0981811837"`
	Make         *string   `json:"make,omitempty" example:"Toyota"`
	Model        *string   `json:"model,omitempty" example:"Vios"`
	Year         *int32    `json:"year,omitempty"  example:"2022"`
	CreatedAt    time.Time `json:"created_at"  example:"2026-06-29T07:45:51Z"`
}
