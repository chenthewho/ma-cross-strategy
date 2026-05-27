import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Activity, Loader2 } from 'lucide-react'
import { useAuth } from '@/app/AuthProvider'
import { loginAPI } from '@/shared/services/auth'
import AppBackground from '@/components/AppBackground'
import { useI18n } from '@/i18n/I18nProvider'

export default function LoginPage() {
  const { login } = useAuth()
  const { t } = useI18n()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await loginAPI(email, password)
      login(res.token, res.user)
      navigate('/')
    } catch (err: any) {
      setError(err.message || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center relative px-4">
      <AppBackground />
      <div className="relative z-10 w-full max-w-[400px] border border-white/10 backdrop-blur-xl bg-slate-900/60 rounded-2xl p-6 lg:p-8">
        <div className="text-center mb-6 lg:mb-8">
          <Activity className="w-8 h-8 lg:w-10 lg:h-10 text-[#ff8c6b] mx-auto mb-3" style={{ filter: 'drop-shadow(0 0 8px rgba(255,140,107,0.4))' }} />
          <h1 className="text-xl lg:text-2xl font-bold text-[#e2e8f0] tracking-wider">{t('login.title')}</h1>
          <p className="text-xs lg:text-sm text-[#64748b] mt-1">{t('login.slogan')}</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3 lg:space-y-4">
          <input
            type="email" required placeholder={t('login.email')}
            value={email} onChange={(e) => setEmail(e.target.value)}
            disabled={loading}
            className="w-full px-3 lg:px-4 py-2.5 bg-slate-900/80 border border-slate-700 rounded-lg text-[#e2e8f0] text-sm focus:border-[#2dd4bf] focus:outline-none transition-colors tracking-wider placeholder:text-[#64748b]"
          />
          <input
            type="password" required placeholder={t('login.password')}
            value={password} onChange={(e) => setPassword(e.target.value)}
            disabled={loading}
            className="w-full px-3 lg:px-4 py-2.5 bg-slate-900/80 border border-slate-700 rounded-lg text-[#e2e8f0] text-sm focus:border-[#2dd4bf] focus:outline-none transition-colors tracking-wider placeholder:text-[#64748b]"
          />
          <button
            type="submit" disabled={loading}
            className="w-full py-2.5 bg-[#2dd4bf] text-[#020617] font-semibold rounded-lg text-sm uppercase tracking-wider hover:bg-[#2dd4bf]/90 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading && <Loader2 className="w-4 h-4 animate-spin" />}
            {t('login.submit')}
          </button>
          {error && <p className="text-[#f87171] text-xs text-center">{error}</p>}
        </form>

        <p className="text-center mt-4 lg:mt-6 text-xs text-[#64748b]">
          {t('login.noAccount')}{' '}
          <Link to="/register" className="text-[#2dd4bf] hover:underline">{t('register.submit')}</Link>
        </p>
      </div>
    </div>
  )
}
