import { apiFetch } from './api'

export async function loginAPI(email: string, password: string) {
  return apiFetch<{ token: string; user: { email: string; role: string } }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function registerAPI(email: string, password: string) {
  return apiFetch<{ token: string; user: { email: string; role: string } }>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function verifyToken() {
  return apiFetch<{ email: string; role: string }>('/api/v1/auth/me')
}
