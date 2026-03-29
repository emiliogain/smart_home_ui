-- +goose Up
CREATE TABLE sensors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sensor_type TEXT NOT NULL,
    location TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    last_seen TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE sensor_data (
    id TEXT PRIMARY KEY,
    sensor_id TEXT NOT NULL REFERENCES sensors (id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL,
    values JSONB NOT NULL,
    quality DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sensor_data_sensor_recorded_at ON sensor_data (sensor_id, recorded_at DESC);

-- +goose Down
DROP TABLE IF EXISTS sensor_data;
DROP TABLE IF EXISTS sensors;
