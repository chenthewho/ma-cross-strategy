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
    `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors duration-150 ${
      isActive
        ? 'text-claude-accent bg-claude-accent-light font-medium'
        : 'text-claude-text-secondary hover:text-claude-text hover:bg-claude-surface-hover'
    }`

  const onNavigate = () => {
    if (window.innerWidth < 1024) close()
  }

  const sidebarContent = (
    <>
      {/* Brand */}
      <div className="h-16 flex items-center gap-3 px-4 border-b border-claude-border shrink-0">
        <div className="w-8 h-8 rounded-lg bg-claude-accent flex items-center justify-center">
          <Activity className="w-5 h-5 text-white" />
        </div>
        <span className="text-lg font-semibold text-claude-text tracking-tight">QuantSaaS</span>
        <button onClick={close} className="lg:hidden ml-auto p-1.5 text-claude-text-secondary hover:text-claude-text rounded-lg hover:bg-claude-border">
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Main nav */}
      <nav className="flex-1 py-4 px-3 space-y-0.5 custom-scrollbar overflow-y-auto">
        {mainItems.map((item) => (
          <NavLink key={item.to} to={item.to} end={item.end} className={linkClass} onClick={onNavigate}>
            <item.icon className="w-5 h-5 shrink-0" />
            <span>{t(item.labelKey)}</span>
          </NavLink>
        ))}
      </nav>

      {/* Footer nav */}
      <div className="p-3 border-t border-claude-border shrink-0">
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
      <aside className="hidden lg:flex h-screen flex-col border-r border-claude-border bg-claude-surface w-60 shrink-0">
        {sidebarContent}
      </aside>

      {/* Mobile: slide-over drawer */}
      <div className={`lg:hidden fixed inset-0 z-50 pointer-events-none ${open ? '' : 'hidden'}`}>
        <div
          className="absolute inset-0 bg-black/30 backdrop-blur-sm pointer-events-auto transition-opacity"
          onClick={close}
        />
        <aside className="absolute left-0 top-0 bottom-0 w-72 flex flex-col border-r border-claude-border bg-claude-surface pointer-events-auto shadow-xl animate-slide-in">
          {sidebarContent}
        </aside>
      </div>
    </>
  )
}
