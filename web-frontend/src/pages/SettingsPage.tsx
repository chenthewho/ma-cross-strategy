import Card from '@/components/Card'
import { useI18n } from '@/i18n/I18nProvider'
import { useAuth } from '@/app/AuthProvider'

export default function SettingsPage() {
  const { t, locale, setLocale } = useI18n()
  const { user } = useAuth()

  return (
    <div className="space-y-4 lg:space-y-6 max-w-xl">
      <h2 className="text-lg lg:text-xl font-semibold text-claude-text">{t('nav.settings')}</h2>
      <Card className="p-4 lg:p-6 space-y-5">
        <div>
          <p className="text-[10px] lg:text-xs text-claude-text-muted mb-1">邮箱</p>
          <p className="font-mono text-xs lg:text-sm text-claude-text break-all">{user?.email}</p>
        </div>
        <div>
          <p className="text-[10px] lg:text-xs text-claude-text-muted mb-1">角色</p>
          <p className="text-xs lg:text-sm text-claude-text">{user?.role}</p>
        </div>
        <div>
          <p className="text-[10px] lg:text-xs text-claude-text-muted mb-2">语言 / Language</p>
          <div className="flex gap-2">
            {(['zh', 'en'] as const).map((l) => (
              <button key={l} onClick={() => setLocale(l)}
                className={`px-3 py-1.5 rounded-lg text-xs border transition-colors font-medium ${locale === l ? 'border-claude-accent bg-claude-accent-light text-claude-accent' : 'border-claude-border text-claude-text-secondary hover:text-claude-text hover:border-claude-border-hover'}`}>
                {l === 'zh' ? '中文' : 'English'}
              </button>
            ))}
          </div>
        </div>
      </Card>
    </div>
  )
}
