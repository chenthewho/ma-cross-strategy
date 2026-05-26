// @ts-nocheck
import { useNavigate } from 'react-router-dom'
import { PlusCircle } from 'lucide-react'
import Card from '@/components/Card'
import strategyCatalog from '@/shared/config/strategyCatalog'
import { useI18n } from '@/i18n/I18nProvider'

export default function TemplatesPage() {
  const { t } = useI18n()
  const navigate = useNavigate()

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-[#e2e8f0]">{t('nav.templates')}</h2>
      <div className="qs-bento-grid">
        {strategyCatalog.map((s) => (
          <Card key={s.id} className="p-5 flex flex-col">
            <div className="h-1 w-12 rounded-full mb-3" style={{ backgroundColor: s.color }} />
            <h3 className="font-semibold text-[#e2e8f0] mb-1">{s.name}</h3>
            <p className="text-xs text-[#64748b] mb-3 flex-1">{s.description}</p>
            <div className="flex items-center justify-between text-xs">
              <span className="text-[#94a3b8] font-mono">{s.symbol}</span>
              {s.supportsEvolution && <span className="text-[#2dd4bf] bg-[#2dd4bf]/10 px-1.5 py-0.5 rounded">进化</span>}
            </div>
            <button
              onClick={() => navigate(`/instances/new?template=${s.id}`)}
              className="mt-3 flex items-center justify-center gap-1.5 py-2 bg-[#2dd4bf]/10 border border-[#2dd4bf]/20 text-[#2dd4bf] rounded-lg text-xs hover:bg-[#2dd4bf]/20 transition-colors"
            >
              <PlusCircle className="w-3.5 h-3.5" />{t('dashboard.newInstance')}
            </button>
          </Card>
        ))}
      </div>
    </div>
  )
}
