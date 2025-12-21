-- TelecomBase: initial schema

CREATE TABLE IF NOT EXISTS users (
    id           BIGSERIAL PRIMARY KEY,
    username     TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role         TEXT NOT NULL DEFAULT 'user',
    approved     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vendors (
    id      BIGSERIAL PRIMARY KEY,
    name    TEXT NOT NULL,
    country TEXT
);

CREATE TABLE IF NOT EXISTS models (
    id         BIGSERIAL PRIMARY KEY,
    vendor_id  BIGINT NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
    name       TEXT NOT NULL,
    device_type TEXT
);

CREATE TABLE IF NOT EXISTS locations (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    note TEXT
);

CREATE TABLE IF NOT EXISTS devices (
    id               BIGSERIAL PRIMARY KEY,
    model_id         BIGINT NOT NULL REFERENCES models(id) ON DELETE RESTRICT,
    location_id      BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    serial_number    TEXT,
    inventory_number TEXT,
    status           TEXT NOT NULL DEFAULT 'active',
    installed_at     DATE,
    description      TEXT,
    CONSTRAINT devices_serial_unique UNIQUE (serial_number)
);

CREATE INDEX IF NOT EXISTS idx_devices_inventory_number ON devices(inventory_number);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
