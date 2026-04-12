/** Two decimal places for numeric sensor readings shown in the UI */
export function formatSensorScalar(value: number): string {
  if (!Number.isFinite(value)) return '—'
  return value.toFixed(2)
}
