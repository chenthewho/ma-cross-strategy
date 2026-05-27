interface StatusBadgeProps {
  status: string
}

const statusMap: Record<string, { label: string; dotColor: string; textColor: string; bgColor: string }> = {
  running: { label: '运行中', dotColor: 'bg-claude-success', textColor: 'text-claude-success', bgColor: 'bg-claude-success-light' },
  stopped: { label: '已停止', dotColor: 'bg-claude-text-muted', textColor: 'text-claude-text-secondary', bgColor: 'bg-claude-border' },
  error:   { label: '异常',   dotColor: 'bg-claude-danger', textColor: 'text-claude-danger', bgColor: 'bg-claude-danger-light' },
  halted:  { label: '已中断', dotColor: 'bg-claude-danger', textColor: 'text-claude-danger', bgColor: 'bg-claude-danger-light' },
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  const key = (status || '').toLowerCase()
  const s = statusMap[key] || statusMap.stopped
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${s.textColor} ${s.bgColor}`}>
      <span className={`inline-block w-1.5 h-1.5 rounded-full ${s.dotColor}`} />
      {s.label}
    </span>
  )
}
