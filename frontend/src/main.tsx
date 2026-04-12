import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { initializeWebSocket } from '@/api/websocket'
import App from './App'
import './index.css'

const rootEl = document.getElementById('root')

try {
  console.log('Smart Home Adaptive UI — Starting...')

  try {
    initializeWebSocket()
  } catch (wsErr) {
    console.warn('WebSocket setup failed; continuing with mock/offline UI.', wsErr)
  }

  if (!rootEl) {
    throw new Error('Root element #root not found')
  }

  createRoot(rootEl).render(
    <StrictMode>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </StrictMode>,
  )
} catch (err) {
  console.error('Smart Home UI failed to start:', err)
}
