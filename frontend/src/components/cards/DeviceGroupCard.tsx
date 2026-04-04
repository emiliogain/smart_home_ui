import { LayoutGroup, motion } from 'framer-motion'
import { clsx } from 'clsx'
import type { Device } from '@/types/device'
import {
  renderDeviceControl,
  type DeviceControlHandlers,
} from '@/components/adaptive/renderDeviceControl'
import { headingPage } from '@/utils/uiClasses'

interface DeviceGroupCardProps {
  title: string
  devices: Device[]
  onToggle: (deviceId: string) => void
  onSetValue: (deviceId: string, value: number) => void
  onToggleLocked?: (deviceId: string) => void
}

export function DeviceGroupCard({
  title,
  devices,
  onToggle,
  onSetValue,
  onToggleLocked,
}: DeviceGroupCardProps) {
  const handlers: DeviceControlHandlers = {
    handleToggle: onToggle,
    handleSetValue: onSetValue,
    handleToggleLocked: onToggleLocked ?? (() => undefined),
  }

  const groupId = `room-${title.replace(/\s+/g, '-').toLowerCase()}`

  return (
    <section>
      <h2 className={clsx(headingPage, 'mb-3')}>{title}</h2>
      {devices.length === 0 ? (
        <p className="text-sm text-[var(--color-text-secondary)]">No devices</p>
      ) : (
        <LayoutGroup id={groupId}>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4">
            {devices.map((device) => (
              <motion.div
                key={device.id}
                layout
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.22, layout: { duration: 0.28 } }}
                className="min-w-0"
              >
                {renderDeviceControl(device, handlers)}
              </motion.div>
            ))}
          </div>
        </LayoutGroup>
      )}
    </section>
  )
}
