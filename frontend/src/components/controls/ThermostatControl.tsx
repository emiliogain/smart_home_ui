import { Minus, Plus, Thermometer } from 'lucide-react'
import { clsx } from 'clsx'
import { ROOM_LABELS } from '@/utils/constants'
import type { Device } from '@/types/device'
import { cardSurface, interactiveControl } from '@/utils/uiClasses'

const MIN_T = 16
const MAX_T = 30

interface ThermostatControlProps {
  device: Device
  onSetValue: (value: number) => void
}

export function ThermostatControl({ device, onSetValue }: ThermostatControlProps) {
  const temp = Math.min(MAX_T, Math.max(MIN_T, device.state.value ?? 20))
  const disabled = !device.controllable
  const roomLabel =
    device.room in ROOM_LABELS
      ? ROOM_LABELS[device.room as keyof typeof ROOM_LABELS]
      : device.room

  const adjust = (delta: number) => {
    const next = Math.min(MAX_T, Math.max(MIN_T, temp + delta))
    if (next !== temp) onSetValue(next)
  }

  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <div className="mb-4 flex items-start gap-2 text-[var(--color-text-primary)]">
        <Thermometer
          className="mt-0.5 h-5 w-5 shrink-0 text-[var(--color-secondary)]"
          strokeWidth={2}
        />
        <div className="min-w-0">
          <p className="text-sm font-semibold">{device.name}</p>
          <p className="text-xs text-[var(--color-text-secondary)]">{roomLabel}</p>
        </div>
      </div>

      <div className="flex items-center justify-center gap-4">
        <button
          type="button"
          disabled={disabled || temp <= MIN_T}
          onClick={() => adjust(-1)}
          className={clsx(
            interactiveControl,
            'flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-white/10 text-[var(--color-text-primary)] hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-white/10 disabled:active:scale-100',
          )}
          aria-label="Decrease temperature"
        >
          <Minus className="h-5 w-5" strokeWidth={2} />
        </button>
        <p className="min-w-[5rem] text-center text-3xl font-semibold tabular-nums text-[var(--color-text-primary)]">
          {temp}°C
        </p>
        <button
          type="button"
          disabled={disabled || temp >= MAX_T}
          onClick={() => adjust(1)}
          className={clsx(
            interactiveControl,
            'flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-white/10 text-[var(--color-text-primary)] hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-white/10 disabled:active:scale-100',
          )}
          aria-label="Increase temperature"
        >
          <Plus className="h-5 w-5" strokeWidth={2} />
        </button>
      </div>

      <p className="mt-3 text-center text-xs text-[var(--color-text-secondary)]">
        Target
      </p>
    </div>
  )
}
