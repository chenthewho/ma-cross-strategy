import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Plus, Play, Pause } from 'lucide-react'
import Card from '@/components/Card'
import StatusBadge from '@/components/StatusBadge'
import PnLChartSkeleton from '@/components/skeletons/PnLChartSkeleton'
import { fetchInstances, updateInstanceStatus, type Instance } from '@/shared/services/instances'
import { fetchEquitySnapshots } from '@/shared/services/dashboard'
import { useI18n } from '@/i18n/I18nProvider'

export default function DashboardPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [selectedId, setSelectedId] = useState<number | null>(null)

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

  useEffect(() => {
    const idParam = searchParams.get('instance')
    if (idParam && !selectedId) {
      const id = parseInt(idParam)
      if (!isNaN(id) && instances.some((i) => i.id === id)) setSelectedId(id)
    }
    if (!idParam && instances.length > 0 && !selectedId) setSelectedId(instances[0].id)
  }, [instances, searchParams])

  const selectInstance = (id: number) => { setSelectedId(id); setSearchParams({ instance: String(id) }) }

  const toggleStatus = async (inst: Instance) => {
    const isRunning = (inst.status || '').toLowerCase() === 'running'
    const newStatus = isRunning ? 'stopped' : 'running'
    await updateInstanceStatus(inst.id, newStatus)
    refetch()
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
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
                  <StatItem label={t('dashboard.totalEquity')} value={formatCNY(selected.total_equity ?? selected.cny_balance ?? 0)} />
                  <StatItem label={t('dashboard.longHold')} value={formatCNY(selected.dead_hold ?? 0)} />
                  <StatItem label={t('dashboard.activeHold')} value={formatCNY(selected.float_hold ?? 0)} />
                  <StatItem label={t('dashboard.availableCash')} value={formatCNY(selected.cny_balance ?? 0)} />
                </div>
              </Card>

              {/* PnL Chart */}
              <Card className="p-4 lg:p-6">
                <h4 className="text-sm font-medium text-claude-text-secondary mb-4">总资产曲线</h4>
                {chartLoading ? <PnLChartSkeleton /> : snapshots.length > 0 ? (
                  <div className="h-48 lg:h-64 flex items-end gap-[2px] lg:gap-1">
                    {snapshots.map((s, i) => {
                      const maxV = Math.max(...snapshots.map(x => x.total_equity))
                      const minV = Math.min(...snapshots.map(x => x.total_equity))
                      const range = maxV - minV || 1
                      const h = ((s.total_equity - minV) / range) * 80 + 10
                      return <div key={i} title={`${formatCNY(s.total_equity)}`} className="flex-1 bg-claude-accent/30 rounded-t hover:bg-claude-accent/50 transition-colors" style={{ height: `${h}%` }} />
                    })}
                  </div>
                ) : (
                  <p className="text-sm text-claude-text-muted text-center py-8 lg:py-12">暂无净值数据</p>
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

function StatItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-[10px] lg:text-xs text-claude-text-muted mb-0.5 lg:mb-1">{label}</p>
      <p className="font-mono text-sm lg:text-lg text-claude-text font-semibold">{value}</p>
    </div>
  )
}

function formatCNY(v: number) {
  return '¥' + v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
