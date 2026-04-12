import { HelpCircle } from 'lucide-react'

interface ConfidenceBadgeProps {
  confidence: number
}

export function ConfidenceBadge({ confidence }: ConfidenceBadgeProps) {
  const pct = Math.round(Math.min(1, Math.max(0, confidence)) * 100)
  const dotClass =
    confidence >= 0.8
      ? 'bg-[var(--color-success)]'
      : confidence >= 0.6
        ? 'bg-[var(--color-warning)]'
        : 'bg-[var(--color-danger)]'

  return (
    <span className="inline-flex items-center gap-1.5 rounded-full border border-[var(--color-text-secondary)]/20 bg-[var(--color-surface)] px-2 py-0.5 text-xs text-[var(--color-text-secondary)] shadow-sm shadow-black/15">
      <span
        className={`h-2 w-2 shrink-0 rounded-full ${dotClass}`}
        aria-hidden
      />
      <span className="tabular-nums">{pct}%</span>
      {confidence < 0.6 ? (
        <HelpCircle className="h-3.5 w-3.5 shrink-0 opacity-90" aria-hidden />
      ) : null}
    </span>
  )
}
