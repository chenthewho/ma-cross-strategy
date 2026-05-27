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

  const engineColor = engine === 'running' ? 'bg-[#34d399]' : engine === 'paused' ? 'bg-[#fbbf24]' : 'bg-[#f87171]'
  const engineLabel = engine === 'running' ? t('engine.running') : engine === 'paused' ? t('engine.paused') : t('engine.halted')

  return (
    <header className="h-14 lg:h-16 border-b border-white/[0.04] flex items-center justify-between px-3 lg:px-6 bg-[#020617]/30 backdrop-blur shrink-0">
      <div className="flex items-center gap-3 lg:gap-6">
        {/* Hamburger on mobile */}
        <button onClick={toggle} className="lg:hidden p-1.5 -ml-1 text-[#94a3b8] hover:text-[#e2e8f0] rounded-lg hover:bg-white/[0.05]">
          <Menu className="w-5 h-5" />
        </button>

        {/* Status indicators — hide on very small screens */}
        <div className="hidden sm:flex items-center gap-6">
          <div className="flex items-center gap-2 text-xs text-[#94a3b8]">
            <span className={`inline-block w-2 h-2 rounded-full ${engineColor}`} />
            <span className="hidden md:inline">{engineLabel}</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-[#94a3b8]">
            <span className={`inline-block w-2 h-2 rounded-full ${api_connected ? 'bg-[#34d399]' : 'bg-[#f87171]'}`} />
            <span className="hidden md:inline">{api_connected ? t('status.online') : t('status.offline')}</span>
          </div>
          {api_configured && (
            <span className="hidden md:inline text-xs text-[#34d399]">API Key ✓</span>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 lg:gap-4">
        <span className="text-xs lg:text-sm text-[#94a3b8] truncate max-w-[120px] lg:max-w-none">{user?.email}</span>
        <button
          onClick={logout}
          className="text-xs text-[#f87171] hover:text-[#fca5a5] transition-colors shrink-0"
        >
          {t('common.cancel')}
        </button>
      </div>
    </header>
  )
}
