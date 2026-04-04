import { motion } from 'framer-motion'
import { ChevronDown } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import { clsx } from 'clsx'
import { cardSurface, headingPage, interactiveSurface } from '@/utils/uiClasses'

interface ExpandableSectionProps {
  title: string
  children: ReactNode
  defaultExpanded?: boolean
}

export function ExpandableSection({
  title,
  children,
  defaultExpanded = false,
}: ExpandableSectionProps) {
  const [expanded, setExpanded] = useState(defaultExpanded)

  return (
    <div className={clsx(cardSurface, 'shadow-black/25')}>
      <button
        type="button"
        onClick={() => setExpanded((e) => !e)}
        className={clsx(
          interactiveSurface,
          'flex w-full items-center justify-between gap-2 rounded-lg px-1 py-1 text-left text-[var(--color-text-secondary)]',
        )}
      >
        <span className={clsx(headingPage, 'font-medium')}>{title}</span>
        <ChevronDown
          className={clsx(
            'h-5 w-5 shrink-0 transition-transform duration-200 ease-in-out',
            expanded && 'rotate-180',
          )}
          aria-hidden
        />
      </button>
      <motion.div
        initial={false}
        animate={{
          height: expanded ? 'auto' : 0,
          opacity: expanded ? 1 : 0,
        }}
        transition={{ duration: 0.28, ease: 'easeInOut' }}
        className="overflow-hidden"
      >
        <div className="pt-3 text-sm text-[var(--color-text-primary)]">
          {children}
        </div>
      </motion.div>
    </div>
  )
}
