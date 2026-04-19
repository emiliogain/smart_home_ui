import { clsx } from 'clsx'
import { useEffect, useMemo, useState } from 'react'
import { useContextStore } from '@/store/contextStore'
import type { SensorSnapshot } from '@/types/sensor'
import { CONTEXT_LABELS, ROOM_LABELS, ROOMS } from '@/utils/constants'
import {
  generateMockHistory,
  generateMockMotionHistory,
} from '@/utils/mockSensorHistory'
import {
  fetchSensorReadings,
  listSensors,
  readingsToHistoryPoints,
} from '@/api/sensors'
import type { BackendSensor } from '@/api/types'
import type { SensorHistoryPoint } from '@/components/cards/SensorCard'
import { ConfidenceBadge } from '@/components/common/ConfidenceBadge'
import { SensorCard } from '@/components/cards/SensorCard'
import { cardSurface, headingPage } from '@/utils/uiClasses'

const READINGS_LIMIT = 80
/** How often to refetch sensor list + reading history from the REST API. */
const SENSORS_POLL_MS = 10_000

function findReadingBySensorId(
  snapshot: SensorSnapshot | null,
  sensorId: string,
) {
  return snapshot?.readings.find((r) => r.sensorId === sensorId)
}

function lastHistoryValue(
  history: SensorHistoryPoint[],
): number | undefined {
  if (!history.length) return undefined
  const v = history[history.length - 1]?.value
  return typeof v === 'number' ? v : undefined
}

/** Append fusion snapshot point when it is newer than REST data (between polls). */
function mergeLiveSnapshot(
  apiHistory: SensorHistoryPoint[],
  live: { value: number; at: string } | null | undefined,
): SensorHistoryPoint[] {
  if (!live) return apiHistory
  const point = { value: live.value, timestamp: live.at }
  if (!apiHistory.length) return [point]
  const last = apiHistory[apiHistory.length - 1]
  if (last.timestamp === point.timestamp && last.value === point.value) {
    return apiHistory
  }
  return [...apiHistory, point].slice(-READINGS_LIMIT)
}

function displayUnit(sensorType: string): string {
  const t = sensorType.toLowerCase()
  if (t.includes('temp')) return '°C'
  if (t.includes('humid')) return '%'
  if (t.includes('light') || t.includes('lux')) return 'lux'
  return ''
}

function isMotionSensor(sensorType: string): boolean {
  return sensorType.toLowerCase().includes('motion')
}

function defaultValueForType(sensorType: string): number {
  const t = sensorType.toLowerCase()
  if (t.includes('temp')) return 21
  if (t.includes('humid')) return 48
  if (t.includes('light')) return 280
  if (t.includes('motion')) return 0
  return 0
}

function fallbackHistory(
  sensorType: string,
  current: number,
): SensorHistoryPoint[] {
  if (isMotionSensor(sensorType)) return generateMockMotionHistory(20)
  return generateMockHistory(current, 20)
}

function typeSortKey(sensorType: string): number {
  const t = sensorType.toLowerCase()
  if (t.includes('temp')) return 0
  if (t.includes('humid')) return 1
  if (t.includes('light')) return 2
  if (t.includes('motion')) return 3
  return 10
}

function sortRoomKeys(locations: Iterable<string>): string[] {
  const order = new Map<string, number>(
    ROOMS.map((r, i) => [r, i]),
  )
  return [...new Set(locations)].sort((a, b) => {
    const ia = order.has(a) ? order.get(a)! : 100
    const ib = order.has(b) ? order.get(b)! : 100
    if (ia !== ib) return ia - ib
    return a.localeCompare(b)
  })
}

