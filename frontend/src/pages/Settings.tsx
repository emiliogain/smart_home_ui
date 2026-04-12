import { useEffect, useState } from 'react'
import { clsx } from 'clsx'
import { ContextType } from '@/types/context'
import { BACKEND_URL, CONTEXT_LABELS } from '@/utils/constants'
import { ExpandableSection } from '@/components/common/ExpandableSection'
import { ToggleSwitch } from '@/components/controls/ToggleSwitch'
import { getSocket } from '@/api/websocket'
import { useContextStore } from '@/store/contextStore'
import { useSettingsStore } from '@/store/settingsStore'
import { FlaskConical } from 'lucide-react'
import {
  cardSurface,
  focusRing,
  headingPage,
  interactiveSurface,
} from '@/utils/uiClasses'

const CONTEXT_OPTIONS = Object.values(ContextType)

function downloadCsv(content: string, filename: string) {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export default function Settings() {
  const adaptiveMode = useSettingsStore((s) => s.adaptiveMode)
  const setAdaptiveMode = useSettingsStore((s) => s.setAdaptiveMode)
  const studyMode = useSettingsStore((s) => s.studyMode)
  const toggleStudyMode = useSettingsStore((s) => s.toggleStudyMode)
  const sessionLogging = useSettingsStore((s) => s.sessionLogging)
  const toggleSessionLogging = useSettingsStore((s) => s.toggleSessionLogging)
  const sessionLogs = useSettingsStore((s) => s.sessionLogs)
  const simulatedContext = useSettingsStore((s) => s.simulatedContext)
  const setSimulatedContext = useSettingsStore((s) => s.setSimulatedContext)
  const exportLogs = useSettingsStore((s) => s.exportLogs)
  const clearLogs = useSettingsStore((s) => s.clearLogs)

  const currentContext = useContextStore((s) => s.currentContext)
  const confidence = useContextStore((s) => s.confidence)
  const sensorSnapshot = useContextStore((s) => s.sensorSnapshot)

  const [wsConnected, setWsConnected] = useState(false)

  useEffect(() => {
    const s = getSocket()
    if (!s) {
      setWsConnected(false)
      return
    }
    const sync = () => setWsConnected(s.connected)
    sync()
    s.on('connect', sync)
    s.on('disconnect', sync)
    return () => {
      s.off('connect', sync)
      s.off('disconnect', sync)
    }
  }, [])

  const sensorCount = sensorSnapshot?.readings.length ?? 0

  return (
    <div className="mx-auto max-w-lg space-y-8 pb-8 pt-2 lg:max-w-xl">
      <h1 className={headingPage}>Settings</h1>

      <section className={clsx(cardSurface, 'border border-white/10')}>
        <h2 className={clsx(headingPage, 'font-semibold')}>Interface Mode</h2>
        <p className="mt-2 text-sm leading-relaxed text-[var(--color-text-secondary)]">
          <strong className="text-[var(--color-text-primary)]">Adaptive</strong>{' '}
          highlights controls and scenes that match the detected context.
          <br />
          <strong className="text-[var(--color-text-primary)]">Static</strong>{' '}
          shows the same layout every time — useful as a baseline in user
          studies.
        </p>
        <div
          className="mt-4 flex rounded-xl bg-white/10 p-1"
          role="group"
          aria-label="Interface mode"
        >
          <button
            type="button"
            onClick={() => setAdaptiveMode(true)}
            className={clsx(
              interactiveSurface,
              'flex-1 rounded-lg py-2.5 text-sm font-semibold',
              adaptiveMode
                ? 'bg-[var(--color-secondary)] text-[var(--color-primary)] hover:bg-[var(--color-secondary)]'
                : 'text-[var(--color-text-secondary)]',
            )}
          >
            Adaptive
          </button>
          <button
            type="button"
            onClick={() => setAdaptiveMode(false)}
            className={clsx(
              interactiveSurface,
              'flex-1 rounded-lg py-2.5 text-sm font-semibold',
              !adaptiveMode
                ? 'bg-[var(--color-secondary)] text-[var(--color-primary)] hover:bg-[var(--color-secondary)]'
                : 'text-[var(--color-text-secondary)]',
            )}
          >
            Static
          </button>
        </div>
      </section>

      <ExpandableSection title="🧪 Study Mode" defaultExpanded={false}>
        <div className="space-y-4">
          <div className="rounded-xl border border-white/10 bg-white/5 p-3">
            <ToggleSwitch
              isOn={studyMode}
              onToggle={() => toggleStudyMode()}
              label="Enable study mode"
              icon={<FlaskConical className="h-5 w-5" strokeWidth={2} />}
            />
          </div>

          {studyMode ? (
            <div className="space-y-3">
              <label className="block text-sm font-medium text-[var(--color-text-secondary)]">
                Simulated context
                <select
                  value={simulatedContext ?? ''}
                  onChange={(e) => {
                    const v = e.target.value
                    setSimulatedContext(
                      v === '' ? null : (v as ContextType),
                    )
                  }}
                  className={clsx(
                    focusRing,
                    'mt-1 w-full cursor-pointer rounded-lg border border-white/10 bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)] transition-all',
                  )}
                >
                  <option value="">(use live context)</option>
                  {CONTEXT_OPTIONS.map((c) => (
                    <option key={c} value={c}>
                      {CONTEXT_LABELS[c].label} ({c})
                    </option>
                  ))}
                </select>
              </label>

              <div className="rounded-xl border border-white/10 bg-white/5 p-3">
                <ToggleSwitch
                  isOn={sessionLogging}
                  onToggle={() => toggleSessionLogging()}
                  label="Session logging"
                  icon={<FlaskConical className="h-5 w-5" strokeWidth={2} />}
                />
              </div>

              {sessionLogging ? (
                <p className="text-sm text-[var(--color-text-secondary)]">
                  Logged events:{' '}
                  <span className="font-semibold text-[var(--color-text-primary)]">
                    {sessionLogs.length}
                  </span>
                </p>
              ) : null}

              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() =>
                    downloadCsv(
                      exportLogs(),
                      `session-logs-${Date.now()}.csv`,
                    )
                  }
                  className={clsx(
                    interactiveSurface,
                    'rounded-lg bg-[var(--color-secondary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary)] hover:opacity-90',
                  )}
                >
                  Export Logs (CSV)
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (
                      sessionLogs.length > 0 &&
                      window.confirm(
                        'Clear all session logs? This cannot be undone.',
                      )
                    ) {
                      clearLogs()
                    }
                  }}
                  className={clsx(
                    interactiveSurface,
                    'rounded-lg bg-white/10 px-4 py-2 text-sm font-medium text-[var(--color-text-primary)]',
                  )}
                >
                  Clear Logs
                </button>
              </div>
            </div>
          ) : null}
        </div>
      </ExpandableSection>

      <section className={clsx(cardSurface, 'border border-white/10')}>
        <h2 className={clsx(headingPage, 'font-semibold')}>System Info</h2>
        <dl className="mt-3 space-y-2 text-sm">
          <div className="flex justify-between gap-4">
            <dt className="text-[var(--color-text-secondary)]">Backend URL</dt>
            <dd className="max-w-[60%] break-all text-right font-mono text-[var(--color-text-primary)]">
              {BACKEND_URL || '(same origin — Vite proxy → backend)'}
            </dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="text-[var(--color-text-secondary)]">WebSocket</dt>
            <dd
              className={
                wsConnected
                  ? 'font-medium text-[var(--color-success)]'
                  : 'font-medium text-[var(--color-danger)]'
              }
            >
              {wsConnected ? 'Connected' : 'Disconnected'}
            </dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="text-[var(--color-text-secondary)]">Context</dt>
            <dd className="text-right text-[var(--color-text-primary)]">
              {CONTEXT_LABELS[currentContext].label}
            </dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="text-[var(--color-text-secondary)]">Confidence</dt>
            <dd className="tabular-nums text-[var(--color-text-primary)]">
              {Math.round(confidence * 100)}%
            </dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="text-[var(--color-text-secondary)]">Sensor count</dt>
            <dd className="tabular-nums text-[var(--color-text-primary)]">
              {sensorCount}
            </dd>
          </div>
        </dl>
      </section>

      <section
        className={clsx(
          cardSurface,
          'border border-white/10 text-center',
        )}
      >
        <h2 className={clsx(headingPage, 'font-semibold')}>About</h2>
        <p className="mt-3 text-sm font-medium text-[var(--color-text-primary)]">
          Smart Home Adaptive UI
        </p>
        <p className="mt-2 text-sm text-[var(--color-text-secondary)]">
          Bachelor&apos;s Thesis — Innopolis University 2025
        </p>
        <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
          Yoqub Davlatov &amp; Emil Gainullin
        </p>
      </section>
    </div>
  )
}
