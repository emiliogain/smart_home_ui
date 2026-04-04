import { DeviceType } from '@/types/device'
import { useContextStore } from '@/store/contextStore'
import { EnvironmentSummaryBar } from '@/components/adaptive/EnvironmentSummaryBar'
import { DeviceGroupCard } from '@/components/cards/DeviceGroupCard'
import { useDevices } from '@/hooks/useDevices'

const SECURITY_TYPES = new Set([
  DeviceType.DOOR_LOCK,
  DeviceType.WINDOW_SENSOR,
  DeviceType.CAMERA,
])

export function StaticDashboard() {
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)
  const {
    devices,
    getDevicesByRoom,
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  } = useDevices()

  const living = getDevicesByRoom('living_room').filter(
    (d) => !SECURITY_TYPES.has(d.type),
  )
  const kitchen = getDevicesByRoom('kitchen').filter(
    (d) => !SECURITY_TYPES.has(d.type),
  )
  const bedroom = getDevicesByRoom('bedroom')
  const security = devices.filter((d) => SECURITY_TYPES.has(d.type))

  return (
    <div className="space-y-8 pt-2 lg:space-y-10">
      <EnvironmentSummaryBar
        snapshot={sensorSnapshot}
        devices={devices}
        roomId="living_room"
      />

      <DeviceGroupCard
        title="Living Room"
        devices={living}
        onToggle={handleToggle}
        onSetValue={handleSetValue}
        onToggleLocked={handleToggleLocked}
      />
      <DeviceGroupCard
        title="Kitchen"
        devices={kitchen}
        onToggle={handleToggle}
        onSetValue={handleSetValue}
        onToggleLocked={handleToggleLocked}
      />
      <DeviceGroupCard
        title="Bedroom"
        devices={bedroom}
        onToggle={handleToggle}
        onSetValue={handleSetValue}
        onToggleLocked={handleToggleLocked}
      />
      <DeviceGroupCard
        title="Security"
        devices={security}
        onToggle={handleToggle}
        onSetValue={handleSetValue}
        onToggleLocked={handleToggleLocked}
      />
    </div>
  )
}
