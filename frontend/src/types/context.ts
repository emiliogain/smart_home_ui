import type { SensorSnapshot } from './sensor'

export type ContextId = string

export enum ContextType {
  NO_ONE_HOME = 'NO_ONE_HOME',
  READING_LIVING_ROOM = 'READING_LIVING_ROOM',
  WATCHING_TV_LIVING_ROOM = 'WATCHING_TV_LIVING_ROOM',
  COOKING_KITCHEN = 'COOKING_KITCHEN',
  SLEEPING = 'SLEEPING',
  ALERT_TOO_HOT = 'ALERT_TOO_HOT',
  ALERT_TOO_COLD = 'ALERT_TOO_COLD',
  COMFORTABLE = 'COMFORTABLE',
  UNKNOWN = 'UNKNOWN',
}

export interface ContextSnapshot {
  id: ContextId
  name: string
}

export interface ContextUpdate {
  currentContext: ContextType
  confidence: number
  lastUpdated: string
  sensorSnapshot: SensorSnapshot | null
}
