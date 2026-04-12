import { useState } from 'react'
import { useContextStore } from '@/store/contextStore'
import { EnvironmentSummaryBar } from '@/components/adaptive/EnvironmentSummaryBar'
import { SceneSuggestion } from '@/components/cards/SceneSuggestion'
import { LightDimmer } from '@/components/controls/LightDimmer'
import { ThermostatControl } from '@/components/controls/ThermostatControl'
import { useAdaptation } from '@/hooks/useAdaptation'
import { useDevices } from '@/hooks/useDevices'

export function ContextWatchingTV() {
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const { config } = useAdaptation()
  const { devices, getDeviceById, handleToggle, handleSetValue } = useDevices()
  const [showScene, setShowScene] = useState(true)

  const ceiling = getDeviceById('ceiling_light_living')
  const thermo = getDeviceById('thermostat_living')
  const scene = config.suggestedScene

  return (
    <div className="relative mx-auto flex w-full max-w-3xl flex-col gap-6 lg:max-w-none">
      <div
        className="pointer-events-none absolute inset-0 -m-4 rounded-2xl bg-black/35"
        aria-hidden
      />

      <div className="relative z-[1] flex flex-col gap-6 lg:max-w-none">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-stretch sm:gap-6">
          {ceiling ? (
            <div className="w-full min-w-0 sm:flex-1">
              <p className="theme-muted mb-2 text-center text-sm text-[var(--color-text-secondary)]">
                Suggested: dim to ~20% for movie watching
              </p>
              <LightDimmer
                device={ceiling}
                onToggle={() => handleToggle(ceiling.id)}
                onSetValue={(v) => handleSetValue(ceiling.id, v)}
              />
            </div>
          ) : null}

          {showScene && scene ? (
            <div className="w-full min-w-0 sm:flex-1">
              <SceneSuggestion
                scene={scene}
                onApply={() => setShowScene(false)}
                onDismiss={() => setShowScene(false)}
              />
            </div>
          ) : null}
        </div>

        {thermo ? (
          <ThermostatControl
            device={thermo}
            onSetValue={(v) => handleSetValue(thermo.id, v)}
          />
        ) : null}

        <EnvironmentSummaryBar
          snapshot={sensorSnapshot}
          devices={devices}
          roomId="living_room"
        />
      </div>
    </div>
  )
}
