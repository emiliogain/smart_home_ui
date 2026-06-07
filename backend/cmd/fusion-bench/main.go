// Command fusion-bench is an offline benchmark of the rule-based and fuzzy fusion
// predictors on the preprocessed HSLU dataset (Zenodo 14243471).
//
// The real dataset has no ground-truth context labels, so this is a LABEL-FREE
// behavioural benchmark: it characterises how the two simulator-tuned models
// behave on unseen real-world streams. It reports, per model, the context-label
// distribution, label stability (transitions per hour and single-tick flips),
// confidence and inference latency, plus the inter-model agreement rate. A second
// pass applies the same temporal smoother used in production (MultiPredictor) to
// each raw label stream and reports the flicker reduction.
//
// Both predictors run AS-IS (no retuning). The replay drives a fixed DATASET-TIME
// grid (default 60s): at each tick a SensorWindow is built from the latest reading
// per sensor and both predictors are evaluated with Now set to the tick's dataset
// timestamp, so motion recency / presence decay is measured in dataset time rather
// than wall-clock time.
//
// Usage:
//
//	# one participant
//	go run ./cmd/fusion-bench -data-dir ../datasets/hslu_processed/user_7 -user-id 7
//
//	# every participant under a root, plus a pooled summary
//	go run ./cmd/fusion-bench -data-root ../datasets/hslu_processed
//
//	# sensor-coverage census of all participants (decides who is benchmarkable)
//	go run ./cmd/fusion-bench -census ../datasets/hslu_processed
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/adapters/secondary/fusion"
	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/hsludata"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

func main() {
	censusRoot := flag.String("census", "", "Scan <root>/user_*/sensors_manifest.json and print a sensor-coverage census, then exit")
	dataDir := flag.String("data-dir", "", "One participant folder user_<id>/ containing merged_timeline.csv")
	dataRoot := flag.String("data-root", "", "Root folder containing user_*/ folders; benchmarks every participant + writes a pooled summary")
	userID := flag.String("user-id", "", "Participant id (required with -data-dir; inferred from folder name with -data-root)")
	gridStr := flag.String("grid", "60s", "Dataset-time spacing between fused decisions")
	windowDays := flag.Float64("window-days", 14, "Benchmark window length in days from the first reading (0 = whole timeline)")
	windowStartStr := flag.String("window-start", "", "Optional dataset-time start (RFC3339); rows before it warm up state but emit no decisions")
	smootherN := flag.Int("smoother-n", 2, "Consecutive ticks a new label must persist before the temporal smoother accepts it (matches production)")
	outDir := flag.String("out", "bench_out", "Output directory for per_tick.csv / summary.json / summary.txt")
	maxTicks := flag.Int("max-ticks", 0, "Stop a participant after this many decisions (0 = unlimited)")
	quiet := flag.Bool("quiet", false, "Suppress the human-readable summary on stdout")
	flag.Parse()

	// Fusion engines log verbosely at every tick; silence them for the benchmark
	// so output is clean and per-call latency is not skewed by I/O.
	log.SetOutput(io.Discard)

	if *censusRoot != "" {
		if err := runCensus(*censusRoot, *outDir); err != nil {
			fatal(err)
		}
		return
	}

	grid, err := time.ParseDuration(strings.TrimSpace(*gridStr))
	if err != nil || grid <= 0 {
		fatal(fmt.Errorf("-grid: need a positive duration (e.g. 60s, 5m): %v", err))
	}
	var windowDur time.Duration
	if *windowDays > 0 {
		windowDur = time.Duration(*windowDays * 24 * float64(time.Hour))
	}
	var windowStart time.Time
	if s := strings.TrimSpace(*windowStartStr); s != "" {
		windowStart, err = hsludata.ParseDateTimeUTC(s)
		if err != nil {
			fatal(fmt.Errorf("-window-start: %v", err))
		}
	}
	if *smootherN < 1 {
		*smootherN = 1
	}

	cfg := benchConfig{
		grid:        grid,
		windowDur:   windowDur,
		windowStart: windowStart,
		smootherN:   *smootherN,
		maxTicks:    *maxTicks,
	}

	switch {
	case *dataRoot != "":
		if err := runRoot(*dataRoot, *outDir, cfg, *quiet); err != nil {
			fatal(err)
		}
	case *dataDir != "":
		uid := strings.TrimSpace(*userID)
		if uid == "" {
			uid = userIDFromDir(*dataDir)
		}
		if uid == "" {
			fatal(fmt.Errorf("-user-id is required with -data-dir (or use a folder named user_<id>)"))
		}
		res, err := benchmarkUser(*dataDir, uid, cfg)
		if err != nil {
			fatal(err)
		}
		if err := writeUserOutputs(*outDir, res); err != nil {
			fatal(err)
		}
		if !*quiet {
			fmt.Print(res.text())
		}
	default:
		fatal(fmt.Errorf("provide one of -data-dir, -data-root, or -census"))
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "fusion-bench:", err)
	os.Exit(1)
}

