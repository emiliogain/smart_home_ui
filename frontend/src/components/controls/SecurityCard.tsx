import { DoorOpen, Lock, Unlock } from 'lucide-react'
import { clsx } from 'clsx'
import { DeviceType, type Device } from '@/types/device'
import { cardSurface, interactiveSurface } from '@/utils/uiClasses'

interface SecurityCardProps {
  device: Device
  onToggle: () => void
}

export function SecurityCard({ device, onToggle }: SecurityCardProps) {
  const disabled = !device.controllable
  const locked = device.state.locked === true

  const isWindow = device.type === DeviceType.WINDOW_SENSOR

  const statusLine = isWindow
    ? locked
      ? 'Closed ✅'
      : 'Open ⚠️'
    : locked
      ? 'Locked ✅'
      : 'Unlocked ⚠️'

  const buttonLabel = isWindow
    ? locked
      ? 'Open'
      : 'Close'
    : locked
      ? 'Unlock'
      : 'Lock'

  return (
    <div className={clsx(cardSurface, 'w-full')}>
      <div className="flex items-start gap-3">
        <div
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white/10 text-[var(--color-secondary)]"
          aria-hidden
        >
          {isWindow ? (
            <DoorOpen className="h-5 w-5" strokeWidth={2} />
          ) : locked ? (
            <Lock className="h-5 w-5" strokeWidth={2} />
          ) : (
            <Unlock className="h-5 w-5" strokeWidth={2} />
          )}
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
            {device.name}
          </h3>
          <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
            {statusLine}
          </p>
        </div>
      </div>

      <button
        type="button"
        disabled={disabled}
        onClick={onToggle}
        className={clsx(
          interactiveSurface,
          'mt-4 w-full rounded-lg py-2.5 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-white/10 disabled:active:scale-100',
          'bg-white/10 text-[var(--color-text-primary)]',
        )}
      >
        {buttonLabel}
      </button>
    </div>
  )
}
