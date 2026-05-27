import { apiFetch } from './api'

export interface Instance {
  id: number
  name: string
  symbol: string
  template_id: string
  status: string
  total_equity?: number
  cny_balance?: number
  dead_hold?: number
  float_hold?: number
  initial_capital?: number
  created_at: string
}

export async function fetchInstances(): Promise<Instance[]> {
  const res = await apiFetch<{ instances: Instance[] }>('/api/v1/instances')
  return res.instances || []
}

export async function fetchInstance(id: number): Promise<Instance> {
  const instances = await fetchInstances()
  const inst = instances.find((i) => i.id === id)
  if (!inst) throw new Error(`Instance ${id} not found`)
  return inst
}

export async function createInstance(data: Record<string, unknown>): Promise<Instance> {
  const res = await apiFetch<{ instance: Instance }>('/api/v1/instances', { method: 'POST', body: JSON.stringify(data) })
  return res.instance
}

export async function updateInstanceStatus(id: number, status: string): Promise<any> {
  const action = status === 'running' ? 'start' : 'stop'
  return apiFetch<any>(`/api/v1/instances/${id}/${action}`, { method: 'POST' })
}

export async function deleteInstance(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/instances/${id}`, { method: 'DELETE' })
}

export interface TradeRecord {
  id: number
  instance_id: number
  action: string
  engine: string
  symbol: string
  filled_qty: number
  filled_price: number
  fee: number
  lot_type: string
  created_at: string
}

export async function fetchTrades(instanceId: number): Promise<TradeRecord[]> {
  const res = await apiFetch<{ trades: TradeRecord[] }>(`/api/v1/instances/${instanceId}/trades`)
  return res.trades || []
}
