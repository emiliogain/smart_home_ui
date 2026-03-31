package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sensorRepository struct {
	pool *pgxpool.Pool
}

// NewSensorRepository creates a PostgreSQL-backed sensor repository.
func NewSensorRepository(pool *pgxpool.Pool) secondary.SensorRepository {
	return &sensorRepository{pool: pool}
}

func (r *sensorRepository) SaveSensor(ctx context.Context, s sensor.Sensor) error {
	const q = `
INSERT INTO sensors (id, name, type, location, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
	name       = EXCLUDED.name,
	type       = EXCLUDED.type,
	location   = EXCLUDED.location,
	status     = EXCLUDED.status,
	updated_at = EXCLUDED.updated_at`
	_, err := r.pool.Exec(ctx, q,
		s.ID, s.Name, string(s.Type), s.Location, s.Status, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *sensorRepository) GetSensor(ctx context.Context, id string) (*sensor.Sensor, error) {
	const q = `SELECT id, name, type, location, status, created_at, updated_at FROM sensors WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)

	var s sensor.Sensor
	var sType string
	err := row.Scan(&s.ID, &s.Name, &sType, &s.Location, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("sensor %s not found", id)
		}
		return nil, err
	}
	s.Type = sensor.SensorType(sType)
	return &s, nil
}

func (r *sensorRepository) ListSensors(ctx context.Context) ([]sensor.Sensor, error) {
	const q = `SELECT id, name, type, location, status, created_at, updated_at FROM sensors ORDER BY created_at`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []sensor.Sensor
	for rows.Next() {
		var s sensor.Sensor
		var sType string
		if err := rows.Scan(&s.ID, &s.Name, &sType, &s.Location, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Type = sensor.SensorType(sType)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *sensorRepository) SaveReading(ctx context.Context, rd sensor.Reading) error {
	const q = `
INSERT INTO sensor_readings (id, sensor_id, value, unit, timestamp)
VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, q, rd.ID, rd.SensorID, rd.Value, rd.Unit, rd.Timestamp)
	return err
}

func (r *sensorRepository) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]sensor.Reading, error) {
	const q = `
SELECT id, sensor_id, value, unit, timestamp
FROM sensor_readings
WHERE sensor_id = $1
ORDER BY timestamp DESC
LIMIT $2`
	rows, err := r.pool.Query(ctx, q, sensorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []sensor.Reading
	for rows.Next() {
		var rd sensor.Reading
		if err := rows.Scan(&rd.ID, &rd.SensorID, &rd.Value, &rd.Unit, &rd.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, rd)
	}
	return out, rows.Err()
}
