import { useMemo } from 'react'
import { useContextStore } from '@/store/contextStore'
import { useSettingsStore } from '@/store/settingsStore'
import type { ContextType } from '@/types/context'
import {
  getAdaptation,
  type AdaptationConfig,
} from '@/utils/adaptationRules'

export function useAdaptation(): {
  config: AdaptationConfig
  effectiveContext: ContextType
} {
  const currentContext = useContextStore((s) => s.currentContext)
  const simulatedContext = useSettingsStore((s) => s.simulatedContext)
  const studyMode = useSettingsStore((s) => s.studyMode)

  const effectiveContext = useMemo((): ContextType => {
    if (studyMode && simulatedContext != null) {
      return simulatedContext
    }
    return currentContext
  }, [studyMode, simulatedContext, currentContext])

  const config = useMemo(
    () => getAdaptation(effectiveContext),
    [effectiveContext],
  )

  return { config, effectiveContext }
}
