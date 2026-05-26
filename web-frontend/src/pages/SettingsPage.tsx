import { useState } from 'react'
import Card from '@/components/Card'
import { useI18n } from '@/i18n/I18nProvider'
import { useAuth } from '@/app/AuthProvider'

export default function SettingsPage() {
  const { t, locale, setLocale } = useI18n()
  const { user } = useAuth()

  return (
    <div className="space-y-6 max-w-xl">
      <h2 className="text-xl font-semibold text-[#e2e8f0]">{t('nav.settings')}</h2>
      <Card className="p-6 space-y-4">
        <div>
          <p className="text-xs text-[#64748b]">邮箱</p>
          <p className="font-mono text-sm text-[#e2e8f0]">{user?.email}</p>
        </div>
        <div>
          <p className="text-xs text-[#64748b]">角色</p>
          <p className="text-sm text-[#e2e8f0]">{user?.role}</p>
        </div>
        <div>
          <p className="text-xs text-[#64748b] mb-2">语言 / Language</p>
          <div className="flex gap-2">
            {(['zh', 'en'] as const).map((l) => (
              <button key={l} onClick={() => setLocale(l)}
                className={`px-3 py-1.5 rounded-lg text-xs border transition-colors ${locale === l ? 'border-[#2dd4bf] bg-[#2dd4bf]/10 text-[#2dd4bf]' : 'border-white/[0.06] text-[#94a3b8] hover:text-[#e2e8f0]'}`}>
                {l === 'zh' ? '中文' : 'English'}
              </button>
            ))}
          </div>
        </div>
      </Card>
    </div>
  )
}
