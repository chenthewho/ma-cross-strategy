import { LayoutDashboard, Cpu, Sliders, Bot, BarChart3, Settings } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

export interface NavItem {
  to: string
  labelKey: string
  icon: LucideIcon
  placement: 'main' | 'footer'
  feature?: string
  end?: boolean
}

export const navItems: NavItem[] = [
  { to: '/', labelKey: 'nav.dashboard', icon: LayoutDashboard, placement: 'main', end: true },
  { to: '/templates', labelKey: 'nav.templates', icon: Cpu, placement: 'main' },
  { to: '/instances', labelKey: 'nav.instances', icon: Sliders, placement: 'main' },
  { to: '/evolution', labelKey: 'nav.evolution', icon: BarChart3, placement: 'main', feature: 'evolution' },
  { to: '/agents', labelKey: 'nav.agents', icon: Bot, placement: 'main' },
  { to: '/backtesting', labelKey: 'nav.backtesting', icon: BarChart3, placement: 'main', feature: 'backtesting' },
  { to: '/settings', labelKey: 'nav.settings', icon: Settings, placement: 'footer' },
]
