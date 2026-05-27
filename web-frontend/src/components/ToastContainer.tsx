import { XCircle, AlertTriangle, CheckCircle, X } from 'lucide-react'
import { useToast } from '@/hooks/useToast'

const iconMap = {
  error: XCircle,
  warning: AlertTriangle,
  success: CheckCircle,
}

const colorMap = {
  error: 'border-[#f87171]/30 bg-[#f87171]/10 text-[#fca5a5]',
  warning: 'border-[#fbbf24]/30 bg-[#fbbf24]/10 text-[#fcd34d]',
  success: 'border-[#34d399]/30 bg-[#34d399]/10 text-[#6ee7b7]',
}

export default function ToastContainer() {
  const { toasts, remove } = useToast()

  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 max-w-[90vw] sm:max-w-sm pointer-events-none">
      {toasts.map((t) => {
        const Icon = iconMap[t.type]
        return (
          <div
            key={t.id}
            className={`pointer-events-auto flex items-start gap-2 px-3 py-2.5 rounded-lg border backdrop-blur-md text-sm animate-toast-in ${colorMap[t.type]}`}
          >
            <Icon className="w-4 h-4 shrink-0 mt-0.5" />
            <span className="flex-1 break-words leading-snug">{t.message}</span>
            <button onClick={() => remove(t.id)} className="shrink-0 opacity-60 hover:opacity-100 transition-opacity">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )
      })}
    </div>
  )
}
