package simulator

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
	"github.com/google/uuid"
)

// ReadingSaver abstracts the ability to persist sensor readings and trigger fusion.
// *app.SensorService satisfies this interface directly.
type ReadingSaver interface {
	SaveReading(ctx context.Context, r sensor.Reading) (*secondary.FusionResult, error)
	SaveReadingsBatch(ctx context.Context, readings []sensor.Reading) (*secondary.FusionResult, error)
}

// Status is a snapshot of the engine state, returned by GetStatus.
type Status struct {
	Running            bool     `json:"running"`
	Paused             bool     `json:"paused"`
	CurrentScenario    string   `json:"currentScenario"`
	IntervalMs         int64    `json:"intervalMs"`
	AvailableScenarios []string `json:"availableScenarios"`
	LastTick           string   `json:"lastTick"`
	TickCount          int64    `json:"tickCount"`
}

// injectRequest is sent on the inject channel for one-off manual readings.
type injectRequest struct {
	SensorName string
	Value      float64
	Unit       string
	Result     chan error
}

// Engine is a controllable simulation engine that runs as a goroutine.
type Engine struct {
	saver     ReadingSaver
	sensorIDs map[string]string // sensor name → database UUID

	mu           sync.RWMutex
	running      bool
	paused       bool
	scenario     Scenario
	allScenarios []Scenario
	interval     time.Duration
	lastTick     time.Time
	tickCount    int64

	// Channels for commands that need goroutine synchronization.
	scenarioCh chan Scenario
	intervalCh chan time.Duration
	injectCh   chan injectRequest
	stopCh     chan struct{}
}

// NewEngine creates a simulation engine. Call Start() to begin generating readings.
func NewEngine(saver ReadingSaver, sensorIDs map[string]string, opts ...Option) *Engine {
	all := AllScenarios()
	e := &Engine{
		saver:        saver,
		sensorIDs:    sensorIDs,
		allScenarios: all,
		scenario:     all[0], // "comfortable" default
		interval:     5 * time.Second,
		scenarioCh:   make(chan Scenario, 1),
		intervalCh:   make(chan time.Duration, 1),
		injectCh:     make(chan injectRequest),
		stopCh:       make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Start launches the simulation loop. It blocks until ctx is cancelled or Stop() is called.
func (e *Engine) Start(ctx context.Context) {
	e.mu.Lock()
	e.running = true
	e.mu.Unlock()

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	log.Printf("[simulator] started — scenario=%s interval=%s", e.scenario.Name, e.interval)

	for {
		select {
		case <-ctx.Done():
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			log.Println("[simulator] stopped (context cancelled)")
			return

		case <-e.stopCh:
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			log.Println("[simulator] stopped")
			return

		case sc := <-e.scenarioCh:
			e.mu.Lock()
			e.scenario = sc
			e.mu.Unlock()
			log.Printf("[simulator] scenario → %s", sc.Name)

		case d := <-e.intervalCh:
			e.mu.Lock()
			e.interval = d
			e.mu.Unlock()
			ticker.Reset(d)
			log.Printf("[simulator] interval → %s", d)

		case req := <-e.injectCh:
			req.Result <- e.doInject(ctx, req.SensorName, req.Value, req.Unit)

		case <-ticker.C:
			// Check pause state without blocking — just skip the tick.
			e.mu.RLock()
			paused := e.paused
			e.mu.RUnlock()
			if paused {
				continue
			}
			e.doTick(ctx)
		}
	}
}

// Stop signals the engine to exit.
func (e *Engine) Stop() {
	select {
	case e.stopCh <- struct{}{}:
	default:
	}
}

// SetScenario switches the active scenario by name. Non-blocking.
func (e *Engine) SetScenario(name string) error {
	for _, sc := range e.allScenarios {
		if strings.EqualFold(sc.Name, name) {
			// Drain any pending scenario change, then send the new one.
			select {
			case <-e.scenarioCh:
			default:
			}
			e.scenarioCh <- sc
			return nil
		}
	}
	return fmt.Errorf("unknown scenario %q", name)
}

// Pause stops generating readings until Resume is called. Non-blocking.
func (e *Engine) Pause() {
	e.mu.Lock()
	e.paused = true
	e.mu.Unlock()
	log.Println("[simulator] paused")
}

// Resume restarts reading generation after a Pause. Non-blocking.
func (e *Engine) Resume() {
	e.mu.Lock()
	e.paused = false
	e.mu.Unlock()
	log.Println("[simulator] resumed")
}

// SetInterval changes the tick rate. Must be between 1s and 30s. Non-blocking.
func (e *Engine) SetInterval(d time.Duration) error {
	if d < time.Second || d > 30*time.Second {
		return fmt.Errorf("interval must be between 1s and 30s, got %s", d)
	}
	select {
	case <-e.intervalCh:
	default:
	}
	e.intervalCh <- d
	return nil
}

// Inject sends a one-off manual reading for the named sensor.
func (e *Engine) Inject(sensorName string, value float64, unit string) error {
	result := make(chan error, 1)
	e.injectCh <- injectRequest{
		SensorName: sensorName,
		Value:      value,
		Unit:       unit,
		Result:     result,
	}
	return <-result
}

// GetStatus returns a snapshot of the current engine state.
func (e *Engine) GetStatus() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()

	lastTick := ""
	if !e.lastTick.IsZero() {
		lastTick = e.lastTick.UTC().Format(time.RFC3339)
	}

	return Status{
		Running:            e.running,
		Paused:             e.paused,
		CurrentScenario:    e.scenario.Name,
		IntervalMs:         e.interval.Milliseconds(),
		AvailableScenarios: ScenarioNames(),
		LastTick:           lastTick,
		TickCount:          e.tickCount,
	}
}

// doTick generates and submits all readings as a batch, running fusion only once.
func (e *Engine) doTick(ctx context.Context) {
	e.mu.RLock()
	scenario := e.scenario
	e.mu.RUnlock()

	// Build all readings first.
	var batch []sensor.Reading
	for sensorName, profile := range scenario.Profiles {
		id, ok := e.sensorIDs[sensorName]
		if !ok {
			continue
		}
		value := math.Round(profile.Generate()*100) / 100
		batch = append(batch, sensor.Reading{
			ID:       uuid.NewString(),
			SensorID: id,
			Value:    value,
			Unit:     profile.Unit,
		})
	}

	// Save all readings and run fusion once.
	if _, err := e.saver.SaveReadingsBatch(ctx, batch); err != nil {
		log.Printf("[simulator] warning: batch save: %v", err)
	}

	e.mu.Lock()
	e.lastTick = time.Now()
	e.tickCount++
	e.mu.Unlock()
}

// doInject submits a single manual reading.
func (e *Engine) doInject(ctx context.Context, sensorName string, value float64, unit string) error {
	id, ok := e.sensorIDs[sensorName]
	if !ok {
		return fmt.Errorf("unknown sensor %q", sensorName)
	}
	reading := sensor.Reading{
		ID:       uuid.NewString(),
		SensorID: id,
		Value:    value,
		Unit:     unit,
	}
	_, err := e.saver.SaveReading(ctx, reading)
	if err != nil {
		return fmt.Errorf("inject %s: %w", sensorName, err)
	}
	log.Printf("[simulator] injected %s = %.2f %s", sensorName, value, unit)
	return nil
}

// DeterministicSensorID generates a stable UUID for a sensor name so
// restarts don't create duplicate sensors in the database.
func DeterministicSensorID(name string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("smarthome.sensor."+name)).String()
}
