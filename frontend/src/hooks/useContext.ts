import { useContextStore } from '@/store/contextStore'

export function useSmartHomeContext() {
  return useContextStore((s) => ({
    currentContext: s.currentContext,
    confidence: s.confidence,
    previousContext: s.previousContext,
    lastUpdated: s.lastUpdated,
    sensorSnapshot: s.sensorSnapshot,
    isConfident: s.confidence >= 0.6,
  }))
}
