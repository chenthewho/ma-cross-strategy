export default function CardSkeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse bg-claude-border rounded-xl h-32 ${className}`} />
}
