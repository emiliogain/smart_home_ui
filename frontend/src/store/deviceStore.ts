import { create } from 'zustand'
import { MOCK_DEVICES } from '@/utils/constants'
import type { Device, DeviceState } from '@/types/device'

interface DeviceStoreState {
  devices: Device[]
  getDevice: (id: string) => Device | undefined
  updateDeviceState: (deviceId: string, newState: Partial<DeviceState>) => void
  toggleDevice: (deviceId: string) => void
  setDevices: (devices: Device[]) => void
}

export const useDeviceStore = create<DeviceStoreState>((set, get) => ({
  devices: MOCK_DEVICES,
  getDevice: (id) => get().devices.find((d) => d.id === id),
  updateDeviceState: (deviceId, newState) =>
    set((s) => ({
      devices: s.devices.map((d) =>
        d.id === deviceId
          ? { ...d, state: { ...d.state, ...newState } }
          : d,
      ),
    })),
  toggleDevice: (deviceId) =>
    set((s) => ({
      devices: s.devices.map((d) =>
        d.id === deviceId
          ? { ...d, state: { ...d.state, on: !d.state.on } }
          : d,
      ),
    })),
  setDevices: (devices) => set({ devices }),
}))
