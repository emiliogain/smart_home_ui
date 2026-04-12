import { Sun } from 'lucide-react'
import { clsx } from 'clsx'
import type { Device } from '@/types/device'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { cardSurface, interactiveSurface } from '@/utils/uiClasses'

interface LightDimmerProps {
  device: Device
  onToggle: () => void
  onSetValue: (value: number) => void
  onModeChange?: (mode: 'warm_white' | 'cool_white') => void
}

export function LightDimmer({
  device,
  onToggle,
  onSetValue,
  onModeChange,
}: LightDimmerProps) {
  const isOn = device.state.on
  const pct = Math.min(100, Math.max(0, device.state.value ?? 0))
  const disabled = !device.controllable
  const showModeButtons =
    device.lightMode != null && onModeChange != null
  const activeMode = device.lightMode

  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <ToggleSwitch
        isOn={isOn}
        onToggle={onToggle}
        label={device.name}
        icon={<Sun className="h-5 w-5" strokeWidth={2} />}
        disabled={disabled}
      />

      {isOn ? (
        <div className="mt-4 space-y-3">
          <div>
            <div className="mb-2 flex items-center justify-between text-xs text-[var(--color-text-secondary)]">
              <span>Brightness</span>
              <span className="tabular-nums font-medium text-[var(--color-text-primary)]">
                {pct}%
              </span>
            </div>
            <div className="relative h-2 rounded-full bg-gray-700">
              <div
                className="pointer-events-none absolute inset-y-0 left-0 rounded-full bg-[var(--color-secondary)]"
                style={{ width: `${pct}%` }}
              />
              <input
                type="range"
                min={0}
                max={100}
                value={pct}
                disabled={disabled}
                onChange={(e) => onSetValue(Number(e.target.value))}
                className="absolute inset-0 h-2 w-full cursor-pointer opacity-0 focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-secondary)] disabled:cursor-not-allowed"
              />
            </div>
          </div>

          {showModeButtons ? (
            <div className="flex gap-2">
              <button
                type="button"
                disabled={disabled}
                onClick={() => onModeChange!('warm_white')}
                className={clsx(
                  interactiveSurface,
                  'flex-1 rounded-lg py-2 text-xs font-medium disabled:opacity-50',
                  activeMode === 'warm_white'
                    ? 'bg-[var(--color-secondary)] text-[var(--color-primary)] hover:bg-[var(--color-secondary)]/90'
                    : 'bg-white/10 text-[var(--color-text-secondary)]',
                )}
              >
                Warm White
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => onModeChange!('cool_white')}
                className={clsx(
                  interactiveSurface,
                  'flex-1 rounded-lg py-2 text-xs font-medium disabled:opacity-50',
                  activeMode === 'cool_white'
                    ? 'bg-[var(--color-secondary)] text-[var(--color-primary)] hover:bg-[var(--color-secondary)]/90'
                    : 'bg-white/10 text-[var(--color-text-secondary)]',
                )}
              >
                Cool White
              </button>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
