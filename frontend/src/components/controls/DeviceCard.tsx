import {
  AlarmClock,
  Camera,
  DoorOpen,
  Fan,
  Flame,
  Lightbulb,
  Lock,
  Sun,
  Thermometer,
  Wind,
} from 'lucide-react'
import { clsx } from 'clsx'
import { DeviceType, type Device } from '@/types/device'
import { ROOM_LABELS } from '@/utils/constants'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { cardSurface } from '@/utils/uiClasses'

function deviceIcon(type: DeviceType) {
  switch (type) {
    case DeviceType.LIGHT:
      return <Sun className="h-5 w-5" strokeWidth={2} />
    case DeviceType.THERMOSTAT:
      return <Thermometer className="h-5 w-5" strokeWidth={2} />
    case DeviceType.EXHAUST_FAN:
      return <Wind className="h-5 w-5" strokeWidth={2} />
    case DeviceType.OVEN:
      return <Flame className="h-5 w-5" strokeWidth={2} />
    case DeviceType.ALARM:
      return <AlarmClock className="h-5 w-5" strokeWidth={2} />
    case DeviceType.DOOR_LOCK:
      return <Lock className="h-5 w-5" strokeWidth={2} />
    case DeviceType.WINDOW_SENSOR:
      return <DoorOpen className="h-5 w-5" strokeWidth={2} />
    case DeviceType.CAMERA:
      return <Camera className="h-5 w-5" strokeWidth={2} />
    case DeviceType.FAN:
      return <Fan className="h-5 w-5" strokeWidth={2} />
    default:
      return <Lightbulb className="h-5 w-5" strokeWidth={2} />
  }
}

interface DeviceCardProps {
  device: Device
  onToggle: () => void
}

export function DeviceCard({ device, onToggle }: DeviceCardProps) {
  const disabled = !device.controllable
  const roomLabel =
    device.room in ROOM_LABELS
      ? ROOM_LABELS[device.room as keyof typeof ROOM_LABELS]
      : device.room
  const hasNumericValue = typeof device.state.value === 'number'

  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <ToggleSwitch
        isOn={device.state.on}
        onToggle={onToggle}
        label={device.name}
        icon={deviceIcon(device.type)}
        disabled={disabled}
      />
      <p className="mt-2 pl-7 text-xs text-[var(--color-text-secondary)]">
        {roomLabel}
      </p>
      {hasNumericValue ? (
        <p className="mt-2 pl-7 text-sm font-medium tabular-nums text-[var(--color-text-primary)]">
          Value: {device.state.value}
        </p>
      ) : null}
    </div>
  )
}
