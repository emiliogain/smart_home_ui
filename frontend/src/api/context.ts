import type { ContextUpdate } from '@/types/context'
import { apiClient } from './client'

export async function fetchCurrentContext(): Promise<ContextUpdate | null> {
  try {
    const { data } = await apiClient.get<ContextUpdate>('/context/current')
    return data
  } catch (err) {
    console.error(err)
    return null
  }
}
