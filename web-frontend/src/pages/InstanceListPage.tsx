import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Plus, Trash2, Eye, ChevronRight } from 'lucide-react'
import Card from '@/components/Card'
import StatusBadge from '@/components/StatusBadge'
import { fetchInstances, deleteInstance, type Instance } from '@/shared/services/instances'
import { useI18n } from '@/i18n/I18nProvider'
import { useState } from 'react'

export default function InstanceListPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const { data: instances = [], refetch } = useQuery({ queryKey: ['instances'], queryFn: fetchInstances })
  const [confirmId, setConfirmId] = useState<number | null>(null)

  const handleDelete = async (id: number) => {
    await deleteInstance(id)
    setConfirmId(null)
    refetch()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-[#e2e8f0]">{t('nav.instances')}</h2>
        <button onClick={() => navigate('/instances/new')}
          className="flex items-center gap-2 px-3 py-2 bg-[#2dd4bf]/10 border border-[#2dd4bf]/20 text-[#2dd4bf] rounded-lg text-sm hover:bg-[#2dd4bf]/20 transition-colors">
          <Plus className="w-4 h-4" />{t('dashboard.newInstance')}
        </button>
      </div>

      <Card className="overflow-hidden">
        <div className="divide-y divide-white/[0.04]">
          {instances.map((inst: Instance) => (
            <div key={inst.id} className="flex items-center justify-between p-4 hover:bg-white/[0.02] transition-colors">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3">
                  <span className="font-medium text-sm text-[#e2e8f0]">{inst.name}</span>
                  <StatusBadge status={inst.status} />
                </div>
                <div className="text-xs text-[#64748b] mt-0.5">{inst.template_id} · {inst.symbol}</div>
              </div>
              <div className="hidden md:block text-right mx-4">
                <span className="font-mono text-sm text-[#e2e8f0]">¥{inst.total_equity?.toLocaleString() ?? '--'}</span>
              </div>
              <div className="flex items-center gap-2">
                <button onClick={() => navigate(`/?instance=${inst.id}`)} className="p-1.5 text-[#94a3b8] hover:text-[#2dd4bf] transition-colors"><Eye className="w-4 h-4" /></button>
                {confirmId === inst.id ? (
                  <span className="flex items-center gap-1">
                    <button onClick={() => handleDelete(inst.id)} className="text-xs text-[#f87171]">确认</button>
                    <button onClick={() => setConfirmId(null)} className="text-xs text-[#94a3b8]">{t('common.cancel')}</button>
                  </span>
                ) : (
                  <button onClick={() => setConfirmId(inst.id)} className="p-1.5 text-[#94a3b8] hover:text-[#f87171] transition-colors"><Trash2 className="w-4 h-4" /></button>
                )}
              </div>
            </div>
          ))}
        </div>
        {instances.length === 0 && <p className="text-sm text-[#64748b] text-center py-12">暂无实例</p>}
      </Card>
    </div>
  )
}
