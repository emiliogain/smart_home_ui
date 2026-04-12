# Developer Handoff — Smart Home Adaptive UI

## Project Overview

A diploma thesis project implementing a smart home system that:
1. Collects data from sensors (PIR motion, temperature, humidity, light)
2. Uses a **rule-based fusion engine** to infer the user's current context (e.g., "cooking in kitchen", "reading", "sleeping")
3. Dynamically adapts the frontend UI based on the inferred context

**Stack:** Go backend (Gin, hexagonal architecture) · React/TypeScript frontend · PostgreSQL · Socket.IO for real-time push

---

## Repository Structure

```
smart_home_ui/
├── backend/
│   ├── cmd/
│   │   ├── smart-home-backend/main.go   # Server entry point
│   │   └── simulator/main.go            # Standalone CLI simulator (fallback)
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── sensor/entity.go         # Sensor, Reading, EnrichedReading
│   │   │   └── context/entity.go        # ContextType, ContextUpdate, SensorSnapshot
│   │   ├── ports/secondary/
│   │   │   ├── sensor_repository.go     # SensorRepository interface
│   │   │   ├── fusion.go                # FusionPredictor interface + SensorWindow
│   │   │   └── websocket.go             # EventBroadcaster interface + NoopBroadcaster
│   │   ├── app/
│   │   │   └── sensor_service.go        # Core orchestration: save → fuse → broadcast
│   │   ├── simulator/                   # Embedded simulator library
│   │   │   ├── engine.go                # Controllable Engine struct
│   │   │   ├── scenario.go              # 7 scenarios with sensor profiles + noise
│   │   │   ├── sensor_defs.go           # DefaultSensors (8 sensors, 3 rooms)
│   │   │   └── options.go               # WithInterval, WithScenario options
│   │   └── adapters/
│   │       ├── primary/
│   │       │   ├── http/
│   │       │   │   ├── sensor_handler.go    # /api/v1/sensors CRUD
│   │       │   │   ├── context_handler.go   # GET /api/context/current
│   │       │   │   ├── admin_handler.go     # /api/admin/simulator/* + /admin page
│   │       │   │   └── admin.html           # Embedded admin panel (//go:embed)
│   │       │   └── websocket/hub.go         # Socket.IO server (googollee/go-socket.io)
│   │       └── secondary/
│   │           ├── database/
│   │           │   ├── postgres.go
│   │           │   └── sensor_repository.go
│   │           └── fusion/
│   │               ├── rule_engine.go       # Rule-based predictor (main implementation)
│   │               └── client.go            # StubPredictor (always returns COMFORTABLE)
│   ├── migrations/
│   │   ├── 001_create_initial_schema.sql    # sensors, sensor_data, devices tables
│   │   └── 002_add_sensor_readings_table.sql # sensor_readings (scalar value + unit)
│   ├── config/config.yaml                   # Local dev config (not committed)
│   └── pkg/client/client.go                 # HTTP client used by standalone CLI
├── frontend/src/
│   ├── App.tsx                  # WebSocket init + mock fallback logic
│   ├── api/
│   │   ├── websocket.ts         # Socket.IO client setup + event listeners
│   │   ├── sensors.ts           # REST: /api/sensors, /api/context/current
│   │   └── devices.ts           # REST: /api/devices, /api/scenes
│   ├── store/
│   │   ├── contextStore.ts      # currentContext, confidence, sensorSnapshot (Zustand)
│   │   ├── deviceStore.ts       # device states
│   │   └── settingsStore.ts     # adaptive/study mode, session logging CSV export
│   ├── pages/
│   │   ├── Dashboard.tsx        # Adaptive context-driven dashboard
│   │   ├── Rooms.tsx            # Per-room sensor + device view
│   │   ├── Sensors.tsx          # Sensor stream cards with mini-charts
│   │   └── Settings.tsx         # Mode toggles, session logging for user study
│   └── utils/
│       ├── mockContextProvider.ts  # Fallback mock context cycle (15s interval)
│       └── adaptationRules.ts      # Context → UI config mapping
└── docker-compose.yml               # postgres, redis, backend, frontend services
```

