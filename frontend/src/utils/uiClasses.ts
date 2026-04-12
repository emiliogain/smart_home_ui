/** Shared interaction + surface styles for thesis UI consistency */
export const focusRing =
  'focus:outline-none focus:ring-2 focus:ring-[var(--color-secondary)]'

export const interactiveSurface =
  `cursor-pointer transition-all hover:bg-white/5 active:scale-95 ${focusRing}`

/** For controls on tinted / filled surfaces (skip generic hover wash). */
export const interactiveControl =
  `cursor-pointer transition-all active:scale-95 ${focusRing}`

export const cardSurface =
  'rounded-xl bg-[var(--color-surface)] p-4 shadow-md shadow-black/20 transition-all hover:ring-1 hover:ring-white/10'

export const headingPage = 'text-lg font-semibold text-[var(--color-text-primary)]'

export const bodyText = 'text-sm text-[var(--color-text-primary)]'

export const bodyMuted = 'text-sm text-[var(--color-text-secondary)]'
