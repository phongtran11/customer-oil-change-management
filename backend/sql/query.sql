-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE refresh_token = $1
LIMIT 1;

-- name: UpdateSessionRevoked :exec
UPDATE sessions
SET is_revoked = TRUE
WHERE refresh_token = $1;

-- name: DeleteAllSessionsForUser :exec
DELETE FROM sessions
WHERE user_id = $1;

-- ── Vehicles ──────────────────────────────────────────────────────────────────

-- name: CreateVehicle :one
INSERT INTO vehicles (license_plate, owner_name, phone_number, make, model, year)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetVehicleByID :one
SELECT * FROM vehicles
WHERE id = $1
LIMIT 1;

-- name: GetVehicleByLicensePlate :one
SELECT * FROM vehicles
WHERE license_plate = $1
LIMIT 1;

-- name: ListVehicles :many
SELECT * FROM vehicles
ORDER BY created_at DESC;

-- name: UpdateVehicle :one
UPDATE vehicles
SET
    license_plate = $2,
    owner_name    = $3,
    phone_number  = $4,
    make          = $5,
    model         = $6,
    year          = $7
WHERE id = $1
RETURNING *;

-- name: DeleteVehicle :exec
DELETE FROM vehicles
WHERE id = $1;

-- ── Oil Change Records ────────────────────────────────────────────────────────

-- name: CreateOilChangeRecord :one
INSERT INTO oil_change_records (
    vehicle_id,
    service_date,
    current_mileage,
    next_service_mileage,
    next_service_date,
    oil_type,
    oil_filter,
    next_oil_filter
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetOilChangeRecordByID :one
SELECT * FROM oil_change_records
WHERE id = $1
LIMIT 1;

-- name: ListOilChangeRecordsByVehicle :many
SELECT * FROM oil_change_records
WHERE vehicle_id = $1
ORDER BY service_date DESC;

-- name: GetLatestOilChangeRecord :one
SELECT * FROM oil_change_records
WHERE vehicle_id = $1
ORDER BY service_date DESC
LIMIT 1;

-- name: DeleteOilChangeRecord :exec
DELETE FROM oil_change_records
WHERE id = $1;
