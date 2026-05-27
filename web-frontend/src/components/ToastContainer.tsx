import { useToast } from '@/hooks/useToast'

export default function ToastContainer() {
  const { toasts } = useToast()

  if (toasts.length === 0) return null

  const typeStyles: Record<string, string> = {
    success: 'bg-claude-success text-white',
    error: 'bg-claude-danger text-white',
    warning: 'bg-claude-warning text-white',
  }

  return (
    <div className="fixed top-4 right-4 z-[100] flex flex-col gap-2 max-w-sm">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`px-4 py-3 rounded-lg shadow-lg text-sm font-medium animate-toast-in ${typeStyles[t.type] || typeStyles.error}`}
        >
          {t.message}
        </div>
      ))}
    </div>
  )
}
