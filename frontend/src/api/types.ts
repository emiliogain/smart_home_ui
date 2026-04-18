/** Sensor row from GET /api/sensors */
export interface BackendSensor {
  id: string
  name: string
  display_label?: string
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
  sensor_name?: string
  sensor_label?: string
  sensor_type?: string
  location?: string
  value: number
  unit: string
  timestamp: string
}