// ── Configuration ─────────────────────────────────────────────────────────────

type benchConfig struct {
	grid        time.Duration
	windowDur   time.Duration // 0 = whole timeline
	windowStart time.Time     // zero = first reading
	smootherN   int
	maxTicks    int
}

// ── Per-participant benchmark ─────────────────────────────────────────────────

// tick is one fused decision at a dataset timestamp.
type tick struct {
	at         time.Time
	ruleLabel  string
	ruleConf   float64
	ruleNs     int64
	fuzzyLabel string
	fuzzyConf  float64
	fuzzyNs    int64
}

// modelMetrics aggregates label-free behaviour for one model.
type modelMetrics struct {
	Model            string             `json:"model"`
	Decisions        int                `json:"decisions"`
	ContextCounts    map[string]int     `json:"contextCounts"`
	ContextSharePct  map[string]float64 `json:"contextSharePct"`
	AvgConfidence    float64            `json:"avgConfidence"`
	LatencyMeanUs    float64            `json:"latencyMeanUs"`
	LatencyP50Us     float64            `json:"latencyP50Us"`
	LatencyP95Us     float64            `json:"latencyP95Us"`
	LatencyMaxUs     float64            `json:"latencyMaxUs"`
	Transitions      int                `json:"transitions"`
	TransitionsPerHr float64            `json:"transitionsPerHour"`
	SingleTickFlips  int                `json:"singleTickFlips"`
	// After the production temporal smoother (windowSize = smootherN).
	SmoothedTransitions int     `json:"smoothedTransitions"`
	SmoothedPerHr       float64 `json:"smoothedTransitionsPerHour"`
	FlickerReductionPct float64 `json:"flickerReductionPct"`
}

type benchResult struct {
	UserID       string       `json:"userId"`
	Decisions    int          `json:"decisions"`
	GridSeconds  float64      `json:"gridSeconds"`
	SmootherN    int          `json:"smootherN"`
	WindowStart  string       `json:"windowStart"`
	WindowEnd    string       `json:"windowEnd"`
	SpanHours    float64      `json:"spanHours"`
	AgreementPct float64      `json:"agreementPct"`
	Sensors      []string     `json:"sensors"`
	Rule         modelMetrics `json:"rule"`
	Fuzzy        modelMetrics `json:"fuzzy"`

	ticks []tick // retained for per_tick.csv (not serialised to summary.json)
}

func benchmarkUser(dataDir, uid string, cfg benchConfig) (*benchResult, error) {
	dir, err := hsludata.ExpandUser(strings.TrimSpace(dataDir))
	if err != nil {
		return nil, err
	}
	mergedPath := filepath.Join(dir, "merged_timeline.csv")
	if _, err := os.Stat(mergedPath); err != nil {
		return nil, fmt.Errorf("merged_timeline.csv missing under %s: %w", dir, err)
	}

	rule := fusion.NewRuleBasedPredictor(fusion.DefaultThresholds())
	fuzzy := fusion.NewFuzzyPredictor()
	ctx := context.Background()

	st := newReplayState()
	var ticks []tick

	haveStart := !cfg.windowStart.IsZero()
	windowStart := cfg.windowStart
	var windowEnd time.Time
	var nextTick time.Time
	stopped := false

	predictAt := func(at time.Time) {
		w := st.window(at)
		rs := time.Now()
		rr, _ := rule.Predict(ctx, w)
		rns := time.Since(rs).Nanoseconds()
		fs := time.Now()
		fr, _ := fuzzy.Predict(ctx, w)
		fns := time.Since(fs).Nanoseconds()
		ticks = append(ticks, tick{
			at:         at,
			ruleLabel:  rr.Label,
			ruleConf:   rr.Confidence,
			ruleNs:     rns,
			fuzzyLabel: fr.Label,
			fuzzyConf:  fr.Confidence,
			fuzzyNs:    fns,
		})
	}

	emitTicksUpTo := func(t time.Time) {
		for haveStart && !nextTick.After(t) {
			if cfg.windowDur > 0 && nextTick.After(windowEnd) {
				stopped = true
				return
			}
			if cfg.maxTicks > 0 && len(ticks) >= cfg.maxTicks {
				stopped = true
				return
			}
			predictAt(nextTick)
			nextTick = nextTick.Add(cfg.grid)
		}
	}

	err = forEachRow(mergedPath, uid, func(at time.Time, stream, room, sensorName, value, avg string) error {
		// Warm-up: rows before an explicit window-start update state but emit nothing.
		if haveStart && at.Before(windowStart) {
			st.apply(at, stream, room, sensorName, value, avg)
			return nil
		}
		if !haveStart {
			windowStart = at
			if cfg.windowDur > 0 {
				windowEnd = windowStart.Add(cfg.windowDur)
			}
			nextTick = windowStart
			haveStart = true
		}
		if cfg.windowDur > 0 && at.After(windowEnd) {
			emitTicksUpTo(windowEnd)
			stopped = true
			return errStop
		}
		st.apply(at, stream, room, sensorName, value, avg)
		emitTicksUpTo(at)
		if stopped {
			return errStop
		}
		return nil
	})
	if err != nil && err != errStop {
		return nil, err
	}

	if len(ticks) == 0 {
		return nil, fmt.Errorf("user %s: no decisions produced (no usable rows in window)", uid)
	}

	res := summarise(uid, cfg, ticks)
	res.Sensors = st.seenSensors()
	return res, nil
}

