import { AnimatePresence, motion } from 'framer-motion'
import { X } from 'lucide-react'
import { clsx } from 'clsx'
import { useContextStore } from '@/store/contextStore'
import { useSettingsStore } from '@/store/settingsStore'
import { useUndoStore } from '@/store/undoStore'
import { focusRing } from '@/utils/uiClasses'

export function UndoBanner() {
  const visible = useUndoStore((s) => s.visible)
  const message = useUndoStore((s) => s.message)
  const dismiss = useUndoStore((s) => s.dismiss)

  return (
    <AnimatePresence>
      {visible ? (
        <motion.div
          className="pointer-events-none fixed inset-x-0 bottom-0 z-50 flex justify-center p-4 pb-[max(1rem,env(safe-area-inset-bottom))]"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
        >
          <motion.div
            className="pointer-events-auto flex w-full max-w-lg items-center gap-3 rounded-xl border border-amber-500/40 bg-amber-600/95 px-4 py-3 text-sm text-amber-950 shadow-lg shadow-black/30"
            initial={{ y: 48, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: 48, opacity: 0 }}
            transition={{ duration: 0.3, ease: 'easeInOut' }}
            role="status"
          >
            <p className="min-w-0 flex-1 font-medium leading-snug">{message}</p>
            <button
              type="button"
              onClick={() => {
                if (useSettingsStore.getState().sessionLogging) {
                  useSettingsStore.getState().addLog(
                    'undo_triggered',
                    useContextStore.getState().currentContext,
                    message,
                  )
                }
                dismiss()
              }}
              className={clsx(
                focusRing,
                'shrink-0 cursor-pointer rounded-lg bg-amber-950/10 px-3 py-1.5 text-xs font-semibold uppercase tracking-wide text-amber-950 transition-all hover:bg-amber-950/20 active:scale-95',
              )}
            >
              Undo
            </button>
            <button
              type="button"
              onClick={() => dismiss()}
              className={clsx(
                focusRing,
                'shrink-0 cursor-pointer rounded-lg p-1 text-amber-950/80 transition-all hover:bg-amber-950/10 hover:text-amber-950 active:scale-95',
              )}
              aria-label="Dismiss"
            >
              <X className="h-4 w-4" strokeWidth={2.5} />
            </button>
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  )
}
