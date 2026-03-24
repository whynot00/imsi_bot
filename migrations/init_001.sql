-- Full schema migration
-- Apply: psql -d <dbname> -f 001_init.sql

BEGIN;

-- -----------------------------------------------------------------
-- devices
-- -----------------------------------------------------------------
CREATE TABLE devices (
    id   BIGSERIAL PRIMARY KEY,
    imsi TEXT,
    imei TEXT,
    CONSTRAINT uq_devices_imsi_imei UNIQUE (imsi, imei)
);

CREATE INDEX idx_devices_imsi         ON devices (imsi)         WHERE imsi IS NOT NULL;
CREATE INDEX idx_devices_imei         ON devices (imei)         WHERE imei IS NOT NULL;
CREATE INDEX idx_devices_imsi_prefix  ON devices (imsi text_pattern_ops) WHERE imsi IS NOT NULL;
CREATE INDEX idx_devices_imei_prefix  ON devices (imei text_pattern_ops) WHERE imei IS NOT NULL;


-- -----------------------------------------------------------------
-- locations_parametr
-- -----------------------------------------------------------------
CREATE TABLE locations_parametr (
    id      BIGSERIAL        PRIMARY KEY,
    seen_at TIMESTAMP        NOT NULL,
    lat     DOUBLE PRECISION NOT NULL,
    lon     DOUBLE PRECISION NOT NULL,
    CONSTRAINT uq_locations_parametr UNIQUE (seen_at, lat, lon)
);

CREATE INDEX idx_locations_parametr_seen_at ON locations_parametr (seen_at);


-- -----------------------------------------------------------------
-- sightings_parametr
-- -----------------------------------------------------------------
CREATE TABLE sightings_parametr (
    id          BIGSERIAL PRIMARY KEY,
    device_id   BIGINT    NOT NULL REFERENCES devices (id),
    seen_at     TIMESTAMP NOT NULL,
    standart    TEXT,
    operator    TEXT,
    event       TEXT,
    location_id BIGINT    REFERENCES locations_parametr (id),
    CONSTRAINT uq_sightings_parametr UNIQUE (device_id, seen_at)
);

CREATE INDEX idx_sightings_parametr_device_id  ON sightings_parametr (device_id);
CREATE INDEX idx_sightings_parametr_seen_at    ON sightings_parametr (seen_at);
CREATE INDEX idx_sightings_parametr_location_id ON sightings_parametr (location_id);


-- -----------------------------------------------------------------
-- sightings_rk
-- -----------------------------------------------------------------
CREATE TABLE sightings_rk (
    id        BIGSERIAL        PRIMARY KEY,
    device_id BIGINT           NOT NULL REFERENCES devices (id),
    seen_at   TIMESTAMP        NOT NULL,
    standart  TEXT,
    lat       DOUBLE PRECISION,
    lon       DOUBLE PRECISION,
    signal    INTEGER,
    CONSTRAINT uq_sightings_rk UNIQUE (device_id, seen_at)
);

CREATE INDEX idx_sightings_rk_device_id ON sightings_rk (device_id);
CREATE INDEX idx_sightings_rk_seen_at   ON sightings_rk (seen_at);


-- -----------------------------------------------------------------
-- users
-- -----------------------------------------------------------------
CREATE TABLE users (
    id         BIGINT    PRIMARY KEY, -- telegram user_id
    username   TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMIT;