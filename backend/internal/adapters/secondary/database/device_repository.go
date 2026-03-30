package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/device"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceRepository struct {
	pool *pgxpool.Pool
}

// NewDeviceRepository creates a PostgreSQL-backed device repository.
func NewDeviceRepository(pool *pgxpool.Pool) secondary.DeviceRepository {
	return &deviceRepository{pool: pool}
}

// deviceRow maps devices table columns for pgx.RowToStructByName.
type deviceRow struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Type         string     `db:"type"`
	Location     string     `db:"location"`
	Description  string     `db:"description"`
	Status       string     `db:"status"`
	PowerState   string     `db:"power_state"`
	LastSeen     *time.Time `db:"last_seen"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	State        []byte     `db:"state"`
	Capabilities []byte     `db:"capabilities"`
}

const deviceSelectCols = `id, name, "type", location, description, status, power_state, last_seen, created_at, updated_at, state, capabilities`

func stateFromJSON(b []byte) (device.State, error) {
	if len(b) == 0 {
		return device.State{Properties: map[string]interface{}{}}, nil
	}
	var s device.State
	if err := json.Unmarshal(b, &s); err != nil {
		return device.State{}, err
	}
	if s.Properties == nil {
		s.Properties = map[string]interface{}{}
	}
	return s, nil
}

func capabilitiesFromJSON(b []byte) (device.Capabilities, error) {
	if len(b) == 0 {
		return device.Capabilities{}, nil
	}
	var c device.Capabilities
	if err := json.Unmarshal(b, &c); err != nil {
		return device.Capabilities{}, err
	}
	if c.SupportedCommands == nil {
		c.SupportedCommands = []string{}
	}
	if c.Properties == nil {
		c.Properties = []device.PropertyDefinition{}
	}
	return c, nil
}

func rowToDevice(row deviceRow) (*device.Device, error) {
	st, err := stateFromJSON(row.State)
	if err != nil {
		return nil, err
	}
	cap, err := capabilitiesFromJSON(row.Capabilities)
	if err != nil {
		return nil, err
	}
	return device.ReconstructDevice(
		row.ID, row.Name, device.Type(row.Type), row.Location, row.Description,
		device.Status(row.Status), device.PowerState(row.PowerState),
		row.LastSeen, row.CreatedAt, row.UpdatedAt, st, cap,
	), nil
}

func (r *deviceRepository) Save(ctx context.Context, d *device.Device) error {
	stateJSON, err := json.Marshal(d.State())
	if err != nil {
		return err
	}
	capJSON, err := json.Marshal(d.Capabilities())
	if err != nil {
		return err
	}
	const q = `
INSERT INTO devices (id, name, "type", location, description, status, power_state, last_seen, created_at, updated_at, state, capabilities)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	"type" = EXCLUDED."type",
	location = EXCLUDED.location,
	description = EXCLUDED.description,
	status = EXCLUDED.status,
	power_state = EXCLUDED.power_state,
	last_seen = EXCLUDED.last_seen,
	updated_at = EXCLUDED.updated_at,
	state = EXCLUDED.state,
	capabilities = EXCLUDED.capabilities`
	_, err = r.pool.Exec(ctx, q,
		d.ID(),
		d.Name(),
		string(d.Type()),
		d.Location(),
		d.Description(),
		string(d.Status()),
		string(d.PowerState()),
		d.LastSeen(),
		d.CreatedAt(),
		d.UpdatedAt(),
		stateJSON,
		capJSON,
	)
	return err
}

func (r *deviceRepository) FindByID(ctx context.Context, id string) (*device.Device, error) {
	q := `SELECT ` + deviceSelectCols + ` FROM devices WHERE id = $1`
	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[deviceRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, device.ErrDeviceNotFound
		}
		return nil, err
	}
	return rowToDevice(row)
}

func (r *deviceRepository) FindAll(ctx context.Context) ([]*device.Device, error) {
	q := `SELECT ` + deviceSelectCols + ` FROM devices ORDER BY id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[deviceRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Device, 0, len(list))
	for _, row := range list {
		d, err := rowToDevice(row)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *deviceRepository) FindByType(ctx context.Context, deviceType device.Type) ([]*device.Device, error) {
	q := `SELECT ` + deviceSelectCols + ` FROM devices WHERE "type" = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, q, string(deviceType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[deviceRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Device, 0, len(list))
	for _, row := range list {
		d, err := rowToDevice(row)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *deviceRepository) FindByLocation(ctx context.Context, location string) ([]*device.Device, error) {
	q := `SELECT ` + deviceSelectCols + ` FROM devices WHERE location = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, q, location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[deviceRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Device, 0, len(list))
	for _, row := range list {
		d, err := rowToDevice(row)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *deviceRepository) Update(ctx context.Context, d *device.Device) error {
	stateJSON, err := json.Marshal(d.State())
	if err != nil {
		return err
	}
	capJSON, err := json.Marshal(d.Capabilities())
	if err != nil {
		return err
	}
	const q = `
UPDATE devices SET
	name = $2,
	"type" = $3,
	location = $4,
	description = $5,
	status = $6,
	power_state = $7,
	last_seen = $8,
	updated_at = $9,
	state = $10,
	capabilities = $11
WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q,
		d.ID(),
		d.Name(),
		string(d.Type()),
		d.Location(),
		d.Description(),
		string(d.Status()),
		string(d.PowerState()),
		d.LastSeen(),
		d.UpdatedAt(),
		stateJSON,
		capJSON,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return device.ErrDeviceNotFound
	}
	return nil
}

func (r *deviceRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return device.ErrDeviceNotFound
	}
	return nil
}

func (r *deviceRepository) Exists(ctx context.Context, id string) (bool, error) {
	var one int
	err := r.pool.QueryRow(ctx, `SELECT 1 FROM devices WHERE id = $1 LIMIT 1`, id).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *deviceRepository) SaveState(ctx context.Context, deviceID string, state *device.State) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, `UPDATE devices SET state = $2, updated_at = NOW() WHERE id = $1`, deviceID, b)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return device.ErrDeviceNotFound
	}
	return nil
}

func (r *deviceRepository) FindState(ctx context.Context, deviceID string) (*device.State, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT state FROM devices WHERE id = $1`, deviceID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, device.ErrDeviceNotFound
		}
		return nil, err
	}
	st, err := stateFromJSON(raw)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// commandRow maps device_commands (SQL column "error" → db tag).
type commandRow struct {
	ID         string     `db:"id"`
	DeviceID   string     `db:"device_id"`
	Command    string     `db:"command"`
	Parameters []byte     `db:"parameters"`
	Status     string     `db:"status"`
	Result     []byte     `db:"result"`
	ErrMsg     *string    `db:"error"`
	CreatedAt  time.Time  `db:"created_at"`
	ExecutedAt *time.Time `db:"executed_at"`
}

const commandSelectCols = `id, device_id, command, parameters, status, result, "error", created_at, executed_at`

func rowToCommand(row commandRow) (*device.Command, error) {
	var params map[string]interface{}
	if len(row.Parameters) > 0 {
		if err := json.Unmarshal(row.Parameters, &params); err != nil {
			return nil, err
		}
	}
	if params == nil {
		params = map[string]interface{}{}
	}
	var res map[string]interface{}
	if len(row.Result) > 0 {
		if err := json.Unmarshal(row.Result, &res); err != nil {
			return nil, err
		}
	}
	errStr := ""
	if row.ErrMsg != nil {
		errStr = *row.ErrMsg
	}
	return device.ReconstructCommand(
		row.ID, row.DeviceID, row.Command, params,
		device.CommandStatus(row.Status), res, errStr,
		row.CreatedAt, row.ExecutedAt,
	), nil
}

func (r *deviceRepository) SaveCommand(ctx context.Context, cmd *device.Command) error {
	paramsJSON, err := json.Marshal(cmd.Parameters())
	if err != nil {
		return err
	}
	var resultJSON interface{}
	if cmd.Result() != nil {
		b, err := json.Marshal(cmd.Result())
		if err != nil {
			return err
		}
		resultJSON = b
	}
	errMsg := cmd.Error()
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	const q = `
INSERT INTO device_commands (id, device_id, command, parameters, status, result, "error", created_at, executed_at)
VALUES ($1, $2, $3, $4::jsonb, $5, $6::jsonb, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
	device_id = EXCLUDED.device_id,
	command = EXCLUDED.command,
	parameters = EXCLUDED.parameters,
	status = EXCLUDED.status,
	result = EXCLUDED.result,
	"error" = EXCLUDED."error",
	executed_at = EXCLUDED.executed_at`
	_, err = r.pool.Exec(ctx, q,
		cmd.ID(),
		cmd.DeviceID(),
		cmd.Command(),
		paramsJSON,
		string(cmd.Status()),
		resultJSON,
		errPtr,
		cmd.CreatedAt(),
		cmd.ExecutedAt(),
	)
	return err
}

func (r *deviceRepository) FindCommandByID(ctx context.Context, id string) (*device.Command, error) {
	q := `SELECT ` + commandSelectCols + ` FROM device_commands WHERE id = $1`
	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[commandRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, device.ErrCommandNotFound
		}
		return nil, err
	}
	return rowToCommand(row)
}

func (r *deviceRepository) FindCommandsByDevice(ctx context.Context, deviceID string, limit int) ([]*device.Command, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT ` + commandSelectCols + ` FROM device_commands WHERE device_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[commandRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Command, 0, len(list))
	for _, row := range list {
		c, err := rowToCommand(row)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *deviceRepository) FindPendingCommands(ctx context.Context) ([]*device.Command, error) {
	q := `SELECT ` + commandSelectCols + ` FROM device_commands WHERE status = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, string(device.CommandPending))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[commandRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Command, 0, len(list))
	for _, row := range list {
		c, err := rowToCommand(row)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *deviceRepository) UpdateCommand(ctx context.Context, cmd *device.Command) error {
	paramsJSON, err := json.Marshal(cmd.Parameters())
	if err != nil {
		return err
	}
	var resultJSON interface{}
	if cmd.Result() != nil {
		b, err := json.Marshal(cmd.Result())
		if err != nil {
			return err
		}
		resultJSON = b
	}
	errMsg := cmd.Error()
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	const q = `
UPDATE device_commands SET
	device_id = $2,
	command = $3,
	parameters = $4::jsonb,
	status = $5,
	result = $6::jsonb,
	"error" = $7,
	executed_at = $8
WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q,
		cmd.ID(),
		cmd.DeviceID(),
		cmd.Command(),
		paramsJSON,
		string(cmd.Status()),
		resultJSON,
		errPtr,
		cmd.ExecutedAt(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return device.ErrCommandNotFound
	}
	return nil
}

func (r *deviceRepository) DeleteCommand(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM device_commands WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return device.ErrCommandNotFound
	}
	return nil
}

type eventRow struct {
	ID        string    `db:"id"`
	DeviceID  string    `db:"device_id"`
	EventType string    `db:"event_type"`
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}

const eventSelectCols = `id, device_id, event_type, data, created_at`

func rowToEvent(row eventRow) (*device.Event, error) {
	var data map[string]interface{}
	if len(row.Data) > 0 {
		if err := json.Unmarshal(row.Data, &data); err != nil {
			return nil, err
		}
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	return &device.Event{
		ID:        row.ID,
		DeviceID:  row.DeviceID,
		Type:      row.EventType,
		Data:      data,
		Timestamp: row.CreatedAt,
	}, nil
}

func (r *deviceRepository) SaveEvent(ctx context.Context, event *device.Event) error {
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	const q = `
INSERT INTO device_events (id, device_id, event_type, data, created_at)
VALUES ($1, $2, $3, $4::jsonb, $5)
ON CONFLICT (id) DO UPDATE SET
	device_id = EXCLUDED.device_id,
	event_type = EXCLUDED.event_type,
	data = EXCLUDED.data,
	created_at = EXCLUDED.created_at`
	_, err = r.pool.Exec(ctx, q, event.ID, event.DeviceID, event.Type, dataJSON, event.Timestamp)
	return err
}

func appendEventTimeFilter(base string, args []interface{}, tsCol string, from, to time.Time) (string, []interface{}) {
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

func (r *deviceRepository) FindEventsByDevice(ctx context.Context, deviceID string, from, to time.Time) ([]*device.Event, error) {
	q := `SELECT ` + eventSelectCols + ` FROM device_events WHERE device_id = $1`
	args := []interface{}{deviceID}
	q, args = appendEventTimeFilter(q, args, "created_at", from, to)
	q += ` ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[eventRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Event, 0, len(list))
	for _, row := range list {
		ev, err := rowToEvent(row)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, nil
}

func (r *deviceRepository) FindEventsByLocation(ctx context.Context, location string, from, to time.Time) ([]*device.Event, error) {
	q := `
SELECT e.id, e.device_id, e.event_type, e.data, e.created_at
FROM device_events e
INNER JOIN devices d ON d.id = e.device_id
WHERE d.location = $1`
	args := []interface{}{location}
	q, args = appendEventTimeFilter(q, args, "e.created_at", from, to)
	q += ` ORDER BY e.created_at ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[eventRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Event, 0, len(list))
	for _, row := range list {
		ev, err := rowToEvent(row)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, nil
}

func (r *deviceRepository) GetCommandHistory(ctx context.Context, deviceID string, days int) ([]*device.Command, error) {
	if days <= 0 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)
	q := `SELECT ` + commandSelectCols + ` FROM device_commands WHERE device_id = $1 AND created_at >= $2 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, deviceID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[commandRow])
	if err != nil {
		return nil, err
	}
	out := make([]*device.Command, 0, len(list))
	for _, row := range list {
		c, err := rowToCommand(row)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *deviceRepository) CalculateUptime(ctx context.Context, deviceID string, from, to time.Time) (time.Duration, error) {
	_ = ctx
	_ = deviceID
	_ = from
	_ = to
	return 0, nil
}
