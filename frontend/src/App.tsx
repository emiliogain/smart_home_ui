import { useEffect } from 'react'
import { Route, Routes } from 'react-router-dom'
import { getSocket } from '@/api/websocket'
import { AppShell } from '@/components/layout/AppShell'
import Dashboard from '@/pages/Dashboard'
import Rooms from '@/pages/Rooms'
import Sensors from '@/pages/Sensors'
import Settings from '@/pages/Settings'
import {
  startMockContextCycle,
  startMockSensorUpdates,
} from '@/utils/mockContextProvider'

export default function App() {
  useEffect(() => {
    let stopCycle: (() => void) | undefined
    let stopSensors: (() => void) | undefined
    let cancelled = false
    let mocksArmed = false

    const armMocks = () => {
      if (cancelled || mocksArmed) return
      mocksArmed = true
      stopCycle = startMockContextCycle()
      stopSensors = startMockSensorUpdates()
      console.info(
        '[mock] Context cycle + sensor drift active (backend optional).',
      )
    }

    if (import.meta.env.DEV) {
      armMocks()
      return () => {
        cancelled = true
        stopCycle?.()
        stopSensors?.()
      }
    }

    const socket = getSocket()
    const delayMs = 2500
    const timer = window.setTimeout(() => {
      if (cancelled || mocksArmed) return
      if (!socket?.connected) {
        armMocks()
      }
    }, delayMs)

    const onConnect = () => {
      window.clearTimeout(timer)
    }

    const onConnectError = () => {
      window.clearTimeout(timer)
      armMocks()
    }

    socket?.once('connect', onConnect)
    socket?.once('connect_error', onConnectError)

    if (!socket) {
      window.clearTimeout(timer)
      armMocks()
    }

    return () => {
      cancelled = true
      window.clearTimeout(timer)
      socket?.off('connect', onConnect)
      socket?.off('connect_error', onConnectError)
      stopCycle?.()
      stopSensors?.()
    }
  }, [])

  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        <Route index element={<Dashboard />} />
        <Route path="rooms" element={<Rooms />} />
        <Route path="sensors" element={<Sensors />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  )
}
