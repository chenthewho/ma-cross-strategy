import {
  Activity,
  BarChart3,
  Settings,
  LogOut,
  Layers,
  TrendingUp,
  Menu,
} from 'lucide-react';
import { useState } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { Link, useLocation, Outlet } from 'react-router-dom';

/* ── Sidebar nav items ── */
const navItems = [
  { to: '/', label: 'Dashboard', icon: Activity },
  { to: '/strategies', label: 'Strategies', icon: Layers },
  { to: '/backtest', label: 'Backtest', icon: TrendingUp },
  { to: '/analytics', label: 'Analytics', icon: BarChart3 },
  { to: '/settings', label: 'Settings', icon: Settings },
];

/* ── AppShell wraps authenticated routes ── */
export default function AppShell() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const location = useLocation();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  return (
    <div className="flex h-screen overflow-hidden bg-[#020617]">
      {/* ── Mobile overlay ── */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/60 lg:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}

      {/* ── Left Sidebar ── */}
      <aside
        className={`
          fixed lg:static inset-y-0 left-0 z-50
          flex flex-col
          transition-all duration-200
          ${sidebarOpen ? 'w-64' : 'w-16'}
          ${mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
          bg-[#0a0d14] border-r border-white/5
        `}
      >
        {/* Logo */}
        <div className="flex items-center h-16 px-4 border-b border-white/5">
          <div className="w-8 h-8 rounded-lg bg-qs-accent/20 flex items-center justify-center flex-shrink-0">
            <TrendingUp className="w-5 h-5 text-qs-accent" />
          </div>
          {sidebarOpen && (
            <span className="ml-3 font-bold text-qs-accent font-mono text-sm">
              QS
            </span>
          )}
          <button
            className="ml-auto p-1 rounded hover:bg-white/5 hidden lg:block cursor-pointer"
            onClick={() => setSidebarOpen(!sidebarOpen)}
          >
            <Menu className="w-4 h-4 text-slate-500" />
          </button>
        </div>

        {/* Nav items */}
        <nav className="flex-1 py-4 space-y-1 px-2 overflow-y-auto custom-scrollbar">
          {navItems.map((item) => {
            const active = location.pathname === item.to;
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm
                  transition-colors
                  ${
                    active
                      ? 'bg-qs-accent/10 text-qs-accent'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-white/5'
                  }
                `}
              >
                <item.icon className="w-5 h-5 flex-shrink-0" />
                {sidebarOpen && <span>{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        {/* Bottom: user + logout */}
        <div className="p-2 border-t border-white/5">
          {sidebarOpen ? (
            <div className="px-3 py-2">
              <p className="text-xs text-slate-500 truncate">
                {user?.email ?? 'trader@example.com'}
              </p>
              <button
                onClick={logout}
                className="flex items-center gap-2 mt-1 text-xs text-slate-500 hover:text-qs-danger transition-colors cursor-pointer"
              >
                <LogOut className="w-3.5 h-3.5" />
                Sign out
              </button>
            </div>
          ) : (
            <button
              onClick={logout}
              className="w-full p-2 flex justify-center text-slate-500 hover:text-qs-danger transition-colors cursor-pointer"
              title="Sign out"
            >
              <LogOut className="w-4 h-4" />
            </button>
          )}
        </div>
      </aside>

      {/* ── Main content area ── */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Topbar */}
        <header className="h-16 flex items-center justify-between px-4 lg:px-6 border-b border-white/5 bg-[#0a0d14]/50 backdrop-blur-sm flex-shrink-0">
          <button
            className="lg:hidden p-2 rounded-lg hover:bg-white/5 cursor-pointer"
            onClick={() => setMobileOpen(true)}
          >
            <Menu className="w-5 h-5 text-slate-400" />
          </button>

          <div className="flex items-center gap-4 ml-auto">
            {/* Status indicator */}
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-qs-safe/10 border border-qs-safe/20">
              <span className="w-2 h-2 rounded-full bg-qs-safe animate-pulse" />
              <span className="text-xs font-medium text-qs-safe font-mono">
                CONNECTED
              </span>
            </div>

            {/* User chip (desktop) */}
            <div className="hidden sm:flex items-center gap-2">
              <div className="w-7 h-7 rounded-full bg-qs-accent/20 flex items-center justify-center">
                <span className="text-xs font-bold text-qs-accent">
                  {user?.email?.charAt(0).toUpperCase() ?? 'T'}
                </span>
              </div>
              <span className="text-sm text-slate-400 hidden md:block">
                {user?.email ?? 'trader@example.com'}
              </span>
            </div>
          </div>
        </header>

        {/* Scrollable page content */}
        <main className="flex-1 overflow-y-auto custom-scrollbar">
          <Outlet />
        </main>
      </div>

      {/* Subtle radial glows */}
      <div className="fixed inset-0 pointer-events-none -z-10">
        <div className="absolute top-0 right-0 w-[600px] h-[600px] bg-qs-accent/3 rounded-full blur-[120px]" />
        <div className="absolute bottom-0 left-0 w-[500px] h-[500px] bg-qs-accent/2 rounded-full blur-[100px]" />
      </div>
    </div>
  );
}