var errStop = fmt.Errorf("stop")

// ── Replay state ──────────────────────────────────────────────────────────────

type scalar struct {
	value float64
	unit  string
}

// replayState forward-fills the latest scalar per sensor and the dataset time of
// the last motion "on" per location (cleared on "off"). At each grid tick it
// builds a SensorWindow whose motion readings carry the last-on timestamp, so the
// engines decay presence relative to the tick's dataset time.
type replayState struct {
	scalars      map[string]scalar    // canonical name -> latest value
	lastMotionOn map[string]time.Time // location -> last "on" dataset time
	seen         map[string]struct{}  // every sensor name ever observed
}

func newReplayState() *replayState {
	return &replayState{
		scalars:      make(map[string]scalar),
		lastMotionOn: make(map[string]time.Time),
		seen:         make(map[string]struct{}),
	}
}

func (s *replayState) apply(at time.Time, stream, room, sensorName, value, avg string) {
	loc, skip := hsludata.NormalizeDatasetRoom(room)
	if skip {
		return
	}
	sensorName = strings.ToLower(strings.TrimSpace(sensorName))
	switch strings.ToLower(strings.TrimSpace(stream)) {
	case "event":
		if sensorName != "movement" {
			return
		}
		name := "motion_" + loc
		s.seen[name] = struct{}{}
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "on":
			s.lastMotionOn[loc] = at
		case "off":
			delete(s.lastMotionOn, loc)
		}
	case "periodic":
		v, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(avg), ",", "."), 64)
		if err != nil {
			return
		}
		switch sensorName {
		case "temperature":
			name := "temp_" + loc
			s.scalars[name] = scalar{v, "°C"}
			s.seen[name] = struct{}{}
		case "humidity":
			name := "humidity_" + loc
			s.scalars[name] = scalar{v, "%"}
			s.seen[name] = struct{}{}
		case "ambient_light":
			name := "light_" + loc
			s.scalars[name] = scalar{v, "lux"}
			s.seen[name] = struct{}{}
		}
	}
}

func (s *replayState) window(at time.Time) secondary.SensorWindow {
	var all []sensor.EnrichedReading
	for name, sc := range s.scalars {
		typ, loc, ok := deriveMeta(name)
		if !ok {
			continue
		}
		all = append(all, sensor.EnrichedReading{
			Reading:    sensor.Reading{SensorID: name, Value: sc.value, Unit: sc.unit, Timestamp: at},
			SensorName: name,
			SensorType: typ,
			Location:   loc,
		})
	}
	for loc, ts := range s.lastMotionOn {
		name := "motion_" + loc
		all = append(all, sensor.EnrichedReading{
			Reading:    sensor.Reading{SensorID: name, Value: 1, Unit: "", Timestamp: ts},
			SensorName: name,
			SensorType: sensor.TypeMotion,
			Location:   loc,
		})
	}
	w := secondary.SensorWindow{
		All:        all,
		ByType:     make(map[sensor.SensorType][]sensor.EnrichedReading),
		ByLocation: make(map[string][]sensor.EnrichedReading),
		Now:        at,
	}
	for _, r := range all {
		w.ByType[r.SensorType] = append(w.ByType[r.SensorType], r)
		w.ByLocation[r.Location] = append(w.ByLocation[r.Location], r)
	}
	return w
}

