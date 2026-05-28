import { apiFetch } from './api';

export interface EvolutionTask {
  id: number;
  strategy_id: string;
  status: string;
  created_at: string;
  progress?: { generation: number; max_generations: number; best_score: number };
  result?: { score_total: number };
}

export interface Challenger {
  id: number;
  role: string;
  score_total: number;
  max_drawdown: number;
  score_6m: number;
  score_2y: number;
  score_5y: number;
  created_at: string;
}

export async function fetchEvolutionTasks(): Promise<{
  tasks: EvolutionTask[];
  challengers: Challenger[];
}> {
  const res = await apiFetch<{ tasks: EvolutionTask[]; challengers: Challenger[] }>(
    '/api/v1/evolution/tasks',
  );
  return {
    tasks: Array.isArray(res?.tasks) ? res.tasks : [],
    challengers: Array.isArray(res?.challengers) ? res.challengers : [],
  };
}

export async function createEvolutionTask(data: {
  strategy_id: string;
  pop_size?: number;
  max_generations?: number;
}) {
  return apiFetch<any>('/api/v1/evolution/tasks', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function promoteEvolutionTask(id: number) {
  return apiFetch<any>(`/api/v1/evolution/tasks/${id}/promote`, {
    method: 'POST',
  });
}
