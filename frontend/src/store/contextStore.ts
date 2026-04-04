import { create } from 'zustand'
import {
  ContextType,
  type ContextUpdate,
} from '@/types/context'
import type { SensorSnapshot } from '@/types/sensor'
import { useSettingsStore } from '@/store/settingsStore'

interface ContextStoreState {
  currentContext: ContextType
  confidence: number
  previousContext: ContextType | null
  lastUpdated: string
  sensorSnapshot: SensorSnapshot | null
  setContext: (update: ContextUpdate) => void
  resetContext: () => void
}

const initialState = {
  currentContext: ContextType.UNKNOWN,
  confidence: 0,
  previousContext: null as ContextType | null,
  lastUpdated: '',
  sensorSnapshot: null as SensorSnapshot | null,
}

export const useContextStore = create<ContextStoreState>((set) => ({
  ...initialState,
  setContext: (update) => {
    set((s) => {
      if (
        useSettingsStore.getState().sessionLogging &&
        update.currentContext !== s.currentContext
      ) {
        useSettingsStore.getState().addLog(
          'context_changed',
          update.currentContext,
          JSON.stringify({
            from: s.currentContext,
            to: update.currentContext,
            confidence: update.confidence,
          }),
        )
      }

      return {
        previousContext: s.currentContext,
        currentContext: update.currentContext,
        confidence: update.confidence,
        lastUpdated: update.lastUpdated,
        sensorSnapshot: update.sensorSnapshot,
      }
    })
  },
  resetContext: () => set({ ...initialState }),
}))
