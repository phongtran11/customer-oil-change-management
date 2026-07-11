package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/lam-thinh/customer-oil-change-management/internal/auth"
	"github.com/lam-thinh/customer-oil-change-management/internal/handler"
	"github.com/lam-thinh/customer-oil-change-management/internal/logger"
)

// New builds and returns the fully-configured chi router.
// It owns all middleware setup and route registration so that main.go
// stays minimal and testable.
func New(h *handler.Handlers, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	// ── Global Middleware ─────────────────────────────────────────────────────
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.Middleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Swagger UI
	r.Get("/api/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/api/swagger/doc.json"),
		httpSwagger.BeforeScript(`const UrlMutatorPlugin = (system) => ({
			rootInjects: {
				setScheme: (scheme) => {
				const jsonSpec = system.getState().toJSON().spec.json;
				const schemes = Array.isArray(scheme) ? scheme : [scheme];
				const newJsonSpec = Object.assign({}, jsonSpec, { schemes });

				return system.specActions.updateJsonSpec(newJsonSpec);
				},
				setHost: (host) => {
				const jsonSpec = system.getState().toJSON().spec.json;
				const newJsonSpec = Object.assign({}, jsonSpec, { host });

				return system.specActions.updateJsonSpec(newJsonSpec);
				},
				setBasePath: (basePath) => {
				const jsonSpec = system.getState().toJSON().spec.json;
				const newJsonSpec = Object.assign({}, jsonSpec, { basePath });

				return system.specActions.updateJsonSpec(newJsonSpec);
				}
			}
			});`),
		httpSwagger.Plugins([]string{"UrlMutatorPlugin"}),
		httpSwagger.UIConfig(map[string]string{
			"onComplete": `() => {
				const scheme = window.location.protocol.replace(':', '');
				const host = window.location.host;
				const pathname = window.location.pathname;
				const basePath = pathname.substring(0, pathname.indexOf("/swagger"));
				window.ui.setScheme(scheme);
				window.ui.setHost(host);
				window.ui.setBasePath(basePath);
			}`,
	}),
	))

	// ── API Routes (Versioned) ────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// ── Public ──
		r.Post("/register", h.Auth.Register)
		r.Post("/login", h.Auth.Login)
		r.Post("/refresh", h.Auth.Refresh)

		// ── Protected ──
		r.Group(func(r chi.Router) {
			r.Use(auth.Authenticator(jwtSecret))

			r.Post("/logout", h.Auth.Logout)

			// ── Vehicles ──
			r.Get("/vehicles", h.Vehicle.ListVehicles)
			r.Post("/vehicles", h.Vehicle.CreateVehicle)
			r.Get("/vehicles/{vehicleID}", h.Vehicle.GetVehicle)
			r.Put("/vehicles/{vehicleID}", h.Vehicle.UpdateVehicle)
			r.Delete("/vehicles/{vehicleID}", h.Vehicle.DeleteVehicle)

			// ── Oil Change Records ──
			r.Post("/vehicles/{vehicleID}/oil-changes", h.OilChange.CreateOilChangeRecord)
			r.Get("/vehicles/{vehicleID}/oil-changes", h.OilChange.ListOilChangeRecords)
			r.Get("/vehicles/{vehicleID}/oil-changes/latest", h.OilChange.GetLatestOilChangeRecord)
			r.Get("/vehicles/{vehicleID}/oil-changes/{recordID}", h.OilChange.GetOilChangeRecord)
			r.Delete("/vehicles/{vehicleID}/oil-changes/{recordID}", h.OilChange.DeleteOilChangeRecord)
		})
	})

	return r
}
