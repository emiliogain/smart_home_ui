import { Camera, Home } from 'lucide-react'
import { clsx } from 'clsx'
import { useState } from 'react'
import { useContextStore } from '@/store/contextStore'
import { EnvironmentSummaryBar } from '@/components/adaptive/EnvironmentSummaryBar'
import { SecurityCard } from '@/components/controls/SecurityCard'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { useDevices } from '@/hooks/useDevices'
import { cardSurface, headingPage } from '@/utils/uiClasses'

export function ContextNoOneHome() {
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const { devices, getDeviceById, handleToggleLocked } = useDevices()
  const [awayArmed, setAwayArmed] = useState(true)

  const front = getDeviceById('front_door')
  const back = getDeviceById('back_door')
  const window = getDeviceById('window_sensor')

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 lg:max-w-none">
      <section>
        <h2 className={clsx(headingPage, 'mb-3')}>Security overview</h2>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
          {front ? (
            <SecurityCard device={front} onToggle={() => handleToggleLocked(front.id)} />
          ) : null}
          {back ? (
            <SecurityCard device={back} onToggle={() => handleToggleLocked(back.id)} />
          ) : null}
          {window ? (
            <SecurityCard
              device={window}
              onToggle={() => handleToggleLocked(window.id)}
            />
          ) : null}
        </div>
      </section>

      <section>
        <h2 className={clsx(headingPage, 'mb-3')}>Camera feeds</h2>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
          <div className="flex aspect-video items-center justify-center rounded-xl bg-gray-800/80 text-[var(--color-text-secondary)] transition-all hover:ring-1 hover:ring-white/10">
            <div className="flex flex-col items-center gap-2">
              <Camera className="h-10 w-10 opacity-60" strokeWidth={1.5} />
              <span className="text-sm">Front camera</span>
            </div>
          </div>
        </div>
      </section>

      <div
        className={clsx(
          cardSurface,
          'mx-auto w-full max-w-sm border border-white/10 p-5 text-center shadow-black/25',
        )}
      >
        <ToggleSwitch
          isOn={awayArmed}
          onToggle={() => setAwayArmed((v) => !v)}
          label="Away Mode"
          icon={<Home className="h-6 w-6" strokeWidth={2} />}
        />
        <p className="mt-3 text-sm text-[var(--color-text-secondary)]">
          Extra alerts and camera monitoring while you&apos;re out.
        </p>
      </div>

      <EnvironmentSummaryBar
        snapshot={sensorSnapshot}
        devices={devices}
        roomId="living_room"
      />
    </div>
  )
}
