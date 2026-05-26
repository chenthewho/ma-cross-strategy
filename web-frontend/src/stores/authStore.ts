import { create } from 'zustand'

interface AuthState {
  token: string | null
  user: { email: string; role: string } | null
  loading: boolean
  restore: () => void
  setAuth: (token: string, user: { email: string; role: string }) => void
  clearAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  loading: true,
  restore: () => {
    const token = localStorage.getItem('token')
    const userRaw = localStorage.getItem('user')
    if (token && userRaw) {
      try {
        const user = JSON.parse(userRaw)
        set({ token, user, loading: false })
      } catch { set({ loading: false }) }
    } else {
      set({ loading: false })
    }
  },
  setAuth: (token, user) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    set({ token, user, loading: false })
  },
  clearAuth: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null, loading: false })
  },
}))
