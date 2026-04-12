import { AlarmClock, Pencil } from 'lucide-react'
import { clsx } from 'clsx'
import { useMemo, useState } from 'react'
import type { Device } from '@/types/device'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { cardSurface, interactiveSurface } from '@/utils/uiClasses'

function hour24ToDisplay(h24: number): string {
  const h = ((Math.floor(h24) % 24) + 24) % 24
  const isPm = h >= 12
  const h12 = h % 12 === 0 ? 12 : h % 12
  return `${String(h12).padStart(2, '0')}:00 ${isPm ? 'PM' : 'AM'}`
}

function hour24ToParts(h24: number): { h12: number; pm: boolean } {
  const h = ((Math.floor(h24) % 24) + 24) % 24
  return {
    h12: h % 12 === 0 ? 12 : h % 12,
    pm: h >= 12,
  }
}

function partsToHour24(h12: number, pm: boolean): number {
  const h = Math.min(12, Math.max(1, Math.round(h12)))
  if (h === 12) return pm ? 12 : 0
  return pm ? h + 12 : h
}

interface AlarmWidgetProps {
  device: Device
  onSetValue: (value: number) => void
  onToggle: () => void
}

export function AlarmWidget({
  device,
  onSetValue,
  onToggle,
}: AlarmWidgetProps) {
  const disabled = !device.controllable
  const hour24 = device.state.value ?? 7
  const [isEditing, setIsEditing] = useState(false)
  const parts = useMemo(() => hour24ToParts(hour24), [hour24])
  const [editH12, setEditH12] = useState(parts.h12)
  const [editPm, setEditPm] = useState(parts.pm)

  const openEdit = () => {
    setEditH12(parts.h12)
    setEditPm(parts.pm)
    setIsEditing(true)
  }

  const applyEdit = () => {
    onSetValue(partsToHour24(editH12, editPm))
    setIsEditing(false)
  }

  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <div className="flex items-center gap-2 text-[var(--color-secondary)]">
        <AlarmClock className="h-5 w-5 shrink-0" strokeWidth={2} />
        <span className="text-sm font-semibold text-[var(--color-text-primary)]">
          {device.name}
        </span>
      </div>

      <p className="mt-3 text-center font-mono text-2xl font-semibold tabular-nums text-[var(--color-text-primary)]">
        {hour24ToDisplay(hour24)}
      </p>

      <div className="mt-2 flex justify-center">
        <button
          type="button"
          disabled={disabled}
          onClick={openEdit}
          className={clsx(
            interactiveSurface,
            'inline-flex items-center gap-1 rounded-lg bg-white/10 px-3 py-1.5 text-xs font-medium text-[var(--color-text-secondary)] disabled:opacity-50 disabled:hover:bg-white/10 disabled:active:scale-100',
          )}
        >
          <Pencil className="h-3.5 w-3.5" strokeWidth={2} />
          Edit
        </button>
      </div>

      {isEditing ? (
        <div className="mt-3 flex flex-col items-center gap-3 rounded-lg bg-white/5 p-3">
          <div className="flex items-center gap-2">
            <label className="sr-only" htmlFor={`alarm-h-${device.id}`}>
              Hour
            </label>
            <input
              id={`alarm-h-${device.id}`}
              type="number"
              min={1}
              max={12}
              value={editH12}
              onChange={(e) =>
                setEditH12(Math.min(12, Math.max(1, Number(e.target.value) || 1)))
              }
              className="w-14 rounded-md border border-white/10 bg-[var(--color-bg)] px-2 py-1 text-center text-sm text-[var(--color-text-primary)]"
            />
            <div className="flex rounded-lg bg-white/10 p-0.5">
              <button
                type="button"
                onClick={() => setEditPm(false)}
                className={clsx(
                  interactiveSurface,
                  'rounded-md px-2 py-1 text-xs font-medium',
                  !editPm
                    ? 'bg-[var(--color-secondary)] text-[var(--color-primary)]'
                    : 'text-[var(--color-text-secondary)]',
                )}
              >
                AM
              </button>
              <button
                type="button"
                onClick={() => setEditPm(true)}
                className={clsx(
                  interactiveSurface,
                  'rounded-md px-2 py-1 text-xs font-medium',
                  editPm
                    ? 'bg-[var(--color-secondary)] text-[var(--color-primary)]'
                    : 'text-[var(--color-text-secondary)]',
                )}
              >
                PM
              </button>
            </div>
          </div>
          <button
            type="button"
            onClick={applyEdit}
            className={clsx(
              interactiveSurface,
              'rounded-lg bg-[var(--color-secondary)] px-4 py-1.5 text-xs font-semibold text-[var(--color-primary)] hover:opacity-90',
            )}
          >
            Save
          </button>
        </div>
      ) : null}

      <div className="mt-4 border-t border-white/10 pt-4">
        <ToggleSwitch
          isOn={device.state.on}
          onToggle={onToggle}
          label="Alarm enabled"
          icon={<AlarmClock className="h-5 w-5" strokeWidth={2} />}
          disabled={disabled}
        />
      </div>
    </div>
  )
}
