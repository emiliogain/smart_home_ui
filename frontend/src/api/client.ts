import axios from 'axios'
import { BACKEND_URL } from '@/utils/constants'

/** REST base: `/api` on the backend (proxied in Vite dev when `BACKEND_URL` is empty). */
export const apiClient = axios.create({
  baseURL: `${BACKEND_URL.replace(/\/$/, '')}/api`,
})
