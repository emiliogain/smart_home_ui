import { AnimatePresence, motion } from 'framer-motion'
import type { ReactNode } from 'react'

interface AnimatedTransitionProps {
  children: ReactNode
  className?: string
  layoutKey: string
}

export function AnimatedTransition({
  children,
  className,
  layoutKey,
}: AnimatedTransitionProps) {
  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={layoutKey}
        className={className}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -20 }}
        transition={{ duration: 0.3, ease: 'easeInOut' }}
      >
        {children}
      </motion.div>
    </AnimatePresence>
  )
}
