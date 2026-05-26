import { apiFetch } from './api'

export interface SystemStatusResponse {
  engine: 'running' | 'paused' | 'halted'
  api_connected: boolean
  api_configured: boolean
}

export async function fetchSystemStatus() {
  return apiFetch<SystemStatusResponse>('/api/v1/system/status')
}
