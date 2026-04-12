import { Fan, VolumeX } from 'lucide-react'
import { clsx } from 'clsx'
import type { ReactNode } from 'react'
import { cardSurface } from '@/utils/uiClasses'
import { AlarmWidget } from '@/components/controls/AlarmWidget'
import { DeviceCard } from '@/components/controls/DeviceCard'
import { LightDimmer } from '@/components/controls/LightDimmer'
import { SecurityCard } from '@/components/controls/SecurityCard'
import { ThermostatControl } from '@/components/controls/ThermostatControl'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { DeviceType, type Device } from '@/types/device'

export interface DeviceControlHandlers {
  handleToggle: (deviceId: string) => void
  handleSetValue: (deviceId: string, value: number) => void
  handleToggleLocked: (deviceId: string) => void
}

export function renderDeviceControl(
  device: Device | undefined,
  h: DeviceControlHandlers,
): ReactNode {
  if (!device) return null

  switch (device.type) {
    case DeviceType.THERMOSTAT:
      return (
        <ThermostatControl
          key={device.id}
          device={device}
          onSetValue={(v) => h.handleSetValue(device.id, v)}
        />
      )
    case DeviceType.LIGHT:
      return (
        <LightDimmer
          key={device.id}
          device={device}
          onToggle={() => h.handleToggle(device.id)}
          onSetValue={(v) => h.handleSetValue(device.id, v)}
        />
      )
    case DeviceType.FAN:
    case DeviceType.EXHAUST_FAN:
      return (
        <div key={device.id} className={clsx(cardSurface, 'w-full')}>
          <ToggleSwitch
            isOn={device.state.on}
            onToggle={() => h.handleToggle(device.id)}
            label={device.name}
            icon={<Fan className="h-5 w-5" strokeWidth={2} />}
            disabled={!device.controllable}
          />
        </div>
      )
    case DeviceType.DOOR_LOCK:
    case DeviceType.WINDOW_SENSOR:
      return (
        <SecurityCard
          key={device.id}
          device={device}
          onToggle={() => h.handleToggleLocked(device.id)}
        />
      )
    case DeviceType.ALARM:
      return (
        <AlarmWidget
          key={device.id}
          device={device}
          onToggle={() => h.handleToggle(device.id)}
          onSetValue={(v) => h.handleSetValue(device.id, v)}
        />
      )
    default:
      return (
        <DeviceCard
          key={device.id}
          device={device}
          onToggle={() => h.handleToggle(device.id)}
        />
      )
  }
}

export function renderFanToggle(
  device: Device | undefined,
  onToggle: () => void,
): ReactNode {
  if (!device) return null
  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <ToggleSwitch
        isOn={device.state.on}
        onToggle={onToggle}
        label={device.name}
        icon={<Fan className="h-5 w-5" strokeWidth={2} />}
        disabled={!device.controllable}
      />
    </div>
  )
}

export function renderSilentToggle(
  silent: boolean,
  onToggle: () => void,
): ReactNode {
  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <ToggleSwitch
        isOn={silent}
        onToggle={onToggle}
        label="Silent Mode"
        icon={<VolumeX className="h-5 w-5" strokeWidth={2} />}
      />
    </div>
  )
}
