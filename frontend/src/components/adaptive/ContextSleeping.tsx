import { BellOff, Moon } from 'lucide-react'
import { useState } from 'react'
import { useContextStore } from '@/store/contextStore'
import { EnvironmentSummaryBar } from '@/components/adaptive/EnvironmentSummaryBar'
import { SceneSuggestion } from '@/components/cards/SceneSuggestion'
import { AlarmWidget } from '@/components/controls/AlarmWidget'
import { ThermostatControl } from '@/components/controls/ThermostatControl'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { DeviceType } from '@/types/device'
import { useAdaptation } from '@/hooks/useAdaptation'
import { useDevices } from '@/hooks/useDevices'

export function ContextSleeping() {
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const { config } = useAdaptation()
  const { devices, getDeviceById, handleToggle, handleSetValue } = useDevices()
  const [showScene, setShowScene] = useState(true)
  const [dnd, setDnd] = useState(true)

  const alarm = getDeviceById('alarm_clock')
  const thermoBed = getDeviceById('thermostat_bedroom')
  const scene = config.suggestedScene

  const lights = devices.filter((d) => d.type === DeviceType.LIGHT)
  const lightsOn = lights.filter((d) => d.state.on).length

  return (
    <div className="min-h-[60vh] rounded-2xl border border-white/5 px-3 py-8 text-[var(--color-text-primary)]">
      <div className="mx-auto flex w-full max-w-3xl flex-col gap-8 lg:max-w-none">
        <header className="flex flex-col items-center gap-3 text-center">
          <Moon
            className="h-12 w-12 text-[var(--color-secondary)] opacity-80 sm:h-14 sm:w-14"
            strokeWidth={1.25}
          />
          <h1 className="text-lg font-semibold tracking-wide text-[var(--color-text-primary)]">
            Good Night
          </h1>
          <p className="theme-muted max-w-sm text-sm text-[var(--color-text-secondary)]">
            Rest well. The home is settling down for sleep.
          </p>
        </header>

        {alarm ? (
          <AlarmWidget
            device={alarm}
            onToggle={() => handleToggle(alarm.id)}
            onSetValue={(v) => handleSetValue(alarm.id, v)}
          />
        ) : null}

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
          <div className="rounded-xl border border-white/5 bg-white/[0.04] p-4 shadow-md shadow-black/30 transition-all hover:ring-1 hover:ring-white/10">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              All Lights Off
            </h3>
            <p className="mt-2 text-2xl font-semibold tabular-nums text-[var(--color-text-primary)] sm:text-3xl">
              {lightsOn === 0 ? 'All off' : `${lightsOn} still on`}
            </p>
            <p className="theme-muted mt-1 text-sm text-[var(--color-text-secondary)]">
              {lights.length} lights total
            </p>
          </div>
          {thermoBed ? (
            <ThermostatControl
              device={thermoBed}
              onSetValue={(v) => handleSetValue(thermoBed.id, v)}
            />
          ) : null}
        </div>

        <div className="rounded-xl border border-white/5 bg-white/[0.04] p-4 transition-all hover:ring-1 hover:ring-white/10">
          <ToggleSwitch
            isOn={dnd}
            onToggle={() => setDnd((d) => !d)}
            label="Do Not Disturb"
            icon={<BellOff className="h-5 w-5" strokeWidth={2} />}
          />
        </div>

        {showScene && scene ? (
          <SceneSuggestion
            scene={scene}
            onApply={() => setShowScene(false)}
            onDismiss={() => setShowScene(false)}
          />
        ) : null}

        <EnvironmentSummaryBar
          snapshot={sensorSnapshot}
          devices={devices}
          roomId="bedroom"
        />
      </div>
    </div>
  )
}
