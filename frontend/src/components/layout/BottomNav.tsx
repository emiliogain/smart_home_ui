import { BarChart3, Home, LayoutGrid, Settings } from 'lucide-react'
import { NavLink } from 'react-router-dom'
import { clsx } from 'clsx'
import { interactiveSurface } from '@/utils/uiClasses'

const tabClass = ({ isActive }: { isActive: boolean }) =>
  clsx(
    interactiveSurface,
    'flex min-h-[3rem] min-w-[3rem] flex-col items-center justify-center gap-0.5 rounded-xl px-2 py-1 text-xs transition-colors sm:gap-1',
    isActive
      ? 'text-[var(--color-secondary)]'
      : 'text-[var(--color-text-secondary)]',
  )

export function BottomNav() {
  return (
    <nav
      className="fixed bottom-0 left-0 right-0 z-30 flex h-16 items-center justify-around border-t border-white/10 bg-[var(--color-surface)] px-2"
      aria-label="Main navigation"
    >
      <NavLink to="/" className={tabClass} end>
        <Home className="h-5 w-5" strokeWidth={2} aria-hidden />
        <span className="hidden max-w-[4.5rem] truncate sm:inline">Home</span>
      </NavLink>
      <NavLink to="/rooms" className={tabClass}>
        <LayoutGrid className="h-5 w-5" strokeWidth={2} aria-hidden />
        <span className="hidden max-w-[4.5rem] truncate sm:inline">Rooms</span>
      </NavLink>
      <NavLink to="/sensors" className={tabClass}>
        <BarChart3 className="h-5 w-5" strokeWidth={2} aria-hidden />
        <span className="hidden max-w-[4.5rem] truncate sm:inline">Sensors</span>
      </NavLink>
      <NavLink to="/settings" className={tabClass}>
        <Settings className="h-5 w-5" strokeWidth={2} aria-hidden />
        <span className="hidden max-w-[4.5rem] truncate sm:inline">
          Settings
        </span>
      </NavLink>
    </nav>
  )
}
