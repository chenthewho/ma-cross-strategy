import { useAuthStore } from '@/stores/authStore'
import { toast } from '@/hooks/useToast'

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
    let msg = res.statusText
    try {
      const body = await res.json()
      msg = body.error || body.message || msg
    } catch {
      msg = await res.text() || msg
    }
    toast.error(msg)
    throw new ApiRequestError(res.status, msg)
  }

  return res.json()
}
