import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Plus, Trash2, Eye } from 'lucide-react'
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
    <div className="space-y-4 lg:space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg lg:text-xl font-semibold text-claude-text">{t('nav.instances')}</h2>
        <button onClick={() => navigate('/instances/new')}
          className="flex items-center gap-1.5 lg:gap-2 px-3 lg:px-4 py-2 bg-claude-accent text-white rounded-lg text-xs lg:text-sm font-medium hover:bg-claude-accent-hover transition-colors">
          <Plus className="w-3.5 h-3.5 lg:w-4 lg:h-4" />{t('dashboard.newInstance')}
        </button>
      </div>

      <Card className="overflow-hidden">
        <div className="divide-y divide-claude-border">
          {instances.map((inst: Instance) => (
            <div key={inst.id} className="flex items-center justify-between p-3 lg:p-4 hover:bg-claude-surface-hover transition-colors">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 lg:gap-3">
                  <span className="font-medium text-xs lg:text-sm text-claude-text truncate">{inst.name}</span>
                  <StatusBadge status={inst.status} />
                </div>
                <div className="text-[10px] lg:text-xs text-claude-text-muted mt-0.5 truncate">{inst.template_id} · {inst.symbol}</div>
              </div>
              <div className="hidden sm:block text-right mx-4">
                <span className="font-mono text-xs lg:text-sm text-claude-text">¥{inst.total_equity?.toLocaleString() ?? '--'}</span>
              </div>
              <div className="flex items-center gap-1 lg:gap-2 shrink-0">
                <button onClick={() => navigate(`/?instance=${inst.id}`)} className="p-1.5 text-claude-text-muted hover:text-claude-accent transition-colors"><Eye className="w-4 h-4" /></button>
                {confirmId === inst.id ? (
                  <span className="flex items-center gap-1">
                    <button onClick={() => handleDelete(inst.id)} className="text-xs text-claude-danger px-1 font-medium">确认</button>
                    <button onClick={() => setConfirmId(null)} className="text-xs text-claude-text-muted px-1">{t('common.cancel')}</button>
                  </span>
                ) : (
                  <button onClick={() => setConfirmId(inst.id)} className="p-1.5 text-claude-text-muted hover:text-claude-danger transition-colors"><Trash2 className="w-4 h-4" /></button>
                )}
              </div>
            </div>
          ))}
        </div>
        {instances.length === 0 && <p className="text-sm text-claude-text-muted text-center py-12">暂无实例</p>}
      </Card>
    </div>
  )
}