---

## Architecture: Data Flow

```
Simulator Engine (embedded goroutine in server)
    │  SaveReadingsBatch(ctx, []Reading)   ← all 8 sensors at once
    ▼
SensorService.SaveReadingsBatch()
    │  1. Persist all readings to sensor_readings table (Postgres)
    │  2. GetAllLatestReadings() → []EnrichedReading (name, type, location included)
    │  3. buildSensorWindow() → SensorWindow{ByType, ByLocation}
    │  4. RuleBasedPredictor.Predict(window) → FusionResult{Label, Confidence}
    │  5. buildSnapshot() → SensorSnapshot   ← sensorId = sensor NAME (not UUID)
    │  6. BroadcastContextUpdate() → Socket.IO → all connected frontend clients
    ▼
Frontend: socket.on('context_update', data => contextStore.setContext(data))
    │  Zustand store: currentContext, confidence, sensorSnapshot
    ▼
React components re-render (Dashboard adapts layout, Rooms shows sensor values)
```

**Critical design note — why batch matters:** The simulator sends 8 readings per tick. If each reading triggered fusion independently, the frontend would receive 8 intermediate context updates per tick (flickering between wrong states). `SaveReadingsBatch()` persists all readings first, then runs fusion once.

---

## The Fusion Engine

**File:** `backend/internal/adapters/secondary/fusion/rule_engine.go`

Priority-ordered rules. The first matching rule wins:

| Priority | Context | Trigger conditions |
|----------|---------|-------------------|
| 1 | `ALERT_TOO_HOT` | temp > 27°C |
| 2 | `ALERT_TOO_COLD` | temp < 17°C |
| 3 | `NO_ONE_HOME` | no motion in any room for > 5 min |
| 4 | `COOKING_KITCHEN` | kitchen motion + (temp > 23°C OR humidity > 50%) |
| 5 | `WATCHING_TV_LIVING_ROOM` | living room motion + **living room** light < 100 lux |
| 6 | `READING_LIVING_ROOM` | living room motion + **living room** light 200–500 lux |
| 7 | `SLEEPING` | no motion > 10 min + all lights < 10 lux + last motion = bedroom |
| 8 | `COMFORTABLE` | motion present, 17–27°C |
| 9 | `UNKNOWN` | nothing matched |

**Critical implementation note — per-location light:** The snapshot tracks light separately per room (`lightByLocation map[string]float64`). Rules 5 and 6 check `lightInLocation("living_room")`, not a global light average. Before this fix, the kitchen's 450 lux reading would prevent the TV/reading contexts from ever firing.

Every fusion pass is logged at INFO level:
```
[fusion] snapshot: temp=24.0 hum=58.0 lights=living_room=200 kitchen=450 bedroom=10 motion=kitchen=0s_ago
[fusion] → COOKING_KITCHEN (90%) reason: kitchen motion + temp=24.0 hum=58.0
```

---

## The Simulator

### Embedded mode (default, started automatically with the server)

Configured in `config/config.yaml`:
```yaml
simulator_enabled: true
simulator_interval: "5s"
simulator_scenario: "comfortable"
```

**8 sensors** across 3 rooms (see `internal/simulator/sensor_defs.go`):
```
temp_living_room     temperature  living_room
humidity_living_room humidity     living_room
light_living_room    light        living_room
motion_living_room   motion       living_room
light_kitchen        light        kitchen
motion_kitchen       motion       kitchen
light_bedroom        light        bedroom
motion_bedroom       motion       bedroom
```

Sensors are registered with **deterministic UUIDs** (`uuid.NewSHA1(uuid.NameSpaceDNS, "smarthome.sensor."+name)`) so restarting the server doesn't create duplicate sensor rows.

