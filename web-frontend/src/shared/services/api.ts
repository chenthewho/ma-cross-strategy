import { useAuthStore } from '@/stores/authStore'

export class ApiRequestError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiRequestError'
  }
}

export async function apiFetch<T>(url: string, options: RequestInit = {}): Promise<T> {
  const { token, clearAuth } = useAuthStore.getState()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(url, { ...options, headers })

  if (res.status === 401) {
    clearAuth()
    window.location.href = '/login'
    throw new ApiRequestError(401, 'Unauthorized')
  }

  if (!res.ok) {
    const body = await res.text()
    throw new ApiRequestError(res.status, body || res.statusText)
  }

  return res.json()
}
