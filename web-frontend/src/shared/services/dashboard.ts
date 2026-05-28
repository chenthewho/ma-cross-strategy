import { apiFetch } from './api'

// type-only
export interface EquitySnapshot {
  recorded_at: string
  total_equity: number
}

export interface KlinePoint {
  open_time: number
  close: number
}

export interface TradeMarker {
  created_at: string
  price: number
  action: 'BUY' | 'SELL'
  engine: string
  qty: number
}

export interface PriceChartData {
  klines: KlinePoint[]
  trades: TradeMarker[]
  avg_buy_price: number
}

export async function fetchEquitySnapshots(instanceId: number) {
  const res = await apiFetch<{ snapshots: EquitySnapshot[] }>(`/api/v1/dashboard/equity-snapshots?instance_id=${instanceId}`)
  return res.snapshots || []
}

export async function fetchPriceChart(instanceId: number) {
  const res = await apiFetch<PriceChartData>(`/api/v1/instances/${instanceId}/price-chart`)
  return res
}
