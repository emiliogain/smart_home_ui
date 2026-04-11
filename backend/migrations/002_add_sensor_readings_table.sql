-- +goose Up
-- Adds a sensor_readings table for scalar readings (value + unit).
-- The existing sensor_data table stores JSONB envelopes; this table
-- is used by the Go repository layer for individual typed readings.

CREATE TABLE IF NOT EXISTS sensor_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sensor_id UUID NOT NULL REFERENCES sensors (id) ON DELETE CASCADE,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50) NOT NULL DEFAULT '',
    "timestamp" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sensor_readings_sensor_time
    ON sensor_readings (sensor_id, "timestamp" DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_sensor_readings_sensor_time;
DROP TABLE IF EXISTS sensor_readings;
