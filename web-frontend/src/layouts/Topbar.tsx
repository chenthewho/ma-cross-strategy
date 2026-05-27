import { useEffect } from 'react'
import { Menu } from 'lucide-react'
import { useAuth } from '@/app/AuthProvider'
import { useSystemStatusStore } from '@/stores/systemStatusStore'
import { fetchSystemStatus } from '@/shared/services/system'
import { useI18n } from '@/i18n/I18nProvider'
import { useSidebar } from '@/hooks/useSidebar'

export default function Topbar() {
  const { user, logout } = useAuth()
  const { engine, api_connected, api_configured, setStatus } = useSystemStatusStore()
  const { t } = useI18n()
  const { toggle } = useSidebar()

  useEffect(() => {
    const poll = () => {
      fetchSystemStatus()
        .then(setStatus)
        .catch(() => {})
    }
    poll()
    const id = setInterval(poll, 30000)
    return () => clearInterval(id)
  }, [])

  const engineColor = engine === 'running' ? 'bg-claude-success' : engine === 'paused' ? 'bg-claude-warning' : 'bg-claude-danger'
  const engineLabel = engine === 'running' ? t('engine.running') : engine === 'paused' ? t('engine.paused') : t('engine.halted')

  return (
    <header className="h-14 lg:h-16 border-b border-claude-border flex items-center justify-between px-3 lg:px-6 bg-claude-bg/80 backdrop-blur-sm shrink-0">
      <div className="flex items-center gap-3 lg:gap-6">
        <button onClick={toggle} className="lg:hidden p-1.5 -ml-1 text-claude-text-secondary hover:text-claude-text rounded-lg hover:bg-claude-border">
          <Menu className="w-5 h-5" />
        </button>

        <div className="hidden sm:flex items-center gap-6">
          <div className="flex items-center gap-2 text-xs text-claude-text-secondary">
            <span className={`inline-block w-2 h-2 rounded-full ${engineColor}`} />
            <span className="hidden md:inline">{engineLabel}</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-claude-text-secondary">
            <span className={`inline-block w-2 h-2 rounded-full ${api_connected ? 'bg-claude-success' : 'bg-claude-danger'}`} />
            <span className="hidden md:inline">{api_connected ? t('status.online') : t('status.offline')}</span>
          </div>
          {api_configured && (
            <span className="hidden md:inline text-xs text-claude-success font-medium">API Key ✓</span>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 lg:gap-4">
        <span className="text-xs lg:text-sm text-claude-text-secondary truncate max-w-[120px] lg:max-w-none">{user?.email}</span>
        <button onClick={logout} className="text-xs text-claude-text-muted hover:text-claude-danger transition-colors shrink-0">
          {t('common.cancel')}
        </button>
      </div>
    </header>
  )
}
