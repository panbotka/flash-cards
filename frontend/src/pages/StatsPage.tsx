import { useQuery } from '@tanstack/react-query'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Cell,
} from 'recharts'
import {
  getStatsSummary,
  getStatsHeatmap,
  getStatsAccuracy,
  getStatsMaturity,
  getStatsForecast,
  getStatsHardest,
} from '../api/client'
import type {
  HeatmapEntry,
  AccuracyEntry,
  ForecastEntry,
  HardestCard,
} from '../api/client'

// ---------------------------------------------------------------------------
// Section label
// ---------------------------------------------------------------------------
function SectionTitle({ children }: { children: string }) {
  return (
    <h2 className="text-xs uppercase tracking-wider text-[#6e6e73] mb-2">
      {children}
    </h2>
  )
}

// ---------------------------------------------------------------------------
// 1. Summary Cards
// ---------------------------------------------------------------------------
function accuracyColor(frac: number): string {
  const pct = frac > 1 ? frac : frac * 100
  if (pct >= 80) return '#30d158'
  if (pct >= 60) return '#ff9f0a'
  return '#ff453a'
}

function SummaryCards() {
  const { data, isLoading } = useQuery({
    queryKey: ['stats', 'summary'],
    queryFn: getStatsSummary,
  })

  const cards: { label: string; value: string; color?: string }[] = data
    ? [
        { label: 'Reviews Today', value: String(data.reviewsToday) },
        { label: 'Streak', value: `${data.streak} ${data.streak === 1 ? 'day' : 'days'}` },
        { label: 'Total Cards', value: String(data.totalCards) },
        {
          label: 'Accuracy Today',
          value: `${Math.round(data.accuracyToday * 100)}%`,
          color: accuracyColor(data.accuracyToday),
        },
      ]
    : []

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div
            key={i}
            className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] animate-pulse h-[76px]"
          />
        ))}
      </div>
    )
  }

  return (
    <div className="grid grid-cols-2 gap-3">
      {cards.map((c) => (
        <div
          key={c.label}
          className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a]"
        >
          <span className="block text-xs text-[#a1a1a6] mb-1">{c.label}</span>
          <span
            className="block text-2xl font-semibold"
            style={c.color ? { color: c.color } : undefined}
          >
            {c.value}
          </span>
        </div>
      ))}
    </div>
  )
}

// ---------------------------------------------------------------------------
// 2. Heatmap (GitHub-style, custom div grid)
// ---------------------------------------------------------------------------
const HEATMAP_COLORS = [
  '#1a1a1a',
  '#1a2a3e',
  '#1a3a5c',
  '#2a5a8c',
  '#3a7abc',
  '#5e9eff',
]

function heatmapLevel(count: number, max: number): number {
  if (count === 0 || max === 0) return 0
  const ratio = count / max
  if (ratio <= 0.15) return 1
  if (ratio <= 0.35) return 2
  if (ratio <= 0.55) return 3
  if (ratio <= 0.8) return 4
  return 5
}

interface HeatmapCell {
  date: string
  count: number
  level: number
}

function buildHeatmapGrid(entries: HeatmapEntry[]) {
  const countMap = new Map<string, number>()
  let max = 0
  for (const e of entries) {
    countMap.set(e.date, e.count)
    if (e.count > max) max = e.count
  }

  const today = new Date()
  const todayDay = today.getDay() // 0 = Sun
  const totalWeeks = 53

  // Start on the Sunday (52 * 7 + todayDay) days ago
  const startDate = new Date(today)
  startDate.setDate(startDate.getDate() - (52 * 7 + todayDay))

  const grid: HeatmapCell[][] = []

  for (let week = 0; week < totalWeeks; week++) {
    const col: HeatmapCell[] = []
    for (let day = 0; day < 7; day++) {
      const d = new Date(startDate)
      d.setDate(d.getDate() + week * 7 + day)
      if (d > today) {
        col.push({ date: '', count: 0, level: -1 })
      } else {
        const ds = d.toISOString().slice(0, 10)
        const count = countMap.get(ds) ?? 0
        col.push({ date: ds, count, level: heatmapLevel(count, max) })
      }
    }
    grid.push(col)
  }

  // Month labels based on the Monday (index 1) of each week
  const monthNames = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
  ]
  const months: { label: string; col: number }[] = []
  let lastMonth = -1
  for (let week = 0; week < totalWeeks; week++) {
    const cell = grid[week][1]
    if (!cell.date) continue
    const m = new Date(cell.date).getMonth()
    if (m !== lastMonth) {
      months.push({ label: monthNames[m], col: week })
      lastMonth = m
    }
  }

  return { grid, months }
}

