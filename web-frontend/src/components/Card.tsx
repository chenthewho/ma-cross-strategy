import type { ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
  onClick?: () => void
}

export default function Card({ children, className = '', onClick }: CardProps) {
  return (
    <div
      className={`bg-claude-surface border border-claude-border rounded-xl shadow-sm
        ${onClick ? 'cursor-pointer hover:border-claude-border-hover hover:shadow-md transition-shadow' : ''}
        ${className}`}
      onClick={onClick}
    >
      {children}
    </div>
  )
}
