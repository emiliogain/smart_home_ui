import { useCallback, useMemo } from 'react'
import { sendDeviceCommand } from '@/api/websocket'
import { useContextStore } from '@/store/contextStore'
import { useDeviceStore } from '@/store/deviceStore'
import { useSettingsStore } from '@/store/settingsStore'

export function useDevices() {
  const devices = useDeviceStore((s) => s.devices)

  const getDevicesByRoom = useCallback(
    (room: string) => devices.filter((d) => d.room === room),
    [devices],
  )

  const getDeviceById = useCallback(
    (id: string) => devices.find((d) => d.id === id),
    [devices],
  )

  const handleToggle = useCallback((deviceId: string) => {
    const device = useDeviceStore.getState().getDevice(deviceId)
    if (!device) return

    const prevOn = device.state.on
    useDeviceStore.getState().toggleDevice(deviceId)
    sendDeviceCommand({ deviceId, action: 'toggle' })

    if (useSettingsStore.getState().sessionLogging) {
      const ctx = useContextStore.getState().currentContext
      useSettingsStore.getState().addLog(
        'device_toggle',
        ctx,
        JSON.stringify({
          deviceId: device.id,
          name: device.name,
          type: device.type,
          room: device.room,
          previousOn: prevOn,
          nextOn: !prevOn,
        }),
      )
    }
  }, [])

  const handleSetValue = useCallback((deviceId: string, value: number) => {
    const device = useDeviceStore.getState().getDevice(deviceId)
    if (!device) return

    const previousValue = device.state.value
    useDeviceStore.getState().updateDeviceState(deviceId, { value })
    sendDeviceCommand({ deviceId, action: 'set_value', value })

    if (useSettingsStore.getState().sessionLogging) {
      const ctx = useContextStore.getState().currentContext
      useSettingsStore.getState().addLog(
        'device_set_value',
        ctx,
        JSON.stringify({
          deviceId: device.id,
          name: device.name,
          type: device.type,
          room: device.room,
          previousValue,
          nextValue: value,
        }),
      )
    }
  }, [])

  const handleToggleLocked = useCallback((deviceId: string) => {
    const device = useDeviceStore.getState().getDevice(deviceId)
    if (!device) return
    const prevLocked = device.state.locked
    if (prevLocked === undefined) return
    const nextLocked = !prevLocked
    useDeviceStore.getState().updateDeviceState(deviceId, { locked: nextLocked })
    sendDeviceCommand({
      deviceId,
      action: 'toggle_lock',
      value: nextLocked ? 1 : 0,
    })
  }, [])

  const heroDevices = useCallback(
    (heroIds: string[]) => {
      const idSet = new Set(heroIds)
      return devices.filter((d) => idSet.has(d.id))
    },
    [devices],
  )

  const visibleDevices = useCallback(
    (hiddenIds: string[]) => {
      const hidden = new Set(hiddenIds)
      return devices.filter((d) => !hidden.has(d.id))
    },
    [devices],
  )

  return useMemo(
    () => ({
      devices,
      getDevicesByRoom,
      getDeviceById,
      handleToggle,
      handleSetValue,
      handleToggleLocked,
      heroDevices,
      visibleDevices,
    }),
    [
      devices,
      getDevicesByRoom,
      getDeviceById,
      handleToggle,
      handleSetValue,
      handleToggleLocked,
      heroDevices,
      visibleDevices,
    ],
  )
}
