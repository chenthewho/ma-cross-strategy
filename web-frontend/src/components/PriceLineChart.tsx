import { useRef, useState, useCallback, useMemo } from 'react'
import type { KlinePoint, TradeMarker } from '@/shared/services/dashboard'

export interface PriceChartData {
  klines: KlinePoint[]
  trades: TradeMarker[]
  avg_buy_price: number
}

const ZOOM_PRESETS = [
  { label: '1天', hours: 24 },
  { label: '3天', hours: 72 },
  { label: '1周', hours: 168 },
  { label: '2周', hours: 336 },
  { label: '全部', hours: 0 },
] as const

export default function PriceLineChart({ data }: { data: PriceChartData }) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [tooltip, setTooltip] = useState<{ idx: number; x: number; y: number } | null>(null)
  const [zoomHours, setZoomHours] = useState(0) // 0 = all
  const pinchRef = useRef<{ dist: number; count: number } | null>(null)

  const W = 800
  const H = 280
  const PAD = { top: 24, right: 20, bottom: 40, left: 70 }
  const innerW = W - PAD.left - PAD.right
  const innerH = H - PAD.top - PAD.bottom

  const { klines: allKlines, trades, avg_buy_price } = data

  // Filter klines by zoom range (from end)
  const klines = useMemo(() => {
    if (zoomHours <= 0) return allKlines
    const cutoff = Date.now() - zoomHours * 3600000
    return allKlines.filter(k => k.open_time >= cutoff)
  }, [allKlines, zoomHours])

  if (allKlines.length < 2) {
    return <div className="text-center text-sm text-claude-text-muted py-8">暂无价格数据</div>
  }

  const closes = klines.map(k => k.close)
  const maxV = Math.max(...closes)
  const minV = Math.min(...closes)
  const range = (maxV - minV) || 1

  const xScale = (i: number) => PAD.left + (i / Math.max(klines.length - 1, 1)) * innerW
  const yScale = (v: number) => PAD.top + innerH - ((v - minV) / range) * innerH

  // Price line
  const points = klines.map((k, i) => `${xScale(i)},${yScale(k.close)}`).join(' ')
  const areaPath = `M${xScale(0)},${PAD.top + innerH} L${points} L${xScale(klines.length - 1)},${PAD.top + innerH} Z`

  // Y-axis ticks
  const yTicks = 5
  const yTickValues = Array.from({ length: yTicks + 1 }, (_, i) => minV + (range * i) / yTicks)

  // X-axis time labels
  const xLabelCount = Math.min(4, klines.length)
  const xLabelIndices = Array.from({ length: xLabelCount }, (_, i) =>
    Math.round((i * Math.max(klines.length - 1, 0)) / Math.max(xLabelCount - 1, 1))
  )

  // Map trade time to nearest kline index
  const getKlineIdx = (time: string) => {
    const ts = new Date(time).getTime()
    let best = 0
    let bestDiff = Infinity
    for (let i = 0; i < klines.length; i++) {
      const diff = Math.abs(klines[i].open_time - ts)
      if (diff < bestDiff) { bestDiff = diff; best = i }
    }
    return best
  }

  // Filter trades to visible range
  const visibleTrades = useMemo(() => {
    if (klines.length === 0) return []
    const minTime = klines[0].open_time
    const maxTime = klines[klines.length - 1].open_time + 3600000
    return trades.filter(t => {
      const ts = new Date(t.created_at).getTime()
      return ts >= minTime && ts <= maxTime
    })
  }, [trades, klines])

  const buys = visibleTrades.filter(t => t.action === 'BUY')
  const sells = visibleTrades.filter(t => t.action === 'SELL')

  // Tooltip on pointer move
  const handlePointer = useCallback((e: React.PointerEvent) => {
    const svg = svgRef.current
    if (!svg) return
    const rect = svg.getBoundingClientRect()
    const scaleX = W / rect.width
    const mx = (e.clientX - rect.left) * scaleX
    const idx = Math.round(((mx - PAD.left) / innerW) * Math.max(klines.length - 1, 0))
    const i = Math.max(0, Math.min(klines.length - 1, idx))
    setTooltip({ idx: i, x: xScale(i), y: yScale(klines[i].close) })
  }, [klines])

  // Pinch-to-zoom
  const getPinchDist = (e: React.TouchEvent) => {
    if (e.touches.length < 2) return 0
    const dx = e.touches[0].clientX - e.touches[1].clientX
    const dy = e.touches[0].clientY - e.touches[1].clientY
    return Math.sqrt(dx * dx + dy * dy)
  }

  const handleTouchStart = (e: React.TouchEvent) => {
    if (e.touches.length === 2) {
      pinchRef.current = { dist: getPinchDist(e), count: zoomHours || 9999 }
    }
  }

  const handleTouchMove = (e: React.TouchEvent) => {
    if (e.touches.length !== 2 || !pinchRef.current) return
    const newDist = getPinchDist(e)
    const ratio = pinchRef.current.dist / Math.max(newDist, 1)
    const newCount = Math.round(pinchRef.current.count * ratio)
    // Map hours inversely: larger count → fewer visible hours
    const newHours = Math.max(4, Math.min(9999, newCount))
    if (newHours > 10000) {
      setZoomHours(0) // ALL
    } else if (newHours > 500) {
      setZoomHours(Math.round(newHours / 24) * 24)
    } else {
      setZoomHours(Math.round(newHours / 4) * 4)
    }
  }

  const handleTouchEnd = () => {
    pinchRef.current = null
  }

  // Mouse wheel zoom
  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault()
    const current = zoomHours || 9999
    const factor = e.deltaY > 0 ? 1.3 : 0.7
    const newHours = Math.round(current * factor)
    if (newHours > 10000) {
      setZoomHours(0)
    } else if (newHours <= 4) {
      setZoomHours(4)
    } else {
      setZoomHours(newHours)
    }
  }, [zoomHours])

  // Average buy price y-position
  const avgY = avg_buy_price > 0 ? yScale(avg_buy_price) : null
  const avgInRange = avgY !== null && avgY >= PAD.top && avgY <= PAD.top + innerH

  return (
    <div className="relative select-none touch-none">
      {/* Zoom preset buttons */}
      <div className="flex gap-1 mb-2">
        {ZOOM_PRESETS.map(p => (
          <button
            key={p.label}
            onClick={() => setZoomHours(p.hours)}
            className={`px-2 py-0.5 text-[12px] rounded border transition-colors ${
              zoomHours === p.hours
                ? 'bg-claude-accent text-white border-claude-accent'
                : 'bg-white text-claude-text-secondary border-claude-border hover:border-claude-accent'
            }`}
          >
            {p.label}
          </button>
        ))}
      </div>

      <svg
        ref={svgRef}
        viewBox={`0 0 ${W} ${H}`}
        className="w-full h-56 lg:h-72"
        onPointerMove={handlePointer}
        onPointerDown={handlePointer}
        onPointerLeave={() => setTooltip(null)}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onWheel={handleWheel}
      >
        {/* Grid lines */}
        {yTickValues.map(v => (
          <g key={v}>
            <line x1={PAD.left} y1={yScale(v)} x2={W - PAD.right} y2={yScale(v)}
              stroke="#e5e0d8" strokeWidth="0.5" />
            <text x={PAD.left - 8} y={yScale(v) + 4} textAnchor="end"
              className="text-[11px] fill-[#9ca3af]">
              ${v.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
          </g>
        ))}

        {/* Area fill */}
        <path d={areaPath} fill="url(#priceGrad)" opacity="0.15" />
        <defs>
          <linearGradient id="priceGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#d97706" />
            <stop offset="100%" stopColor="#d97706" stopOpacity="0" />
          </linearGradient>
        </defs>

        {/* Price line */}
        <polyline points={points} fill="none" stroke="#d97706" strokeWidth="2"
          strokeLinejoin="round" strokeLinecap="round" />
        {/* Buy markers — positioned exactly on price line */}
        {buys.map((b, i) => {
          const idx = getKlineIdx(b.created_at)
          const cx = xScale(idx)
          const cy = yScale(klines[idx].close) // on the line
          if (cy < PAD.top + 6 || cy > PAD.top + innerH - 6) return null
          return (
            <g key={`buy-${i}`}>
              <polygon
                points={`${cx},${cy + 10} ${cx - 4},${cy + 2} ${cx + 4},${cy + 2}`}
                fill="#22c55e" stroke="white" strokeWidth="0.8"
              />
            </g>
          )
        })}

        {/* Sell markers — positioned exactly on price line */}
        {sells.map((s, i) => {
          const idx = getKlineIdx(s.created_at)
          const cx = xScale(idx)
          const cy = yScale(klines[idx].close) // on the line
          if (cy < PAD.top + 6 || cy > PAD.top + innerH - 6) return null
          return (
            <g key={`sell-${i}`}>
              <polygon
                points={`${cx},${cy - 10} ${cx - 4},${cy - 2} ${cx + 4},${cy - 2}`}
                fill="#ef4444" stroke="white" strokeWidth="0.8"
              />
            </g>
          )
        })}

        {/* Average buy price line */}
        {avgInRange && (
          <g>
            <line x1={PAD.left} y1={avgY} x2={W - PAD.right} y2={avgY}
              stroke="#22c55e" strokeWidth="1" strokeDasharray="6 3" opacity="0.7" />
            <text x={W - PAD.right} y={avgY - 6} textAnchor="end"
              className="text-[11px] fill-[#22c55e] font-medium">
              均价 ${avg_buy_price.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
          </g>
        )}

        {/* Tooltip */}
        {tooltip && klines[tooltip.idx] && (
          <g>
            <line x1={tooltip.x} y1={PAD.top} x2={tooltip.x} y2={PAD.top + innerH}
              stroke="#d97706" strokeWidth="1" strokeDasharray="3 2" opacity="0.6" />
            <circle cx={tooltip.x} cy={tooltip.y} r="4" fill="white" stroke="#d97706" strokeWidth="2" />
            <rect
              x={tooltip.x > W / 2 ? tooltip.x - 140 : tooltip.x + 10}
              y={PAD.top + 2}
              width="130" height="42" rx="6"
              fill="white" stroke="#e5e0d8" strokeWidth="1"
            />
            <text
              x={tooltip.x > W / 2 ? tooltip.x - 75 : tooltip.x + 75}
              y={PAD.top + 18}
              textAnchor="middle" className="text-[12px] font-semibold fill-[#d97706]"
            >
              ${klines[tooltip.idx].close.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
            <text
              x={tooltip.x > W / 2 ? tooltip.x - 75 : tooltip.x + 75}
              y={PAD.top + 35}
              textAnchor="middle" className="text-[11px] fill-[#9ca3af]"
            >
              {new Date(klines[tooltip.idx].open_time).toLocaleString('zh-CN', {
                month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit'
              })}
            </text>
          </g>
        )}

        {/* X-axis time labels */}
        {xLabelIndices.map(i => (
          <text key={i} x={xScale(i)} y={H - 8} textAnchor="middle"
            className="text-[12px] fill-[#9ca3af]">
            {new Date(klines[i].open_time).toLocaleString('zh-CN', {
              month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit'
            })}
          </text>
        ))}

        {/* Legend */}
        <g transform={`translate(${PAD.left + 8}, ${PAD.top - 10})`}>
          <line x1={0} y1={0} x2={16} y2={0} stroke="#d97706" strokeWidth="2" />
          <text x={20} y={3.5} className="text-[11px] fill-[#9ca3af]">价格</text>
          <polygon points="28,10 24,2 32,2" fill="#22c55e" stroke="white" strokeWidth="0.5" />
          <text x={36} y={3.5} className="text-[11px] fill-[#9ca3af]">买入</text>
          <polygon points="48,-2 44,6 52,6" fill="#ef4444" stroke="white" strokeWidth="0.5" />
          <text x={56} y={3.5} className="text-[11px] fill-[#9ca3af]">卖出</text>
        </g>
      </svg>

      {/* Zoom hint */}
      <p className="text-[11px] text-claude-text-muted text-center mt-1">
        双指缩放 · 滚轮缩放 · 手指滑动查看价格
      </p>
    </div>
  )
}
