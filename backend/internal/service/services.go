package service

import (
	"time"

	db "github.com/lam-thinh/customer-oil-change-management/internal/db/sqlc"
)

// Config holds config parameters required by the services.
type Config struct {
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// Services holds all initialized service instances.
type Services struct {
	Auth      *AuthService
	Vehicle   *VehicleService
	OilChange *OilChangeService
}

// NewServices initializes and registers all application services.
func NewServices(queries *db.Queries, cfg Config) *Services {
	return &Services{
		Auth: NewAuthService(
			queries,
			cfg.JWTSecret,
			cfg.AccessTokenExpiry,
			cfg.RefreshTokenExpiry,
		),
		Vehicle:   NewVehicleService(queries),
		OilChange: NewOilChangeService(queries, queries),
	}
}