func (s *replayState) seenSensors() []string {
	out := make([]string, 0, len(s.seen))
	for n := range s.seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// deriveMeta maps a canonical sensor name to its (type, location).
func deriveMeta(name string) (sensor.SensorType, string, bool) {
	switch {
	case strings.HasPrefix(name, "temp_"):
		return sensor.TypeTemperature, strings.TrimPrefix(name, "temp_"), true
	case strings.HasPrefix(name, "humidity_"):
		return sensor.TypeHumidity, strings.TrimPrefix(name, "humidity_"), true
	case strings.HasPrefix(name, "light_"):
		return sensor.TypeLight, strings.TrimPrefix(name, "light_"), true
	case strings.HasPrefix(name, "motion_"):
		return sensor.TypeMotion, strings.TrimPrefix(name, "motion_"), true
	default:
		return "", "", false
	}
}

// ── Metrics ───────────────────────────────────────────────────────────────────

func summarise(uid string, cfg benchConfig, ticks []tick) *benchResult {
	n := len(ticks)
	ruleLabels := make([]string, n)
	fuzzyLabels := make([]string, n)
	var agree int
	for i, t := range ticks {
		ruleLabels[i] = t.ruleLabel
		fuzzyLabels[i] = t.fuzzyLabel
		if t.ruleLabel == t.fuzzyLabel {
			agree++
		}
	}

	span := ticks[n-1].at.Sub(ticks[0].at)
	spanHours := span.Hours()

	res := &benchResult{
		UserID:       uid,
		Decisions:    n,
		GridSeconds:  cfg.grid.Seconds(),
		SmootherN:    cfg.smootherN,
		WindowStart:  ticks[0].at.UTC().Format(time.RFC3339),
		WindowEnd:    ticks[n-1].at.UTC().Format(time.RFC3339),
		SpanHours:    round2(spanHours),
		AgreementPct: round2(100 * float64(agree) / float64(n)),
		Rule:         modelStats("rule_based", ruleLabels, confs(ticks, true), latUs(ticks, true), spanHours, cfg.smootherN),
		Fuzzy:        modelStats("fuzzy", fuzzyLabels, confs(ticks, false), latUs(ticks, false), spanHours, cfg.smootherN),
		ticks:        ticks,
	}
	return res
}

func modelStats(name string, labels []string, confidences, latencyUs []float64, spanHours float64, smootherN int) modelMetrics {
	n := len(labels)
	counts := map[string]int{}
	for _, l := range labels {
		counts[l]++
	}
	shares := map[string]float64{}
	for l, c := range counts {
		shares[l] = round2(100 * float64(c) / float64(n))
	}
	transitions := countTransitions(labels)
	smoothed := countTransitions(smoothStream(labels, smootherN))
	perHr := 0.0
	smPerHr := 0.0
	if spanHours > 0 {
		perHr = float64(transitions) / spanHours
		smPerHr = float64(smoothed) / spanHours
	}
	reduction := 0.0
	if transitions > 0 {
		reduction = 100 * float64(transitions-smoothed) / float64(transitions)
	}
	return modelMetrics{
		Model:               name,
		Decisions:           n,
		ContextCounts:       counts,
		ContextSharePct:     shares,
		AvgConfidence:       round3(mean(confidences)),
		LatencyMeanUs:       round2(mean(latencyUs)),
		LatencyP50Us:        round2(percentile(latencyUs, 50)),
		LatencyP95Us:        round2(percentile(latencyUs, 95)),
		LatencyMaxUs:        round2(maxf(latencyUs)),
		Transitions:         transitions,
		TransitionsPerHr:    round2(perHr),
		SingleTickFlips:     countSingleFlips(labels),
		SmoothedTransitions: smoothed,
		SmoothedPerHr:       round2(smPerHr),
		FlickerReductionPct: round2(reduction),
	}
}

func confs(ticks []tick, ruleSide bool) []float64 {
	out := make([]float64, len(ticks))
	for i, t := range ticks {
		if ruleSide {
			out[i] = t.ruleConf
		} else {
			out[i] = t.fuzzyConf
		}
	}
	return out
}

func latUs(ticks []tick, ruleSide bool) []float64 {
	out := make([]float64, len(ticks))
	for i, t := range ticks {
		if ruleSide {
			out[i] = float64(t.ruleNs) / 1000.0
		} else {
			out[i] = float64(t.fuzzyNs) / 1000.0
		}
	}
	return out
}

// smoothStream replicates fusion.TemporalSmoother: a new label must persist for
// `n` consecutive ticks before it is accepted; until then the confirmed label is
// returned. Used to count flicker after smoothing.
func smoothStream(raw []string, n int) []string {
	out := make([]string, len(raw))
	confirmed, pending := "", ""
	pendingCount := 0
	for i, r := range raw {
		switch {
		case confirmed == "":
			confirmed = r
		case r == confirmed:
			pending, pendingCount = "", 0
		default:
			if r == pending {
				pendingCount++
			} else {
				pending, pendingCount = r, 1
			}
			if pendingCount >= n {
				confirmed, pending, pendingCount = pending, "", 0
			}
		}
		out[i] = confirmed
	}
	return out
}

func countTransitions(s []string) int {
	c := 0
	for i := 1; i < len(s); i++ {
		if s[i] != s[i-1] {
			c++
		}
	}
	return c
}

func countSingleFlips(s []string) int {
	c := 0
	for i := 1; i < len(s)-1; i++ {
		if s[i] != s[i-1] && s[i-1] == s[i+1] {
			c++
		}
	}
	return c
}

// ── CSV row reader ────────────────────────────────────────────────────────────

// forEachRow streams merged_timeline.csv rows for one user in file order.
func forEachRow(path, userID string, fn func(at time.Time, stream, room, sensorName, value, avg string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	hdr, err := r.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	idx := hsludata.HeaderIndex(hdr)
	for _, col := range []string{"datetime_utc", "stream"} {
		if _, ok := idx[col]; !ok {
			return fmt.Errorf("merged_timeline.csv missing column %q", col)
		}
	}
	get := func(rec []string, key string) string {
		i, ok := idx[key]
		if !ok || i >= len(rec) {
			return ""
		}
		return rec[i]
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if userID != "" && strings.TrimSpace(get(rec, "id")) != userID {
			continue
		}
		at, err := hsludata.ParseDateTimeUTC(strings.TrimSpace(get(rec, "datetime_utc")))
		if err != nil {
			continue
		}
		if e := fn(at, get(rec, "stream"), get(rec, "room"), get(rec, "sensor"), get(rec, "value"), get(rec, "average_value")); e != nil {
			return e
		}
	}
}

// ── Root (all participants + pooled) ──────────────────────────────────────────

func runRoot(root, outDir string, cfg benchConfig, quiet bool) error {
	root, err := hsludata.ExpandUser(root)
	if err != nil {
		return err
	}
	dirs, err := userDirs(root)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no user_*/merged_timeline.csv found under %s", root)
	}

	var results []*benchResult
	for _, d := range dirs {
		uid := userIDFromDir(d)
		res, err := benchmarkUser(d, uid, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip user %s: %v\n", uid, err)
			continue
		}
		if err := writeUserOutputs(outDir, res); err != nil {
			return err
		}
		results = append(results, res)
		fmt.Printf("user %-6s decisions=%-7d agreement=%5.1f%%  rule_flicks/h=%6.2f  fuzzy_flicks/h=%6.2f\n",
			uid, res.Decisions, res.AgreementPct, res.Rule.TransitionsPerHr, res.Fuzzy.TransitionsPerHr)
	}
	if len(results) == 0 {
		return fmt.Errorf("no participants produced results")
	}

	pooled := poolResults(results, cfg)
	if err := writePooledOutputs(outDir, pooled, results); err != nil {
		return err
	}
	if !quiet {
		fmt.Print(pooled.text())
	}
	return nil
}

// poolResults aggregates per-tick totals across participants into one summary
// (UserID="POOLED"). Rates are recomputed from summed counts and total span.
func poolResults(results []*benchResult, cfg benchConfig) *benchResult {
	var allRule, allFuzzy []string
	var ruleConf, fuzzyConf, ruleLat, fuzzyLat []float64
	var agree, total int
	var totalSpanHours float64
	for _, r := range results {
		totalSpanHours += r.SpanHours
		for _, t := range r.ticks {
			allRule = append(allRule, t.ruleLabel)
			allFuzzy = append(allFuzzy, t.fuzzyLabel)
			ruleConf = append(ruleConf, t.ruleConf)
			fuzzyConf = append(fuzzyConf, t.fuzzyConf)
			ruleLat = append(ruleLat, float64(t.ruleNs)/1000.0)
			fuzzyLat = append(fuzzyLat, float64(t.fuzzyNs)/1000.0)
			total++
			if t.ruleLabel == t.fuzzyLabel {
				agree++
			}
		}
	}
	// Pooled transitions are summed per participant (cross-participant joins are
	// not real transitions), but per-hour uses the combined span.
	var ruleTrans, fuzzyTrans, ruleSm, fuzzySm, ruleFlip, fuzzyFlip int
	for _, r := range results {
		ruleTrans += r.Rule.Transitions
		fuzzyTrans += r.Fuzzy.Transitions
		ruleSm += r.Rule.SmoothedTransitions
		fuzzySm += r.Fuzzy.SmoothedTransitions
		ruleFlip += r.Rule.SingleTickFlips
		fuzzyFlip += r.Fuzzy.SingleTickFlips
	}

	mk := func(name string, labels []string, conf, lat []float64, trans, sm, flip int) modelMetrics {
		m := modelStats(name, labels, conf, lat, totalSpanHours, cfg.smootherN)
		m.Transitions = trans
		m.SmoothedTransitions = sm
		m.SingleTickFlips = flip
		if totalSpanHours > 0 {
			m.TransitionsPerHr = round2(float64(trans) / totalSpanHours)
			m.SmoothedPerHr = round2(float64(sm) / totalSpanHours)
		}
		if trans > 0 {
			m.FlickerReductionPct = round2(100 * float64(trans-sm) / float64(trans))
		}
		return m
	}

	return &benchResult{
		UserID:       fmt.Sprintf("POOLED(%d users)", len(results)),
		Decisions:    total,
		GridSeconds:  cfg.grid.Seconds(),
		SmootherN:    cfg.smootherN,
		SpanHours:    round2(totalSpanHours),
		AgreementPct: round2(100 * float64(agree) / float64(total)),
		Rule:         mk("rule_based", allRule, ruleConf, ruleLat, ruleTrans, ruleSm, ruleFlip),
		Fuzzy:        mk("fuzzy", allFuzzy, fuzzyConf, fuzzyLat, fuzzyTrans, fuzzySm, fuzzyFlip),
	}
}

// ── Coverage census ───────────────────────────────────────────────────────────

type manifest struct {
	Sensors []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Location string `json:"location"`
	} `json:"sensors"`
}

