import { apiFetch } from './api'

export async function fetchEvolutionTasks() {
  return apiFetch<any[]>('/api/v1/evolution/tasks')
}

export async function createEvolutionTask(data: Record<string, unknown>) {
  return apiFetch<any>('/api/v1/evolution/tasks', { method: 'POST', body: JSON.stringify(data) })
}

export async function fetchGenomes() {
  return apiFetch<any[]>('/api/v1/evolution/genomes')
}
