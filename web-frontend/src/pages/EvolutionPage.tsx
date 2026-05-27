import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Loader2, Play } from 'lucide-react'
import Card from '@/components/Card'
import { fetchEvolutionTasks, createEvolutionTask, fetchGenomes } from '@/shared/services/evolution'
import { useI18n } from '@/i18n/I18nProvider'

export default function EvolutionPage() {
  const { t } = useI18n()
  const [tab, setTab] = useState<'optimize' | 'library'>('optimize')
  const [loading, setLoading] = useState(false)
  const [popSize, setPopSize] = useState('300')
  const [maxGen, setMaxGen] = useState('25')

  const { data: tasks = [] } = useQuery({ queryKey: ['evolution-tasks'], queryFn: fetchEvolutionTasks, refetchInterval: 5000 })
  const { data: genomes = [] } = useQuery({ queryKey: ['genomes'], queryFn: fetchGenomes, enabled: tab === 'library' })

  const runningTask = tasks.find((t: any) => t.status === 'running')

  const startEvo = async () => {
    setLoading(true)
    try { await createEvolutionTask({ strategy_id: 'golden_cross', pop_size: parseInt(popSize), max_generations: parseInt(maxGen) }) }
    catch (e) {} finally { setLoading(false) }
  }

  const tabs = [
    { key: 'optimize' as const, label: t('evolution.optimize') },
    { key: 'library' as const, label: t('evolution.library') },
  ]

  return (
    <div className="space-y-4 lg:space-y-6">
      <h2 className="text-lg lg:text-xl font-semibold text-[#e2e8f0]">{t('nav.evolution')}</h2>
      <div className="flex gap-4 border-b border-white/[0.04] pb-0 overflow-x-auto">
        {tabs.map((tb) => (
          <button key={tb.key} onClick={() => setTab(tb.key)}
            className={`px-3 lg:px-4 py-2 text-xs lg:text-sm border-b-2 transition-colors whitespace-nowrap ${tab === tb.key ? 'border-[#2dd4bf] text-[#2dd4bf]' : 'border-transparent text-[#94a3b8] hover:text-[#e2e8f0]'}`}>
            {tb.label}
          </button>
        ))}
      </div>

      {tab === 'optimize' && (
        <div className="space-y-4 lg:space-y-6">
          {!runningTask ? (
            <Card className="p-4 lg:p-6 space-y-4">
              <h3 className="font-medium text-[#e2e8f0] text-sm lg:text-base">{t('evolution.startEvo')}</h3>
              <div className="grid grid-cols-2 gap-3 lg:gap-4">
                <div>
                  <label className="text-[10px] lg:text-xs text-[#94a3b8] block mb-1">{t('evolution.popSize')}</label>
                  <input type="number" value={popSize} onChange={(e) => setPopSize(e.target.value)}
                    className="w-full px-2.5 lg:px-3 py-2 bg-slate-900/80 border border-slate-700 rounded-lg text-xs lg:text-sm text-[#e2e8f0] font-mono focus:border-[#2dd4bf] focus:outline-none" />
                </div>
                <div>
                  <label className="text-[10px] lg:text-xs text-[#94a3b8] block mb-1">{t('evolution.maxGen')}</label>
                  <input type="number" value={maxGen} onChange={(e) => setMaxGen(e.target.value)}
                    className="w-full px-2.5 lg:px-3 py-2 bg-slate-900/80 border border-slate-700 rounded-lg text-xs lg:text-sm text-[#e2e8f0] font-mono focus:border-[#2dd4bf] focus:outline-none" />
                </div>
              </div>
              <button onClick={startEvo} disabled={loading}
                className="flex items-center justify-center gap-2 w-full lg:w-auto px-4 py-2 lg:py-2.5 bg-[#2dd4bf] text-[#020617] rounded-lg text-xs lg:text-sm font-semibold hover:bg-[#2dd4bf]/90 disabled:opacity-50">
                {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                {t('evolution.startEvo')}
              </button>
            </Card>
          ) : (
            <Card className="p-4 lg:p-6 space-y-3">
              <h3 className="font-medium text-[#e2e8f0] text-sm lg:text-base">进化进行中</h3>
              <div className="text-xs text-[#94a3b8] font-mono">
                代 {runningTask.progress?.generation || 0} / {runningTask.progress?.max_generations || '--'}
              </div>
              <div className="w-full bg-slate-800 rounded-full h-2"><div className="bg-[#2dd4bf] h-2 rounded-full" style={{ width: `${((runningTask.progress?.generation || 0) / (runningTask.progress?.max_generations || 1)) * 100}%` }} /></div>
              <div className="font-mono text-xs lg:text-sm text-[#e2e8f0]">评分: {runningTask.progress?.best_score?.toFixed(4) || '--'}</div>
            </Card>
          )}

          <div className="text-[10px] lg:text-xs text-[#64748b]">历史任务</div>
          {tasks.map((t: any) => (
            <Card key={t.id} className="p-3 flex items-center justify-between">
              <span className="text-xs lg:text-sm text-[#94a3b8]">#{t.id} {t.status}</span>
              <span className="text-xs font-mono text-[#94a3b8]">{t.result?.score_total?.toFixed(4) || '--'}</span>
            </Card>
          ))}
        </div>
      )}

      {tab === 'library' && (
        <div className="space-y-3">
          {genomes.map((g: any) => (
            <Card key={g.id} className={`p-3 lg:p-4 ${g.role === 'champion' ? 'border-[#2dd4bf] bg-[#2dd4bf]/[0.04]' : ''}`}>
              <div className="flex items-center justify-between">
                <span className="text-[10px] lg:text-xs px-1.5 lg:px-2 py-0.5 rounded border text-[#94a3b8] border-white/[0.08]">{g.role === 'champion' ? t('evolution.champion') : '候选'}</span>
                <span className="font-mono text-xs lg:text-sm text-[#e2e8f0]">{g.score_total?.toFixed(4) || '--'}</span>
              </div>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 mt-2 text-[10px] lg:text-xs text-[#64748b]">
                <span>6M: {g.score_6m?.toFixed(2) || '--'}</span>
                <span>2Y: {g.score_2y?.toFixed(2) || '--'}</span>
                <span>5Y: {g.score_5y?.toFixed(2) || '--'}</span>
                <span>MaxDD: {(g.max_drawdown * 100)?.toFixed(1) || '--'}%</span>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
