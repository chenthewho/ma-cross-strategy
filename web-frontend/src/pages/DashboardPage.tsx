import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Plus, Play, Pause, Settings } from 'lucide-react'
import Card from '@/components/Card'
import StatusBadge from '@/components/StatusBadge'
import PnLChartSkeleton from '@/components/skeletons/PnLChartSkeleton'
import { fetchInstances, updateInstanceStatus, type Instance } from '@/shared/services/instances'
import { fetchEquitySnapshots, type EquitySnapshot } from '@/shared/services/dashboard'
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
    const newStatus = inst.status === 'running' ? 'stopped' : 'running'
    await updateInstanceStatus(inst.id, newStatus)
    refetch()
  }

  const selected = selectedId ? instances.find((i) => i.id === selectedId) : null

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-[#e2e8f0]">{t('dashboard.title')}</h2>
        <button onClick={() => navigate('/instances/new')} className="flex items-center gap-2 px-3 py-2 bg-[#2dd4bf]/10 border border-[#2dd4bf]/20 text-[#2dd4bf] rounded-lg text-sm hover:bg-[#2dd4bf]/20 transition-colors">
          <Plus className="w-4 h-4" />{t('dashboard.newInstance')}
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Instance list sidebar */}
        <div className="lg:col-span-1 space-y-3">
          {instances.map((inst) => (
            <Card
              key={inst.id}
              onClick={() => selectInstance(inst.id)}
              className={`p-4 cursor-pointer transition-all ${selectedId === inst.id ? 'border-l-2 border-[#2dd4bf] bg-[#2dd4bf]/[0.04]' : ''}`}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="font-medium text-sm text-[#e2e8f0]">{inst.name}</span>
                <StatusBadge status={inst.status} />
              </div>
              <div className="text-xs text-[#64748b] font-mono">{inst.symbol} · {inst.template_id}</div>
              <button
                onClick={(e) => { e.stopPropagation(); toggleStatus(inst) }}
                className="mt-2 flex items-center gap-1 text-xs text-[#94a3b8] hover:text-[#e2e8f0] transition-colors"
              >
                {inst.status === 'running' ? <Pause className="w-3 h-3" /> : <Play className="w-3 h-3" />}
                {inst.status === 'running' ? t('common.stop') : t('common.start')}
              </button>
            </Card>
          ))}
          {instances.length === 0 && (
            <p className="text-sm text-[#64748b] text-center py-8">暂无实例，点击上方按钮创建</p>
          )}
        </div>

        {/* Main display area */}
        <div className="lg:col-span-3 space-y-6">
          {selected ? (
            <>
              <Card className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-semibold text-[#e2e8f0]">{selected.name}</h3>
                  <StatusBadge status={selected.status} />
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <StatItem label={t('dashboard.totalEquity')} value={formatCNY(selected.total_equity ?? 0)} />
                  <StatItem label={t('dashboard.longHold')} value="--" />
                  <StatItem label={t('dashboard.activeHold')} value="--" />
                  <StatItem label={t('dashboard.availableCash')} value="--" />
                </div>
              </Card>

              {/* PnL Chart */}
              <Card className="p-6">
                <h4 className="text-sm font-medium text-[#94a3b8] mb-4">总资产曲线</h4>
                {chartLoading ? <PnLChartSkeleton /> : snapshots.length > 0 ? (
                  <div className="h-64 flex items-end gap-1">
                    {snapshots.map((s, i) => {
                      const maxV = Math.max(...snapshots.map(x => x.total_equity))
                      const minV = Math.min(...snapshots.map(x => x.total_equity))
                      const range = maxV - minV || 1
                      const h = ((s.total_equity - minV) / range) * 80 + 10
                      return <div key={i} title={`${formatCNY(s.total_equity)}`} className="flex-1 bg-[#2dd4bf]/30 rounded-t" style={{ height: `${h}%` }} />
                    })}
                  </div>
                ) : (
                  <p className="text-sm text-[#64748b] text-center py-12">暂无净值数据</p>
                )}
              </Card>

              {/* Strategy Journey */}
              <Card className="p-6">
                <h4 className="text-sm font-medium text-[#94a3b8] mb-3">策略运行概况</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div><span className="text-[#64748b]">创建时间</span><p className="font-mono text-[#e2e8f0]">{selected.created_at?.slice(0, 10)}</p></div>
                  <div><span className="text-[#64748b]">标的代码</span><p className="font-mono text-[#e2e8f0]">{selected.symbol}</p></div>
                  <div><span className="text-[#64748b]">策略模板</span><p className="font-mono text-[#e2e8f0]">{selected.template_id}</p></div>
                  <div><span className="text-[#64748b]">当前状态</span><p className="font-mono text-[#e2e8f0]"><StatusBadge status={selected.status} /></p></div>
                </div>
              </Card>
            </>
          ) : (
            <Card className="p-12 text-center text-[#64748b]">
              请选择一个实例查看详情
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
      <p className="text-xs text-[#64748b] mb-1">{label}</p>
      <p className="font-mono text-lg text-[#e2e8f0] font-semibold">{value}</p>
    </div>
  )
}

function formatCNY(v: number) {
  return '¥' + v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
