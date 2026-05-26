import { apiFetch } from './api'

export interface Instance {
  id: number
  name: string
  symbol: string
  template_id: string
  status: string
  total_equity?: number
  created_at: string
}

export async function fetchInstances(): Promise<Instance[]> {
  return apiFetch<Instance[]>('/api/v1/instances')
}

export async function fetchInstance(id: number): Promise<Instance> {
  return apiFetch<Instance>(`/api/v1/instances/${id}`)
}

export async function createInstance(data: Record<string, unknown>): Promise<Instance> {
  return apiFetch<Instance>('/api/v1/instances', { method: 'POST', body: JSON.stringify(data) })
}

export async function updateInstanceStatus(id: number, status: string): Promise<Instance> {
  return apiFetch<Instance>(`/api/v1/instances/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}

export async function deleteInstance(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/instances/${id}`, { method: 'DELETE' })
}
