import { clsx } from 'clsx'
import type { SensorSnapshot } from '@/types/sensor'
import { DeviceType, type Device } from '@/types/device'
import { ROOM_LABELS } from '@/utils/constants'
import { formatSensorScalar } from '@/utils/formatSensorValue'
import { cardSurface } from '@/utils/uiClasses'

function readingLabel(
  snapshot: SensorSnapshot | null,
  predicate: (r: { sensorId: string; unit?: string }) => boolean,
): string | null {
  if (!snapshot?.readings?.length) return null
  const r = snapshot.readings.find(predicate)
  if (!r) return null
  const v = formatSensorScalar(r.value)
  return r.unit ? `${v}${r.unit}` : v
}

export interface EnvironmentSummaryBarProps {
  snapshot: SensorSnapshot | null
  devices: Device[]
  /** Optional room id to label the bar (e.g. primary room) */
  roomId?: string
  /** Hide humidity slot when false */
  showHumidity?: boolean
  /** Hide lights summary when false */
  showLights?: boolean
}

export function EnvironmentSummaryBar({
  snapshot,
  devices,
  roomId,
  showHumidity = true,
  showLights = true,
}: EnvironmentSummaryBarProps) {
  let temp =
    readingLabel(snapshot, (r) => {
      const s = r.sensorId.toLowerCase()
      return s.includes('temp') || r.unit === '°C' || r.unit === 'C'
    }) ?? null

  let humidity =
    readingLabel(snapshot, (r) => {
      const s = r.sensorId.toLowerCase()
      return s.includes('humid') || (r.unit === '%' && !s.includes('light'))
    }) ?? null

  if (!temp) {
    const t = devices.find((d) => d.type === DeviceType.THERMOSTAT)
    temp =
      t != null && t.state.value != null
        ? `${formatSensorScalar(t.state.value)}°C`
        : '—'
  }

  if (!humidity && showHumidity) {
    humidity = '—'
  }

  const lights = devices.filter((d) => d.type === DeviceType.LIGHT)
  const lightsOn = lights.filter((d) => d.state.on).length
  const lightsLabel =
    lights.length > 0 ? `${lightsOn}/${lights.length} lights on` : '—'

  const roomLabel =
    roomId && roomId in ROOM_LABELS
      ? ROOM_LABELS[roomId as keyof typeof ROOM_LABELS]
      : roomId

  return (
    <div
      className={clsx(
        cardSurface,
        'flex flex-wrap items-center gap-3 border border-white/10 py-3 text-sm shadow-sm shadow-black/20',
      )}
    >
      <div className="flex min-w-[5rem] flex-col gap-0.5">
        <span className="text-xs text-[var(--color-text-secondary)] sm:text-sm">
          Temperature
        </span>
        <span className="font-semibold tabular-nums text-[var(--color-text-primary)]">
          {temp}
        </span>
      </div>
      {showHumidity ? (
        <div className="flex min-w-[5rem] flex-col gap-0.5 border-l border-white/10 pl-3">
          <span className="text-xs text-[var(--color-text-secondary)] sm:text-sm">
            Humidity
          </span>
          <span className="font-semibold tabular-nums text-[var(--color-text-primary)]">
            {humidity ?? '—'}
          </span>
        </div>
      ) : null}
      {showLights ? (
        <div className="flex min-w-[6rem] flex-col gap-0.5 border-l border-white/10 pl-3">
          <span className="text-xs text-[var(--color-text-secondary)] sm:text-sm">
            Lights
          </span>
          <span className="font-medium text-[var(--color-text-primary)]">
            {lightsLabel}
          </span>
        </div>
      ) : null}
      {roomLabel ? (
        <div className="ml-auto flex flex-col gap-0.5 text-right">
          <span className="text-xs text-[var(--color-text-secondary)] sm:text-sm">
            Room
          </span>
          <span className="font-medium text-[var(--color-text-primary)]">
            {roomLabel}
          </span>
        </div>
      ) : null}
    </div>
  )
}