type coverageRow struct {
	UserID   string
	Channels int
	hasType  map[string]map[string]bool // type -> location -> present
}

func (c coverageRow) has(typ, loc string) bool {
	if m, ok := c.hasType[typ]; ok {
		return m[loc]
	}
	return false
}

func (c coverageRow) hasTypeAny(typ string) bool {
	if m, ok := c.hasType[typ]; ok {
		return len(m) > 0
	}
	return false
}

// usable: a comfort/alert baseline can fire (temp + humidity somewhere) and at
// least one activity room is observable (light + motion in the living room).
func (c coverageRow) usable() bool {
	return c.hasTypeAny("temperature") && c.hasTypeAny("humidity") &&
		c.has("light", "living_room") && c.has("motion", "living_room")
}

// full: every context class the models can emit is structurally reachable.
func (c coverageRow) full() bool {
	return c.usable() &&
		c.has("light", "kitchen") && c.has("motion", "kitchen") &&
		c.has("light", "bedroom") && c.has("motion", "bedroom")
}

func runCensus(root, outDir string) error {
	root, err := hsludata.ExpandUser(root)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	var rows []coverageRow
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "user_") || !isDir(filepath.Join(root, e.Name())) {
			continue
		}
		mp := filepath.Join(root, e.Name(), "sensors_manifest.json")
		b, err := os.ReadFile(mp)
		if err != nil {
			continue
		}
		var m manifest
		if json.Unmarshal(b, &m) != nil {
			continue
		}
		row := coverageRow{UserID: strings.TrimPrefix(e.Name(), "user_"), hasType: map[string]map[string]bool{}}
		for _, s := range m.Sensors {
			if row.hasType[s.Type] == nil {
				row.hasType[s.Type] = map[string]bool{}
			}
			row.hasType[s.Type][s.Location] = true
			row.Channels++
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return fmt.Errorf("no user_*/sensors_manifest.json under %s", root)
	}
	sort.Slice(rows, func(i, j int) bool { return numLess(rows[i].UserID, rows[j].UserID) })

	var usable, full int
	var b strings.Builder
	fmt.Fprintf(&b, "Sensor-coverage census — %d participants under %s\n", len(rows), root)
	fmt.Fprintf(&b, "%-8s %-4s %-5s %-5s %-6s  %-7s %-7s\n", "user", "chan", "temp", "humid", "rooms", "usable", "full")
	for _, r := range rows {
		if r.usable() {
			usable++
		}
		if r.full() {
			full++
		}
		fmt.Fprintf(&b, "%-8s %-4d %-5s %-5s %-6s  %-7s %-7s\n",
			r.UserID, r.Channels,
			yn(r.hasTypeAny("temperature")), yn(r.hasTypeAny("humidity")),
			roomsCovered(r), yn(r.usable()), yn(r.full()))
	}
	fmt.Fprintf(&b, "\nusable (temp+humidity + living light+motion): %d/%d\n", usable, len(rows))
	fmt.Fprintf(&b, "full   (+ kitchen & bedroom light+motion):     %d/%d\n", full, len(rows))
	fmt.Print(b.String())

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// machine-readable census
	cf, err := os.Create(filepath.Join(outDir, "census.csv"))
	if err != nil {
		return err
	}
	defer func() { _ = cf.Close() }()
	w := csv.NewWriter(cf)
	_ = w.Write([]string{"user", "channels", "has_temp", "has_humidity",
		"light_living", "motion_living", "light_kitchen", "motion_kitchen",
		"light_bedroom", "motion_bedroom", "usable", "full"})
	for _, r := range rows {
		_ = w.Write([]string{
			r.UserID, strconv.Itoa(r.Channels),
			b01(r.hasTypeAny("temperature")), b01(r.hasTypeAny("humidity")),
			b01(r.has("light", "living_room")), b01(r.has("motion", "living_room")),
			b01(r.has("light", "kitchen")), b01(r.has("motion", "kitchen")),
			b01(r.has("light", "bedroom")), b01(r.has("motion", "bedroom")),
			b01(r.usable()), b01(r.full()),
		})
	}
	w.Flush()
	fmt.Printf("\nwrote %s\n", filepath.Join(outDir, "census.csv"))
	return nil
}

