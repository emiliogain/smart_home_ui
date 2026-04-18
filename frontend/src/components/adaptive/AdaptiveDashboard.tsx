import { LayoutGroup, motion } from 'framer-motion'
import { clsx } from 'clsx'
import { ContextType } from '@/types/context'
import { DeviceType } from '@/types/device'
import { ROOMS, ROOM_LABELS } from '@/utils/constants'
import type { AdaptationConfig } from '@/utils/adaptationRules'
import { AnimatedTransition } from '@/components/common/AnimatedTransition'
import { ExpandableSection } from '@/components/common/ExpandableSection'
import { ContextAlert } from '@/components/adaptive/ContextAlert'
import { ContextCooking } from '@/components/adaptive/ContextCooking'
import { ContextNoOneHome } from '@/components/adaptive/ContextNoOneHome'
import { ContextReading } from '@/components/adaptive/ContextReading'
import { ContextSleeping } from '@/components/adaptive/ContextSleeping'
import { ContextWatchingTV } from '@/components/adaptive/ContextWatchingTV'
import {
  renderDeviceControl,
  type DeviceControlHandlers,
} from '@/components/adaptive/renderDeviceControl'
import { useAdaptation } from '@/hooks/useAdaptation'
import { useDevices } from '@/hooks/useDevices'
import { DeviceGroupCard } from '@/components/cards/DeviceGroupCard'
import { StaticDashboard } from '@/components/adaptive/StaticDashboard'
import { useSettingsStore } from '@/store/settingsStore'
import { useContextStore } from '@/store/contextStore'
import { formatSensorScalar } from '@/utils/formatSensorValue'
import { cardSurface, headingPage } from '@/utils/uiClasses'

const SECURITY_TYPES = new Set([
  DeviceType.DOOR_LOCK,
  DeviceType.WINDOW_SENSOR,
  DeviceType.CAMERA,
])

function DashboardSensorSidebar() {
  const snapshot = useContextStore((s) => s.sensorSnapshot)
  const readings = snapshot?.readings ?? []

  return (
    <aside
      className={clsx(
        cardSurface,
        'sticky top-24 hidden max-h-[calc(100vh-8rem)] w-full max-w-[300px] shrink-0 overflow-y-auto border border-white/10 lg:block lg:p-5',
      )}
    >
      <h3
        className={clsx(
          headingPage,
          'mb-3 text-[var(--color-text-secondary)]',
        )}
      >
        Live sensors
      </h3>
      {readings.length === 0 ? (
        <p className="text-sm text-[var(--color-text-secondary)]">
          No snapshot yet. Connect the backend or enable mock context.
        </p>
      ) : (
        <ul className="space-y-2.5">
          {readings.map((r) => (
            <li
              key={`${r.sensorId}-${r.at}`}
              className="flex items-baseline justify-between gap-2 text-sm"
            >
              <span className="truncate text-[var(--color-text-secondary)]">
                {r.sensorLabel ?? r.sensorId}
              </span>
              <span className="shrink-0 tabular-nums text-[var(--color-text-primary)]">
                {formatSensorScalar(r.value)}
                {r.unit ? ` ${r.unit}` : ''}
              </span>
            </li>
          ))}
        </ul>
      )}
    </aside>
  )
}

function GenericContextDashboard() {
  const {
    devices,
    getDevicesByRoom,
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  } = useDevices()

  const securityDevices = devices.filter((d) => SECURITY_TYPES.has(d.type))

  return (
    <div className="space-y-6">
      {ROOMS.map((room) => {
        const roomDevs = getDevicesByRoom(room).filter(
          (d) => !SECURITY_TYPES.has(d.type),
        )
        return (
          <DeviceGroupCard
            key={room}
            title={ROOM_LABELS[room]}
            devices={roomDevs}
            onToggle={handleToggle}
            onSetValue={handleSetValue}
            onToggleLocked={handleToggleLocked}
          />
        )
      })}
      <DeviceGroupCard
        title="Security"
        devices={securityDevices}
        onToggle={handleToggle}
        onSetValue={handleSetValue}
        onToggleLocked={handleToggleLocked}
      />
    </div>
  )
}

function AllControlsPanel() {
  const {
    devices,
    getDevicesByRoom,
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  } = useDevices()

  const handlers: DeviceControlHandlers = {
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  }

  const securityDevices = devices.filter((d) => SECURITY_TYPES.has(d.type))

  return (
    <LayoutGroup id="all-controls">
      <div className="space-y-8">
        {ROOMS.map((room) => {
          const roomDevs = getDevicesByRoom(room).filter(
            (d) => !SECURITY_TYPES.has(d.type),
          )
          if (roomDevs.length === 0) return null
          return (
            <div key={room}>
              <h3
                className={clsx(
                  headingPage,
                  'mb-3 uppercase tracking-wide text-[var(--color-text-secondary)]',
                )}
              >
                {ROOM_LABELS[room]}
              </h3>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
                {roomDevs.map((d) => (
                  <motion.div
                    key={d.id}
                    layout
                    initial={{ opacity: 0, y: 6 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.22, layout: { duration: 0.28 } }}
                    className="min-w-0"
                  >
                    {renderDeviceControl(d, handlers)}
                  </motion.div>
                ))}
              </div>
            </div>
          )
        })}
        {securityDevices.length > 0 ? (
          <div>
            <h3
              className={clsx(
                headingPage,
                'mb-3 uppercase tracking-wide text-[var(--color-text-secondary)]',
              )}
            >
              Security
            </h3>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
              {securityDevices.map((d) => (
                <motion.div
                  key={d.id}
                  layout
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.22, layout: { duration: 0.28 } }}
                  className="min-w-0"
                >
                  {renderDeviceControl(d, handlers)}
                </motion.div>
              ))}
            </div>
          </div>
        ) : null}
      </div>
    </LayoutGroup>
  )
}

function ContextSwitch({
  effectiveContext,
  config,
}: {
  effectiveContext: ContextType
  config: AdaptationConfig
}) {
  switch (effectiveContext) {
    case ContextType.NO_ONE_HOME:
      return <ContextNoOneHome />
    case ContextType.READING_LIVING_ROOM:
      return <ContextReading />
    case ContextType.WATCHING_TV_LIVING_ROOM:
      return <ContextWatchingTV />
    case ContextType.COOKING_KITCHEN:
      return <ContextCooking />
    case ContextType.SLEEPING:
      return <ContextSleeping />
    case ContextType.ALERT_TOO_HOT:
    case ContextType.ALERT_TOO_COLD:
      return (
        <ContextAlert
          config={config}
          effectiveContext={effectiveContext}
        />
      )
    default:
      return <GenericContextDashboard />
  }
}

export function AdaptiveDashboard() {
  const adaptiveMode = useSettingsStore((s) => s.adaptiveMode)
  const { config, effectiveContext } = useAdaptation()

  if (!adaptiveMode) {
    return <StaticDashboard />
  }

  const themeClass =
    config.theme === 'dark'
      ? 'theme-dark'
      : config.theme === 'alert'
        ? 'theme-alert'
        : ''

  return (
    <div className={clsx('space-y-8 pb-4', themeClass)}>
      <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:gap-10">
        <div className="min-w-0 flex-1 space-y-8">
          <AnimatedTransition layoutKey={effectiveContext}>
            <ContextSwitch effectiveContext={effectiveContext} config={config} />
          </AnimatedTransition>

          <ExpandableSection title="All Controls" defaultExpanded={false}>
            <AllControlsPanel />
          </ExpandableSection>
        </div>

        <DashboardSensorSidebar />
      </div>
    </div>
  )
}
