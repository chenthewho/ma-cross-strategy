import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { ArrowLeft, ArrowRight } from 'lucide-react'
import Card from '@/components/Card'
import strategyCatalog from '@/shared/config/strategyCatalog'
import { createInstance } from '@/shared/services/instances'
import { useI18n } from '@/i18n/I18nProvider'

export default function InstanceCreatePage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const initialTemplate = searchParams.get('template') || ''

  const [step, setStep] = useState(0)
  const [templateId, setTemplateId] = useState(initialTemplate)
  const [name, setName] = useState('')
  const [capital, setCapital] = useState('100000')
  const [monthlyInject, setMonthlyInject] = useState('5000')
  const [risk, setRisk] = useState('30')

  const handleSubmit = async () => {
    try {
      const res = await createInstance({
        template_id: templateId,
        name,
        initial_capital: parseFloat(capital),
        monthly_inject: parseFloat(monthlyInject),
        max_drawdown: parseFloat(risk) / 100,
      })
      navigate(`/?instance=${res.id}`, { state: { success: true } })
    } catch (e) {}
  }

  return (
    <div className="max-w-xl mx-auto space-y-6">
      <h2 className="text-xl font-semibold text-[#e2e8f0]">{t('dashboard.newInstance')}</h2>

      {step === 0 && (
        <div className="space-y-4">
          <p className="text-sm text-[#94a3b8]">选择策略模板</p>
          <div className="grid grid-cols-1 gap-3">
            {strategyCatalog.map((s) => (
              <Card key={s.id}
                onClick={() => setTemplateId(s.id)}
                className={`p-4 cursor-pointer transition-all ${templateId === s.id ? 'border-[#2dd4bf] bg-[#2dd4bf]/[0.04]' : ''}`}>
                <div className="h-1 w-8 rounded-full mb-2" style={{ backgroundColor: s.color }} />
                <p className="font-medium text-sm text-[#e2e8f0]">{s.name}</p>
                <p className="text-xs text-[#64748b] mt-1">{s.description}</p>
              </Card>
            ))}
          </div>
          <button onClick={() => setStep(1)} disabled={!templateId} className="w-full flex items-center justify-center gap-2 py-2.5 bg-[#2dd4bf] text-[#020617] font-semibold rounded-lg text-sm hover:bg-[#2dd4bf]/90 disabled:opacity-50">
            {t('common.save')} <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      )}

      {step === 1 && (
        <div className="space-y-4">
          <button onClick={() => setStep(0)} className="flex items-center gap-1 text-sm text-[#94a3b8] hover:text-[#e2e8f0]"><ArrowLeft className="w-3 h-3" />{t('common.back')}</button>
          <div>
            <label className="text-xs text-[#94a3b8] block mb-1">实例名称</label>
            <input value={name} onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-900/80 border border-slate-700 rounded-lg text-[#e2e8f0] text-sm focus:border-[#2dd4bf] focus:outline-none" />
          </div>
          <div>
            <label className="text-xs text-[#94a3b8] block mb-1">初始资金 (CNY)</label>
            <input type="number" value={capital} onChange={(e) => setCapital(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-900/80 border border-slate-700 rounded-lg text-[#e2e8f0] text-sm focus:border-[#2dd4bf] focus:outline-none font-mono" />
          </div>
          <div>
            <label className="text-xs text-[#94a3b8] block mb-1">月注资金额 (CNY)</label>
            <input type="number" value={monthlyInject} onChange={(e) => setMonthlyInject(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-900/80 border border-slate-700 rounded-lg text-[#e2e8f0] text-sm focus:border-[#2dd4bf] focus:outline-none font-mono" />
          </div>
          <div>
            <label className="text-xs text-[#94a3b8] block mb-1">最大可接受回撤: {risk}%</label>
            <input type="range" min="5" max="80" value={risk} onChange={(e) => setRisk(e.target.value)}
              className="w-full accent-[#2dd4bf]" />
          </div>
          <button onClick={handleSubmit} disabled={!name} className="w-full flex items-center justify-center gap-2 py-2.5 bg-[#2dd4bf] text-[#020617] font-semibold rounded-lg text-sm hover:bg-[#2dd4bf]/90 disabled:opacity-50">
            {t('common.create')}
          </button>
        </div>
      )}
    </div>
  )
}
