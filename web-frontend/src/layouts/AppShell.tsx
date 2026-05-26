import { Outlet } from 'react-router-dom'
import AppBackground from '@/components/AppBackground'
import Sidebar from '@/layouts/Sidebar'
import Topbar from '@/layouts/Topbar'

export default function AppShell() {
  return (
    <div className="flex h-screen overflow-hidden">
      <AppBackground />
      <Sidebar />
      <div className="flex-1 flex flex-col relative z-10">
        <Topbar />
        <main className="flex-1 overflow-y-auto p-4 lg:p-6 max-w-[1800px] w-full mx-auto custom-scrollbar">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
