import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Loader2, Play } from 'lucide-react'
import Card from '@/components/Card'
import { fetchEvolutionTasks, createEvolutionTask, promoteEvolutionTask } from '@/shared/services/evolution'

export default function EvolutionPage() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'optimize' | 'library'>('optimize')
  const [loading, setLoading] = useState(false)
  const [popSize, setPopSize] = useState('300')
  const [maxGen, setMaxGen] = useState('25')

  const { data: evoData } = useQuery({ queryKey: ['evolution-tasks'], queryFn: fetchEvolutionTasks, refetchInterval: 5000 })
  const tasks = evoData?.tasks || []
  const challengers = evoData?.challengers || []

  const runningTask = tasks.find((t: any) => t.status === 'running')

  const startEvo = async () => {
    setLoading(true)
    try { await createEvolutionTask({ strategy_id: 'golden_cross', pop_size: parseInt(popSize), max_generations: parseInt(maxGen) }) }
    catch (e) {} finally { setLoading(false) }
  }

  const promote = async (id: number) => {
    try {
      await promoteEvolutionTask(id)
      queryClient.invalidateQueries({ queryKey: ['evolution-tasks'] })
    } catch (e) {}
  }

  const tabs = [
    { key: 'optimize' as const, label: '参数优化' },
    { key: 'library' as const, label: '基因库' },
  ]

  return (
    <div className="space-y-4 lg:space-y-6">
      <h2 className="text-lg lg:text-xl font-semibold text-claude-text">进化实验室</h2>
      <div className="flex gap-4 border-b border-claude-border pb-0 overflow-x-auto">
        {tabs.map((tb) => (
          <button key={tb.key} onClick={() => setTab(tb.key)}
            className={`px-3 lg:px-4 py-2 text-xs lg:text-sm border-b-2 transition-colors whitespace-nowrap ${tab === tb.key ? 'border-claude-accent text-claude-accent font-medium' : 'border-transparent text-claude-text-secondary hover:text-claude-text'}`}>
            {tb.label}
          </button>
        ))}
      </div>

      {tab === 'optimize' && (
        <div className="space-y-4 lg:space-y-6">
          {!runningTask ? (
            <Card className="p-4 lg:p-6 space-y-4">
              <h3 className="font-medium text-claude-text text-sm lg:text-base">启动新一轮优化</h3>
              <div className="grid grid-cols-2 gap-3 lg:gap-4">
                <div>
                  <label className="text-[10px] lg:text-xs text-claude-text-secondary block mb-1">种群大小</label>
                  <input type="number" value={popSize} onChange={(e) => setPopSize(e.target.value)}
                    className="w-full px-3 py-2 bg-claude-bg border border-claude-border rounded-lg text-xs lg:text-sm text-claude-text font-mono focus:border-claude-accent focus:ring-1 focus:ring-claude-accent/30 outline-none transition-colors" />
                </div>
                <div>
                  <label className="text-[10px] lg:text-xs text-claude-text-secondary block mb-1">最大代数</label>
                  <input type="number" value={maxGen} onChange={(e) => setMaxGen(e.target.value)}
                    className="w-full px-3 py-2 bg-claude-bg border border-claude-border rounded-lg text-xs lg:text-sm text-claude-text font-mono focus:border-claude-accent focus:ring-1 focus:ring-claude-accent/30 outline-none transition-colors" />
                </div>
              </div>
              <button onClick={startEvo} disabled={loading}
                className="flex items-center justify-center gap-2 w-full lg:w-auto px-4 py-2 lg:py-2.5 bg-claude-accent text-white rounded-lg text-xs lg:text-sm font-medium hover:bg-claude-accent-hover disabled:opacity-50 transition-colors">
                {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                启动优化
              </button>
            </Card>
          ) : (
            <Card className="p-4 lg:p-6 space-y-3">
              <h3 className="font-medium text-claude-text text-sm lg:text-base">进化进行中</h3>
              <div className="text-xs text-claude-text-secondary font-mono">
                代 -- / --
              </div>
              <div className="w-full bg-claude-border rounded-full h-2"><div className="bg-claude-accent h-2 rounded-full transition-all" style={{ width: '0%' }} /></div>
              <div className="font-mono text-xs lg:text-sm text-claude-text">评分: --</div>
            </Card>
          )}

          <div className="text-[10px] lg:text-xs text-claude-text-muted font-medium">历史任务</div>
          {tasks.map((t: any) => (
            <Card key={t.id} className="p-3 flex items-center justify-between">
              <span className="text-xs lg:text-sm text-claude-text-secondary">#{t.id} {t.status}</span>
              <span className="text-xs font-mono text-claude-text">--</span>
            </Card>
          ))}
        </div>
      )}

      {tab === 'library' && (
        <div className="space-y-3">
          {challengers.map((g: any) => (
            <Card key={g.id} className={`p-3 lg:p-4 ${g.role === 'champion' ? 'border-claude-accent bg-claude-accent-light' : ''}`}>
              <div className="flex items-center justify-between">
                <span className="text-[10px] lg:text-xs px-1.5 lg:px-2 py-0.5 rounded border text-claude-text-secondary border-claude-border">{g.role === 'champion' ? '当前最优' : '候选参数'}</span>
                <span className="font-mono text-xs lg:text-sm text-claude-text font-medium">{g.score_total?.toFixed(4) || '--'}</span>
              </div>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 mt-2 text-[10px] lg:text-xs text-claude-text-muted">
                <span>6M: {g.score_6m?.toFixed(2) || '--'}</span>
                <span>2Y: {g.score_2y?.toFixed(2) || '--'}</span>
                <span>5Y: {g.score_5y?.toFixed(2) || '--'}</span>
                <span>MaxDD: {(g.max_drawdown * 100)?.toFixed(1) || '--'}%</span>
              </div>
              {g.role !== 'champion' && (
                <button onClick={() => promote(g.id)}
                  className="mt-2 text-[10px] lg:text-xs px-2 py-1 bg-claude-accent text-white rounded hover:bg-claude-accent-hover transition-colors">
                  晋升为最优
                </button>
              )}
            </Card>
          ))}
          {challengers.length === 0 && (
            <p className="text-sm text-claude-text-muted text-center py-8">暂无基因记录</p>
          )}
        </div>
      )}
    </div>
  )
}
