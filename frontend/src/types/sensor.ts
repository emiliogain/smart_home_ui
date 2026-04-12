export type SensorId = string

export interface SensorReading {
  sensorId: SensorId
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