export default function Sensors() {
  const currentContext = useContextStore((s) => s.currentContext)
  const confidence = useContextStore((s) => s.confidence)
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)

  const [backendSensors, setBackendSensors] = useState<BackendSensor[]>([])
  const [historiesById, setHistoriesById] = useState<
    Record<string, SensorHistoryPoint[]>
  >({})

  useEffect(() => {
    let cancelled = false
    let pollWhileHidden = false

    const tick = async () => {
      if (pollWhileHidden) return
      const list = await listSensors()
      if (cancelled) return
      setBackendSensors(list)
      if (!list.length) {
        setHistoriesById({})
        return
      }
      const readingsLists = await Promise.all(
        list.map((s) => fetchSensorReadings(s.id, READINGS_LIMIT)),
      )
      if (cancelled) return
      const next: Record<string, SensorHistoryPoint[]> = {}
      list.forEach((s, i) => {
        next[s.id] = readingsToHistoryPoints(readingsLists[i] ?? [])
      })
      setHistoriesById(next)
    }

    const onVisibility = () => {
      pollWhileHidden = document.visibilityState === 'hidden'
      if (!pollWhileHidden && !cancelled) void tick()
    }
    onVisibility()
    document.addEventListener('visibilitychange', onVisibility)

    void tick()
    const intervalId = window.setInterval(() => void tick(), SENSORS_POLL_MS)

    return () => {
      cancelled = true
      window.clearInterval(intervalId)
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }, [])

  const sensorsByRoom = useMemo(() => {
    const map = new Map<string, BackendSensor[]>()
    for (const s of backendSensors) {
      const room = s.location || 'unknown'
      if (!map.has(room)) map.set(room, [])
      map.get(room)!.push(s)
    }
    for (const arr of map.values()) {
      arr.sort(
        (a, b) =>
          typeSortKey(a.type) - typeSortKey(b.type) ||
          a.name.localeCompare(b.name),
      )
    }
    return map
  }, [backendSensors])

  const roomKeys = useMemo(
    () => sortRoomKeys(sensorsByRoom.keys()),
    [sensorsByRoom],
  )

  const ctxLabel = CONTEXT_LABELS[currentContext]

  const lastUpdated = useContextStore((s) => s.lastUpdated)
  const sensorCount = sensorSnapshot?.readings.length ?? 0
  const apiHistoryCount = useMemo(
    () =>
      Object.values(historiesById).filter((h) => h.length > 0).length,
    [historiesById],
  )

  return (
    <div className="space-y-8 pt-2 lg:space-y-10">
      <div
        className={clsx(
          cardSurface,
          'flex flex-col items-center gap-3 border border-white/10 p-6 text-center shadow-black/20',
        )}
      >
        <div className="text-5xl leading-none sm:text-6xl" aria-hidden>
          {ctxLabel.emoji}
        </div>
        <div>
          <p className={headingPage}>{ctxLabel.label}</p>
          <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
            {ctxLabel.description}
          </p>
        </div>
        <div className="scale-125">
          <ConfidenceBadge confidence={confidence} />
        </div>
      </div>

      <section>
        <h2 className={clsx(headingPage, 'mb-2')}>Context fusion</h2>
        <div
          className={clsx(
            cardSurface,
            'border border-white/10 px-4 py-3 text-sm text-[var(--color-text-secondary)] shadow-sm shadow-black/15',
          )}
        >
          <span className="font-medium text-[var(--color-text-primary)]">
            Fuzzy Logic
          </span>
          <span className="text-[var(--color-text-secondary)]">
            {' '}
            — sensor weighting and membership functions (backend). This UI
            visualizes the fused context and raw signals for demos and debugging.
          </span>
        </div>
      </section>

      <section>
        <h2 className={clsx(headingPage, 'mb-3')}>Sensor streams</h2>
        <p className="mb-4 text-sm text-[var(--color-text-secondary)]">
          All sensors from{' '}
          <code className="text-xs text-[var(--color-text-primary)]">
            GET /api/sensors
          </code>
          , grouped by room. Values follow the live fusion snapshot when
          present; charts merge that snapshot with history from{' '}
          <code className="text-xs text-[var(--color-text-primary)]">
            GET /api/sensors/:id/readings
          </code>{' '}
          (refreshed every {SENSORS_POLL_MS / 1000}s while the tab is visible).
          Otherwise a short mock series is used. Fusion snapshot:{' '}
          {sensorCount} reading{sensorCount === 1 ? '' : 's'}
          {lastUpdated ? (
            <span className="text-[var(--color-text-secondary)]">
              {' '}
              · last context update{' '}
              <time dateTime={lastUpdated} className="tabular-nums">
                {new Date(lastUpdated).toLocaleTimeString()}
              </time>
            </span>
          ) : null}
          {apiHistoryCount > 0 ? (
            <span className="text-[var(--color-success)]">
              {' '}
              · {apiHistoryCount} sensor{apiHistoryCount === 1 ? '' : 's'} with API
              history
            </span>
          ) : null}
        </p>

        {!backendSensors.length ? (
          <p className="rounded-xl border border-white/10 bg-white/5 px-4 py-6 text-center text-sm text-[var(--color-text-secondary)]">
            No sensors returned from the API yet. Start the backend (and
            simulator) so sensors are registered, then refresh this page.
          </p>
        ) : (
          <div className="space-y-10">
            {roomKeys.map((room) => {
              const sensors = sensorsByRoom.get(room) ?? []
              if (!sensors.length) return null
              const roomLabel =
                room in ROOM_LABELS
                  ? ROOM_LABELS[room as keyof typeof ROOM_LABELS]
                  : room.replace(/_/g, ' ')

              return (
                <div key={room}>
                  <h3
                    className={clsx(
                      headingPage,
                      'mb-4 text-lg font-semibold text-[var(--color-text-primary)]',
                    )}
                  >
                    {roomLabel}
                  </h3>
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 lg:gap-5">
                    {sensors.map((s) => {
                      const apiHistory = historiesById[s.id] ?? []
                      const fromSnapshot = findReadingBySensorId(
                        sensorSnapshot,
                        s.id,
                      )
                      const live =
                        fromSnapshot != null
                          ? {
                              value: fromSnapshot.value,
                              at: fromSnapshot.at,
                            }
                          : null
                      const current =
                        fromSnapshot?.value ??
                        lastHistoryValue(apiHistory) ??
                        defaultValueForType(s.type)
                      const chartHistory =
                        apiHistory.length > 0
                          ? mergeLiveSnapshot(apiHistory, live)
                          : live
                            ? mergeLiveSnapshot([], live)
                            : fallbackHistory(s.type, current)
                      const unit = displayUnit(s.type)
                      return (
                        <SensorCard
                          key={s.id}
                          sensorId={s.id}
                          sensorLabel={s.display_label ?? s.name}
                          type={s.type}
                          value={current}
                          unit={unit}
                          room={s.location}
                          history={chartHistory}
                          chartVariant={
                            isMotionSensor(s.type) ? 'motion' : 'line'
                          }
                        />
                      )
                    })}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </section>
    </div>
  )
}
