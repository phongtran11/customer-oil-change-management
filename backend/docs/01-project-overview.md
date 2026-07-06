# Project Overview

## Customer Oil Change Management — Backend API

A production-ready Go REST API for managing customer oil change records. Built with a clean service-oriented architecture (Handlers → Services → Repositories).

---

## Technology Stack

| Layer | Library | Version |
|---|---|---|
| Routing | `go-chi/chi` | v5 |
| Database Driver | `jackc/pgx` | v5 |
| Data Access | `sqlc` | v1.27+ |
| Migrations | `pressly/goose` | v3 |
| Validation | `go-playground/validator` | v10 |
| Configuration | `spf13/viper` | v1 |
| Logging | `log/slog` | stdlib |
| Auth (JWT) | `golang-jwt/jwt` | v5 |
| Auth (Password) | `crypto/bcrypt` | stdlib |

---

## Directory Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── auth/                       # JWT, bcrypt, chi middleware
│   ├── config/                     # Viper config loader
│   ├── db/
│   │   ├── db.go                   # pgxpool connection helper
│   │   └── sqlc/                   # Generated repository layer
│   │       ├── models.go           # User, Session, Vehicle, OilChangeRecord
│   │       ├── query.sql.go        # All generated query functions
│   │       └── db.go
│   ├── dto/                        # Shared request/response structs
│   │   ├── auth_dto.go
│   │   ├── vehicle_dto.go
│   │   └── oil_change_dto.go
│   ├── handler/                    # HTTP controllers (handler logic only)
│   │   ├── auth_handler.go
│   │   ├── vehicle_handler.go
│   │   ├── oil_change_handler.go
│   │   └── response.go
│   ├── router/                     # Route registration + middleware wiring
│   │   └── router.go
│   └── service/                    # Business logic
│       ├── auth_service.go
│       ├── vehicle_service.go
│       └── oil_change_service.go
├── migrations/                     # Goose SQL migrations
│   ├── 00001_init.sql              # users + sessions
│   └── 00002_vehicles_and_oil_change_records.sql
├── sql/                            # sqlc schema + annotated queries
│   ├── schema.sql
│   └── query.sql
├── docs/                           # Project documentation (you are here)
├── docker-compose.yml              # Production Compose config
├── docker-compose.override.yml     # Dev override (auto-merged, uses Dockerfile.dev)
├── Dockerfile                      # Multi-stage production image
├── Dockerfile.dev                  # Dev image with air hot-reload
├── .air.toml                       # air watcher configuration
├── .env.example
├── .gitignore
├── go.mod
└── sqlc.yaml
```

---

## Architecture

```
HTTP Request
    │
    ▼
router.New()  ←  internal/router/router.go
(RequestID, Logger, Recoverer, Timeout)
    │
    ├── Public Routes
    │   POST /api/v1/register
    │   POST /api/v1/login
    │   POST /api/v1/refresh
    │
    └── Protected Routes (auth.Authenticator JWT middleware)
        POST   /api/v1/logout
        │
        GET    /api/v1/vehicles
        POST   /api/v1/vehicles
        GET    /api/v1/vehicles/{vehicleID}
        PUT    /api/v1/vehicles/{vehicleID}
        DELETE /api/v1/vehicles/{vehicleID}
        │
        POST   /api/v1/vehicles/{vehicleID}/oil-changes
        GET    /api/v1/vehicles/{vehicleID}/oil-changes
        GET    /api/v1/vehicles/{vehicleID}/oil-changes/latest
        GET    /api/v1/vehicles/{vehicleID}/oil-changes/{recordID}
        DELETE /api/v1/vehicles/{vehicleID}/oil-changes/{recordID}
    │
    ▼
Handlers  (internal/handler/)
    AuthHandler  ← dto.Auth*
    VehicleHandler  ← dto.Vehicle*
    OilChangeHandler  ← dto.OilChangeRecord*
    │
    ▼
Services  (internal/service/)
    AuthService
    VehicleService
    OilChangeService
    │
    ▼
db.Queries  (internal/db/sqlc/)  →  PostgreSQL
```

---

## Quick Start

See [02-getting-started.md](./02-getting-started.md) for setup instructions.

## API Reference

- Authentication: [03-api-auth.md](./03-api-auth.md)
- Vehicles: [07-api-vehicles.md](./07-api-vehicles.md)
- Oil Change Records: [08-api-oil-changes.md](./08-api-oil-changes.md)

## Database Schema

See [04-database-schema.md](./04-database-schema.md) for schema details.

## Configuration

See [05-configuration.md](./05-configuration.md) for all environment variables.

## Hot Reload (Development)

See [06-hot-reload-dev.md](./06-hot-reload-dev.md) for the Docker + air development workflow.

## Deployment (VPS)

See [09-deployment.md](./09-deployment.md) for GitHub Actions + GHCR + VPS deployment.
