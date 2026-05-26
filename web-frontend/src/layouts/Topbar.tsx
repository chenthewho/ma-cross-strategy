import { useEffect } from 'react'
import { useAuth } from '@/app/AuthProvider'
import { useSystemStatusStore } from '@/stores/systemStatusStore'
import { fetchSystemStatus } from '@/shared/services/system'
import { useI18n } from '@/i18n/I18nProvider'

export default function Topbar() {
  const { user, logout } = useAuth()
  const { engine, api_connected, api_configured, setStatus } = useSystemStatusStore()
  const { t } = useI18n()

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
    <header className="h-16 border-b border-white/[0.04] flex items-center justify-between px-6 bg-[#020617]/30 backdrop-blur">
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2 text-xs text-[#94a3b8]">
          <span className={`inline-block w-2 h-2 rounded-full ${engineColor}`} />
          {engineLabel}
        </div>
        <div className="flex items-center gap-2 text-xs text-[#94a3b8]">
          <span className={`inline-block w-2 h-2 rounded-full ${api_connected ? 'bg-[#34d399]' : 'bg-[#f87171]'}`} />
          {api_connected ? t('status.online') : t('status.offline')}
        </div>
        {api_configured && (
          <span className="text-xs text-[#34d399]">API Key ✓</span>
        )}
      </div>

      <div className="flex items-center gap-4">
        <span className="text-sm text-[#94a3b8]">{user?.email}</span>
        <button
          onClick={logout}
          className="text-xs text-[#f87171] hover:text-[#fca5a5] transition-colors"
        >
          {t('common.cancel')}
        </button>
      </div>
    </header>
  )
}
