import type { SensorHistoryPoint } from '@/components/cards/SensorCard'

/**
 * Generates synthetic history points with small random drift around `current`.
 */
export function generateMockHistory(
  current: number,
  points: number,
): SensorHistoryPoint[] {
  const out: SensorHistoryPoint[] = []
  const now = Date.now()
  const stepMs = 3_600_000

  for (let i = 0; i < points; i++) {
    const jitter = (Math.random() - 0.5) * Math.max(0.5, Math.abs(current) * 0.04 + 0.5)
    const value = Math.round((current + jitter) * 10) / 10
    out.push({
      value,
      timestamp: new Date(now - (points - 1 - i) * stepMs).toISOString(),
    })
  }

  return out
}

/** Binary motion-style timeline for dot charts (0 = no motion, 1 = event). */
export function generateMockMotionHistory(points: number): SensorHistoryPoint[] {
  const out: SensorHistoryPoint[] = []
  const now = Date.now()
  const stepMs = 600_000

  for (let i = 0; i < points; i++) {
    out.push({
      value: Math.random() > 0.82 ? 1 : 0,
      timestamp: new Date(now - (points - 1 - i) * stepMs).toISOString(),
    })
  }

  return out
}
