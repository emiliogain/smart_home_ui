import type { ContextUpdate } from '@/types/context'
import type { SensorHistory, SensorReading } from '@/types/sensor'
import { apiClient } from './devices'

export async function fetchCurrentSensors(): Promise<SensorReading[]> {
  try {
    const { data } = await apiClient.get<SensorReading[]>('/sensors')
    return data
  } catch (err) {
    console.error(err)
    return []
  }
}

export async function fetchSensorHistory(
  sensorId: string,
  hours?: number,
): Promise<SensorHistory | null> {
  try {
    const { data } = await apiClient.get<SensorHistory>(
      `/sensors/${sensorId}/history`,
      { params: hours != null ? { hours } : undefined },
    )
    return data
  } catch (err) {
    console.error(err)
    return null
  }
}

export async function fetchCurrentContext(): Promise<ContextUpdate | null> {
  try {
    const { data } = await apiClient.get<ContextUpdate>('/context/current')
    return data
  } catch (err) {
    console.error(err)
    return null
  }
}
