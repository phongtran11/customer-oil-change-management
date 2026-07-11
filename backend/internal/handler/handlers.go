package handler

import (
	"github.com/lam-thinh/customer-oil-change-management/internal/service"
)

// Config holds configuration parameters required by the handlers.
type Config struct {
	IsProd bool
}

// Handlers bundles all HTTP handler dependencies.
type Handlers struct {
	Auth      *AuthHandler
	Vehicle   *VehicleHandler
	OilChange *OilChangeHandler
}

// NewHandlers initializes and registers all application handlers.
func NewHandlers(svcs *service.Services, cfg Config) *Handlers {
	return &Handlers{
		Auth:      NewAuthHandler(svcs.Auth, cfg.IsProd),
		Vehicle:   NewVehicleHandler(svcs.Vehicle),
		OilChange: NewOilChangeHandler(svcs.OilChange),
	}
}
