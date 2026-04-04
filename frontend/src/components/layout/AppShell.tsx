import { Home, Settings } from 'lucide-react'
import { NavLink, Outlet } from 'react-router-dom'
import { clsx } from 'clsx'
import { UndoBanner } from '@/components/common/UndoBanner'
import { BottomNav } from '@/components/layout/BottomNav'
import { ContextBanner } from '@/components/layout/ContextBanner'
import { interactiveSurface } from '@/utils/uiClasses'

export function AppShell() {
  return (
    <div className="min-h-screen bg-[var(--color-bg)]">
      <header className="sticky top-0 z-40 flex h-14 items-center bg-[var(--color-surface)] px-3 lg:px-8">
        <div className="mx-auto flex w-full max-w-5xl items-center gap-2 lg:max-w-7xl">
          <div className="flex shrink-0 items-center gap-2 text-[var(--color-text-primary)]">
            <Home className="h-5 w-5 text-[var(--color-secondary)]" aria-hidden />
            <span className="truncate text-sm font-semibold sm:text-lg">
              Smart Home
            </span>
          </div>
          <div className="flex min-w-0 flex-1 justify-center px-1">
            <ContextBanner />
          </div>
          <NavLink
            to="/settings"
            className={clsx(
              interactiveSurface,
              'flex shrink-0 rounded-lg p-2 text-[var(--color-text-secondary)] hover:text-[var(--color-secondary)]',
            )}
            aria-label="Settings"
          >
            <Settings className="h-5 w-5" strokeWidth={2} />
          </NavLink>
        </div>
      </header>

      <main className="mx-auto w-full max-w-5xl px-4 pb-20 pt-16 lg:max-w-7xl lg:px-8">
        <Outlet />
      </main>

      <BottomNav />
      <UndoBanner />
    </div>
  )
}
