import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await login(email, password);
      navigate('/');
    } catch {
      setError('Invalid credentials');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#020617]">
      <div className="w-full max-w-md p-8 rounded-2xl border border-white/10 bg-qs-surface backdrop-blur-sm">
        <h1 className="text-2xl font-bold text-qs-accent mb-2 font-mono">
          QuantStrategy
        </h1>
        <p className="text-slate-400 mb-8 text-sm">Sign in to your dashboard</p>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg bg-white/5 border border-white/10 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-qs-accent/50 focus:ring-1 focus:ring-qs-accent/20 transition-colors"
              placeholder="trader@example.com"
              required
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg bg-white/5 border border-white/10 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-qs-accent/50 focus:ring-1 focus:ring-qs-accent/20 transition-colors"
              placeholder="••••••••"
              required
            />
          </div>

          {error && (
            <p className="text-qs-danger text-sm">{error}</p>
          )}

          <button
            type="submit"
            className="w-full py-2.5 rounded-lg bg-qs-accent text-slate-900 font-semibold hover:bg-qs-accent/80 transition-colors cursor-pointer"
          >
            Sign In
          </button>
        </form>
      </div>
    </div>
  );
}
