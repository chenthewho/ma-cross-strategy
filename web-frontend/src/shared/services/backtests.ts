import { apiFetch } from './api'

export async function createBacktest(data: Record<string, unknown>) {
  return apiFetch<any>('/api/v1/backtests', { method: 'POST', body: JSON.stringify(data) })
}

export async function fetchBacktest(id: string) {
  return apiFetch<any>(`/api/v1/backtests/${id}`)
}
