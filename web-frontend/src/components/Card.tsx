import type { ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
  onClick?: () => void
}

export default function Card({ children, className = '', onClick }: CardProps) {
  return (
    <div
      className={`border border-white/[0.04] bg-slate-900/20 backdrop-blur rounded-xl ${onClick ? 'cursor-pointer hover:border-white/[0.08]' : ''} ${className}`}
      onClick={onClick}
    >
      {children}
    </div>
  )
}
