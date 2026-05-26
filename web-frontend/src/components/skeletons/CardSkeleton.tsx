export default function CardSkeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse bg-slate-800/40 rounded-xl h-32 ${className}`} />
}
