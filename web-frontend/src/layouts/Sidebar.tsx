import { NavLink } from 'react-router-dom'
import { useEffect } from 'react'
import { Activity, X } from 'lucide-react'
import { navItems } from '@/shared/config/navigation'
import { hasFeature } from '@/shared/config/features'
import { useI18n } from '@/i18n/I18nProvider'
import { useSidebar } from '@/hooks/useSidebar'

export default function Sidebar() {
  const { t } = useI18n()
  const { open, close } = useSidebar()

  // Lock body scroll when mobile drawer is open
  useEffect(() => {
    if (open) {
      document.body.classList.add('sidebar-open')
    } else {
      document.body.classList.remove('sidebar-open')
    }
    return () => document.body.classList.remove('sidebar-open')
  }, [open])

  const mainItems = navItems.filter((i) => !i.feature || hasFeature(i.feature as any)).filter((i) => i.placement === 'main')
  const footerItems = navItems.filter((i) => !i.feature || hasFeature(i.feature as any)).filter((i) => i.placement === 'footer')

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-150 ${
      isActive
        ? 'text-[#2dd4bf] border border-[#2dd4bf]/10 bg-[#2dd4bf]/[0.06]'
        : 'text-[#94a3b8] hover:text-[#cbd5e1] hover:bg-white/[0.03]'
    }`

  const onNavigate = () => {
    // close drawer on mobile after navigation
    if (window.innerWidth < 1024) close()
  }

  const sidebarContent = (
    <>
      {/* Brand */}
      <div className="h-16 flex items-center gap-3 px-3 border-b border-white/[0.04] shrink-0">
        <Activity className="w-6 h-6 text-[#ff8c6b]" style={{ filter: 'drop-shadow(0 0 8px rgba(255,140,107,0.4))' }} />
        <span className="text-lg font-bold text-[#e2e8f0] tracking-wider">QuantSaaS</span>
        {/* Close button on mobile */}
        <button onClick={close} className="lg:hidden ml-auto p-1.5 text-[#94a3b8] hover:text-[#e2e8f0] rounded-lg hover:bg-white/[0.05]">
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Main nav */}
      <nav className="flex-1 py-4 px-2 space-y-1 custom-scrollbar overflow-y-auto">
        {mainItems.map((item) => (
          <NavLink key={item.to} to={item.to} end={item.end} className={linkClass} onClick={onNavigate}>
            <item.icon className="w-5 h-5 shrink-0" />
            <span>{t(item.labelKey)}</span>
          </NavLink>
        ))}
      </nav>

      {/* Footer nav */}
      <div className="p-2 border-t border-white/[0.04] shrink-0">
        {footerItems.map((item) => (
          <NavLink key={item.to} to={item.to} className={linkClass} onClick={onNavigate}>
            <item.icon className="w-5 h-5 shrink-0" />
            <span>{t(item.labelKey)}</span>
          </NavLink>
        ))}
      </div>
    </>
  )

  return (
    <>
      {/* Desktop: persistent sidebar */}
      <aside className="hidden lg:flex h-screen flex-col border-r border-[#0a0f1c] bg-[#020617]/40 backdrop-blur-xl w-64 shrink-0">
        {sidebarContent}
      </aside>

      {/* Mobile: slide-over drawer */}
      <div className={`lg:hidden fixed inset-0 z-50 pointer-events-none ${open ? '' : 'hidden'}`}>
        {/* Backdrop */}
        <div
          className="absolute inset-0 bg-black/60 backdrop-blur-sm pointer-events-auto transition-opacity"
          onClick={close}
        />
        {/* Drawer */}
        <aside className="absolute left-0 top-0 bottom-0 w-72 flex flex-col border-r border-[#0a0f1c] bg-[#020617]/95 backdrop-blur-xl pointer-events-auto shadow-2xl animate-slide-in">
          {sidebarContent}
        </aside>
      </div>
    </>
  )
}
