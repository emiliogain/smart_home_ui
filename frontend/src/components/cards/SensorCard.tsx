import { Activity, Droplets, Sun, Thermometer } from 'lucide-react'
import { clsx } from 'clsx'
import type { DotProps } from 'recharts'
import { Line, LineChart, ResponsiveContainer, XAxis, YAxis } from 'recharts'
import { ROOM_LABELS } from '@/utils/constants'
import { formatSensorScalar } from '@/utils/formatSensorValue'
import { cardSurface } from '@/utils/uiClasses'

export interface SensorHistoryPoint {
  value: number
  timestamp: string
}

interface SensorCardProps {
  sensorId: string
  type: string
  value: number | boolean
  unit: string
  room: string
  history?: SensorHistoryPoint[]
  chartVariant?: 'line' | 'motion'
}

function typeIcon(sensorType: string) {
  const t = sensorType.toLowerCase()
  if (t.includes('temp')) {
    return <Thermometer className="h-5 w-5" strokeWidth={2} />
  }
  if (t.includes('humid')) {
    return <Droplets className="h-5 w-5" strokeWidth={2} />
  }
  if (t.includes('light') || t.includes('lux') || t.includes('bright')) {
    return <Sun className="h-5 w-5" strokeWidth={2} />
  }
  if (t.includes('motion')) {
    return <Activity className="h-5 w-5" strokeWidth={2} />
  }
  return <Activity className="h-5 w-5" strokeWidth={2} />
}

function formatValue(value: number | boolean): string {
  if (typeof value === 'boolean') {
    return value ? 'On' : 'Off'
  }
  return formatSensorScalar(value)
}

function motionHeadline(
  value: number | boolean,
  history: SensorHistoryPoint[] | undefined,
): string {
  if (history?.length) {
    const n = history.slice(-20).filter((h) => h.value >= 1).length
    return `${n} events (window)`
  }
  if (typeof value === 'number') {
    return value >= 1 ? 'Motion' : 'Clear'
  }
  return formatValue(value)
}

function MotionEventDot(
  props: DotProps & { payload?: { value: number; timestamp?: string } },
) {
  const { cx, cy, payload } = props
  if (payload == null || payload.value < 1 || cx == null || cy == null) {
    return null
  }
  return (
    <circle
      cx={cx}
      cy={cy}
      r={5}
      fill="var(--color-secondary)"
      stroke="var(--color-bg)"
      strokeWidth={1}
    />
  )
}

export function SensorCard({
  sensorId,
  type,
  value,
  unit,
  room,
  history,
  chartVariant = 'line',
}: SensorCardProps) {
  const roomLabel =
    room in ROOM_LABELS
      ? ROOM_LABELS[room as keyof typeof ROOM_LABELS]
      : room

  const chartData = (history ?? []).slice(-20).map((p, i) => ({
    i,
    value: typeof p.value === 'number' ? p.value : p.value ? 1 : 0,
    timestamp: p.timestamp,
  }))

  const displayPrimary =
    chartVariant === 'motion' ? motionHeadline(value, history) : formatValue(value)

  return (
    <div className={clsx(cardSurface, 'shadow-black/25')}>
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white/10 text-[var(--color-secondary)]">
          {typeIcon(type)}
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm text-[var(--color-text-secondary)]">{sensorId}</p>
          <div className="mt-1 flex flex-wrap items-baseline gap-x-1.5 gap-y-0">
            <span className="text-2xl font-semibold tabular-nums text-[var(--color-text-primary)]">
              {displayPrimary}
            </span>
            {unit && chartVariant !== 'motion' ? (
              <span className="text-sm text-[var(--color-text-secondary)]">
                {unit}
              </span>
            ) : null}
          </div>
        </div>
      </div>

      {chartData.length > 0 ? (
        <div className="mt-3 h-[60px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 4, right: 4, left: 4, bottom: 4 }}>
              <XAxis dataKey="i" hide />
              <YAxis
                hide
                domain={
                  chartVariant === 'motion' ? [-0.2, 1.2] : ['dataMin - 1', 'dataMax + 1']
                }
              />
              <Line
                type={chartVariant === 'motion' ? 'stepAfter' : 'monotone'}
                dataKey="value"
                stroke="var(--color-secondary)"
                strokeOpacity={chartVariant === 'motion' ? 0.35 : 1}
                strokeWidth={chartVariant === 'motion' ? 1.5 : 2}
                dot={
                  chartVariant === 'motion'
                    ? (dotProps) => <MotionEventDot {...dotProps} />
                    : false
                }
                isAnimationActive={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      ) : null}

      <p className="mt-2 text-sm text-[var(--color-text-secondary)]">{roomLabel}</p>
    </div>
  )
}
