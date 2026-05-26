import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { Outlet, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

interface AuthContextType {
  user: { email: string; role: string } | null
  loading: boolean
  login: (token: string, user: { email: string; role: string }) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const { token, user, loading, restore, setAuth, clearAuth } = useAuthStore()

  useEffect(() => { restore() }, [])

  const login = (t: string, u: { email: string; role: string }) => setAuth(t, u)
  const logout = () => clearAuth()

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {!loading && children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export function AuthGate() {
  const { user, loading } = useAuth()
  if (loading) return <div className="flex items-center justify-center h-screen bg-[#020617]"><div className="animate-spin h-8 w-8 border-2 border-[#2dd4bf] border-t-transparent rounded-full" /></div>
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}