**7 scenarios** with calibrated values (see `internal/simulator/scenario.go`). Each profile has a base value and Gaussian noise:

| Scenario | temp | humidity | living light | kitchen light | motion |
|----------|------|----------|--------------|---------------|--------|
| comfortable | 22°C | 45% | 300 lux | 100 lux | living_room |
| reading | 21°C | 42% | 320 lux | 20 lux | living_room |
| watching_tv | 22°C | 44% | 50 lux | 10 lux | living_room |
| cooking | 24°C | 58% | 200 lux | 450 lux | kitchen |
| sleeping | 19°C | 40% | 2 lux | 0 lux | none |
| no_one_home | 21°C | 45% | 0 lux | 0 lux | none |
| alert_too_hot | 29°C | 55% | 400 lux | 200 lux | living_room |

### Admin panel (real-time control)

URL: `http://localhost:8080/admin`

REST API under `/api/admin/simulator/`:
- `GET /status` — full simulator state + current fusion context + sensor snapshot
- `POST /scenario {"scenario":"cooking"}` — switch scenario immediately
- `POST /control {"action":"pause"|"resume"|"toggle"}` — play/pause
- `POST /interval {"interval":"3s"}` — change tick rate (1s–30s)
- `POST /inject {"sensorName":"temp_living_room","value":30,"unit":"°C"}` — one-off injection

**Pause/Resume implementation note:** Pause/Resume are non-blocking — they just set a mutex-protected boolean (`e.paused`). The tick loop checks this flag at the start of each tick and skips if true. Earlier implementation used unbuffered channels for pause, which caused the HTTP handler to hang while `doTick()` was busy with DB writes.

### Standalone CLI (for testing without the server running)
```bash
cd backend
go run ./cmd/simulator --scenario cooking
go run ./cmd/simulator --cycle --cycle-duration 2m
```

---

## Frontend: Real Data vs Mocks

**File:** `frontend/src/App.tsx`

The app always tries the real WebSocket first. Mocks are a fallback:

1. `main.tsx` calls `initializeWebSocket()` — socket starts connecting
2. `App.tsx` useEffect starts a 3-second timer
3. **If socket connects within 3s** → mocks never arm
4. **If `context_update` event arrives** → any armed mocks are stopped immediately
5. **If 3s passes with no connection** → mock context cycle starts (15s intervals)

**Previous bug (now fixed):** There was an `if (import.meta.env.DEV) { armMocks(); return }` block that caused development mode to always use mocks, bypassing the WebSocket entirely. This was removed.

**Mock context cycle** (`mockContextProvider.ts`): Rotates through COMFORTABLE → READING → COOKING → SLEEPING → NO_ONE_HOME → ALERT_TOO_HOT every 15 seconds, with random sensor value drift every 5 seconds. Used only when backend is unreachable.

---

## Frontend: Sensor Display

The frontend reads sensor data from `contextStore.sensorSnapshot` which is populated by WebSocket `context_update` events.

**`sensorId` in snapshot = sensor name** (e.g. `temp_living_room`), NOT the database UUID.

This matters because both `Rooms.tsx` and `Sensors.tsx` use substring matching:

```typescript
// Rooms.tsx — filter sensors for current room tab
const key = roomId.split('_')[0]  // "living_room" → "living"
readings.filter(r => r.sensorId.toLowerCase().includes(key))
// "temp_living_room".includes("living") → ✅

// Sensors.tsx — find temperature value
findReading(snapshot, 'temp')?.value
// "temp_living_room".includes("temp") → ✅
```

**How the name gets into the snapshot:** `buildSnapshot()` in `sensor_service.go` uses `r.SensorName` (from `EnrichedReading`, which is populated by the SQL lateral join that selects `s.name`). If `SensorName` is empty, it falls back to the UUID.

---

## Database

```
postgres://postgres:password@localhost:5432/smarthome
```

