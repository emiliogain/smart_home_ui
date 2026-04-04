import io from 'socket.io-client'
import type { Socket } from 'socket.io-client'
import { useContextStore } from '@/store/contextStore'
import { useDeviceStore } from '@/store/deviceStore'
import type { ContextUpdate } from '@/types/context'
import type { DeviceCommand, DeviceState } from '@/types/device'

const backendUrl =
  import.meta.env.VITE_BACKEND_URL ?? 'http://localhost:8000'

let socket: Socket | null = null

export function getSocket(): Socket | null {
  return socket
}

export function disconnectWebSocket(): void {
  socket?.disconnect()
  socket = null
}

export function initializeWebSocket(): Socket | null {
  try {
    disconnectWebSocket()

    socket = io(backendUrl, {
      transports: ['websocket'],
    })

    socket.on('connect', () => {
      console.log('WebSocket connected')
    })

    socket.on('disconnect', () => {
      console.log('WebSocket disconnected')
    })

    socket.on('connect_error', (err) => {
      console.warn('WebSocket connection failed:', err)
    })

    socket.on('context_update', (data: ContextUpdate) => {
      useContextStore.getState().setContext(data)
    })

    socket.on(
      'device_state_update',
      (data: { deviceId: string; state: Partial<DeviceState> }) => {
        useDeviceStore.getState().updateDeviceState(data.deviceId, data.state)
      },
    )

    return socket
  } catch (err) {
    console.warn('WebSocket initialization failed:', err)
    return null
  }
}

export function sendDeviceCommand(command: DeviceCommand): void {
  if (!socket) {
    console.warn('WebSocket not initialized; device command not sent')
    return
  }
  socket.emit('device_command', command)
}
