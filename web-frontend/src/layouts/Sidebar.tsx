import { NavLink } from 'react-router-dom'
import { Activity } from 'lucide-react'
import { navItems } from '@/shared/config/navigation'
import { hasFeature } from '@/shared/config/features'
import { useI18n } from '@/i18n/I18nProvider'

export default function Sidebar() {
  const { t } = useI18n()
  const mainItems = navItems.filter((i) => !i.feature || hasFeature(i.feature as any)).filter((i) => i.placement === 'main')
  const footerItems = navItems.filter((i) => !i.feature || hasFeature(i.feature as any)).filter((i) => i.placement === 'footer')

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-150 ${
      isActive
        ? 'text-[#2dd4bf] border border-[#2dd4bf]/10 bg-[#2dd4bf]/[0.06]'
        : 'text-[#94a3b8] hover:text-[#cbd5e1] hover:bg-white/[0.03]'
    }`

  return (
    <aside className="h-screen flex flex-col border-r border-[#0a0f1c] bg-[#020617]/40 backdrop-blur-xl w-16 lg:w-64 transition-all">
      {/* Brand */}
      <div className="h-16 flex items-center gap-3 px-3 border-b border-white/[0.04]">
        <Activity className="w-6 h-6 text-[#ff8c6b]" style={{ filter: 'drop-shadow(0 0 8px rgba(255,140,107,0.4))' }} />
        <span className="hidden lg:block text-lg font-bold text-[#e2e8f0] tracking-wider">QuantSaaS</span>
      </div>

      {/* Main nav */}
      <nav className="flex-1 py-4 px-2 space-y-1 custom-scrollbar overflow-y-auto">
        {mainItems.map((item) => (
          <NavLink key={item.to} to={item.to} end={item.end} className={linkClass}>
            <item.icon className="w-5 h-5 shrink-0" />
            <span className="hidden lg:block">{t(item.labelKey)}</span>
          </NavLink>
        ))}
      </nav>

      {/* Footer nav */}
      <div className="p-2 border-t border-white/[0.04]">
        {footerItems.map((item) => (
          <NavLink key={item.to} to={item.to} className={linkClass}>
            <item.icon className="w-5 h-5 shrink-0" />
            <span className="hidden lg:block">{t(item.labelKey)}</span>
          </NavLink>
        ))}
      </div>
    </aside>
  )
}
