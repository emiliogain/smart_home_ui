import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { ContextType } from '@/types/context'

export interface SessionLogEntry {
  action: string
  timestamp: string
  context: ContextType
  details?: string
}

interface SettingsStoreState {
  adaptiveMode: boolean
  studyMode: boolean
  sessionLogging: boolean
  sessionLogs: SessionLogEntry[]
  simulatedContext: ContextType | null
  toggleAdaptiveMode: () => void
  setAdaptiveMode: (value: boolean) => void
  toggleStudyMode: () => void
  toggleSessionLogging: () => void
  setSimulatedContext: (context: ContextType | null) => void
  addLog: (action: string, context: ContextType, details?: string) => void
  clearLogs: () => void
  exportLogs: () => string
}

function csvEscape(value: string): string {
  if (/[",\n\r]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}

export const useSettingsStore = create<SettingsStoreState>()(
  persist(
    (set, get) => ({
      adaptiveMode: true,
      studyMode: false,
      sessionLogging: false,
      sessionLogs: [],
      simulatedContext: null,
      toggleAdaptiveMode: () =>
        set((s) => ({ adaptiveMode: !s.adaptiveMode })),
      setAdaptiveMode: (adaptiveMode) => set({ adaptiveMode }),
      toggleStudyMode: () => set((s) => ({ studyMode: !s.studyMode })),
      toggleSessionLogging: () =>
        set((s) => ({ sessionLogging: !s.sessionLogging })),
      setSimulatedContext: (context) => set({ simulatedContext: context }),
      addLog: (action, context, details) =>
        set((s) => ({
          sessionLogs: [
            ...s.sessionLogs,
            {
              action,
              timestamp: new Date().toISOString(),
              context,
              details,
            },
          ],
        })),
      clearLogs: () => set({ sessionLogs: [] }),
      exportLogs: () => {
        const logs = get().sessionLogs
        const header = 'timestamp,action,context,details'
        const lines = logs.map((log) =>
          [
            csvEscape(log.timestamp),
            csvEscape(log.action),
            csvEscape(log.context),
            csvEscape(log.details ?? ''),
          ].join(','),
        )
        return [header, ...lines].join('\n')
      },
    }),
    { name: 'smart-home-settings' },
  ),
)