func roomsCovered(r coverageRow) string {
	set := map[string]struct{}{}
	for _, m := range r.hasType {
		for loc := range m {
			set[loc] = struct{}{}
		}
	}
	var out []string
	for _, loc := range []string{"living_room", "kitchen", "bedroom"} {
		if _, ok := set[loc]; ok {
			out = append(out, loc[:3])
		}
	}
	if len(out) == 0 {
		return "-"
	}
	return strings.Join(out, ",")
}

// ── Output writers ────────────────────────────────────────────────────────────

func writeUserOutputs(outDir string, res *benchResult) error {
	dir := filepath.Join(outDir, "user_"+res.UserID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// per_tick.csv
	pf, err := os.Create(filepath.Join(dir, "per_tick.csv"))
	if err != nil {
		return err
	}
	w := csv.NewWriter(pf)
	_ = w.Write([]string{"datetime_utc", "rule_label", "rule_conf", "rule_latency_us",
		"fuzzy_label", "fuzzy_conf", "fuzzy_latency_us", "agree"})
	for _, t := range res.ticks {
		_ = w.Write([]string{
			t.at.UTC().Format(time.RFC3339),
			t.ruleLabel, fmt.Sprintf("%.3f", t.ruleConf), fmt.Sprintf("%.2f", float64(t.ruleNs)/1000.0),
			t.fuzzyLabel, fmt.Sprintf("%.3f", t.fuzzyConf), fmt.Sprintf("%.2f", float64(t.fuzzyNs)/1000.0),
			b01(t.ruleLabel == t.fuzzyLabel),
		})
	}
	w.Flush()
	_ = pf.Close()

	if err := writeJSON(filepath.Join(dir, "summary.json"), res); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.txt"), []byte(res.text()), 0o644)
}

func writePooledOutputs(outDir string, pooled *benchResult, all []*benchResult) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "pooled_summary.json"), pooled); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "pooled_summary.txt"), []byte(pooled.text()), 0o644); err != nil {
		return err
	}
	// one-row-per-user comparison table for the thesis
	cf, err := os.Create(filepath.Join(outDir, "per_user_table.csv"))
	if err != nil {
		return err
	}
	defer func() { _ = cf.Close() }()
	w := csv.NewWriter(cf)
	_ = w.Write([]string{"user", "decisions", "span_hours", "agreement_pct",
		"rule_avg_conf", "fuzzy_avg_conf",
		"rule_latency_mean_us", "fuzzy_latency_mean_us",
		"rule_flicks_per_hr", "fuzzy_flicks_per_hr",
		"rule_single_flips", "fuzzy_single_flips",
		"rule_flick_reduction_pct", "fuzzy_flick_reduction_pct"})
	for _, r := range all {
		_ = w.Write([]string{
			r.UserID, strconv.Itoa(r.Decisions), f2(r.SpanHours), f2(r.AgreementPct),
			f3(r.Rule.AvgConfidence), f3(r.Fuzzy.AvgConfidence),
			f2(r.Rule.LatencyMeanUs), f2(r.Fuzzy.LatencyMeanUs),
			f2(r.Rule.TransitionsPerHr), f2(r.Fuzzy.TransitionsPerHr),
			strconv.Itoa(r.Rule.SingleTickFlips), strconv.Itoa(r.Fuzzy.SingleTickFlips),
			f2(r.Rule.FlickerReductionPct), f2(r.Fuzzy.FlickerReductionPct),
		})
	}
	w.Flush()
	return nil
}

