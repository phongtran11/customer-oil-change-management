CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) UNIQUE NOT NULL,
    is_revoked    BOOLEAN      NOT NULL DEFAULT FALSE,
    expires_at    TIMESTAMPTZ  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vehicles (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    license_plate VARCHAR(50)  NOT NULL,
    owner_name    VARCHAR(255) NOT NULL,
    phone_number  VARCHAR(20)  NOT NULL,
    make          VARCHAR(100),
    model         VARCHAR(100),
    year          INTEGER,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oil_change_records (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id           UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    service_date         TIMESTAMPTZ  NOT NULL,
    current_mileage      INTEGER      NOT NULL,
    next_service_mileage INTEGER,
    next_service_date    TIMESTAMPTZ,
    oil_type             VARCHAR(50),
    oil_filter           VARCHAR(100),
    next_oil_filter      VARCHAR(100),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
