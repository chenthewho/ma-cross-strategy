import { useState } from 'react'
import { Loader2, Play } from 'lucide-react'
import Card from '@/components/Card'
import { createBacktest } from '@/shared/services/backtests'

export default function BacktestingPage() {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)

  const startBacktest = async () => {
    setLoading(true)
    try {
      const r = await createBacktest({ strategy_id: 'golden_cross', mode: 'champion' })
      setResult(r)
    } catch (e) {} finally { setLoading(false) }
  }

  return (
    <div className="space-y-4 lg:space-y-6 max-w-2xl">
      <h2 className="text-lg lg:text-xl font-semibold text-claude-text">回测</h2>

      <Card className="p-4 lg:p-6 space-y-4">
        <p className="text-xs lg:text-sm text-claude-text-secondary">使用当前最优参数进行历史回测</p>
        <button onClick={startBacktest} disabled={loading}
          className="flex items-center gap-2 px-4 py-2 lg:py-2.5 bg-claude-accent text-white rounded-lg text-xs lg:text-sm font-medium hover:bg-claude-accent-hover disabled:opacity-50 transition-colors">
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
          开始回测
        </button>
      </Card>

      {result && (
        <Card className="p-4 lg:p-6 space-y-3">
          <h3 className="font-medium text-claude-text text-sm lg:text-base">回测结果</h3>
          <div className="grid grid-cols-2 gap-3 text-xs lg:text-sm">
            <div className="font-mono"><span className="text-claude-text-muted">总收益率</span><p className="text-claude-success font-medium">{result.roi ? (result.roi * 100).toFixed(2) + '%' : '--'}</p></div>
            <div className="font-mono"><span className="text-claude-text-muted">最大回撤</span><p className="text-claude-danger font-medium">{result.max_drawdown ? (result.max_drawdown * 100).toFixed(2) + '%' : '--'}</p></div>
            <div className="font-mono"><span className="text-claude-text-muted">Alpha</span><p className="text-claude-success font-medium">{result.alpha ? (result.alpha * 100).toFixed(2) + '%' : '--'}</p></div>
          </div>
        </Card>
      )}
    </div>
  )
}
