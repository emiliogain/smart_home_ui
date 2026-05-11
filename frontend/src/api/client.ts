import axios from 'axios'
import { BACKEND_URL } from '@/utils/constants'

/** REST client; base URL is `BACKEND_URL` + `/api` (dev default `http://localhost:8080/api`). */
export const apiClient = axios.create({
  baseURL: `${BACKEND_URL.replace(/\/$/, '')}/api`,
})
