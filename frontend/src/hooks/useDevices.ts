import { useCallback, useMemo } from 'react'
import { controlDevice } from '@/api/devices'
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

  const handleToggle = useCallback(async (deviceId: string) => {
    const device = useDeviceStore.getState().getDevice(deviceId)
    if (!device) return

    const prevOn = device.state.on
    useDeviceStore.getState().toggleDevice(deviceId)

    const failed =
      (await controlDevice({ deviceId, action: 'toggle' })) === null

    if (failed) {
      useDeviceStore.getState().updateDeviceState(deviceId, { on: prevOn })
      return
    }

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

  const handleSetValue = useCallback(async (deviceId: string, value: number) => {
    const device = useDeviceStore.getState().getDevice(deviceId)
    if (!device) return

    const prevState = { ...device.state }
    useDeviceStore.getState().updateDeviceState(deviceId, { value })

    const command = { deviceId, action: 'set_value', value }
    const failed = (await controlDevice(command)) === null

    if (failed) {
      useDeviceStore.getState().updateDeviceState(deviceId, prevState)
      return
    }

    sendDeviceCommand(command)

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
          previousValue: prevState.value,
          nextValue: value,
        }),
      )
    }
  }, [])

  const handleToggleLocked = useCallback(async (deviceId: string) => {
    const device = useDeviceStore.getState().getDevice(deviceId)
    if (!device) return
    const prevLocked = device.state.locked
    if (prevLocked === undefined) return
    const nextLocked = !prevLocked
    useDeviceStore.getState().updateDeviceState(deviceId, { locked: nextLocked })

    const failed =
      (await controlDevice({
        deviceId,
        action: 'toggle_lock',
        value: nextLocked ? 1 : 0,
      })) === null

    if (failed) {
      useDeviceStore.getState().updateDeviceState(deviceId, { locked: prevLocked })
      return
    }

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
