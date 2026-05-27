import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, AuthGate } from '@/app/AuthProvider'
import { I18nProvider } from '@/i18n/I18nProvider'
import { ToastProvider } from '@/hooks/useToast'
import AppShell from '@/layouts/AppShell'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import DashboardPage from '@/pages/DashboardPage'
import TemplatesPage from '@/pages/TemplatesPage'
import InstanceListPage from '@/pages/InstanceListPage'
import InstanceCreatePage from '@/pages/InstanceCreatePage'
import EvolutionPage from '@/pages/EvolutionPage'
import AgentsPage from '@/pages/AgentsPage'
import BacktestingPage from '@/pages/BacktestingPage'
import SettingsPage from '@/pages/SettingsPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <I18nProvider>
          <ToastProvider>
          <AuthProvider>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route element={<AuthGate />}>
                <Route element={<AppShell />}>
                  <Route index element={<DashboardPage />} />
                  <Route path="templates" element={<TemplatesPage />} />
                  <Route path="instances" element={<InstanceListPage />} />
                  <Route path="instances/new" element={<InstanceCreatePage />} />
                  <Route path="evolution" element={<EvolutionPage />} />
                  <Route path="agents" element={<AgentsPage />} />
                  <Route path="backtesting" element={<BacktestingPage />} />
                  <Route path="settings" element={<SettingsPage />} />
                </Route>
              </Route>
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </AuthProvider>
          </ToastProvider>
        </I18nProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