function Heatmap() {
  const { data, isLoading } = useQuery({
    queryKey: ['stats', 'heatmap'],
    queryFn: getStatsHeatmap,
  })

  if (isLoading || !data) {
    return (
      <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] h-[140px] animate-pulse" />
    )
  }

  const { grid, months } = buildHeatmapGrid(data)
  const cellSize = 10
  const gap = 2

  return (
    <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] overflow-x-auto">
      {/* Month labels */}
      <div
        className="relative mb-1"
        style={{ height: 14, width: grid.length * (cellSize + gap) }}
      >
        {months.map((m, i) => (
          <span
            key={`${m.label}-${i}`}
            className="absolute text-[10px] text-[#6e6e73]"
            style={{ left: m.col * (cellSize + gap) }}
          >
            {m.label}
          </span>
        ))}
      </div>

      {/* Grid */}
      <div
        className="flex gap-[2px]"
        style={{ width: grid.length * (cellSize + gap) }}
      >
        {grid.map((week, wi) => (
          <div key={wi} className="flex flex-col gap-[2px]">
            {week.map((cell, di) => (
              <div
                key={`${wi}-${di}`}
                className="rounded-[2px]"
                style={{
                  width: cellSize,
                  height: cellSize,
                  backgroundColor:
                    cell.level === -1
                      ? 'transparent'
                      : HEATMAP_COLORS[cell.level],
                }}
                title={
                  cell.date ? `${cell.date}: ${cell.count} reviews` : ''
                }
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// 3. Accuracy Chart (Line)
// ---------------------------------------------------------------------------
interface AccuracyTooltipProps {
  active?: boolean
  payload?: { value: number; payload: AccuracyEntry }[]
  label?: string
}

function AccuracyTooltip({ active, payload, label }: AccuracyTooltipProps) {
  if (!active || !payload || payload.length === 0) return null
  const entry = payload[0].payload
  return (
    <div className="bg-[#1a1a1a] border border-[#2a2a2a] rounded-lg px-3 py-2 text-xs">
      <p className="text-[#a1a1a6] mb-1">{label}</p>
      <p className="text-white font-medium">{Math.round(entry.accuracy)}%</p>
      <p className="text-[#6e6e73]">{entry.total} reviews</p>
    </div>
  )
}

function AccuracyChart() {
  const { data, isLoading } = useQuery({
    queryKey: ['stats', 'accuracy'],
    queryFn: getStatsAccuracy,
  })

  if (isLoading || !data) {
    return (
      <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] h-[220px] animate-pulse" />
    )
  }

  const chartData = data.map((e) => ({
    ...e,
    date: e.date.slice(5), // "MM-DD"
    accuracy: e.accuracy * 100,
  }))

  return (
    <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a]">
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={chartData}>
          <XAxis
            dataKey="date"
            tick={{ fill: '#6e6e73', fontSize: 10 }}
            axisLine={false}
            tickLine={false}
            interval="preserveStartEnd"
          />
          <YAxis
            domain={[0, 100]}
            tick={{ fill: '#6e6e73', fontSize: 10 }}
            axisLine={false}
            tickLine={false}
            width={30}
            tickFormatter={(v: number) => `${v}%`}
          />
          <Tooltip content={<AccuracyTooltip />} />
          <Line
            type="monotone"
            dataKey="accuracy"
            stroke="#5e9eff"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, fill: '#5e9eff' }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

// ---------------------------------------------------------------------------
// 4. Maturity Distribution (Horizontal Bar)
// ---------------------------------------------------------------------------
const MATURITY_COLORS: Record<string, string> = {
  New: '#5e9eff',
  Learning: '#ff9f0a',
  Young: '#30d158',
  Mature: '#34c759',
}

interface MaturityTooltipProps {
  active?: boolean
  payload?: { value: number; payload: { name: string; count: number } }[]
}

function MaturityTooltip({ active, payload }: MaturityTooltipProps) {
  if (!active || !payload || payload.length === 0) return null
  const entry = payload[0].payload
  return (
    <div className="bg-[#1a1a1a] border border-[#2a2a2a] rounded-lg px-3 py-2 text-xs">
      <p className="text-white font-medium">
        {entry.name}: {entry.count}
      </p>
    </div>
  )
}

function MaturityChart() {
  const { data, isLoading } = useQuery({
    queryKey: ['stats', 'maturity'],
    queryFn: getStatsMaturity,
  })

  if (isLoading || !data) {
    return (
      <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] h-[200px] animate-pulse" />
    )
  }

  const chartData = [
    { name: 'New', count: data.new },
    { name: 'Learning', count: data.learning },
    { name: 'Young', count: data.young },
    { name: 'Mature', count: data.mature },
  ]

  return (
    <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a]">
      <ResponsiveContainer width="100%" height={180}>
        <BarChart data={chartData} layout="vertical" barCategoryGap="20%">
          <XAxis
            type="number"
            tick={{ fill: '#6e6e73', fontSize: 10 }}
            axisLine={false}
            tickLine={false}
          />
          <YAxis
            type="category"
            dataKey="name"
            tick={{ fill: '#a1a1a6', fontSize: 12 }}
            axisLine={false}
            tickLine={false}
            width={70}
          />
          <Tooltip content={<MaturityTooltip />} cursor={false} />
          <Bar dataKey="count" radius={[0, 4, 4, 0]} barSize={20}>
            {chartData.map((entry) => (
              <Cell key={entry.name} fill={MATURITY_COLORS[entry.name]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

// ---------------------------------------------------------------------------
// 5. Forecast (Vertical Bar)
// ---------------------------------------------------------------------------
interface ForecastTooltipProps {
  active?: boolean
  payload?: {
    value: number
    payload: ForecastEntry & { label: string }
  }[]
}

function ForecastTooltip({ active, payload }: ForecastTooltipProps) {
  if (!active || !payload || payload.length === 0) return null
  const entry = payload[0].payload
  return (
    <div className="bg-[#1a1a1a] border border-[#2a2a2a] rounded-lg px-3 py-2 text-xs">
      <p className="text-[#a1a1a6] mb-1">{entry.date}</p>
      <p className="text-white font-medium">{entry.count} reviews</p>
    </div>
  )
}

function ForecastChart() {
  const { data, isLoading } = useQuery({
    queryKey: ['stats', 'forecast'],
    queryFn: getStatsForecast,
  })

  if (isLoading || !data) {
    return (
      <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] h-[220px] animate-pulse" />
    )
  }

  const chartData = data.map((e) => ({
    ...e,
    label: e.date.slice(5), // "MM-DD"
  }))

  return (
    <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a]">
      <ResponsiveContainer width="100%" height={200}>
        <BarChart data={chartData}>
          <defs>
            <linearGradient
              id="forecastGradient"
              x1="0"
              y1="0"
              x2="0"
              y2="1"
            >
              <stop offset="0%" stopColor="#5e9eff" stopOpacity={1} />
              <stop offset="100%" stopColor="#5e9eff" stopOpacity={0.4} />
            </linearGradient>
          </defs>
          <XAxis
            dataKey="label"
            tick={{ fill: '#6e6e73', fontSize: 10 }}
            axisLine={false}
            tickLine={false}
            interval={4}
          />
          <YAxis
            tick={{ fill: '#6e6e73', fontSize: 10 }}
            axisLine={false}
            tickLine={false}
            width={30}
          />
          <Tooltip content={<ForecastTooltip />} cursor={false} />
          <Bar
            dataKey="count"
            fill="url(#forecastGradient)"
            radius={[3, 3, 0, 0]}
            barSize={12}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

// ---------------------------------------------------------------------------
// 6. Hardest Cards Table
// ---------------------------------------------------------------------------
function HardestCardsTable() {
  const { data, isLoading } = useQuery({
    queryKey: ['stats', 'hardest'],
    queryFn: getStatsHardest,
  })

  if (isLoading) {
    return (
      <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] h-[200px] animate-pulse" />
    )
  }

  if (!data || data.length === 0) {
    return (
      <div className="bg-[#1a1a1a] rounded-xl p-4 border border-[#2a2a2a] text-[#6e6e73] text-sm">
        No review data yet.
      </div>
    )
  }

  return (
    <div className="bg-[#1a1a1a] rounded-xl border border-[#2a2a2a] overflow-hidden">
      <div className="overflow-x-auto max-h-[360px] overflow-y-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[#2a2a2a] text-[#6e6e73] text-xs uppercase tracking-wider">
              <th className="text-left p-3 font-medium">Czech</th>
              <th className="text-left p-3 font-medium">English</th>
              <th className="text-right p-3 font-medium">Reviews</th>
              <th className="text-right p-3 font-medium">Accuracy</th>
            </tr>
          </thead>
          <tbody>
            {data.slice(0, 10).map((card: HardestCard) => (
              <tr
                key={card.cardId}
                className="border-b border-[#2a2a2a] last:border-b-0"
              >
                <td className="p-3 text-white">{card.czech}</td>
                <td className="p-3 text-[#a1a1a6]">{card.english}</td>
                <td className="p-3 text-right text-[#a1a1a6]">
                  {card.totalReviews}
                </td>
                <td
                  className="p-3 text-right font-medium"
                  style={{ color: accuracyColor(card.accuracy) }}
                >
                  {Math.round(card.accuracy * 100)}%
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------
export function StatsPage() {
  return (
    <div className="p-4 pb-20 max-w-2xl mx-auto flex flex-col gap-4">
      <section>
        <SectionTitle>Overview</SectionTitle>
        <SummaryCards />
      </section>

      <section>
        <SectionTitle>Activity</SectionTitle>
        <Heatmap />
      </section>

      <section>
        <SectionTitle>Accuracy (30 Days)</SectionTitle>
        <AccuracyChart />
      </section>

      <section>
        <SectionTitle>Card Maturity</SectionTitle>
        <MaturityChart />
      </section>

      <section>
        <SectionTitle>Forecast</SectionTitle>
        <ForecastChart />
      </section>

      <section>
        <SectionTitle>Hardest Cards</SectionTitle>
        <HardestCardsTable />
      </section>
    </div>
  )
}
