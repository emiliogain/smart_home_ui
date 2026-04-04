import axios from 'axios'
import { BACKEND_URL } from '@/utils/constants'
import type { Device, DeviceCommand } from '@/types/device'

export const apiClient = axios.create({
  baseURL: `${BACKEND_URL}/api`,
})

export async function fetchDevices(): Promise<Device[] | null> {
  try {
    const { data } = await apiClient.get<Device[]>('/devices')
    return data
  } catch (err) {
    console.error(err)
    return null
  }
}

export async function controlDevice(
  command: DeviceCommand,
): Promise<void | null> {
  try {
    await apiClient.post(`/devices/${command.deviceId}/control`, {
      action: command.action,
      value: command.value,
    })
  } catch (err) {
    console.error(err)
    return null
  }
}

export async function fetchScenes(): Promise<unknown[]> {
  try {
    const { data } = await apiClient.get<unknown[]>('/scenes')
    return data
  } catch (err) {
    console.error(err)
    return []
  }
}

export async function applyScene(sceneId: string): Promise<void | null> {
  try {
    await apiClient.post(`/scenes/${sceneId}/apply`)
  } catch (err) {
    console.error(err)
    return null
  }
}
