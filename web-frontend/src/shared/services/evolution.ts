import { apiFetch } from './api'

export interface EvolutionTask {
  id: number
  status: string
  progress?: { generation: number; max_generations: number; best_score: number }
  result?: { score_total: number }
}

export async function fetchEvolutionTasks(): Promise<EvolutionTask[]> {
  const res = await apiFetch<{ tasks: EvolutionTask[] }>('/api/v1/evolution/tasks')
  return res.tasks || []
}

export async function createEvolutionTask(data: Record<string, unknown>) {
  const res = await apiFetch<{ task: any }>('/api/v1/evolution/tasks', { method: 'POST', body: JSON.stringify(data) })
  return res.task
}

export async function fetchGenomes() {
  const res = await apiFetch<{ challengers: any[] }>('/api/v1/genome/challengers')
  return res.challengers || []
}
