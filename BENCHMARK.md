# Fusion Benchmark on Real Data (HSLU 14243471)

Offline, reproducible comparison of the **rule-based** and **fuzzy** fusion
predictors on the real HSLU smart-home dataset
([Zenodo 14243471](https://zenodo.org/records/14243471): 62 elderly-resident homes,
4 countries, 359 days; PIR + door events and periodic temperature / humidity /
ambient-light / CO₂ / VOC / sound at 2–5 min).

Tool: [`backend/cmd/fusion-bench`](backend/cmd/fusion-bench/main.go).

---

## Why this benchmark is *label-free*

The dataset ships **no activity / ground-truth context annotations**. The
simulator-based evaluation in Chapter 5 measures *accuracy* against a known
scenario hint (`MultiPredictor.SetScenarioHint`); that hint does not exist for
real data, so accuracy **cannot** be computed here.

Instead the benchmark characterises how the two models — calibrated for the
simulator and run **as-is, no retuning** — *behave* on unseen real streams:

| Metric | Needs labels? | What it tells you |
|--------|:---:|-------------------|
| Inter-model **agreement** (% ticks rule == fuzzy) | no | how often the models converge; disagreement concentrates in ambiguous states |
| Context-label **distribution** (% per context) | no | which contexts each model gravitates to on real data |
| **Transitions / hour** + **single-tick flips** | no | label stability (flicker) — the core robustness claim from Ch5 |
| **Confidence** (mean) | no | how decisive each model is |
| **Latency** (mean / p95 / max, µs) | no | real-time suitability |
| **Smoothing** pass (flicker before vs after `TemporalSmoother`) | no | how much the production anti-flicker window helps each model |

This makes the real-data result a **generalisation / transfer** study, complementing
the simulator accuracy result rather than replacing it.

> **Note on motion/presence:** PIR fires repeatedly while a room is occupied; a
> room's presence decays from the **last** `on` event using *dataset* time
> (the benchmark sets `SensorWindow.Now` to each tick's dataset timestamp). If the
> dataset also has `off` events, they clear presence immediately. This is robust
> whether or not the export contains `off` rows.

---

## Step 1 — Download + preprocess (one time)

Download the two archives from Zenodo and unzip:

- `event_data.csv` (PIR + door events)
- `periodic_data_monthly_csv/periodic_data_*.csv` (environmental series)

Then preprocess into per-participant folders:

```bash
python3 scripts/preprocess_hslu_14243471.py \
  --events ~/Downloads/event_data.csv \
  --periodic-dir ~/Downloads/periodic_data_monthly_csv \
  --output-dir ./datasets/hslu_processed
```

This writes `datasets/hslu_processed/user_<id>/{merged_timeline.csv,sensors_manifest.json,...}`
for each participant (large CSVs are gitignored).

## Step 2 — Coverage census (decide who is benchmarkable)

The models need certain channels per room. The census reports who has them:

```bash
make bench-census        # DATAROOT defaults to ../datasets/hslu_processed
# or: cd backend && go run ./cmd/fusion-bench -census ../datasets/hslu_processed
```

- **usable** = temperature + humidity present, plus living-room light + motion
  (comfort / alerts / TV / reading can fire).
- **full** = usable **and** kitchen light+motion **and** bedroom light+motion
  (every context class is structurally reachable).

Output: a table on stdout + `bench_out/census.csv`. Use it to pick the
coverage-filtered subset.

## Step 3 — Run the benchmark

One participant:

```bash
make bench PARTICIPANT=7
# tune: make bench PARTICIPANT=7 GRID=60s WINDOW_DAYS=14
```

All participants + a pooled summary (run after filtering with the census, or on
everything):

```bash
make bench-all                 # GRID=60s WINDOW_DAYS=14 by default
make bench-all WINDOW_DAYS=0   # whole timeline per user (heavier)
```

Flags (see `go run ./cmd/fusion-bench -h`):

| Flag | Default | Meaning |
|------|---------|---------|
| `-data-dir` / `-data-root` | — | one `user_<id>/` folder, or a root of them |
| `-user-id` | inferred | participant id (matches the CSV `id` column) |
| `-grid` | `60s` | dataset-time spacing between fused decisions |
| `-window-days` | `14` | window length from first reading (`0` = whole timeline) |
| `-window-start` | first row | optional RFC3339 start; earlier rows warm up state only |
| `-smoother-n` | `2` | smoother window (matches production `MultiPredictor`) |
| `-out` | `bench_out` | output directory |

## Step 4 — Outputs

```
bench_out/
├── census.csv                 # coverage of every participant
├── per_user_table.csv         # one row per user — ready for a thesis table
├── pooled_summary.{json,txt}  # aggregate across the benchmarked subset
└── user_<id>/
    ├── per_tick.csv           # every decision: both labels, confidences, latency, agree
    ├── summary.json           # machine-readable metrics
    └── summary.txt            # human-readable summary (same as stdout)
```

`per_user_table.csv` columns: `decisions, span_hours, agreement_pct,
rule_avg_conf, fuzzy_avg_conf, rule_latency_mean_us, fuzzy_latency_mean_us,
rule_flicks_per_hr, fuzzy_flicks_per_hr, rule_single_flips, fuzzy_single_flips,
rule_flick_reduction_pct, fuzzy_flick_reduction_pct`.

---

## How it maps to the thesis

- **Chapter 5 (Technical Evaluation):** add a *real-data* subsection alongside the
  simulator results. Report the census (e.g. *N of 62 homes are benchmarkable*),
  then the pooled agreement, per-model distribution, flicker (raw and smoothed),
  confidence, and latency. Discuss disagreement concentrating in ambiguous states
  and whether the fuzzy boundary-smoothness advantage persists under real noise.
- **Chapter 6 (Conclusion / Limitations):** the as-is transfer framing and the
  no-labels limitation are already consistent with the "simulator rather than
  long-term physical deployment" limitation.

---

## Validation

A committed synthetic fixture (`backend/cmd/fusion-bench/testdata/bench/user_99`,
regenerate with `python3 .../gen_fixture.py`) drives a smoke test:

```bash
cd backend && go test ./cmd/fusion-bench/
```

It asserts the harness runs and reproduces the headline divergence: the rule
engine's priority-2 `NO_ONE_HOME` short-circuits before its priority-6
`SLEEPING` (so it never labels sleep), while the fuzzy engine's sleep-temp band
separates the two.

## Implementation note (production safety)

`SensorWindow` gained an optional `Now time.Time`. Production
(`buildSensorWindow`) leaves it zero, so the engines fall back to `time.Now()` and
**live behaviour is unchanged**; only the benchmark sets it (to dataset time) so
motion-recency is measured in dataset time rather than wall-clock time.
