import io from 'socket.io-client'
import { BACKEND_URL } from '@/utils/constants'
import { useContextStore } from '@/store/contextStore'
import { useDeviceStore } from '@/store/deviceStore'
import type { ContextUpdate } from '@/types/context'
import type { DeviceCommand, DeviceState } from '@/types/device'

let socket: SocketIOClient.Socket | null = null

export function getSocket(): SocketIOClient.Socket | null {
  return socket
}

export function disconnectWebSocket(): void {
  socket?.disconnect()
  socket = null
}

export function initializeWebSocket(): SocketIOClient.Socket | null {
  try {
    disconnectWebSocket()

    const opts: SocketIOClient.ConnectOpts = {
      transports: ['websocket'],
      path: '/socket.io',
    }
    // go-socket.io speaks the Socket.IO v2 / Engine.IO v3 protocol (socket.io-client@2).
    socket = BACKEND_URL ? io(BACKEND_URL, opts) : io(opts)

    socket.on('connect', () => {
      console.log('WebSocket connected')
    })

    socket.on('disconnect', () => {
      console.log('WebSocket disconnected')
    })

    socket.on('connect_error', (err: Error) => {
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
