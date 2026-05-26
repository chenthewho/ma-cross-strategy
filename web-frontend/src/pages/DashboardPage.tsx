import { Activity, TrendingUp, BarChart3, Shield } from 'lucide-react';

const stats = [
  { label: 'Total PnL', value: '$12,450.32', change: '+3.2%', icon: TrendingUp, positive: true },
  { label: 'Win Rate', value: '67.4%', change: '+2.1%', icon: Activity, positive: true },
  { label: 'Sharpe', value: '1.84', change: '-0.1', icon: BarChart3, positive: false },
  { label: 'Drawdown', value: '-4.2%', change: '', icon: Shield, positive: true },
];

export default function DashboardPage() {
  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white font-mono">Dashboard</h1>
          <p className="text-slate-400 text-sm mt-1">
            Strategy performance overview
          </p>
        </div>
        <span className="px-3 py-1 rounded-full bg-qs-safe/10 text-qs-safe text-xs font-medium border border-qs-safe/20">
          Live
        </span>
      </div>

      {/* Stats Grid */}
      <div className="qs-bento-grid">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="p-5 rounded-xl bg-qs-surface border border-white/5 hover:border-white/10 transition-colors"
          >
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">
                {stat.label}
              </span>
              <stat.icon className="w-4 h-4 text-slate-500" />
            </div>
            <div className="text-2xl font-bold text-white font-mono mb-1">
              {stat.value}
            </div>
            {stat.change && (
              <span
                className={`text-xs font-medium ${
                  stat.positive ? 'text-qs-safe' : 'text-qs-danger'
                }`}
              >
                {stat.change}
              </span>
            )}
          </div>
        ))}
      </div>

      {/* Chart Placeholder */}
      <div className="rounded-xl bg-qs-surface border border-white/5 p-6 h-64 flex items-center justify-center">
        <div className="text-center">
          <BarChart3 className="w-10 h-10 text-slate-600 mx-auto mb-3" />
          <p className="text-slate-500 text-sm font-mono">
            Equity Curve — Coming Soon
          </p>
        </div>
      </div>
    </div>
  );
}
