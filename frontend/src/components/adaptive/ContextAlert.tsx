import { LayoutGroup, motion } from 'framer-motion'
import { clsx } from 'clsx'
import { useState } from 'react'
import { ContextType } from '@/types/context'
import type { AdaptationConfig } from '@/utils/adaptationRules'
import { AlertCard } from '@/components/cards/AlertCard'
import {
  renderDeviceControl,
  type DeviceControlHandlers,
} from '@/components/adaptive/renderDeviceControl'
import { useDevices } from '@/hooks/useDevices'
import { headingPage } from '@/utils/uiClasses'

interface ContextAlertProps {
  config: AdaptationConfig
  effectiveContext: ContextType
}

export function ContextAlert({ config, effectiveContext }: ContextAlertProps) {
  const { getDeviceById, handleToggle, handleSetValue, handleToggleLocked } =
    useDevices()

  const handlers: DeviceControlHandlers = {
    handleToggle,
    handleSetValue,
    handleToggleLocked,
  }

  const message = config.alertMessage ?? ''
  const scene = config.suggestedScene
  const severity =
    effectiveContext === ContextType.ALERT_TOO_COLD ? 'danger' : 'warning'

  const [showSceneActions, setShowSceneActions] = useState(true)

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6">
      {message ? (
        <AlertCard
          message={message}
          suggestedScene={showSceneActions ? scene : null}
          onApply={() => setShowSceneActions(false)}
          onDismiss={() => setShowSceneActions(false)}
          severity={severity}
          pulse
        />
      ) : null}

      <section>
        <h2 className={clsx(headingPage, 'mb-3')}>Suggested controls</h2>
        <LayoutGroup id="context-alert-hero">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
            {config.heroControls.map((id) => (
              <motion.div
                key={id}
                layout
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.22, layout: { duration: 0.28 } }}
                className="min-w-0"
              >
                {renderDeviceControl(getDeviceById(id), handlers)}
              </motion.div>
            ))}
          </div>
        </LayoutGroup>
      </section>
    </div>
  )
}
