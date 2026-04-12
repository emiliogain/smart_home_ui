import { ContextType } from '@/types/context'
import { DeviceType, type Device } from '@/types/device'

/** API + Socket.IO target. In local `npm run dev`, defaults to same origin so Vite can proxy to the backend (see vite.config.ts). */
export const BACKEND_URL: string =
  import.meta.env.VITE_BACKEND_URL ??
  (import.meta.env.DEV ? '' : 'http://localhost:8080')

/** Backend HTML admin (`GET /admin`). Uses Vite proxy in dev when `BACKEND_URL` is empty. */
export function adminPanelURL(): string {
  const base = BACKEND_URL.replace(/\/$/, '')
  return base ? `${base}/admin` : '/admin'
}

export const ROOMS = ['living_room', 'bedroom', 'kitchen'] as const

type RoomId = (typeof ROOMS)[number]

export const ROOM_LABELS: Record<RoomId, string> = {
  living_room: 'Living Room',
  bedroom: 'Bedroom',
  kitchen: 'Kitchen',
}

export const CONTEXT_LABELS: Record<
  ContextType,
  { label: string; emoji: string; description: string }
> = {
  [ContextType.NO_ONE_HOME]: {
    label: 'Away Mode',
    emoji: '🔴',
    description: 'No one is home',
  },
  [ContextType.READING_LIVING_ROOM]: {
    label: 'Reading Mode',
    emoji: '📖',
    description: 'Reading in the living room',
  },
  [ContextType.WATCHING_TV_LIVING_ROOM]: {
    label: 'TV Mode',
    emoji: '📺',
    description: 'Watching TV in the living room',
  },
  [ContextType.COOKING_KITCHEN]: {
    label: 'Cooking Mode',
    emoji: '🍳',
    description: 'Cooking in the kitchen',
  },
  [ContextType.SLEEPING]: {
    label: 'Sleep Mode',
    emoji: '🌙',
    description: 'Sleeping',
  },
  [ContextType.ALERT_TOO_HOT]: {
    label: 'Too Hot',
    emoji: '🔥',
    description: 'Temperature above comfort range',
  },
  [ContextType.ALERT_TOO_COLD]: {
    label: 'Too Cold',
    emoji: '❄️',
    description: 'Temperature below comfort range',
  },
  [ContextType.COMFORTABLE]: {
    label: 'Comfortable',
    emoji: '✅',
    description: 'Environment is comfortable',
  },
  [ContextType.UNKNOWN]: {
    label: 'Detecting...',
    emoji: '❓',
    description: 'Determining context',
  },
}

export const MOCK_DEVICES: Device[] = [
  {
    id: 'ceiling_light_living',
    name: 'Ceiling Light',
    type: DeviceType.LIGHT,
    room: 'living_room',
    state: { on: true, value: 80 },
    controllable: true,
  },
  {
    id: 'reading_lamp',
    name: 'Reading Lamp',
    type: DeviceType.LIGHT,
    room: 'living_room',
    state: { on: false, value: 0 },
    controllable: true,
  },
  {
    id: 'thermostat_living',
    name: 'Thermostat',
    type: DeviceType.THERMOSTAT,
    room: 'living_room',
    state: { on: true, value: 22 },
    controllable: true,
  },
  {
    id: 'kitchen_light',
    name: 'Kitchen Light',
    type: DeviceType.LIGHT,
    room: 'kitchen',
    state: { on: false, value: 0 },
    controllable: true,
  },
  {
    id: 'exhaust_fan',
    name: 'Exhaust Fan',
    type: DeviceType.EXHAUST_FAN,
    room: 'kitchen',
    state: { on: false },
    controllable: true,
  },
  {
    id: 'oven',
    name: 'Oven',
    type: DeviceType.OVEN,
    room: 'kitchen',
    state: { on: false, value: 0 },
    controllable: true,
  },
  {
    id: 'bedroom_light',
    name: 'Bedroom Light',
    type: DeviceType.LIGHT,
    room: 'bedroom',
    state: { on: false, value: 0 },
    controllable: true,
  },
  {
    id: 'thermostat_bedroom',
    name: 'Thermostat',
    type: DeviceType.THERMOSTAT,
    room: 'bedroom',
    state: { on: true, value: 20 },
    controllable: true,
  },
  {
    id: 'alarm_clock',
    name: 'Alarm',
    type: DeviceType.ALARM,
    room: 'bedroom',
    state: { on: true, value: 7 },
    controllable: true,
  },
  {
    id: 'front_door',
    name: 'Front Door',
    type: DeviceType.DOOR_LOCK,
    room: 'living_room',
    state: { on: true, locked: true },
    controllable: true,
  },
  {
    id: 'back_door',
    name: 'Back Door',
    type: DeviceType.DOOR_LOCK,
    room: 'kitchen',
    state: { on: true, locked: true },
    controllable: true,
  },
  {
    id: 'window_sensor',
    name: 'Window',
    type: DeviceType.WINDOW_SENSOR,
    room: 'living_room',
    state: { on: true, locked: true },
    controllable: true,
  },
  {
    id: 'camera_front',
    name: 'Front Camera',
    type: DeviceType.CAMERA,
    room: 'living_room',
    state: { on: true },
    controllable: true,
  },
  {
    id: 'fan_living',
    name: 'Fan',
    type: DeviceType.FAN,
    room: 'living_room',
    state: { on: false, value: 0 },
    controllable: true,
  },
]
