export type DeviceId = string

export enum DeviceType {
  LIGHT = 'LIGHT',
  THERMOSTAT = 'THERMOSTAT',
  EXHAUST_FAN = 'EXHAUST_FAN',
  OVEN = 'OVEN',
  ALARM = 'ALARM',
  DOOR_LOCK = 'DOOR_LOCK',
  WINDOW_SENSOR = 'WINDOW_SENSOR',
  CAMERA = 'CAMERA',
  FAN = 'FAN',
}

export interface DeviceState {
  on: boolean
  value?: number
  locked?: boolean
}

export interface Device {
  id: DeviceId
  name: string
  type: DeviceType
  room: string
  state: DeviceState
  controllable: boolean
  /** When set, LightDimmer shows warm/cool mode controls */
  lightMode?: 'warm_white' | 'cool_white'
}

export interface DeviceCommand {
  deviceId: string
  action: string
  value?: number
}
