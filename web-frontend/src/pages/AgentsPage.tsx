import { useEffect } from 'react'
import Card from '@/components/Card'
import { useSystemStatusStore } from '@/stores/systemStatusStore'
import { fetchSystemStatus } from '@/shared/services/system'
import { useI18n } from '@/i18n/I18nProvider'

export default function AgentsPage() {
  const { t } = useI18n()
  const { api_connected, api_configured, setStatus } = useSystemStatusStore()

  useEffect(() => {
    fetchSystemStatus().then(setStatus).catch(() => {})
    const id = setInterval(() => fetchSystemStatus().then(setStatus).catch(() => {}), 30000)
    return () => clearInterval(id)
  }, [])

  return (
    <div className="space-y-4 lg:space-y-6 max-w-2xl">
      <h2 className="text-lg lg:text-xl font-semibold text-claude-text">{t('agents.title')}</h2>

      <Card className="p-4 lg:p-6 text-center">
        {api_connected ? (
          <>
            <div className="w-4 h-4 rounded-full bg-claude-success mx-auto mb-3 animate-pulse" />
            <p className="text-claude-success font-semibold text-sm lg:text-base">{t('agents.connected')}</p>
            <p className="text-[10px] lg:text-xs text-claude-text-muted mt-1 font-mono">最后心跳: 刚刚</p>
          </>
        ) : (
          <>
            <div className="w-4 h-4 rounded-full bg-claude-text-muted mx-auto mb-3" />
            <p className="text-claude-danger font-semibold text-sm lg:text-base">{t('agents.disconnected')}</p>
          </>
        )}
      </Card>

      <div className="space-y-3 lg:space-y-4">
        {[
          { step: t('agents.step1'), desc: '下载对应平台的 Agent 可执行文件' },
          { step: t('agents.step2'), desc: '在本地创建 config.agent.yaml，填入券商 API Key', hint: t('agents.apiKeyHint') },
          { step: t('agents.step3'), desc: '运行 Agent 后，此处状态应自动更新为已连接' },
        ].map((s, i) => (
          <Card key={i} className="p-3 lg:p-4 flex items-start gap-3 lg:gap-4">
            <span className="w-7 h-7 lg:w-8 lg:h-8 rounded-full bg-claude-accent-light border border-claude-accent-border flex items-center justify-center text-xs lg:text-sm font-bold text-claude-accent shrink-0">{i + 1}</span>
            <div>
              <p className="font-medium text-xs lg:text-sm text-claude-text">{s.step}</p>
              <p className="text-[10px] lg:text-xs text-claude-text-muted mt-0.5">{s.desc}</p>
              {'hint' in s && <p className="text-[10px] lg:text-xs text-claude-warning mt-1 italic">⚠ {s.hint}</p>}
            </div>
          </Card>
        ))}
      </div>

      <Card className="p-3 lg:p-4">
        <p className="text-[10px] lg:text-xs text-claude-text-secondary">API Key 状态</p>
        <p className={`font-mono text-xs lg:text-sm mt-1 ${api_configured ? 'text-claude-success' : 'text-claude-danger'}`}>
          {api_configured ? '已配置' : '未配置'}
        </p>
      </Card>
    </div>
  )
}
