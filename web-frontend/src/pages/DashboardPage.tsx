import { useEffect, useState, useRef, useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Play, Pause, RefreshCw, ChevronDown, ChevronUp } from 'lucide-react'
import Card from '@/components/Card'
import StatusBadge from '@/components/StatusBadge'
import PnLChartSkeleton from '@/components/skeletons/PnLChartSkeleton'
import PriceLineChart from '@/components/PriceLineChart'
import { fetchInstances, updateInstanceStatus, fetchTrades, type Instance, type TradeRecord } from '@/shared/services/instances'
import { fetchEquitySnapshots, fetchPriceChart } from '@/shared/services/dashboard'
import { useI18n } from '@/i18n/I18nProvider'

export default function DashboardPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [expandedTradeId, setExpandedTradeId] = useState<number | null>(null)

  const { data: instances = [], refetch } = useQuery({
    queryKey: ['instances'],
    queryFn: fetchInstances,
    refetchInterval: 60000,
  })

  const { data: snapshots = [], isLoading: chartLoading } = useQuery({
    queryKey: ['equity', selectedId],
    queryFn: () => selectedId ? fetchEquitySnapshots(selectedId) : Promise.resolve([]),
    enabled: !!selectedId,
    refetchInterval: 60000,
  })

  const { data: trades = [] } = useQuery({
    queryKey: ['trades', selectedId],
    queryFn: () => selectedId ? fetchTrades(selectedId) : Promise.resolve([]),
    enabled: !!selectedId,
    refetchInterval: 60000,
  })

  const { data: priceChart } = useQuery({
    queryKey: ['priceChart', selectedId],
    queryFn: () => selectedId ? fetchPriceChart(selectedId) : Promise.resolve(null),
    enabled: !!selectedId,
    refetchInterval: 60000,
  })

  useEffect(() => {
    const idParam = searchParams.get('instance')
    const safeInstances = instances || []
    if (idParam && !selectedId) {
      const id = parseInt(idParam)
      if (!isNaN(id) && safeInstances.some((i) => i.id === id)) setSelectedId(id)
    }
    if (!idParam && safeInstances.length > 0 && !selectedId) setSelectedId(safeInstances[0].id)
  }, [instances, searchParams])

  const selectInstance = (id: number) => { setSelectedId(id); setSearchParams({ instance: String(id) }) }

  const toggleStatus = async (inst: Instance) => {
    const isRunning = (inst.status || '').toLowerCase() === 'running'
    const newStatus = isRunning ? 'stopped' : 'running'
    await updateInstanceStatus(inst.id, newStatus)
    refetch()
  }

  const queryClient = useQueryClient()
  const [refreshing, setRefreshing] = useState(false)
  const refreshAll = async () => {
    setRefreshing(true)
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['instances'] }),
      queryClient.invalidateQueries({ queryKey: ['equity', selectedId] }),
      queryClient.invalidateQueries({ queryKey: ['trades', selectedId] }),
    ])
    refetch()
    setTimeout(() => setRefreshing(false), 600)
  }

  const selected = selectedId ? instances.find((i) => i.id === selectedId) : null

  return (
    <div className="space-y-4 lg:space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg lg:text-xl font-semibold text-claude-text">{t('dashboard.title')}</h2>
        <button onClick={() => navigate('/instances/new')} className="flex items-center gap-1.5 lg:gap-2 px-3 lg:px-4 py-2 bg-claude-accent text-white rounded-lg text-xs lg:text-sm font-medium hover:bg-claude-accent-hover transition-colors">
          <Plus className="w-3.5 h-3.5 lg:w-4 lg:h-4" />{t('dashboard.newInstance')}
        </button>
      </div>

      {/* Instance selector — horizontal scroll on mobile */}
      <div className="lg:hidden -mx-3 sm:-mx-4">
        <div className="flex gap-2 overflow-x-auto px-3 sm:px-4 pb-2 custom-scrollbar snap-x">
          {instances.map((inst) => (
            <button
              key={inst.id}
              onClick={() => selectInstance(inst.id)}
              className={`shrink-0 snap-start flex items-center gap-2 px-3 py-2 rounded-lg border text-xs whitespace-nowrap transition-colors ${
                selectedId === inst.id
                  ? 'border-claude-accent bg-claude-accent-light text-claude-accent'
                  : 'border-claude-border text-claude-text-secondary hover:border-claude-border-hover'
              }`}
            >
              <StatusBadge status={inst.status} />
              {inst.name}
            </button>
          ))}
          {instances.length === 0 && (
            <p className="text-xs text-claude-text-muted py-2">暂无实例</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 lg:gap-6">
        {/* Instance list sidebar — hidden on mobile */}
        <div className="hidden lg:block lg:col-span-1 space-y-2">
          {instances.map((inst) => (
            <Card
              key={inst.id}
              onClick={() => selectInstance(inst.id)}
              className={`p-4 cursor-pointer transition-all ${selectedId === inst.id ? 'border-l-[3px] border-l-claude-accent bg-claude-accent-light/40' : ''}`}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="font-medium text-sm text-claude-text">{inst.name}</span>
                <StatusBadge status={inst.status} />
              </div>
              <div className="text-xs text-claude-text-muted font-mono">{inst.symbol} · {inst.template_id}</div>
              <button
                onClick={(e) => { e.stopPropagation(); toggleStatus(inst) }}
                className="mt-2 flex items-center gap-1 text-xs text-claude-text-secondary hover:text-claude-accent transition-colors"
              >
                {(inst.status || '').toLowerCase() === 'running' ? <Pause className="w-3 h-3" /> : <Play className="w-3 h-3" />}
                {(inst.status || '').toLowerCase() === 'running' ? t('common.stop') : t('common.start')}
              </button>
            </Card>
          ))}
          {instances.length === 0 && (
            <p className="text-sm text-claude-text-muted text-center py-8">暂无实例</p>
          )}
        </div>

        {/* Main display area */}
        <div className="lg:col-span-3 space-y-4 lg:space-y-6">
          {selected ? (
            <>
              <Card className="p-4 lg:p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-base lg:text-lg font-semibold text-claude-text">{selected.name}</h3>
                  <div className="flex items-center gap-3">
                    <button
                      onClick={refreshAll}
                      disabled={refreshing}
                      className="p-1.5 rounded-lg text-claude-text-secondary hover:text-claude-accent hover:bg-claude-accent-light transition-colors"
                      title="刷新数据"
                    >
                      <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
                    </button>
                    <StatusBadge status={selected.status} />
                    <button
                      onClick={() => toggleStatus(selected)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                        (selected.status || '').toLowerCase() === 'running'
                          ? 'bg-claude-danger-light border border-claude-danger/20 text-claude-danger hover:bg-claude-danger/10'
                          : 'bg-claude-accent text-white hover:bg-claude-accent-hover'
                      }`}
                    >
                      {(selected.status || '').toLowerCase() === 'running' ? <><Pause className="w-3 h-3 inline mr-1" />{t('common.stop')}</> : <><Play className="w-3 h-3 inline mr-1" />{t('common.start')}</>}
                    </button>
                  </div>
                </div>
                <div className="grid grid-cols-2 lg:grid-cols-5 gap-3 lg:gap-4">
                  <StatItem label={t('dashboard.totalEquity')} value={formatCNY(selected.total_equity ?? selected.cny_balance ?? 0)} />
                  <StatItem label="底仓 Dead" value={formatCNY(selected.dead_hold ?? 0)} colorClass="text-amber-600" />
                  <StatItem label="浮动仓 Float" value={formatCNY((selected as any).float_hold ?? 0)} colorClass="text-blue-600" />
                  <StatItem label="冷封仓 Cold" value={formatCNY((selected as any).cold_sealed_hold ?? 0)} colorClass="text-slate-500" />
                  <StatItem label="可用现金" value={formatCNY(selected.cny_balance ?? 0)} />
                </div>
                {/* 仓位分布条 */}
                <PositionBar
                  dead={(selected as any).dead_hold ?? 0}
                  floating={(selected as any).float_hold ?? 0}
                  cold={(selected as any).cold_sealed_hold ?? 0}
                  cash={selected.cny_balance ?? 0}
                  className="mt-3"
                />
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4 mt-3 pt-3 border-t border-claude-border">
                  <StatItem label="总成本" value={formatCNY((selected as any).total_cost ?? selected.initial_capital ?? 0)}
                    desc={`初始 ${formatCNY(selected.initial_capital ?? 0)} + 定投 ${formatCNY((selected as any).cumulative_injected ?? 0)}`} />
                  <StatItem label="盈利金额" value={formatPnL(selected.total_equity ?? 0, (selected as any).total_cost ?? selected.initial_capital ?? 0)}
                    colorClass={(selected.total_equity ?? 0) >= ((selected as any).total_cost ?? selected.initial_capital ?? 0) ? 'text-claude-success' : 'text-claude-danger'} />
                  <StatItem label="盈亏百分比" value={formatPnLPct(selected.total_equity ?? 0, (selected as any).total_cost ?? selected.initial_capital ?? 0)}
                    colorClass={(selected.total_equity ?? 0) >= ((selected as any).total_cost ?? selected.initial_capital ?? 0) ? 'text-claude-success' : 'text-claude-danger'} />
                  <StatItem label="已实现盈亏" value={formatPnL((selected as any).realized_pnl ?? 0, 0)}
                    colorClass={((selected as any).realized_pnl ?? 0) >= 0 ? 'text-claude-success' : 'text-claude-danger'} />
                </div>
              </Card>

              {/* PnL Chart */}
              <Card className="p-4 lg:p-6">
                <h4 className="text-sm font-medium text-claude-text-secondary mb-4">总资产曲线</h4>
                {chartLoading ? <PnLChartSkeleton /> : snapshots.length > 0 ? (
                  <EquityLineChart snapshots={snapshots} />
                ) : (
                  <p className="text-sm text-claude-text-muted text-center py-8 lg:py-12">暂无净值数据</p>
                )}
              </Card>

              {/* Price Chart with Buy/Sell Markers */}
              <Card className="p-4 lg:p-6">
                <h4 className="text-sm font-medium text-claude-text-secondary mb-4">
                  BTC 价格走势
                  {priceChart && (
                    <span className="ml-2 text-xs font-normal text-claude-text-muted">
                      当前 ${priceChart.klines[priceChart.klines.length - 1]?.close?.toLocaleString('zh-CN', { maximumFractionDigits: 0 }) || '--'}
                    </span>
                  )}
                </h4>
                {priceChart && priceChart.klines.length > 0 ? (
                  <PriceLineChart data={priceChart} />
                ) : (
                  <p className="text-sm text-claude-text-muted text-center py-8 lg:py-12">
                    {priceChart && priceChart.klines.length === 0 ? '暂无价格数据' : '加载中...'}
                  </p>
                )}
              </Card>

              {/* Strategy info */}
              <Card className="p-4 lg:p-6">
                <h4 className="text-sm font-medium text-claude-text-secondary mb-3">策略运行概况</h4>
                <div className="grid grid-cols-2 gap-3 lg:gap-4 text-xs lg:text-sm">
                  <div><span className="text-claude-text-muted">创建时间</span><p className="font-mono text-claude-text mt-0.5">{selected.created_at?.slice(0, 10)}</p></div>
                  <div><span className="text-claude-text-muted">标的代码</span><p className="font-mono text-claude-text mt-0.5">{selected.symbol}</p></div>
                  <div><span className="text-claude-text-muted">策略模板</span><p className="font-mono text-claude-text mt-0.5">{selected.template_id}</p></div>
                  <div><span className="text-claude-text-muted">当前状态</span><p className="font-mono mt-0.5"><StatusBadge status={selected.status} /></p></div>
                </div>
              </Card>

              {/* Trade history */}
              <Card className="p-4 lg:p-6">
                <h4 className="text-sm font-medium text-claude-text-secondary mb-3">
                  交易记录
                  {trades.length > 0 && <span className="ml-2 text-claude-text-muted font-normal">({trades.length} 笔)</span>}
                </h4>
                {trades.length > 0 ? (
                  <div className="-mx-4 lg:-mx-6">
                    {/* 表头 */}
                    <div className="px-4 lg:px-6 py-2 grid grid-cols-[auto_1fr_1fr_1fr_auto] gap-3 text-[10px] lg:text-xs text-claude-text-muted border-b border-claude-border">
                      <span className="w-12">方向</span>
                      <span className="text-right">数量</span>
                      <span className="text-right">单价</span>
                      <span className="text-right">成交金额</span>
                      <span className="w-14 text-right">时间</span>
                    </div>
                    {trades.slice(0, 20).map((tr: TradeRecord) => (
                      <div key={tr.id}>
                        <div
                          className={`px-4 lg:px-6 py-2.5 grid grid-cols-[auto_1fr_1fr_1fr_auto] gap-3 items-center text-xs lg:text-sm border-b border-claude-border last:border-0 ${
                            tr.action === 'SELL' ? 'cursor-pointer hover:bg-claude-accent-light/30' : ''
                          }`}
                          onClick={() => {
                            if (tr.action === 'SELL') {
                              setExpandedTradeId(expandedTradeId === tr.id ? null : tr.id)
                            }
                          }}
                        >
                          <span className="w-12 flex items-center gap-1">
                            <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] lg:text-xs font-medium ${
                              tr.action === 'BUY' 
                                ? 'bg-green-50 text-green-600' 
                                : 'bg-red-50 text-red-600'
                            }`}>
                              {tr.action === 'BUY' ? '买入' : '卖出'}
                            </span>
                            {tr.action === 'SELL' && (
                              expandedTradeId === tr.id 
                                ? <ChevronUp className="w-3 h-3 text-claude-text-muted" />
                                : <ChevronDown className="w-3 h-3 text-claude-text-muted" />
                            )}
                          </span>
                          <span className="font-mono text-claude-text text-right">
                            {tr.filled_qty > 0 ? Number(tr.filled_qty).toFixed(4) : '--'}
                          </span>
                          <span className="font-mono text-claude-accent text-right font-medium">
                            ${tr.filled_price > 0 ? Number(tr.filled_price).toLocaleString('zh-CN', { maximumFractionDigits: 0 }) : '--'}
                          </span>
                          <span className={`font-mono text-right font-medium ${tr.action === 'BUY' ? 'text-red-500' : 'text-green-600'}`}>
                            {tr.action === 'BUY' ? '-' : '+'}${((tr.filled_qty || 0) * (tr.filled_price || 0)).toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
                          </span>
                          <span className="text-claude-text-muted text-[10px] lg:text-xs w-14 text-right">
                            {new Date(tr.created_at).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>
                        {/* Expanded detail for SELL */}
                        {tr.action === 'SELL' && expandedTradeId === tr.id && (
                          <div className="px-4 lg:px-6 py-3 bg-claude-accent-light/20 border-b border-claude-border">
                            <SellDetail trade={tr} />
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-claude-text-muted text-center py-4">暂无交易记录，启动实例后将自动记录</p>
                )}
              </Card>
            </>
          ) : (
            <Card className="p-8 lg:p-12 text-center text-claude-text-muted">
              {instances.length === 0 ? '暂无实例，点击右上角按钮创建' : '请选择一个实例查看详情'}
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}

function StatItem({ label, value, colorClass, desc }: { label: string; value: string; colorClass?: string; desc?: string }) {
  return (
    <div>
      <p className="text-[10px] lg:text-xs text-claude-text-muted mb-0.5 lg:mb-1">{label}</p>
      <p className={`font-mono text-sm lg:text-lg font-semibold ${colorClass || 'text-claude-text'}`}>{value}</p>
      {desc && <p className="text-[9px] lg:text-[10px] text-claude-text-muted mt-0.5">{desc}</p>}
    </div>
  )
}

function formatCNY(v: number) {
  return '¥' + v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

interface EquityPoint {
  total_equity: number
  recorded_at: string
}

function EquityLineChart({ snapshots }: { snapshots: EquityPoint[] }) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [tooltip, setTooltip] = useState<{ i: number; x: number; y: number } | null>(null)

  const W = 600
  const H = 240
  const PAD = { top: 24, right: 20, bottom: 36, left: 70 }
  const innerW = W - PAD.left - PAD.right
  const innerH = H - PAD.top - PAD.bottom

  const values = snapshots.map(s => s.total_equity)
  const maxV = Math.max(...values)
  const minV = Math.min(...values)
  const range = (maxV - minV) || 1

  const xScale = (i: number) => PAD.left + (i / (snapshots.length - 1)) * innerW
  const yScale = (v: number) => PAD.top + innerH - ((v - minV) / range) * innerH

  const points = snapshots.map((s, i) => `${xScale(i)},${yScale(s.total_equity)}`).join(' ')
  const areaPath = `M${xScale(0)},${PAD.top + innerH} L${points} L${xScale(snapshots.length - 1)},${PAD.top + innerH} Z`

  const yTicks = 4
  const yTickValues = Array.from({ length: yTicks + 1 }, (_, i) => minV + (range * i) / yTicks)

  const handlePointer = useCallback((e: React.PointerEvent) => {
    const svg = svgRef.current
    if (!svg) return
    const rect = svg.getBoundingClientRect()
    const scaleX = W / rect.width
    const mx = (e.clientX - rect.left) * scaleX
    const idx = Math.round(((mx - PAD.left) / innerW) * (snapshots.length - 1))
    const i = Math.max(0, Math.min(snapshots.length - 1, idx))
    setTooltip({ i, x: xScale(i), y: yScale(snapshots[i].total_equity) })
  }, [snapshots])

  return (
    <div className="relative select-none touch-none">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${W} ${H}`}
        className="w-full h-48 lg:h-64"
        onPointerMove={handlePointer}
        onPointerDown={handlePointer}
        onPointerLeave={() => setTooltip(null)}
      >
        {/* Grid lines */}
        {yTickValues.map(v => (
          <g key={v}>
            <line x1={PAD.left} y1={yScale(v)} x2={W - PAD.right} y2={yScale(v)} stroke="#e5e0d8" strokeWidth="0.5" />
            <text x={PAD.left - 8} y={yScale(v) + 5} textAnchor="end" className="text-[11px] fill-[#9ca3af]">
              ¥{v.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
          </g>
        ))}

        {/* Area fill */}
        <path d={areaPath} fill="url(#equityGrad)" opacity="0.3" />
        <defs>
          <linearGradient id="equityGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#d97706" />
            <stop offset="100%" stopColor="#d97706" stopOpacity="0" />
          </linearGradient>
        </defs>

        {/* Line */}
        <polyline points={points} fill="none" stroke="#d97706" strokeWidth="2" strokeLinejoin="round" strokeLinecap="round" />

        {/* Dot at each point (small, visible on hover area) */}
        {snapshots.map((s, i) => (
          <circle key={i} cx={xScale(i)} cy={yScale(s.total_equity)} r="2" fill="#d97706" />
        ))}

        {/* Tooltip */}
        {tooltip && (
          <g>
            <line x1={tooltip.x} y1={PAD.top} x2={tooltip.x} y2={PAD.top + innerH} stroke="#d97706" strokeWidth="1" strokeDasharray="3 2" opacity="0.6" />
            <circle cx={tooltip.x} cy={tooltip.y} r="5" fill="white" stroke="#d97706" strokeWidth="2" />
            <rect x={tooltip.x > W / 2 ? tooltip.x - 130 : tooltip.x + 10} y={PAD.top + 4} width="120" height="28" rx="6" fill="white" stroke="#e5e0d8" strokeWidth="1" />
            <text
              x={tooltip.x > W / 2 ? tooltip.x - 70 : tooltip.x + 70}
              y={PAD.top + 23}
              textAnchor="middle"
              className="text-[12px] font-semibold fill-[#d97706]"
            >
              ¥{snapshots[tooltip.i].total_equity.toLocaleString('zh-CN', { maximumFractionDigits: 0 })}
            </text>
          </g>
        )}

        {/* Time axis labels */}
        {snapshots.length > 1 && (
          <>
            <text x={PAD.left} y={H - 8} textAnchor="start" className="text-[11px] fill-[#9ca3af]">
              {new Date(snapshots[0].recorded_at).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
            </text>
            <text x={W - PAD.right} y={H - 8} textAnchor="end" className="text-[11px] fill-[#9ca3af]">
              {new Date(snapshots[snapshots.length - 1].recorded_at).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
            </text>
          </>
        )}
      </svg>
    </div>
  )
}

function formatPnL(equity: number, initial: number) {
  if (!initial) return '¥0.00'
  const pnl = equity - initial
  const sign = pnl >= 0 ? '+' : ''
  return sign + '¥' + pnl.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatPnLPct(equity: number, initial: number) {
  if (!initial) return '0.00%'
  const pct = ((equity - initial) / initial) * 100
  const sign = pct >= 0 ? '+' : ''
  return sign + pct.toFixed(2) + '%'
}

function PositionBar({ dead, floating, cold, cash, className }: {
  dead: number; floating: number; cold: number; cash: number; className?: string
}) {
  const total = dead + floating + cold + cash
  if (total <= 0) return null

  const segments = [
    { label: '底仓', value: dead, color: 'bg-amber-500', text: 'text-amber-700' },
    { label: '浮动', value: floating, color: 'bg-blue-500', text: 'text-blue-700' },
    { label: '冷封', value: cold, color: 'bg-slate-400', text: 'text-slate-600' },
    { label: '现金', value: cash, color: 'bg-emerald-400', text: 'text-emerald-700' },
  ].filter(s => s.value > 0)

  return (
    <div className={className}>
      <div className="flex h-3 rounded-full overflow-hidden bg-claude-border">
        {segments.map((s, i) => (
          <div
            key={i}
            className={`${s.color} transition-all duration-500`}
            style={{ width: `${(s.value / total) * 100}%`, minWidth: s.value > 0 ? '4px' : 0 }}
            title={`${s.label}: ¥${s.value.toLocaleString()} (${((s.value/total)*100).toFixed(1)}%)`}
          />
        ))}
      </div>
      <div className="flex gap-3 mt-1.5 text-[10px]">
        {segments.map((s, i) => (
          <span key={i} className={`flex items-center gap-1 ${s.text}`}>
            <span className={`w-2 h-2 rounded-sm ${s.color}`} />
            {s.label} {((s.value/total)*100).toFixed(0)}%
          </span>
        ))}
      </div>
    </div>
  )
}

function SellDetail({ trade }: { trade: TradeRecord }) {
  const sellUsd = trade.filled_qty * trade.filled_price
  const costUsd = trade.cost_basis   // cost_basis is already in USD
  const profitUsd = sellUsd - costUsd
  const profitRate = costUsd > 0 ? (profitUsd / costUsd) * 100 : 0

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 text-xs">
      <div>
        <span className="text-claude-text-muted">卖出数量</span>
        <p className="font-mono text-claude-text font-medium mt-0.5">{trade.filled_qty.toFixed(6)} BTC</p>
      </div>
      <div>
        <span className="text-claude-text-muted">卖出金额</span>
        <p className="font-mono text-green-600 font-medium mt-0.5">+${sellUsd.toLocaleString('zh-CN', { maximumFractionDigits: 2 })}</p>
      </div>
      <div>
        <span className="text-claude-text-muted">卖出时间</span>
        <p className="font-mono text-claude-text mt-0.5">{new Date(trade.created_at).toLocaleString('zh-CN', { month:'2-digit', day:'2-digit', hour:'2-digit', minute:'2-digit' })}</p>
      </div>
      <div>
        <span className="text-claude-text-muted">成本(USD)</span>
        <p className="font-mono text-claude-text mt-0.5">${costUsd.toLocaleString('zh-CN', { maximumFractionDigits: 2 })}</p>
      </div>
      <div>
        <span className="text-claude-text-muted">盈利金额</span>
        <p className={`font-mono font-medium mt-0.5 ${profitUsd >= 0 ? 'text-green-600' : 'text-red-500'}`}>
          {profitUsd >= 0 ? '+' : ''}${profitUsd.toLocaleString('zh-CN', { maximumFractionDigits: 2 })}
        </p>
      </div>
      <div>
        <span className="text-claude-text-muted">盈利率</span>
        <p className={`font-mono font-medium mt-0.5 ${profitRate >= 0 ? 'text-green-600' : 'text-red-500'}`}>
          {profitRate >= 0 ? '+' : ''}{profitRate.toFixed(2)}%
        </p>
      </div>
      <div>
        <span className="text-claude-text-muted">引擎</span>
        <p className="font-mono text-claude-text mt-0.5">{trade.engine}</p>
      </div>
    </div>
  )
}
