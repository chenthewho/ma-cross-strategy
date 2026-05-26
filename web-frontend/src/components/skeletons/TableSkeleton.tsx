export default function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="animate-pulse bg-slate-800/40 rounded-lg h-12" />
      ))}
    </div>
  )
}
