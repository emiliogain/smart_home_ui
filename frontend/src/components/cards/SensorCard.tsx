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
  /** Stable id (UUID); not shown unless no label. */
  sensorId: string
  /** Primary title shown above the value (human-friendly). */
  sensorLabel?: string
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

function formatChartTimeLabel(timestamp: string | undefined, index: number) {
  if (!timestamp) return String(index)
  const d = new Date(timestamp)
  if (Number.isNaN(d.getTime())) return String(index)
  return d.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  })
}

function axisTickStyle() {
  return {
    fontSize: 10,
    fill: 'var(--color-text-secondary)',
  } as const
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
  sensorLabel,
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

  const title = sensorLabel?.trim() || sensorId

  const chartData = (history ?? []).slice(-20).map((p, i) => ({
    i,
    value: typeof p.value === 'number' ? p.value : p.value ? 1 : 0,
    timestamp: p.timestamp,
  }))

  const displayPrimary =
    chartVariant === 'motion' ? motionHeadline(value, history) : formatValue(value)

  const n = chartData.length
  const xAxisTicks =
    n <= 0 ? [] : n === 1 ? [0] : n === 2 ? [0, 1] : [0, Math.floor((n - 1) / 2), n - 1]

  const axisStroke = 'rgba(255,255,255,0.12)'
  const chartMargins =
    chartVariant === 'motion'
      ? { top: 6, right: 8, left: 30, bottom: 24 }
      : { top: 8, right: 10, left: 42, bottom: 30 }
  const chartHeightClass = chartVariant === 'motion' ? 'h-[92px]' : 'h-[132px]'

  return (
    <div className={clsx(cardSurface, 'shadow-black/25')}>
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white/10 text-[var(--color-secondary)]">
          {typeIcon(type)}
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm text-[var(--color-text-secondary)]">{title}</p>
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
        <div className={clsx('mt-3 w-full', chartHeightClass)}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={chartMargins}>
              <XAxis
                dataKey="i"
                type="number"
                domain={['dataMin', 'dataMax']}
                ticks={xAxisTicks}
                tick={axisTickStyle()}
                tickLine={{ stroke: axisStroke }}
                axisLine={{ stroke: axisStroke }}
                tickFormatter={(tickVal: number) => {
                  const idx = Math.round(Number(tickVal))
                  return formatChartTimeLabel(chartData[idx]?.timestamp, idx)
                }}
              />
              <YAxis
                width={chartVariant === 'motion' ? 28 : 38}
                domain={
                  chartVariant === 'motion' ? [-0.2, 1.2] : ['dataMin - 1', 'dataMax + 1']
                }
                ticks={chartVariant === 'motion' ? [0, 1] : undefined}
                tickCount={chartVariant === 'motion' ? undefined : 5}
                tick={axisTickStyle()}
                tickLine={{ stroke: axisStroke }}
                axisLine={{ stroke: axisStroke }}
                tickFormatter={(v: number) => {
                  if (chartVariant === 'motion') {
                    return String(Math.round(v))
                  }
                  return Number.isFinite(v) ? v.toFixed(1) : ''
                }}
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
