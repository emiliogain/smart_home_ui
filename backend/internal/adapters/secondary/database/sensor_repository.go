package database

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type sensorRepository struct {
	pool *pgxpool.Pool
}

// NewSensorRepository creates a PostgreSQL-backed sensor repository.
func NewSensorRepository(pool *pgxpool.Pool) secondary.SensorRepository {
	return &sensorRepository{pool: pool}
}

func (r *sensorRepository) SaveSensor(ctx context.Context, s sensor.Sensor) error {
	q, args, err := psq.
		Insert("sensors").
		Columns("id", "name", "type", "location", "status", "created_at", "updated_at").
		Values(s.ID, s.Name, string(s.Type), s.Location, s.Status, s.CreatedAt, s.UpdatedAt).
		Suffix(`ON CONFLICT (id) DO UPDATE SET
			name       = EXCLUDED.name,
			type       = EXCLUDED.type,
			location   = EXCLUDED.location,
			status     = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}
	_, err = r.pool.Exec(ctx, q, args...)
	return err
}

func (r *sensorRepository) GetSensor(ctx context.Context, id string) (*sensor.Sensor, error) {
	q, args, err := psq.
		Select("id", "name", "type", "location", "status", "created_at", "updated_at").
		From("sensors").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var s sensor.Sensor
	var sType string
	err = r.pool.QueryRow(ctx, q, args...).
		Scan(&s.ID, &s.Name, &sType, &s.Location, &s.Status, &s.CreatedAt, &s.UpdatedAt)
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
	q, args, err := psq.
		Select("id", "name", "type", "location", "status", "created_at", "updated_at").
		From("sensors").
		OrderBy("created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, q, args...)
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
	q, args, err := psq.
		Insert("sensor_readings").
		Columns("id", "sensor_id", "value", "unit", "timestamp").
		Values(rd.ID, rd.SensorID, rd.Value, rd.Unit, rd.Timestamp).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}
	_, err = r.pool.Exec(ctx, q, args...)
	return err
}

func (r *sensorRepository) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]sensor.Reading, error) {
	q, args, err := psq.
		Select("id", "sensor_id", "value", "unit", "timestamp").
		From("sensor_readings").
		Where(sq.Eq{"sensor_id": sensorID}).
		OrderBy("timestamp DESC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, q, args...)
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
