package fusion

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// ── Public types used by the admin handler ───────────────────────────────────

// PredictionRecord holds the outcome of a single model prediction for one tick.
type PredictionRecord struct {
	ModelName    string    `json:"modelName"`
	PredictedCtx string    `json:"predictedCtx"`
	Confidence   float64   `json:"confidence"`
	LatencyUs    int64     `json:"latencyUs"`
	ExpectedCtx  string    `json:"expectedCtx,omitempty"`
	Matches      bool      `json:"matches"`
	Timestamp    time.Time `json:"timestamp"`
}

// ComparisonRow is a single tick's worth of predictions across all models.
type ComparisonRow struct {
	Timestamp time.Time                   `json:"timestamp"`
	PerModel  map[string]PredictionRecord `json:"perModel"`
	Agree     bool                        `json:"agree"`
}

// ModelStats aggregates metrics for one model over the recorded history.
type ModelStats struct {
	ModelName        string         `json:"modelName"`
	TotalPredictions int            `json:"totalPredictions"`
	AvgConfidence    float64        `json:"avgConfidence"`
	AvgLatencyUs     float64        `json:"avgLatencyUs"`
	ContextCounts    map[string]int `json:"contextCounts"`
	// Accuracy is the fraction of predictions that matched the expected context.
	// Only meaningful when the simulator provides a scenario hint.
	Accuracy float64 `json:"accuracy"`
	// AgreementRate is the fraction of ticks where this model agreed with
	// the other model(s).
	AgreementRate float64 `json:"agreementRate"`
}

// ComparisonSnapshot is the full payload returned by the comparison endpoint.
type ComparisonSnapshot struct {
	ActiveModel string          `json:"activeModel"`
	Models      []string        `json:"models"`
	Recent      []ComparisonRow `json:"recent"`
	Stats       []ModelStats    `json:"stats"`
}

// ── Ring buffer ──────────────────────────────────────────────────────────────

const ringSize = 500

type metricsStore struct {
	mu   sync.RWMutex
	rows [ringSize]ComparisonRow
	head int // next write position
	size int // number of valid entries
}

func (m *metricsStore) add(row ComparisonRow) {
	m.mu.Lock()
	m.rows[m.head] = row
	m.head = (m.head + 1) % ringSize
	if m.size < ringSize {
		m.size++
	}
	m.mu.Unlock()
}

// last returns the n most recent rows, newest last.
func (m *metricsStore) last(n int) []ComparisonRow {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n > m.size {
		n = m.size
	}
	if n == 0 {
		return nil
	}

	out := make([]ComparisonRow, n)
	// head points to the next write position, so head-1 is the newest entry
	for i := 0; i < n; i++ {
		idx := (m.head - 1 - i + ringSize) % ringSize
		out[n-1-i] = m.rows[idx]
	}
	return out
}

func (m *metricsStore) all() []ComparisonRow {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.size == 0 {
		return nil
	}
	out := make([]ComparisonRow, m.size)
	for i := 0; i < m.size; i++ {
		idx := (m.head - m.size + i + ringSize*2) % ringSize
		out[i] = m.rows[idx]
	}
	return out
}

// ── MultiPredictor ───────────────────────────────────────────────────────────

// NamedPredictor pairs a FusionPredictor with a human-readable name.
type NamedPredictor struct {
	Name      string
	Predictor secondary.FusionPredictor
}

// MultiPredictor runs all registered predictors on every call to Predict,
// records metrics for each, and returns the result from the active model.
// It satisfies the secondary.FusionPredictor interface and can therefore
// be dropped in wherever a single predictor was used before.
type MultiPredictor struct {
	predictors   []NamedPredictor
	store        *metricsStore
	mu           sync.RWMutex
	activeModel  string
	hintMu       sync.RWMutex
	scenarioHint string // expected context label (from simulator)
}

// NewMultiPredictor creates a MultiPredictor. The first predictor in the list
// becomes the initial active model.
func NewMultiPredictor(predictors ...NamedPredictor) *MultiPredictor {
	active := ""
	if len(predictors) > 0 {
		active = predictors[0].Name
	}
	return &MultiPredictor{
		predictors:  predictors,
		store:       &metricsStore{},
		activeModel: active,
	}
}

// SetActiveModel switches which model's result is returned by Predict.
func (m *MultiPredictor) SetActiveModel(name string) error {
	for _, p := range m.predictors {
		if p.Name == name {
			m.mu.Lock()
			m.activeModel = name
			m.mu.Unlock()
			log.Printf("[fusion] active model → %s", name)
			return nil
		}
	}
	return fmt.Errorf("unknown model %q; available: %s", name, strings.Join(m.GetModelNames(), ", "))
}

// GetModelNames returns the names of all registered predictors.
func (m *MultiPredictor) GetModelNames() []string {
	names := make([]string, len(m.predictors))
	for i, p := range m.predictors {
		names[i] = p.Name
	}
	return names
}

// GetActiveModel returns the name of the currently active model.
func (m *MultiPredictor) GetActiveModel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeModel
}

