import { motion } from 'framer-motion'
import { Sparkles } from 'lucide-react'
import { clsx } from 'clsx'
import { interactiveSurface } from '@/utils/uiClasses'
import { useContextStore } from '@/store/contextStore'
import { useSettingsStore } from '@/store/settingsStore'

export interface SuggestedScenePayload {
  name: string
  description: string
  actions: string
}

interface SceneSuggestionProps {
  scene: SuggestedScenePayload
  onApply: () => void
  onDismiss: () => void
}

export function SceneSuggestion({
  scene,
  onApply,
  onDismiss,
}: SceneSuggestionProps) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.35, ease: 'easeOut' }}
      className="rounded-xl bg-gradient-to-br from-[var(--color-secondary)] via-[var(--color-secondary)]/70 to-[var(--color-secondary)]/30 p-px shadow-lg shadow-black/25"
    >
      <div className="rounded-[11px] bg-[var(--color-surface)] p-4 transition-all hover:ring-1 hover:ring-white/10">
        <div className="mb-3 flex items-center gap-2 text-[var(--color-secondary)]">
          <Sparkles className="h-5 w-5 shrink-0" strokeWidth={2} />
          <span className="text-xs font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
            Suggested Scene
          </span>
        </div>
        <h3 className="text-lg font-bold text-[var(--color-text-primary)]">
          {scene.name}
        </h3>
        <p className="mt-2 text-sm leading-relaxed text-[var(--color-text-secondary)]">
          {scene.description}
        </p>
        <p className="mt-1 text-sm text-[var(--color-text-secondary)]/80">
          {scene.actions}
        </p>
        <div className="mt-4 flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => {
              if (useSettingsStore.getState().sessionLogging) {
                useSettingsStore.getState().addLog(
                  'scene_applied',
                  useContextStore.getState().currentContext,
                  scene.name,
                )
              }
              onApply()
            }}
            className={clsx(
              interactiveSurface,
              'rounded-lg bg-[var(--color-secondary)] px-4 py-2 text-sm font-semibold text-white hover:opacity-90',
            )}
          >
            Apply Scene
          </button>
          <button
            type="button"
            onClick={() => {
              if (useSettingsStore.getState().sessionLogging) {
                useSettingsStore.getState().addLog(
                  'scene_dismissed',
                  useContextStore.getState().currentContext,
                  scene.name,
                )
              }
              onDismiss()
            }}
            className={clsx(
              interactiveSurface,
              'rounded-lg px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]',
            )}
          >
            Dismiss
          </button>
        </div>
      </div>
    </motion.div>
  )
}
