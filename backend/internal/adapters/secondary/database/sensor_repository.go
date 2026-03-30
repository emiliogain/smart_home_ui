// Package database contains database adapters (secondary adapters)
package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

// sensorRow maps sensors table columns for pgx.RowToStructByName (db tags match SQL identifiers).
type sensorRow struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Type        string     `db:"type"`
	Location    string     `db:"location"`
	Description string     `db:"description"`
	Status      string     `db:"status"`
	LastSeen    *time.Time `db:"last_seen"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	Config      []byte     `db:"config"`
}

type sensorConfigJSON struct {
	ReportingIntervalNs int64                  `json:"reporting_interval_ns"`
	Thresholds          map[string]interface{} `json:"thresholds"`
	Calibration         map[string]float64     `json:"calibration"`
	Enabled             bool                   `json:"enabled"`
}

func configToJSON(c sensor.Config) ([]byte, error) {
	th := c.Thresholds
	if th == nil {
		th = map[string]interface{}{}
	}
	cal := c.Calibration
	if cal == nil {
		cal = map[string]float64{}
	}
	return json.Marshal(sensorConfigJSON{
		ReportingIntervalNs: int64(c.ReportingInterval),
		Thresholds:          th,
		Calibration:         cal,
		Enabled:             c.Enabled,
	})
}

func configFromJSON(b []byte) (sensor.Config, error) {
	if len(b) == 0 {
		return sensor.Config{
			ReportingInterval: 30 * time.Second,
			Thresholds:        map[string]interface{}{},
			Calibration:       map[string]float64{},
			Enabled:           true,
		}, nil
	}
	var c sensorConfigJSON
	if err := json.Unmarshal(b, &c); err != nil {
		return sensor.Config{}, err
	}
	if c.Thresholds == nil {
		c.Thresholds = map[string]interface{}{}
	}
	if c.Calibration == nil {
		c.Calibration = map[string]float64{}
	}
	if c.ReportingIntervalNs == 0 {
		c.ReportingIntervalNs = int64(30 * time.Second)
	}
	return sensor.Config{
		ReportingInterval: time.Duration(c.ReportingIntervalNs),
		Thresholds:        c.Thresholds,
		Calibration:       c.Calibration,
		Enabled:           c.Enabled,
	}, nil
}

func rowToSensor(row sensorRow) (*sensor.Sensor, error) {
	cfg, err := configFromJSON(row.Config)
	if err != nil {
		return nil, err
	}
	return sensor.ReconstructSensor(
		row.ID, row.Name, sensor.SensorType(row.Type), row.Location, row.Description,
		sensor.Status(row.Status), row.LastSeen, row.CreatedAt, row.UpdatedAt, cfg,
	), nil
}

// sensorDataRow maps sensor_data table (column "timestamp" → db tag timestamp).
type sensorDataRow struct {
	ID        string    `db:"id"`
	SensorID  string    `db:"sensor_id"`
	Timestamp time.Time `db:"timestamp"`
	Values    []byte    `db:"values"`
	Quality   float64   `db:"quality"`
	CreatedAt time.Time `db:"created_at"`
}

func rowToData(row sensorDataRow) (*sensor.Data, error) {
	var vals map[string]interface{}
	if len(row.Values) > 0 {
		if err := json.Unmarshal(row.Values, &vals); err != nil {
			return nil, err
		}
	}
	if vals == nil {
		vals = map[string]interface{}{}
	}
	return sensor.ReconstructData(row.ID, row.SensorID, row.Timestamp, vals, row.Quality, row.CreatedAt), nil
}

const sensorSelectCols = `id, name, "type", location, description, status, last_seen, created_at, updated_at, config`

func (r *sensorRepository) Save(ctx context.Context, s *sensor.Sensor) error {
	cfgBytes, err := configToJSON(s.Config())
	if err != nil {
		return err
	}
	const q = `
INSERT INTO sensors (id, name, "type", location, description, status, last_seen, created_at, updated_at, config)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	"type" = EXCLUDED."type",
	location = EXCLUDED.location,
	description = EXCLUDED.description,
	status = EXCLUDED.status,
	last_seen = EXCLUDED.last_seen,
	updated_at = EXCLUDED.updated_at,
	config = EXCLUDED.config`
	_, err = r.pool.Exec(ctx, q,
		s.ID(),
		s.Name(),
		string(s.Type()),
		s.Location(),
		s.Description(),
		string(s.Status()),
		s.LastSeen(),
		s.CreatedAt(),
		s.UpdatedAt(),
		cfgBytes,
	)
	return err
}

func (r *sensorRepository) FindByID(ctx context.Context, id string) (*sensor.Sensor, error) {
	q := `SELECT ` + sensorSelectCols + ` FROM sensors WHERE id = $1`
	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sensorRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sensor.ErrSensorNotFound
		}
		return nil, err
	}
	return rowToSensor(row)
}

func (r *sensorRepository) FindAll(ctx context.Context) ([]*sensor.Sensor, error) {
	q := `SELECT ` + sensorSelectCols + ` FROM sensors ORDER BY id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[sensorRow])
	if err != nil {
		return nil, err
	}
	out := make([]*sensor.Sensor, 0, len(list))
	for _, row := range list {
		s, err := rowToSensor(row)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *sensorRepository) FindByType(ctx context.Context, sensorType sensor.SensorType) ([]*sensor.Sensor, error) {
	q := `SELECT ` + sensorSelectCols + ` FROM sensors WHERE "type" = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, q, string(sensorType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[sensorRow])
	if err != nil {
		return nil, err
	}
	out := make([]*sensor.Sensor, 0, len(list))
	for _, row := range list {
		s, err := rowToSensor(row)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *sensorRepository) FindByLocation(ctx context.Context, location string) ([]*sensor.Sensor, error) {
	q := `SELECT ` + sensorSelectCols + ` FROM sensors WHERE location = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, q, location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[sensorRow])
	if err != nil {
		return nil, err
	}
	out := make([]*sensor.Sensor, 0, len(list))
	for _, row := range list {
		s, err := rowToSensor(row)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *sensorRepository) Update(ctx context.Context, s *sensor.Sensor) error {
	cfgBytes, err := configToJSON(s.Config())
	if err != nil {
		return err
	}
	const q = `
UPDATE sensors SET
	name = $2,
	"type" = $3,
	location = $4,
	description = $5,
	status = $6,
	last_seen = $7,
	updated_at = $8,
	config = $9
WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q,
		s.ID(),
		s.Name(),
		string(s.Type()),
		s.Location(),
		s.Description(),
		string(s.Status()),
		s.LastSeen(),
		s.UpdatedAt(),
		cfgBytes,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return sensor.ErrSensorNotFound
	}
	return nil
}

func (r *sensorRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sensors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return sensor.ErrSensorNotFound
	}
	return nil
}

func (r *sensorRepository) Exists(ctx context.Context, id string) (bool, error) {
	var one int
	err := r.pool.QueryRow(ctx, `SELECT 1 FROM sensors WHERE id = $1 LIMIT 1`, id).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *sensorRepository) SaveData(ctx context.Context, data *sensor.Data) error {
	valuesJSON, err := json.Marshal(data.Values())
	if err != nil {
		return err
	}
	const q = `
INSERT INTO sensor_data (id, sensor_id, "timestamp", "values", quality, created_at)
VALUES ($1, $2, $3, $4::jsonb, $5, $6)
ON CONFLICT (id) DO UPDATE SET
	sensor_id = EXCLUDED.sensor_id,
	"timestamp" = EXCLUDED."timestamp",
	"values" = EXCLUDED."values",
	quality = EXCLUDED.quality,
	created_at = EXCLUDED.created_at`
	_, err = r.pool.Exec(ctx, q,
		data.ID(),
		data.SensorID(),
		data.Timestamp(),
		valuesJSON,
		data.Quality(),
		data.CreatedAt(),
	)
	return err
}

const sensorDataSelectCols = `id, sensor_id, "timestamp", "values", quality, created_at`

func (r *sensorRepository) FindDataByID(ctx context.Context, id string) (*sensor.Data, error) {
	q := `SELECT ` + sensorDataSelectCols + ` FROM sensor_data WHERE id = $1`
	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sensorDataRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sensor.ErrDataNotFound
		}
		return nil, err
	}
	return rowToData(row)
}

func appendSensorDataTimeFilter(base string, args []interface{}, tsCol string, from, to time.Time) (string, []interface{}) {
	n := len(args) + 1
	if !from.IsZero() {
		base += fmt.Sprintf(" AND %s >= $%d", tsCol, n)
		args = append(args, from)
		n++
	}
	if !to.IsZero() {
		base += fmt.Sprintf(" AND %s <= $%d", tsCol, n)
		args = append(args, to)
	}
	return base, args
}

func (r *sensorRepository) FindDataBySensor(ctx context.Context, sensorID string, from, to time.Time) ([]*sensor.Data, error) {
	q := `SELECT ` + sensorDataSelectCols + ` FROM sensor_data WHERE sensor_id = $1`
	args := []interface{}{sensorID}
	q, args = appendSensorDataTimeFilter(q, args, `"timestamp"`, from, to)
	q += ` ORDER BY "timestamp" ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[sensorDataRow])
	if err != nil {
		return nil, err
	}
	out := make([]*sensor.Data, 0, len(list))
	for _, row := range list {
		d, err := rowToData(row)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *sensorRepository) FindLatestData(ctx context.Context, sensorID string) (*sensor.Data, error) {
	q := `SELECT ` + sensorDataSelectCols + ` FROM sensor_data WHERE sensor_id = $1 ORDER BY "timestamp" DESC LIMIT 1`
	rows, err := r.pool.Query(ctx, q, sensorID)
	if err != nil {
		return nil, err
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sensorDataRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sensor.ErrDataNotFound
		}
		return nil, err
	}
	return rowToData(row)
}

func (r *sensorRepository) FindDataByLocation(ctx context.Context, location string, from, to time.Time) ([]*sensor.Data, error) {
	q := `
SELECT d.id, d.sensor_id, d."timestamp", d."values", d.quality, d.created_at
FROM sensor_data d
INNER JOIN sensors s ON s.id = d.sensor_id
WHERE s.location = $1`
	args := []interface{}{location}
	q, args = appendSensorDataTimeFilter(q, args, `d."timestamp"`, from, to)
	q += ` ORDER BY d."timestamp" ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[sensorDataRow])
	if err != nil {
		return nil, err
	}
	out := make([]*sensor.Data, 0, len(list))
	for _, row := range list {
		d, err := rowToData(row)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *sensorRepository) FindAggregatedData(ctx context.Context, sensorID string, from, to time.Time, _ time.Duration) ([]*sensor.Data, error) {
	return r.FindDataBySensor(ctx, sensorID, from, to)
}

func (r *sensorRepository) GetDataCount(ctx context.Context, sensorID string, from, to time.Time) (int64, error) {
	q := `SELECT COUNT(*) FROM sensor_data WHERE sensor_id = $1`
	args := []interface{}{sensorID}
	q, args = appendSensorDataTimeFilter(q, args, `"timestamp"`, from, to)
	var n int64
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *sensorRepository) DeleteOldData(ctx context.Context, before time.Time) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sensor_data WHERE "timestamp" < $1`, before)
	return err
}

func (r *sensorRepository) CalculateStorageSize(ctx context.Context, sensorID string) (int64, error) {
	const q = `SELECT COALESCE(COUNT(*) * 1024, 0)::bigint FROM sensor_data WHERE sensor_id = $1`
	var n int64
	if err := r.pool.QueryRow(ctx, q, sensorID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