func writeJSON(path string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func (r *benchResult) text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n=== fusion-bench: user %s ===\n", r.UserID)
	fmt.Fprintf(&b, "decisions: %d   grid: %.0fs   span: %.1f h   window: %s .. %s\n",
		r.Decisions, r.GridSeconds, r.SpanHours, r.WindowStart, r.WindowEnd)
	if len(r.Sensors) > 0 {
		fmt.Fprintf(&b, "sensors: %s\n", strings.Join(r.Sensors, ", "))
	}
	fmt.Fprintf(&b, "inter-model agreement: %.1f%%\n\n", r.AgreementPct)
	fmt.Fprintf(&b, "%-26s %14s %14s\n", "metric", "rule_based", "fuzzy")
	row := func(label, a, c string) { fmt.Fprintf(&b, "%-26s %14s %14s\n", label, a, c) }
	row("avg confidence", f3(r.Rule.AvgConfidence), f3(r.Fuzzy.AvgConfidence))
	row("latency mean (us)", f2(r.Rule.LatencyMeanUs), f2(r.Fuzzy.LatencyMeanUs))
	row("latency p95 (us)", f2(r.Rule.LatencyP95Us), f2(r.Fuzzy.LatencyP95Us))
	row("transitions", strconv.Itoa(r.Rule.Transitions), strconv.Itoa(r.Fuzzy.Transitions))
	row("transitions / hour", f2(r.Rule.TransitionsPerHr), f2(r.Fuzzy.TransitionsPerHr))
	row("single-tick flips", strconv.Itoa(r.Rule.SingleTickFlips), strconv.Itoa(r.Fuzzy.SingleTickFlips))
	row(fmt.Sprintf("smoothed trans (N=%d)", r.SmootherN), strconv.Itoa(r.Rule.SmoothedTransitions), strconv.Itoa(r.Fuzzy.SmoothedTransitions))
	row("flicker reduction %", f2(r.Rule.FlickerReductionPct), f2(r.Fuzzy.FlickerReductionPct))
	b.WriteString("\ncontext distribution (% of decisions):\n")
	b.WriteString(distText(r.Rule.ContextSharePct, r.Fuzzy.ContextSharePct))
	return b.String()
}

