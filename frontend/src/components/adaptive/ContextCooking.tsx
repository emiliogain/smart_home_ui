import { clsx } from 'clsx'
import { Fan } from 'lucide-react'
import { useContextStore } from '@/store/contextStore'
import { EnvironmentSummaryBar } from '@/components/adaptive/EnvironmentSummaryBar'
import { DeviceCard } from '@/components/controls/DeviceCard'
import { LightDimmer } from '@/components/controls/LightDimmer'
import { TimerWidget } from '@/components/controls/TimerWidget'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { useDevices } from '@/hooks/useDevices'
import { cardSurface } from '@/utils/uiClasses'

export function ContextCooking() {
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const { devices, getDeviceById, handleToggle, handleSetValue } = useDevices()

  const exhaust = getDeviceById('exhaust_fan')
  const oven = getDeviceById('oven')
  const kitchenLight = getDeviceById('kitchen_light')

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 lg:max-w-none">
      <div className="w-full">
        <TimerWidget />
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
        {exhaust ? (
          <div className={clsx(cardSurface, 'w-full')}>
            <ToggleSwitch
              isOn={exhaust.state.on}
              onToggle={() => handleToggle(exhaust.id)}
              label={exhaust.name}
              icon={<Fan className="h-5 w-5" strokeWidth={2} />}
              disabled={!exhaust.controllable}
            />
          </div>
        ) : null}
        {oven ? (
          <DeviceCard device={oven} onToggle={() => handleToggle(oven.id)} />
        ) : null}
      </div>

      {kitchenLight ? (
        <LightDimmer
          device={kitchenLight}
          onToggle={() => handleToggle(kitchenLight.id)}
          onSetValue={(v) => handleSetValue(kitchenLight.id, v)}
        />
      ) : null}

      <EnvironmentSummaryBar
        snapshot={sensorSnapshot}
        devices={devices}
        roomId="kitchen"
        showLights={false}
      />
    </div>
  )
}
