interface StatusBadgeProps {
  status: string
}

const statusMap: Record<string, { label: string; dotColor: string; textColor: string }> = {
  running: { label: '运行中', dotColor: 'bg-[#34d399]', textColor: 'text-[#34d399]' },
  stopped: { label: '已停止', dotColor: 'bg-[#94a3b8]', textColor: 'text-[#94a3b8]' },
  error: { label: '异常', dotColor: 'bg-[#f87171]', textColor: 'text-[#f87171]' },
  halted: { label: '已中断', dotColor: 'bg-[#f87171]', textColor: 'text-[#f87171]' },
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  // Normalize to lowercase for case-insensitive matching
  const key = (status || '').toLowerCase()
  const s = statusMap[key] || statusMap.stopped
  return (
    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${s.textColor} bg-white/[0.03] border border-white/[0.06]`}>
      <span className={`inline-block w-1.5 h-1.5 rounded-full ${s.dotColor}`} />
      {s.label}
    </span>
  )
}
