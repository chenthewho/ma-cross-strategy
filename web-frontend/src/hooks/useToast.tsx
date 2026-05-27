import { createContext, useContext, useState, useCallback, useEffect } from 'react'
import type { ReactNode } from 'react'

export type ToastType = 'error' | 'warning' | 'success'

interface Toast {
  id: number
  message: string
  type: ToastType
}

// ── Module-level event emitter for imperative calls outside React ──

let toastId = 0
type Listener = (toast: Toast) => void
const listeners = new Set<Listener>()

export const toast = {
  push(message: string, type: ToastType = 'error') {
    const t: Toast = { id: ++toastId, message, type }
    listeners.forEach((fn) => fn(t))
  },
  error(msg: string) { toast.push(msg, 'error') },
  warning(msg: string) { toast.push(msg, 'warning') },
  success(msg: string) { toast.push(msg, 'success') },
}

// ── React context for rendering ──

const ToastContext = createContext({
  toasts: [] as Toast[],
  remove: (_id: number) => {},
})

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  useEffect(() => {
    const handler = (t: Toast) => {
      setToasts((prev) => [...prev.slice(-4), t])
      setTimeout(() => remove(t.id), 4000)
    }
    listeners.add(handler)
    return () => { listeners.delete(handler) }
  }, [remove])

  return (
    <ToastContext.Provider value={{ toasts, remove }}>
      {children}
    </ToastContext.Provider>
  )
}

export function useToast() {
  return useContext(ToastContext)
}
