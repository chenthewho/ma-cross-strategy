import { apiFetch } from './api'

// type-only
export interface EquitySnapshot {
  recorded_at: string
  total_equity: number
}

export async function fetchEquitySnapshots(instanceId: number) {
  const res = await apiFetch<{ snapshots: EquitySnapshot[] }>(`/api/v1/dashboard/equity-snapshots?instance_id=${instanceId}`)
  return res.snapshots || []
}