Key tables:
| Table | Purpose |
|-------|---------|
| `sensors` | Registered sensors: id (UUID), name, type, location, status |
| `sensor_readings` | Scalar readings: sensor_id FK, value DOUBLE, unit, timestamp |
| `sensor_data` | Original JSONB readings (unused by current code, kept from initial schema) |
| `devices` | Smart home devices with JSONB state |
| `device_commands` | Command log |

Run migrations: `cd backend && make migrate-up`

---

## Running Locally

### All-in-one
```bash
make demo         # starts postgres+redis, migrations, frontend (bg), simulator (bg), backend (fg)
make demo-stop    # kill background processes
```

### Manual
```bash
docker-compose up -d postgres

cd backend
make migrate-up
make run          # backend on :8080, simulator starts automatically

# separate terminal
cd frontend
npm run dev       # frontend on :5173

open http://localhost:5173   # main frontend
open http://localhost:8080/admin  # simulator control panel
```

---

## API Reference

| Method | Path | Notes |
|--------|------|-------|
| POST | `/api/v1/sensors` | Register sensor `{name, type, location}` |
| GET | `/api/v1/sensors` | List all sensors |
| POST | `/api/v1/sensors/:id/readings` | Submit reading `{value, unit}` |
| GET | `/api/sensors` | Same list, no version prefix (frontend uses this) |
| POST | `/api/sensors/:id/readings` | Same submit, no version prefix |
| GET | `/api/context/current` | Latest `ContextUpdate` JSON |
| GET | `/api/admin/simulator/status` | Simulator state + context + snapshot |
| POST | `/api/admin/simulator/scenario` | Switch scenario |
| POST | `/api/admin/simulator/control` | pause/resume/toggle |
| POST | `/api/admin/simulator/interval` | Change tick rate |
| POST | `/api/admin/simulator/inject` | One-off sensor reading |
| GET | `/admin` | Admin panel HTML |
| GET/POST | `/socket.io/*any` | Socket.IO WebSocket |

**Socket.IO events:**
- `context_update` (server → client): `ContextUpdate` payload
- `device_state_update` (server → client): `{deviceId, state}`
- `device_command` (client → server): logged, not yet handled

---

## Known Issues / Remaining Work

### Sensor cards on Sensors page show only living room sensors
`Sensors.tsx` hardcodes 4 sensor cards (temp, humidity, light, motion) all labelled `living_room`. Kitchen and bedroom sensors are not shown as separate cards — they exist in the snapshot but are not displayed. The page needs to be extended to render one card per sensor (or per sensor type per room).

### Device endpoints not implemented in backend
The frontend expects `GET /api/devices`, `POST /api/devices/:id/control`, `GET /api/scenes`, `POST /api/scenes/:id/apply`. These routes don't exist yet. The frontend falls back silently to the mock device list in `constants.ts` (14 hardcoded devices). Device controls visually work but don't actually hit the backend.

### Socket.IO version compatibility
The backend uses `github.com/googollee/go-socket.io v1.7.0`, which supports Socket.IO v3/v4 with WebSocket transport. The frontend uses `socket.io-client@4.8.1` with `transports: ['websocket']`. If WebSocket connection issues appear, verify that CORS is allowing the frontend origin and that the `/socket.io/` path is not being blocked.

### No authentication
No JWT or session management. Intended for local network / thesis demo use only.

---

## Thesis Context

The thesis compares adaptive vs static UI for smart home control. Key metrics collected via `Settings.tsx`:
- **Session Logging** — records every context change + device interaction to localStorage, exportable as CSV
- **Study Mode** — overrides context locally for controlled A/B testing without needing the backend to produce a specific scenario
- **SUS scores** collected post-study via questionnaire (offline)

The methodology (`methodology_temp.md`) plans to compare at least two fusion approaches. The current rule-based engine is the primary implementation. A fuzzy logic or ML-based predictor could be added as an alternative `FusionPredictor` implementation behind the same port interface without changing any other code.
