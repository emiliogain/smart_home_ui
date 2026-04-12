import { useState } from 'react'
import { useContextStore } from '@/store/contextStore'
import { EnvironmentSummaryBar } from '@/components/adaptive/EnvironmentSummaryBar'
import { SceneSuggestion } from '@/components/cards/SceneSuggestion'
import { LightDimmer } from '@/components/controls/LightDimmer'
import { ThermostatControl } from '@/components/controls/ThermostatControl'
import { useAdaptation } from '@/hooks/useAdaptation'
import { useDevices } from '@/hooks/useDevices'
import { renderSilentToggle } from '@/components/adaptive/renderDeviceControl'

export function ContextReading() {
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const { config } = useAdaptation()
  const { devices, getDeviceById, handleToggle, handleSetValue } = useDevices()
  const [showScene, setShowScene] = useState(true)
  const [silent, setSilent] = useState(false)

  const lamp = getDeviceById('reading_lamp')
  const thermo = getDeviceById('thermostat_living')
  const scene = config.suggestedScene

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 lg:max-w-none">
      <div className="flex flex-col gap-6 sm:flex-row sm:items-stretch sm:gap-6">
        {lamp ? (
          <div className="w-full min-w-0 sm:flex-1 sm:max-w-none md:scale-[1.02] md:transform">
            <LightDimmer
              device={lamp}
              onToggle={() => handleToggle(lamp.id)}
              onSetValue={(v) => handleSetValue(lamp.id, v)}
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

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
        {thermo ? (
          <ThermostatControl
            device={thermo}
            onSetValue={(v) => handleSetValue(thermo.id, v)}
          />
        ) : null}
        {renderSilentToggle(silent, () => setSilent((s) => !s))}
      </div>

      <EnvironmentSummaryBar
        snapshot={sensorSnapshot}
        devices={devices}
        roomId="living_room"
      />
    </div>
  )
}