func distText(rule, fuzzy map[string]float64) string {
	keys := map[string]struct{}{}
	for k := range rule {
		keys[k] = struct{}{}
	}
	for k := range fuzzy {
		keys[k] = struct{}{}
	}
	var ks []string
	for k := range keys {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return rule[ks[i]]+fuzzy[ks[i]] > rule[ks[j]]+fuzzy[ks[j]] })
	var b strings.Builder
	fmt.Fprintf(&b, "  %-28s %10s %10s\n", "context", "rule", "fuzzy")
	for _, k := range ks {
		fmt.Fprintf(&b, "  %-28s %9.1f%% %9.1f%%\n", k, rule[k], fuzzy[k])
	}
	return b.String()
}

// ── small helpers ─────────────────────────────────────────────────────────────

func userDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "user_") || !isDir(filepath.Join(root, e.Name())) {
			continue
		}
		d := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(d, "merged_timeline.csv")); err == nil {
			dirs = append(dirs, d)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return numLess(userIDFromDir(dirs[i]), userIDFromDir(dirs[j])) })
	return dirs, nil
}

func userIDFromDir(dir string) string {
	base := filepath.Base(strings.TrimRight(dir, "/"))
	return strings.TrimPrefix(base, "user_")
}

// isDir reports whether path is a directory, following symlinks (so subset roots
// built from symlinks to user_<id>/ folders are accepted).
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func maxf(xs []float64) float64 {
	m := 0.0
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}

func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	c := append([]float64(nil), xs...)
	sort.Float64s(c)
	if p <= 0 {
		return c[0]
	}
	if p >= 100 {
		return c[len(c)-1]
	}
	rank := p / 100 * float64(len(c)-1)
	lo := int(rank)
	if lo >= len(c)-1 {
		return c[len(c)-1]
	}
	frac := rank - float64(lo)
	return c[lo] + frac*(c[lo+1]-c[lo])
}

func numLess(a, b string) bool {
	ai, ae := strconv.Atoi(a)
	bi, be := strconv.Atoi(b)
	if ae == nil && be == nil {
		return ai < bi
	}
	return a < b
}

func round2(f float64) float64 { return float64(int64(f*100+sign(f)*0.5)) / 100 }
func round3(f float64) float64 { return float64(int64(f*1000+sign(f)*0.5)) / 1000 }
func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}

func f2(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) }
func f3(f float64) string { return strconv.FormatFloat(f, 'f', 3, 64) }

func yn(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func b01(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
