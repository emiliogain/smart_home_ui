import { LayoutGroup, motion } from 'framer-motion'
import { clsx } from 'clsx'
import { useEffect, useMemo, useState } from 'react'
import { useContextStore } from '@/store/contextStore'
import type { SensorReading } from '@/types/sensor'
import { ROOMS, ROOM_LABELS } from '@/utils/constants'
import {
  renderDeviceControl,
  type DeviceControlHandlers,
} from '@/components/adaptive/renderDeviceControl'
import { useDevices } from '@/hooks/useDevices'
import { formatSensorScalar } from '@/utils/formatSensorValue'
import { cardSurface, headingPage, interactiveSurface } from '@/utils/uiClasses'
import { listSensors } from '@/api/sensors'
import type { BackendSensor } from '@/api/types'

type RoomTab = (typeof ROOMS)[number]

/** Row shown in the sensor table — either live (from snapshot) or registered-but-pending. */
interface SensorRow {
  key: string
  label: string
  value: string
  unit: string
  pending: boolean
}

function buildSensorRows(
  registered: BackendSensor[],
  readings: SensorReading[] | undefined,
  roomId: RoomTab,
): SensorRow[] {
  const byName = new Map<string, SensorReading>()
  for (const r of readings ?? []) {
    const inRoom = r.location
      ? r.location === roomId
      : r.sensorId.toLowerCase().includes(roomId.split('_')[0])
    if (inRoom) byName.set(r.sensorId, r)
  }

  const roomSensors = registered.filter((s) => s.location === roomId)

  // If the backend returned sensors for this room, show them all (with live data when available).
  if (roomSensors.length > 0) {
    return roomSensors.map((s) => {
      const live = byName.get(s.id)
      return {
        key: s.id,
        label: s.display_label ?? s.name,
        value: live != null ? formatSensorScalar(live.value) : '—',
        unit: live?.unit?.trim() ?? '',
        pending: live == null,
      }
    })
  }

  // No registered sensors for this room — fall back to snapshot-only rows.
  return [...byName.values()].map((r) => ({
    key: `${r.sensorId}-${r.at}`,
    label: r.sensorLabel ?? r.sensorId,
    value: formatSensorScalar(r.value),
    unit: r.unit?.trim() ?? '',
    pending: false,
  }))
}

export default function Rooms() {
  const [active, setActive] = useState<RoomTab>('living_room')
  const [registeredSensors, setRegisteredSensors] = useState<BackendSensor[]>([])
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const {
    getDevicesByRoom,
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  } = useDevices()

  const handlers: DeviceControlHandlers = {
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  }

  useEffect(() => {
    listSensors().then(setRegisteredSensors)
    const id = setInterval(() => listSensors().then(setRegisteredSensors), 15_000)
    return () => clearInterval(id)
  }, [])

  const roomRows = useMemo(
    () => buildSensorRows(registeredSensors, sensorSnapshot?.readings, active),
    [registeredSensors, sensorSnapshot, active],
  )

  const devices = getDevicesByRoom(active)

  return (
    <div className="space-y-6 pt-2">
      <div>
        <h1 className={headingPage}>Rooms</h1>
        {/*
          overflow-x-auto makes overflow-y clip in CSS; pad vertically so focus rings
          and active:scale-95 are not clipped under the heading.
        */}
        <div className="-mx-1 mt-4 overflow-x-auto px-1 py-3 sm:mt-5 sm:py-3.5">
          <div className="flex min-w-min gap-2 px-1">
            {ROOMS.map((roomId) => (
              <button
                key={roomId}
                type="button"
                onClick={() => setActive(roomId)}
                className={clsx(
                  interactiveSurface,
                  'shrink-0 rounded-full px-4 py-2.5 text-sm font-medium ring-offset-2 ring-offset-[var(--color-bg)]',
                  active === roomId
                    ? 'bg-[var(--color-secondary)] text-[var(--color-primary)] hover:bg-[var(--color-secondary)]'
                    : 'bg-white/10 text-[var(--color-text-secondary)]',
                )}
              >
                {ROOM_LABELS[roomId]}
              </button>
            ))}
          </div>
        </div>
      </div>

      <section>
        <h2 className="mb-2 text-lg font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
          Sensors in {ROOM_LABELS[active]}
        </h2>
        {roomRows.length === 0 ? (
          <p className="rounded-xl border border-dashed border-white/15 bg-[var(--color-surface)]/50 px-3 py-4 text-sm text-[var(--color-text-secondary)]">
            No sensors registered for this room yet.
          </p>
        ) : (
          <ul
            className={clsx(
              cardSurface,
              'divide-y divide-white/10 border border-white/10',
            )}
          >
            {roomRows.map((row) => (
              <li
                key={row.key}
                className="flex flex-col gap-1 px-4 py-3 first:pt-3.5 last:pb-3.5"
              >
                <span className="text-sm leading-snug text-[var(--color-text-secondary)]">
                  {row.label}
                </span>
                <span className="flex flex-wrap items-baseline gap-x-1.5 gap-y-0">
                  <span
                    className={clsx(
                      'text-xl font-semibold tabular-nums leading-none',
                      row.pending
                        ? 'text-[var(--color-text-secondary)]'
                        : 'text-[var(--color-text-primary)]',
                    )}
                  >
                    {row.value}
                  </span>
                  {row.unit ? (
                    <span className="text-sm font-medium text-[var(--color-text-secondary)]">
                      {row.unit}
                    </span>
                  ) : null}
                  {row.pending && (
                    <span className="text-xs text-[var(--color-text-secondary)] opacity-60">
                      no data yet
                    </span>
                  )}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <h2 className="mb-3 text-lg font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
          Devices
        </h2>
        {devices.length === 0 ? (
          <p className="text-sm text-[var(--color-text-secondary)]">
            No devices in this room.
          </p>
        ) : (
          <LayoutGroup id={`room-devices-${active}`}>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
              {devices.map((d) => (
                <motion.div
                  key={d.id}
                  layout
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.22, layout: { duration: 0.28 } }}
                  className="min-w-0"
                >
                  {renderDeviceControl(d, handlers)}
                </motion.div>
              ))}
            </div>
          </LayoutGroup>
        )}
      </section>
    </div>
  )
}
