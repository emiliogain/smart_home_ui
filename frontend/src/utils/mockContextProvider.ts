import { useContextStore } from '@/store/contextStore'
import { ContextType } from '@/types/context'
import type { SensorReading, SensorSnapshot } from '@/types/sensor'

const MOCK_SEQUENCE: ContextType[] = [
  ContextType.COMFORTABLE,
  ContextType.READING_LIVING_ROOM,
  ContextType.COOKING_KITCHEN,
  ContextType.SLEEPING,
  ContextType.NO_ONE_HOME,
  ContextType.ALERT_TOO_HOT,
]

interface MockSensorFields {
  temp: number
  humidity: number
  light: number
  motion: boolean
  motionRoom: string | null
}

function mockFieldsForContext(ctx: ContextType): MockSensorFields {
  switch (ctx) {
    case ContextType.READING_LIVING_ROOM:
      return {
        temp: 21,
        humidity: 42,
        light: 320,
        motion: true,
        motionRoom: 'living_room',
      }
    case ContextType.COOKING_KITCHEN:
      return {
        temp: 24,
        humidity: 58,
        light: 450,
        motion: true,
        motionRoom: 'kitchen',
      }
    case ContextType.SLEEPING:
      return {
        temp: 19,
        humidity: 40,
        light: 5,
        motion: false,
        motionRoom: 'bedroom',
      }
    case ContextType.NO_ONE_HOME:
      return {
        temp: 21,
        humidity: 45,
        light: 0,
        motion: false,
        motionRoom: null,
      }
    case ContextType.ALERT_TOO_HOT:
      return {
        temp: 29,
        humidity: 55,
        light: 400,
        motion: true,
        motionRoom: 'living_room',
      }
    case ContextType.COMFORTABLE:
    default:
      return {
        temp: 22,
        humidity: 45,
        light: 300,
        motion: true,
        motionRoom: 'living_room',
      }
  }
}

function buildReadings(fields: MockSensorFields, at: string): SensorReading[] {
  const list: SensorReading[] = [
    {
      sensorId: 'temperature.primary',
      sensorLabel: 'Temperature · Living room',
      sensorType: 'temperature',
      location: 'living_room',
      value: fields.temp,
      unit: '°C',
      at,
    },
    {
      sensorId: 'humidity.primary',
      sensorLabel: 'Humidity · Living room',
      sensorType: 'humidity',
      location: 'living_room',
      value: fields.humidity,
      unit: '%',
      at,
    },
    {
      sensorId: 'light.level',
      sensorLabel: 'Light · Living room',
      sensorType: 'light',
      location: 'living_room',
      value: fields.light,
      unit: 'lux',
      at,
    },
    {
      sensorId: 'motion.pir',
      sensorLabel: 'Motion · Living room',
      sensorType: 'motion',
      location: 'living_room',
      value: fields.motion ? 1 : 0,
      at,
    },
  ]
  if (fields.motionRoom) {
    list.push({
      sensorId: 'motion.room',
      sensorLabel: `Motion · ${fields.motionRoom.replace(/_/g, ' ')}`,
      sensorType: 'motion',
      location: fields.motionRoom,
      value: 0,
      unit: fields.motionRoom,
      at,
    })
  }
  return list
}

function randomConfidence(): number {
  return Math.round((0.65 + Math.random() * 0.3) * 100) / 100
}

function snapshotForContext(ctx: ContextType): SensorSnapshot {
  const at = new Date().toISOString()
  return {
    readings: buildReadings(mockFieldsForContext(ctx), at),
  }
}

function buildContextUpdate(ctx: ContextType) {
  return {
    currentContext: ctx,
    confidence: randomConfidence(),
    lastUpdated: new Date().toISOString(),
    sensorSnapshot: snapshotForContext(ctx),
  }
}

/**
 * Cycles mock contexts every 15s for offline / demo use.
 * Applies an initial update immediately, then on the interval.
 */
export function startMockContextCycle(): () => void {
  let index = 0

  const tick = () => {
    const ctx = MOCK_SEQUENCE[index % MOCK_SEQUENCE.length]
    index += 1
    useContextStore.getState().setContext(buildContextUpdate(ctx))
  }

  tick()
  const id = window.setInterval(tick, 15_000)

  return () => window.clearInterval(id)
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}

function varyReadings(readings: SensorReading[]): SensorReading[] {
  const at = new Date().toISOString()
  return readings.map((r) => {
    const id = (r.sensorType ?? r.sensorId).toLowerCase()
    if (id.includes('temp')) {
      return {
        ...r,
        value: clamp(r.value + (Math.random() - 0.5), 15, 35),
        at,
      }
    }
    if (id.includes('humid')) {
      return {
        ...r,
        value: clamp(r.value + (Math.random() - 0.5) * 4, 20, 75),
        at,
      }
    }
    if (id.includes('light')) {
      return {
        ...r,
        value: clamp(r.value + (Math.random() - 0.5) * 40, 0, 2000),
        at,
      }
    }
    if (id.includes('motion') && r.sensorId === 'motion.pir') {
      return {
        ...r,
        value: Math.random() > 0.85 ? (r.value >= 1 ? 0 : 1) : r.value,
        at,
      }
    }
    return { ...r, at }
  })
}

/**
 * Nudges numeric readings every 5s while keeping the current context/confidence.
 */
export function startMockSensorUpdates(): () => void {
  const id = window.setInterval(() => {
    const s = useContextStore.getState()
    if (!s.sensorSnapshot?.readings.length) return

    useContextStore.getState().setContext({
      currentContext: s.currentContext,
      confidence: s.confidence,
      lastUpdated: new Date().toISOString(),
      sensorSnapshot: {
        readings: varyReadings(s.sensorSnapshot.readings),
      },
    })
  }, 5_000)

  return () => window.clearInterval(id)
}
