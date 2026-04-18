export type SensorId = string

export interface SensorReading {
  sensorId: SensorId
  /** Human-readable title from the backend (context snapshot or readings API). */
  sensorLabel?: string
  sensorType?: string
  location?: string
  value: number
  unit?: string
  at: string
}

export interface SensorSnapshot {
  readings: SensorReading[]
}

export interface SensorHistory {
  sensorId: string
  readings: SensorReading[]
}
