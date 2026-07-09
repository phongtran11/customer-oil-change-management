# Database Schema

## Migration Tool

Migrations are managed by [pressly/goose](https://github.com/pressly/goose) and live in `migrations/`.  
They are applied **automatically on application startup when `APP_ENV=development`**. On production (`APP_ENV=production`), they must be run manually.

Migration file naming convention: `{5-digit-sequence}_{name}.sql`

---

## Tables

### `users`

Stores registered user accounts.

| Column | Type | Constraints | Default |
|---|---|---|---|
| `id` | `UUID` | PRIMARY KEY | `gen_random_uuid()` |
| `email` | `VARCHAR(255)` | UNIQUE, NOT NULL | — |
| `password_hash` | `VARCHAR(255)` | NOT NULL | — |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` |

**Triggers**

- `set_users_updated_at` — automatically updates `updated_at` to `NOW()` on every `UPDATE`.

---

### `sessions`

Stores active refresh token sessions. One user can have multiple concurrent sessions (e.g. multiple devices).

| Column | Type | Constraints | Default |
|---|---|---|---|
| `id` | `UUID` | PRIMARY KEY | `gen_random_uuid()` |
| `user_id` | `UUID` | NOT NULL, FK → `users(id)` ON DELETE CASCADE | — |
| `refresh_token` | `VARCHAR(512)` | UNIQUE, NOT NULL | — |
| `is_revoked` | `BOOLEAN` | NOT NULL | `FALSE` |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL | — |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` |

**Indexes**

| Index | Column | Reason |
|---|---|---|
| `idx_sessions_user_id` | `user_id` | Fast lookup of all sessions for a user |
| `idx_sessions_refresh_token` | `refresh_token` | Fast token validation on every `/refresh` and `/logout` |

**Foreign Keys**

- `user_id` references `users(id)` with `ON DELETE CASCADE` — deleting a user removes all their sessions.

---

### `vehicles`

Stores customer vehicle information.

| Column | Type | Constraints | Default |
|---|---|---|---|
| `id` | `UUID` | PRIMARY KEY | `gen_random_uuid()` |
| `license_plate` | `VARCHAR(50)` | NOT NULL | — |
| `owner_name` | `VARCHAR(255)` | NOT NULL | — |
| `phone_number` | `VARCHAR(20)` | NOT NULL | — |
| `make` | `VARCHAR(100)` | nullable | — |
| `model` | `VARCHAR(100)` | nullable | — |
| `year` | `INTEGER` | nullable | — |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` |

**Indexes**

| Index | Column | Reason |
|---|---|---|
| `idx_vehicles_license_plate` | `license_plate` | UNIQUE — enforces no duplicate plates |
| `idx_vehicles_owner_name` | `owner_name` | Search vehicles by owner name |
| `idx_vehicles_phone_number` | `phone_number` | Search vehicles by phone number |

> `make`, `model`, and `year` are **optional** — older records may not have this information.

---

### `oil_change_records`

Stores per-vehicle oil change service history.

| Column | Type | Constraints | Default |
|---|---|---|---|
| `id` | `UUID` | PRIMARY KEY | `gen_random_uuid()` |
| `vehicle_id` | `UUID` | NOT NULL, FK → `vehicles(id)` ON DELETE CASCADE | — |
| `service_date` | `TIMESTAMPTZ` | NOT NULL | — |
| `current_mileage` | `INTEGER` | NOT NULL | — |
| `next_service_mileage` | `INTEGER` | nullable | — |
| `next_service_date` | `TIMESTAMPTZ` | nullable | — |
| `oil_type` | `VARCHAR(50)` | nullable | — |
| `oil_filter` | `VARCHAR(100)` | nullable | — |
| `next_oil_filter` | `VARCHAR(100)` | nullable | — |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` |

**Indexes**

| Index | Column | Reason |
|---|---|---|
| `idx_oil_change_records_vehicle_id` | `vehicle_id` | List all records for a vehicle |
| `idx_oil_change_records_service_date` | `service_date` | Sort/filter by service date |

**Foreign Keys**

- `vehicle_id` references `vehicles(id)` with `ON DELETE CASCADE` — deleting a vehicle removes all its oil change records.

---

## Entity-Relationship Diagram

```
┌─────────────────────────┐        ┌──────────────────────────────────┐
│          users          │        │            sessions              │
├─────────────────────────┤        ├──────────────────────────────────┤
│ id           UUID  (PK) │◄──┐    │ id           UUID  (PK)         │
│ email        VARCHAR    │   └────│ user_id      UUID  (FK)         │
│ password_hash VARCHAR   │        │ refresh_token VARCHAR (UNIQUE)   │
│ created_at   TIMESTAMPTZ│        │ is_revoked   BOOLEAN             │
│ updated_at   TIMESTAMPTZ│        │ expires_at   TIMESTAMPTZ         │
└─────────────────────────┘        │ created_at   TIMESTAMPTZ         │
                                   └──────────────────────────────────┘

┌──────────────────────────┐        ┌──────────────────────────────────┐
│         vehicles         │        │        oil_change_records        │
├──────────────────────────┤        ├──────────────────────────────────┤
│ id           UUID  (PK)  │◄──┐    │ id                UUID  (PK)    │
│ license_plate VARCHAR    │   └────│ vehicle_id        UUID  (FK)    │
│ owner_name   VARCHAR     │        │ service_date      TIMESTAMPTZ   │
│ phone_number VARCHAR     │        │ current_mileage   INTEGER       │
│ make         VARCHAR     │        │ next_service_mileage INTEGER    │
│ model        VARCHAR     │        │ next_service_date TIMESTAMPTZ   │
│ year         INTEGER     │        │ oil_type          VARCHAR       │
│ created_at   TIMESTAMPTZ │        │ oil_filter        VARCHAR       │
└──────────────────────────┘        │ next_oil_filter   VARCHAR       │
                                    │ created_at        TIMESTAMPTZ   │
                                    └──────────────────────────────────┘
```

---

## Migration Files

| File | Description |
|---|---|
| `00001_init.sql` | Creates `users` and `sessions` tables with triggers and indexes |
| `00002_vehicles_and_oil_change_records.sql` | Creates `vehicles` and `oil_change_records` tables with indexes |

---

## Running Migrations

**Up (apply all pending)**

```bash
goose -dir migrations postgres "$DB_URL" up
```

**Down (rollback last)**

```bash
goose -dir migrations postgres "$DB_URL" down
```

**Status**

```bash
goose -dir migrations postgres "$DB_URL" status
```

---

## Adding a New Migration

```bash
goose -dir migrations create add_new_feature sql
# Creates: migrations/00003_add_new_feature.sql
```

Then fill in the `-- +goose Up` and `-- +goose Down` sections.
