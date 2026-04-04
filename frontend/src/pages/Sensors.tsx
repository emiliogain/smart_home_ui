import { clsx } from 'clsx'
import { useMemo } from 'react'
import { useContextStore } from '@/store/contextStore'
import type { SensorSnapshot } from '@/types/sensor'
import { CONTEXT_LABELS } from '@/utils/constants'
import {
  generateMockHistory,
  generateMockMotionHistory,
} from '@/utils/mockSensorHistory'
import { ConfidenceBadge } from '@/components/common/ConfidenceBadge'
import { SensorCard } from '@/components/cards/SensorCard'
import { useDevices } from '@/hooks/useDevices'
import { cardSurface, headingPage } from '@/utils/uiClasses'

function findReading(snapshot: SensorSnapshot | null, part: string) {
  return snapshot?.readings.find((r) =>
    r.sensorId.toLowerCase().includes(part.toLowerCase()),
  )
}

export default function Sensors() {
  const currentContext = useContextStore((s) => s.currentContext)
  const confidence = useContextStore((s) => s.confidence)
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const { getDeviceById } = useDevices()

  const ctxLabel = CONTEXT_LABELS[currentContext]

  const tempCurrent =
    findReading(sensorSnapshot, 'temp')?.value ??
    getDeviceById('thermostat_living')?.state.value ??
    22
  const humidCurrent =
    findReading(sensorSnapshot, 'humid')?.value ?? 48
  const lightCurrent =
    findReading(sensorSnapshot, 'light')?.value ??
    findReading(sensorSnapshot, 'lux')?.value ??
    320
  const motionCurrent =
    findReading(sensorSnapshot, 'motion')?.value ?? 0

  const tempHistory = useMemo(
    () => generateMockHistory(Number(tempCurrent), 20),
    [tempCurrent],
  )
  const humidHistory = useMemo(
    () => generateMockHistory(Number(humidCurrent), 20),
    [humidCurrent],
  )
  const lightHistory = useMemo(
    () => generateMockHistory(Number(lightCurrent), 20),
    [lightCurrent],
  )
  const motionHistory = useMemo(() => generateMockMotionHistory(20), [])

  const sensorCount = sensorSnapshot?.readings.length ?? 0

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
          History below is mocked around the current reading when the snapshot
          is sparse ({sensorCount} reading{sensorCount === 1 ? '' : 's'} in
          snapshot).
        </p>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 lg:gap-5">
          <SensorCard
            sensorId="temperature.primary"
            type="temperature"
            value={tempCurrent}
            unit="°C"
            room="living_room"
            history={tempHistory}
          />
          <SensorCard
            sensorId="humidity.primary"
            type="humidity"
            value={humidCurrent}
            unit="%"
            room="living_room"
            history={humidHistory}
          />
          <SensorCard
            sensorId="light.level"
            type="light"
            value={lightCurrent}
            unit="lux"
            room="living_room"
            history={lightHistory}
          />
          <SensorCard
            sensorId="motion.pir"
            type="motion"
            value={motionCurrent}
            unit=""
            room="living_room"
            history={motionHistory}
            chartVariant="motion"
          />
        </div>
      </section>
    </div>
  )
}
