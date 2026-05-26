import { create } from 'zustand'

interface SystemStatus {
  engine: 'running' | 'paused' | 'halted'
  api_connected: boolean
  api_configured: boolean
}

interface SystemStatusState extends SystemStatus {
  setStatus: (status: SystemStatus) => void
}

export const useSystemStatusStore = create<SystemStatusState>((set) => ({
  engine: 'running',
  api_connected: false,
  api_configured: false,
  setStatus: (status) => set(status),
}))
