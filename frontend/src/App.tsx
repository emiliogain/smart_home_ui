import { useEffect } from 'react'
import { Route, Routes } from 'react-router-dom'
import { fetchCurrentContext } from '@/api/context'
import { getSocket } from '@/api/websocket'
import { useContextStore } from '@/store/contextStore'
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
    void fetchCurrentContext().then((update) => {
      if (update) useContextStore.getState().setContext(update)
    })
  }, [])

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
      console.info('[mock] Backend unreachable — mock context cycle active.')
    }

    const disarmMocks = () => {
      if (!mocksArmed) return
      mocksArmed = false
      stopCycle?.()
      stopSensors?.()
      stopCycle = undefined
      stopSensors = undefined
      console.info('[ws] Real backend connected — mocks stopped.')
    }

    const socket = getSocket()

    // Fall back to mocks if the socket hasn't connected within 3 seconds.
    const timer = window.setTimeout(() => {
      if (cancelled || mocksArmed) return
      if (!socket?.connected) {
        armMocks()
      }
    }, 3000)

    const onConnect = () => {
      window.clearTimeout(timer)
      disarmMocks()
      void fetchCurrentContext().then((update) => {
        if (update) useContextStore.getState().setContext(update)
      })
    }

    const onConnectError = () => {
      window.clearTimeout(timer)
      armMocks()
    }

    // Stop mocks as soon as the first real context update arrives.
    const onContextUpdate = () => {
      disarmMocks()
    }

    socket?.once('connect', onConnect)
    socket?.once('connect_error', onConnectError)
    socket?.on('context_update', onContextUpdate)

    if (!socket) {
      window.clearTimeout(timer)
      armMocks()
    }

    return () => {
      cancelled = true
      window.clearTimeout(timer)
      socket?.off('connect', onConnect)
      socket?.off('connect_error', onConnectError)
      socket?.off('context_update', onContextUpdate)
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
