import { useState } from 'react'
import { Loader2, Play } from 'lucide-react'
import Card from '@/components/Card'
import { createBacktest } from '@/shared/services/backtests'
import { useI18n } from '@/i18n/I18nProvider'

export default function BacktestingPage() {
  const { t } = useI18n()
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
      <h2 className="text-lg lg:text-xl font-semibold text-[#e2e8f0]">{t('backtesting.title')}</h2>

      <Card className="p-4 lg:p-6 space-y-4">
        <p className="text-xs lg:text-sm text-[#94a3b8]">{t('backtesting.currentParams')}</p>
        <button onClick={startBacktest} disabled={loading}
          className="flex items-center gap-2 px-4 py-2 lg:py-2.5 bg-[#2dd4bf] text-[#020617] rounded-lg text-xs lg:text-sm font-semibold hover:bg-[#2dd4bf]/90 disabled:opacity-50">
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
          {t('backtesting.start')}
        </button>
      </Card>

      {result && (
        <Card className="p-4 lg:p-6 space-y-3">
          <h3 className="font-medium text-[#e2e8f0] text-sm lg:text-base">回测结果</h3>
          <div className="grid grid-cols-2 gap-3 text-xs lg:text-sm">
            <div className="font-mono"><span className="text-[#64748b]">总收益率</span><p className="text-[#34d399]">{result.roi ? (result.roi * 100).toFixed(2) + '%' : '--'}</p></div>
            <div className="font-mono"><span className="text-[#64748b]">最大回撤</span><p className="text-[#f87171]">{result.max_drawdown ? (result.max_drawdown * 100).toFixed(2) + '%' : '--'}</p></div>
            <div className="font-mono"><span className="text-[#64748b]">Alpha</span><p className="text-[#34d399]">{result.alpha ? (result.alpha * 100).toFixed(2) + '%' : '--'}</p></div>
          </div>
        </Card>
      )}
    </div>
  )
}
