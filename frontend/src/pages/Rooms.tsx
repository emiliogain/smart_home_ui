import { LayoutGroup, motion } from 'framer-motion'
import { clsx } from 'clsx'
import { useMemo, useState } from 'react'
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

type RoomTab = (typeof ROOMS)[number]

function readingsForRoom(
  readings: SensorReading[] | undefined,
  roomId: RoomTab,
): SensorReading[] {
  if (!readings?.length) return []
  const key = roomId.split('_')[0].toLowerCase()
  return readings.filter((r) => r.sensorId.toLowerCase().includes(key))
}

export default function Rooms() {
  const [active, setActive] = useState<RoomTab>('living_room')
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

  const roomReadings = useMemo(
    () => readingsForRoom(sensorSnapshot?.readings, active),
    [sensorSnapshot, active],
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
        {roomReadings.length === 0 ? (
          <p className="rounded-xl border border-dashed border-white/15 bg-[var(--color-surface)]/50 px-3 py-4 text-sm text-[var(--color-text-secondary)]">
            No readings matched this room in the current snapshot. Values appear
            when the backend publishes sensor data for this zone.
          </p>
        ) : (
          <ul
            className={clsx(
              cardSurface,
              'space-y-2 border border-white/10 py-3',
            )}
          >
            {roomReadings.map((r) => (
              <li
                key={`${r.sensorId}-${r.at}`}
                className="flex flex-wrap items-baseline justify-between gap-2 text-sm"
              >
                <span className="text-[var(--color-text-secondary)]">
                  {r.sensorId}
                </span>
                <span className="font-medium tabular-nums text-[var(--color-text-primary)]">
                  {formatSensorScalar(r.value)}
                  {r.unit ? ` ${r.unit}` : ''}
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
