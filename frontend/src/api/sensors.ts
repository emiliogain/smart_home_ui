import type { SensorHistory, SensorReading } from '@/types/sensor'
import { apiClient } from './client'
import type { BackendReading, BackendSensor } from './types'

export async function listSensors(): Promise<BackendSensor[]> {
  try {
    const { data } = await apiClient.get<{ sensors: BackendSensor[] }>(
      '/sensors',
    )
    return data.sensors ?? []
  } catch (err) {
    console.error(err)
    return []
  }
}

export async function getSensor(sensorId: string): Promise<BackendSensor | null> {
  try {
    const { data } = await apiClient.get<BackendSensor>(`/sensors/${sensorId}`)
    return data
  } catch (err) {
    console.error(err)
    return null
  }
}

export async function createSensor(payload: {
  name: string
  type: string
  location: string
}): Promise<BackendSensor | null> {
  try {
    const { data } = await apiClient.post<BackendSensor>('/sensors', payload)
    return data
  } catch (err) {
    console.error(err)
    return null
  }
}

function readingToSensorReading(r: BackendReading): SensorReading {
  const at =
    typeof r.timestamp === 'string'
      ? r.timestamp
      : new Date(r.timestamp).toISOString()
  return {
    sensorId: r.sensor_id,
    value: r.value,
    unit: r.unit || undefined,
    at,
  }
}

/** Latest readings for a sensor (newest first from API); returned chronological for charts. */
export async function fetchSensorReadings(
  sensorId: string,
  limit = 100,
): Promise<SensorReading[]> {
  try {
    const { data } = await apiClient.get<{ readings: BackendReading[] }>(
      `/sensors/${sensorId}/readings`,
      { params: { limit } },
    )
    const rows = data.readings ?? []
    return [...rows].reverse().map(readingToSensorReading)
  } catch (err) {
    console.error(err)
    return []
  }
}

export async function submitSensorReading(
  sensorId: string,
  value: number,
  unit: string,
): Promise<boolean> {
  try {
    await apiClient.post(`/sensors/${sensorId}/readings`, { value, unit })
    return true
  } catch (err) {
    console.error(err)
    return false
  }
}

/** Maps API readings to chart points (ascending time). */
export function readingsToHistoryPoints(
  readings: SensorReading[],
): { value: number; timestamp: string }[] {
  return readings.map((r) => ({
    value: r.value,
    timestamp: r.at,
  }))
}

export async function fetchSensorHistory(
  sensorId: string,
  limit?: number,
): Promise<SensorHistory | null> {
  const readings = await fetchSensorReadings(
    sensorId,
    limit != null ? limit : 100,
  )
  if (!readings.length) return null
  return { sensorId, readings }
}

/** Flat list of latest values per sensor (from list + optional last reading not available without N+1 — use context snapshot or readings per id). */
export async function fetchCurrentSensorsFromList(): Promise<SensorReading[]> {
  const sensors = await listSensors()
  const out: SensorReading[] = []
  for (const s of sensors) {
    const latest = await fetchSensorReadings(s.id, 1)
    const last = latest[latest.length - 1]
    if (last) out.push(last)
  }
  return out
}
