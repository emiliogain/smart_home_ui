import { clsx } from 'clsx'
import { useEffect, useState } from 'react'
import { cardSurface, interactiveSurface } from '@/utils/uiClasses'

function formatMmSs(total: number): string {
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export function TimerWidget() {
  const [seconds, setSeconds] = useState(0)
  const [isRunning, setIsRunning] = useState(false)
  const [timeUp, setTimeUp] = useState(false)

  useEffect(() => {
    if (!isRunning) return

    const id = window.setInterval(() => {
      setSeconds((s) => {
        if (s <= 1) {
          setIsRunning(false)
          setTimeUp(true)
          return 0
        }
        return s - 1
      })
    }, 1000)

    return () => window.clearInterval(id)
  }, [isRunning])

  const addMinute = () => {
    setTimeUp(false)
    setSeconds((s) => s + 60)
  }

  const toggleRun = () => {
    if (timeUp) setTimeUp(false)
    if (seconds === 0 && !isRunning) return
    setIsRunning((r) => !r)
  }

  const reset = () => {
    setIsRunning(false)
    setSeconds(0)
    setTimeUp(false)
  }

  return (
    <div
      className={clsx(
        cardSurface,
        'w-full transition-shadow',
        timeUp &&
          'ring-2 ring-red-500 ring-offset-2 ring-offset-[var(--color-bg)] animate-pulse hover:ring-red-500',
      )}
    >
      <p
        className={clsx(
          'text-center font-mono text-4xl font-semibold tabular-nums',
          timeUp
            ? 'text-red-400'
            : 'text-[var(--color-accent)]',
        )}
      >
        {formatMmSs(seconds)}
      </p>

      {timeUp ? (
        <p className="mt-2 text-center text-sm font-medium text-red-400">
          Time&apos;s up!
        </p>
      ) : null}

      <div className="mt-4 flex flex-wrap justify-center gap-2">
        <button
          type="button"
          onClick={addMinute}
          className={clsx(
            interactiveSurface,
            'rounded-lg bg-white/10 px-3 py-2 text-xs font-medium text-[var(--color-text-primary)]',
          )}
        >
          +1 min
        </button>
        <button
          type="button"
          onClick={toggleRun}
          disabled={seconds === 0 && !isRunning}
          className={clsx(
            interactiveSurface,
            'rounded-lg bg-[var(--color-secondary)] px-4 py-2 text-xs font-semibold text-[var(--color-primary)] hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:opacity-40 disabled:active:scale-100',
          )}
        >
          {isRunning ? 'Pause' : 'Start'}
        </button>
        <button
          type="button"
          onClick={reset}
          className={clsx(
            interactiveSurface,
            'rounded-lg bg-white/10 px-3 py-2 text-xs font-medium text-[var(--color-text-primary)]',
          )}
        >
          Reset
        </button>
      </div>
    </div>
  )
}
