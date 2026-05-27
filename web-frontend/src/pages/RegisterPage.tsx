import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Activity, Loader2 } from 'lucide-react'
import { useAuth } from '@/app/AuthProvider'
import { registerAPI } from '@/shared/services/auth'
import { useI18n } from '@/i18n/I18nProvider'

export default function RegisterPage() {
  const { login } = useAuth()
  const { t } = useI18n()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (password !== confirm) { setError('Passwords do not match'); return }
    setLoading(true)
    try {
      const res = await registerAPI(email, password)
      login(res.token, res.user)
      navigate('/')
    } catch (err: any) {
      setError(err.message || 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-claude-bg px-4">
      <div className="w-full max-w-[400px] bg-claude-surface border border-claude-border rounded-2xl shadow-lg p-6 lg:p-8">
        <div className="text-center mb-6 lg:mb-8">
          <div className="w-10 h-10 rounded-xl bg-claude-accent flex items-center justify-center mx-auto mb-3">
            <Activity className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-xl lg:text-2xl font-semibold text-claude-text">{t('register.title')}</h1>
          <p className="text-xs lg:text-sm text-claude-text-muted mt-1">{t('login.slogan')}</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3 lg:space-y-4">
          <input type="email" required placeholder={t('login.email')}
            value={email} onChange={(e) => setEmail(e.target.value)} disabled={loading}
            className="w-full px-4 py-2.5 bg-claude-bg border border-claude-border rounded-lg text-claude-text text-sm focus:border-claude-accent focus:ring-1 focus:ring-claude-accent/30 outline-none transition-colors placeholder:text-claude-text-muted" />
          <input type="password" required placeholder={t('login.password')}
            value={password} onChange={(e) => setPassword(e.target.value)} disabled={loading}
            className="w-full px-4 py-2.5 bg-claude-bg border border-claude-border rounded-lg text-claude-text text-sm focus:border-claude-accent focus:ring-1 focus:ring-claude-accent/30 outline-none transition-colors placeholder:text-claude-text-muted" />
          <input type="password" required placeholder={t('register.confirmPassword')}
            value={confirm} onChange={(e) => setConfirm(e.target.value)} disabled={loading}
            className="w-full px-4 py-2.5 bg-claude-bg border border-claude-border rounded-lg text-claude-text text-sm focus:border-claude-accent focus:ring-1 focus:ring-claude-accent/30 outline-none transition-colors placeholder:text-claude-text-muted" />
          <button type="submit" disabled={loading}
            className="w-full py-2.5 bg-claude-accent text-white font-semibold rounded-lg text-sm hover:bg-claude-accent-hover transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
            {loading && <Loader2 className="w-4 h-4 animate-spin" />}
            {t('register.submit')}
          </button>
          {error && <p className="text-claude-danger text-xs text-center">{error}</p>}
        </form>

        <p className="text-center mt-4 lg:mt-6 text-xs text-claude-text-muted">
          {t('login.hasAccount')}{' '}
          <Link to="/login" className="text-claude-accent font-medium hover:underline">{t('login.submit')}</Link>
        </p>
      </div>
    </div>
  )
}
