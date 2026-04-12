import { AnimatePresence, motion } from 'framer-motion'
import { ConfidenceBadge } from '@/components/common/ConfidenceBadge'
import { useContextStore } from '@/store/contextStore'
import { useSettingsStore } from '@/store/settingsStore'
import { CONTEXT_LABELS } from '@/utils/constants'

export function ContextBanner() {
  const currentContext = useContextStore((s) => s.currentContext)
  const confidence = useContextStore((s) => s.confidence)
  const studyMode = useSettingsStore((s) => s.studyMode)
  const simulatedContext = useSettingsStore((s) => s.simulatedContext)

  const displayContext =
    studyMode && simulatedContext != null ? simulatedContext : currentContext

  const { emoji, label, description } = CONTEXT_LABELS[displayContext]
  const showSimBadge = studyMode && simulatedContext != null

  return (
    <div className="inline-flex max-w-full items-center gap-2 rounded-full bg-white/10 px-3 py-1.5 sm:px-3.5">
      {showSimBadge ? (
        <span className="shrink-0 rounded bg-orange-500/35 px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-orange-300">
          SIM
        </span>
      ) : null}
      <AnimatePresence mode="wait">
        <motion.span
          key={displayContext}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.22, ease: 'easeInOut' }}
          className="inline-flex min-w-0 flex-col gap-0.5 sm:flex-row sm:items-center sm:gap-1.5"
        >
          <span className="inline-flex min-w-0 items-center gap-1.5">
            <span className="shrink-0 text-base leading-none sm:text-lg" aria-hidden>
              {emoji}
            </span>
            <span className="truncate text-sm font-medium text-[var(--color-text-primary)]">
              {label}
            </span>
          </span>
          <span className="hidden max-w-[14rem] truncate text-sm text-[var(--color-text-secondary)] sm:inline sm:max-w-[20rem]">
            {description}
          </span>
        </motion.span>
      </AnimatePresence>
      <ConfidenceBadge confidence={confidence} />
    </div>
  )
}
