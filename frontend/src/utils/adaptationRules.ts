import { ContextType } from '@/types/context'

export interface AdaptationConfig {
  heroControls: string[]
  hiddenControls: string[]
  suggestedScene: {
    name: string
    description: string
    actions: string
  } | null
  alertMessage: string | null
  theme: 'default' | 'dark' | 'alert'
  showTimer: boolean
  showSecurity: boolean
  showCameras: boolean
}

const neutralConfig: AdaptationConfig = {
  heroControls: [],
  hiddenControls: [],
  suggestedScene: null,
  alertMessage: null,
  theme: 'default',
  showTimer: false,
  showSecurity: false,
  showCameras: false,
}

export function getAdaptation(context: ContextType): AdaptationConfig {
  switch (context) {
    case ContextType.NO_ONE_HOME:
      return {
        heroControls: [
          'front_door',
          'back_door',
          'window_sensor',
          'camera_front',
        ],
        hiddenControls: [
          'reading_lamp',
          'oven',
          'exhaust_fan',
          'alarm_clock',
        ],
        suggestedScene: null,
        alertMessage: null,
        theme: 'default',
        showTimer: false,
        showSecurity: true,
        showCameras: true,
      }
    case ContextType.READING_LIVING_ROOM:
      return {
        heroControls: ['reading_lamp', 'thermostat_living'],
        hiddenControls: [
          'oven',
          'exhaust_fan',
          'kitchen_light',
          'camera_front',
        ],
        suggestedScene: {
          name: 'Cozy Reading',
          description:
            'Lamp at 72% warm white, ceiling off, 21°C',
          actions: 'Optimized for comfortable reading',
        },
        alertMessage: null,
        theme: 'default',
        showTimer: false,
        showSecurity: false,
        showCameras: false,
      }
    case ContextType.WATCHING_TV_LIVING_ROOM:
      return {
        heroControls: ['ceiling_light_living', 'thermostat_living'],
        hiddenControls: [
          'oven',
          'exhaust_fan',
          'kitchen_light',
          'reading_lamp',
        ],
        suggestedScene: {
          name: 'Movie Night',
          description: 'Lights dimmed to 20%, temperature 22°C',
          actions: 'Optimized for watching',
        },
        alertMessage: null,
        theme: 'dark',
        showTimer: false,
        showSecurity: false,
        showCameras: false,
      }
    case ContextType.COOKING_KITCHEN:
      return {
        heroControls: ['oven', 'exhaust_fan', 'kitchen_light'],
        hiddenControls: [
          'reading_lamp',
          'bedroom_light',
          'alarm_clock',
          'camera_front',
        ],
        suggestedScene: null,
        alertMessage: null,
        theme: 'default',
        showTimer: true,
        showSecurity: false,
        showCameras: false,
      }
    case ContextType.SLEEPING:
      return {
        heroControls: ['alarm_clock', 'thermostat_bedroom'],
        hiddenControls: [
          'kitchen_light',
          'oven',
          'exhaust_fan',
          'ceiling_light_living',
          'camera_front',
        ],
        suggestedScene: {
          name: 'Good Night',
          description: 'All lights off, thermostat 19°C, DND on',
          actions: 'Optimized for sleep',
        },
        alertMessage: null,
        theme: 'dark',
        showTimer: false,
        showSecurity: false,
        showCameras: false,
      }
    case ContextType.ALERT_TOO_HOT:
      return {
        heroControls: ['thermostat_living', 'fan_living', 'window_sensor'],
        hiddenControls: [],
        suggestedScene: {
          name: 'Cool Down',
          description: 'Lower temperature to 23°C, turn on fan',
          actions: 'Reduce temperature',
        },
        alertMessage:
          'Temperature is above the comfortable range. Recommended action: lower thermostat to 23°C and turn on the fan.',
        theme: 'alert',
        showTimer: false,
        showSecurity: false,
        showCameras: false,
      }
    case ContextType.ALERT_TOO_COLD:
      return {
        heroControls: ['thermostat_living', 'thermostat_bedroom'],
        hiddenControls: [],
        suggestedScene: {
          name: 'Warm Up',
          description: 'Raise temperature to 23°C',
          actions: 'Increase temperature',
        },
        alertMessage:
          'Temperature is below the comfortable range. Recommended action: raise thermostat to 23°C.',
        theme: 'alert',
        showTimer: false,
        showSecurity: false,
        showCameras: false,
      }
    case ContextType.COMFORTABLE:
    case ContextType.UNKNOWN:
    default:
      return { ...neutralConfig }
  }
}
