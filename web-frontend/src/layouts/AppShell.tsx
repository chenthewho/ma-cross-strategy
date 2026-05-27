import { Outlet } from 'react-router-dom'
import AppBackground from '@/components/AppBackground'
import Sidebar from '@/layouts/Sidebar'
import Topbar from '@/layouts/Topbar'
import { SidebarProvider } from '@/hooks/useSidebar'
import ToastContainer from '@/components/ToastContainer'

export default function AppShell() {
  return (
    <SidebarProvider>
      <div className="flex h-screen overflow-hidden">
        <AppBackground />
        <Sidebar />
        <div className="flex-1 flex flex-col relative z-10 min-w-0">
          <Topbar />
          <main className="flex-1 overflow-y-auto p-3 sm:p-4 lg:p-6 custom-scrollbar">
            <Outlet />
          </main>
        </div>
      </div>
      <ToastContainer />
    </SidebarProvider>
  )
}
