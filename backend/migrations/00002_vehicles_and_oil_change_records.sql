-- +goose Up
-- +goose StatementBegin

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

CREATE UNIQUE INDEX idx_vehicles_license_plate ON vehicles(license_plate);
CREATE INDEX idx_vehicles_owner_name ON vehicles(owner_name);
CREATE INDEX idx_vehicles_phone_number ON vehicles(phone_number);

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

CREATE INDEX idx_oil_change_records_vehicle_id ON oil_change_records(vehicle_id);
CREATE INDEX idx_oil_change_records_service_date ON oil_change_records(service_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oil_change_records;
DROP TABLE IF EXISTS vehicles;
-- +goose StatementEnd
