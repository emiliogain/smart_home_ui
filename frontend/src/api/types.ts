/** Sensor row from GET /api/sensors */
export interface BackendSensor {
  id: string
  name: string
  type: string
  location: string
  status: string
  created_at: string
  updated_at: string
}

/** Reading row from GET /api/sensors/:id/readings */
export interface BackendReading {
  id: string
  sensor_id: string
  value: number
  unit: string
  timestamp: string
}