// SetScenarioHint records the expected context for the current simulator
// scenario. This is used to compute accuracy metrics.
// The expected string should be the ContextType constant, e.g. "COOKING_KITCHEN".
func (m *MultiPredictor) SetScenarioHint(expected string) {
	m.hintMu.Lock()
	m.scenarioHint = expected
	m.hintMu.Unlock()
}

// Predict runs all models, records metrics, and returns the active model's result.
func (m *MultiPredictor) Predict(ctx context.Context, window secondary.SensorWindow) (*secondary.FusionResult, error) {
	m.hintMu.RLock()
	hint := m.scenarioHint
	m.hintMu.RUnlock()

	m.mu.RLock()
	active := m.activeModel
	m.mu.RUnlock()

	type outcome struct {
		name   string
		result *secondary.FusionResult
		us     int64
		err    error
	}

	ch := make(chan outcome, len(m.predictors))
	for _, np := range m.predictors {
		np := np
		go func() {
			start := time.Now()
			res, err := np.Predictor.Predict(ctx, window)
			us := time.Since(start).Microseconds()
			ch <- outcome{name: np.Name, result: res, us: us, err: err}
		}()
	}

	row := ComparisonRow{
		Timestamp: time.Now(),
		PerModel:  make(map[string]PredictionRecord),
	}
	var activeResult *secondary.FusionResult
	var activeErr error

	for range m.predictors {
		o := <-ch
		if o.err != nil {
			log.Printf("[fusion] %s error: %v", o.name, o.err)
			if o.name == active {
				activeErr = o.err
			}
			continue
		}
		matches := hint != "" && o.result.Label == hint
		row.PerModel[o.name] = PredictionRecord{
			ModelName:    o.name,
			PredictedCtx: o.result.Label,
			Confidence:   o.result.Confidence,
			LatencyUs:    o.us,
			ExpectedCtx:  hint,
			Matches:      matches,
			Timestamp:    row.Timestamp,
		}
		if o.name == active {
			activeResult = o.result
		}
	}

	// Check agreement (all models that returned results agree).
	labels := make(map[string]struct{})
	for _, rec := range row.PerModel {
		labels[rec.PredictedCtx] = struct{}{}
	}
	row.Agree = len(labels) == 1

	m.store.add(row)

	// Build log line showing all model results.
	parts := make([]string, 0, len(row.PerModel))
	for _, rec := range row.PerModel {
		parts = append(parts, fmt.Sprintf("%s→%s(%.0f%%)", rec.ModelName, rec.PredictedCtx, rec.Confidence*100))
	}
	log.Printf("[fusion] %s active=%s", strings.Join(parts, " "), active)

	if activeResult == nil {
		return &secondary.FusionResult{Label: "UNKNOWN", Confidence: 0.5}, activeErr
	}
	return activeResult, activeErr
}

// GetComparison returns the last n ticks with per-model predictions and aggregate stats.
func (m *MultiPredictor) GetComparison(n int) ComparisonSnapshot {
	rows := m.store.last(n)
	all := m.store.all()

	m.mu.RLock()
	active := m.activeModel
	m.mu.RUnlock()

	stats := m.computeStats(all)

	return ComparisonSnapshot{
		ActiveModel: active,
		Models:      m.GetModelNames(),
		Recent:      rows,
		Stats:       stats,
	}
}

func (m *MultiPredictor) computeStats(rows []ComparisonRow) []ModelStats {
	type accumulator struct {
		totalConf     float64
		totalLatency  int64
		count         int
		matchCount    int
		hintedCount   int
		contextCounts map[string]int
		agreeCount    int
	}

	accs := make(map[string]*accumulator)
	for _, np := range m.predictors {
		accs[np.Name] = &accumulator{contextCounts: make(map[string]int)}
	}

	for _, row := range rows {
		for name, rec := range row.PerModel {
			a := accs[name]
			if a == nil {
				continue
			}
			a.count++
			a.totalConf += rec.Confidence
			a.totalLatency += rec.LatencyUs
			a.contextCounts[rec.PredictedCtx]++
			if rec.ExpectedCtx != "" {
				a.hintedCount++
				if rec.Matches {
					a.matchCount++
				}
			}
			if row.Agree {
				a.agreeCount++
			}
		}
	}

	stats := make([]ModelStats, 0, len(m.predictors))
	for _, np := range m.predictors {
		a := accs[np.Name]
		s := ModelStats{
			ModelName:     np.Name,
			ContextCounts: a.contextCounts,
		}
		if a.count > 0 {
			s.TotalPredictions = a.count
			s.AvgConfidence = a.totalConf / float64(a.count)
			s.AvgLatencyUs = float64(a.totalLatency) / float64(a.count)
			s.AgreementRate = float64(a.agreeCount) / float64(a.count)
		}
		if a.hintedCount > 0 {
			s.Accuracy = float64(a.matchCount) / float64(a.hintedCount)
		}
		stats = append(stats, s)
	}
	return stats
}
