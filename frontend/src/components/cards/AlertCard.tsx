import { motion } from 'framer-motion'
import { AlertTriangle } from 'lucide-react'
import { clsx } from 'clsx'
import type { SuggestedScenePayload } from '@/components/cards/SceneSuggestion'
import { useContextStore } from '@/store/contextStore'
import { useSettingsStore } from '@/store/settingsStore'
import { cardSurface, interactiveSurface } from '@/utils/uiClasses'

interface AlertCardProps {
  message: string
  suggestedScene: SuggestedScenePayload | null
  onApply: () => void
  onDismiss: () => void
  /** Left border uses warning (amber) or danger (red) */
  severity?: 'warning' | 'danger'
  /** Pulse border when inside `.theme-alert` (active alert context) */
  pulse?: boolean
}

export function AlertCard({
  message,
  suggestedScene,
  onApply,
  onDismiss,
  severity = 'warning',
  pulse = false,
}: AlertCardProps) {
  const borderColor =
    severity === 'danger'
      ? 'var(--color-danger)'
      : 'var(--color-warning)'

  return (
    <div
      className={clsx(
        cardSurface,
        'w-full border border-white/10',
        pulse && 'theme-alert-pulse',
      )}
      style={{ borderLeftWidth: 4, borderLeftColor: borderColor }}
    >
      <div className="flex gap-3">
        <div className="relative shrink-0 pt-0.5">
          <motion.span
            className="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-[var(--color-danger)]"
            animate={{ opacity: [1, 0.35, 1], scale: [1, 1.15, 1] }}
            transition={{ duration: 1.6, repeat: Infinity, ease: 'easeInOut' }}
            aria-hidden
          />
          <AlertTriangle
            className="h-6 w-6 text-[var(--color-warning)]"
            strokeWidth={2}
            aria-hidden
          />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="text-lg font-bold text-[var(--color-text-primary)]">
            Alert
          </h3>
          <p className="mt-1 text-sm leading-relaxed text-[var(--color-text-secondary)]">
            {message}
          </p>

          {!suggestedScene ? (
            <button
              type="button"
              onClick={() => {
                if (useSettingsStore.getState().sessionLogging) {
                  useSettingsStore.getState().addLog(
                    'alert_dismissed',
                    useContextStore.getState().currentContext,
                  )
                }
                onDismiss()
              }}
              className={clsx(
                interactiveSurface,
                'mt-3 rounded-lg px-1 py-1 text-left text-sm font-medium text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]',
              )}
            >
              Dismiss
            </button>
          ) : null}

          {suggestedScene ? (
            <div className="mt-4 rounded-lg border border-white/10 bg-white/5 p-3 transition-all hover:ring-1 hover:ring-white/10">
              <p className="text-sm font-semibold text-[var(--color-text-primary)]">
                {suggestedScene.name}
              </p>
              <p className="mt-1 text-xs text-[var(--color-text-secondary)]">
                {suggestedScene.description}
              </p>
              <p className="mt-1 text-xs text-[var(--color-text-secondary)]/80">
                {suggestedScene.actions}
              </p>
              <div className="mt-3 flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => {
                    if (useSettingsStore.getState().sessionLogging) {
                      useSettingsStore.getState().addLog(
                        'alert_action_applied',
                        useContextStore.getState().currentContext,
                        suggestedScene.name,
                      )
                    }
                    onApply()
                  }}
                  className={clsx(
                    interactiveSurface,
                    'rounded-lg bg-[var(--color-secondary)] px-3 py-1.5 text-xs font-semibold text-white hover:opacity-90',
                  )}
                >
                  Apply
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (useSettingsStore.getState().sessionLogging) {
                      useSettingsStore.getState().addLog(
                        'alert_dismissed',
                        useContextStore.getState().currentContext,
                        suggestedScene.name,
                      )
                    }
                    onDismiss()
                  }}
                  className={clsx(
                    interactiveSurface,
                    'rounded-lg px-3 py-1.5 text-xs font-medium text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]',
                  )}
                >
                  Dismiss
                </button>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}
