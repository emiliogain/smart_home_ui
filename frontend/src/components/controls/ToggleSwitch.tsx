import { motion } from 'framer-motion'
import { clsx } from 'clsx'
import type { ReactNode } from 'react'
import { interactiveSurface } from '@/utils/uiClasses'

interface ToggleSwitchProps {
  isOn: boolean
  onToggle: () => void
  label: string
  icon: ReactNode
  disabled?: boolean
}

export function ToggleSwitch({
  isOn,
  onToggle,
  label,
  icon,
  disabled = false,
}: ToggleSwitchProps) {
  return (
    <div className="flex items-center justify-between gap-3">
      <div className="flex min-w-0 items-center gap-2 text-[var(--color-text-primary)]">
        <span className="shrink-0 text-[var(--color-secondary)]">{icon}</span>
        <span className="truncate text-sm font-medium">{label}</span>
      </div>
      <div className="flex shrink-0 flex-col items-center gap-1">
        <button
          type="button"
          role="switch"
          aria-checked={isOn}
          disabled={disabled}
          onClick={onToggle}
          className={clsx(
            interactiveSurface,
            'relative h-6 w-11 shrink-0 rounded-full disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-transparent disabled:active:scale-100',
            isOn ? 'bg-[var(--color-secondary)]' : 'bg-gray-600',
          )}
        >
          <motion.span
            className="absolute top-1 h-4 w-4 rounded-full bg-white shadow-sm"
            initial={false}
            animate={{ left: isOn ? 24 : 4 }}
            transition={{ type: 'spring', stiffness: 500, damping: 32 }}
          />
        </button>
        <span className="text-[10px] font-medium uppercase tracking-wide text-[var(--color-text-secondary)]">
          {isOn ? 'ON' : 'OFF'}
        </span>
      </div>
    </div>
  )
}
