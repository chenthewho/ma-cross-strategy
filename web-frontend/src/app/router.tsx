import { createBrowserRouter } from 'react-router-dom';
import AppShell from '@/components/AppShell';
import LoginPage from '@/pages/LoginPage';
import DashboardPage from '@/pages/DashboardPage';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    element: <AppShell />,
    children: [
      {
        path: '/',
        element: <DashboardPage />,
      },
      {
        path: '/strategies',
        element: <PlaceholderPage title="Strategies" />,
      },
      {
        path: '/backtest',
        element: <PlaceholderPage title="Backtest" />,
      },
      {
        path: '/analytics',
        element: <PlaceholderPage title="Analytics" />,
      },
      {
        path: '/settings',
        element: <PlaceholderPage title="Settings" />,
      },
    ],
  },
]);

function PlaceholderPage({ title }: { title: string }) {
  return (
    <div className="p-6 flex items-center justify-center h-full">
      <p className="text-slate-500 font-mono text-sm">{title} — Coming Soon</p>
    </div>
  );
}
