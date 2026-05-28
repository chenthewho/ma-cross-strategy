import { useRef, useState, useCallback } from 'react'
import type { KlinePoint, TradeMarker } from '@/shared/services/dashboard'

export interface PriceChartData {
  klines: KlinePoint[]
  trades: TradeMarker[]
  avg_buy_price: number
}

export default function PriceLineChart({ data }: { data: PriceChartData }) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [tooltip, setTooltip] = useState<{ idx: number; x: number; y: number } | null>(null)

  const W = 800
  const H = 280
  const PAD = { top: 24, right: 20, bottom: 40, left: 70 }
  const innerW = W - PAD.left - PAD.right
  const innerH = H - PAD.top - PAD.bottom

  const { klines, trades, avg_buy_price } = data

  if (klines.length < 2) {
    return <div className="text-center text-sm text-claude-text-muted py-8">暂无价格数据</div>
  }

  const closes = klines.map(k => k.close)
  const maxV = Math.max(...closes)
  const minV = Math.min(...closes)
  const range = (maxV - minV) || 1

  const xScale = (i: number) => PAD.left + (i / (klines.length - 1)) * innerW
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
    Math.round((i * (klines.length - 1)) / (xLabelCount - 1))
  )

  // Map trade time to nearest kline index for marker placement
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

  const buys = trades.filter(t => t.action === 'BUY')
  const sells = trades.filter(t => t.action === 'SELL')

  const handlePointer = useCallback((e: React.PointerEvent) => {
    const svg = svgRef.current
    if (!svg) return
    const rect = svg.getBoundingClientRect()
    const scaleX = W / rect.width
    const mx = (e.clientX - rect.left) * scaleX
    const idx = Math.round(((mx - PAD.left) / innerW) * (klines.length - 1))
    const i = Math.max(0, Math.min(klines.length - 1, idx))
    setTooltip({ idx: i, x: xScale(i), y: yScale(klines[i].close) })
  }, [klines])

  // Average buy price y-position
  const avgY = avg_buy_price > 0 ? yScale(avg_buy_price) : null
  const avgInRange = avgY !== null && avgY >= PAD.top && avgY <= PAD.top + innerH

  return (
    <div className="relative select-none touch-none">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${W} ${H}`}
        className="w-full h-56 lg:h-72"
        onPointerMove={handlePointer}
        onPointerDown={handlePointer}
        onPointerLeave={() => setTooltip(null)}
      >
        {/* Grid lines */}
        {yTickValues.map(v => (
          <g key={v}>
            <line x1={PAD.left} y1={yScale(v)} x2={W - PAD.right} y2={yScale(v)}
              stroke="#e5e0d8" strokeWidth="0.5" />
            <text x={PAD.left - 8} y={yScale(v) + 4} textAnchor="end"
              className="text-[9px] fill-[#9ca3af]">
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

        {/* Buy markers (green up triangle) */}
        {buys.map((b, i) => {
          const idx = getKlineIdx(b.created_at)
          const cx = xScale(idx)
          const cy = yScale(b.price) - 3
          // Constrain to chart area
          if (cy < PAD.top || cy > PAD.top + innerH) return null
          return (
            <g key={`buy-${i}`}>
              <polygon
                points={`${cx},${cy - 8} ${cx - 5},${cy} ${cx + 5},${cy}`}
                fill="#22c55e" opacity="0.85"
              />
            </g>
          )
        })}

        {/* Sell markers (red down triangle) */}
        {sells.map((s, i) => {
          const idx = getKlineIdx(s.created_at)
          const cx = xScale(idx)
          const cy = yScale(s.price) + 3
          if (cy < PAD.top || cy > PAD.top + innerH) return null
          return (
            <g key={`sell-${i}`}>
              <polygon
                points={`${cx},${cy + 8} ${cx - 5},${cy} ${cx + 5},${cy}`}
                fill="#ef4444" opacity="0.85"
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
              className="text-[9px] fill-[#22c55e] font-medium">
              均价 ${avg_buy_price.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
          </g>
        )}

        {/* Tooltip */}
        {tooltip && (
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
              textAnchor="middle" className="text-[10px] font-semibold fill-[#d97706]"
            >
              ${klines[tooltip.idx].close.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
            <text
              x={tooltip.x > W / 2 ? tooltip.x - 75 : tooltip.x + 75}
              y={PAD.top + 35}
              textAnchor="middle" className="text-[9px] fill-[#9ca3af]"
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
            className="text-[8px] fill-[#9ca3af]">
            {new Date(klines[i].open_time).toLocaleString('zh-CN', {
              month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit'
            })}
          </text>
        ))}

        {/* Legend */}
        <g transform={`translate(${PAD.left + 8}, ${PAD.top - 10})`}>
          <line x1={0} y1={0} x2={16} y2={0} stroke="#d97706" strokeWidth="2" />
          <text x={20} y={3.5} className="text-[9px] fill-[#9ca3af]">价格</text>
          <polygon points="28,-5 24,2 32,2" fill="#22c55e" />
          <text x={36} y={3.5} className="text-[9px] fill-[#9ca3af]">买入</text>
          <polygon points="48,5 44,-2 52,-2" fill="#ef4444" />
          <text x={56} y={3.5} className="text-[9px] fill-[#9ca3af]">卖出</text>
        </g>
      </svg>
    </div>
  )
}
