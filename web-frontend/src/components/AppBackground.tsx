import { useEffect, useState } from 'react'

export default function AppBackground() {
  const [reducedMotion, setReducedMotion] = useState(false)

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    setReducedMotion(mq.matches)
    const handler = (e: MediaQueryListEvent) => setReducedMotion(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  return (
    <div className="fixed inset-0 overflow-hidden pointer-events-none" style={{ zIndex: 0 }}>
      {/* Glow orbs */}
      <div
        className="absolute -top-20 -left-20 w-[600px] h-[600px] rounded-full opacity-30"
        style={{
          background: 'radial-gradient(circle, #ff8c6b 0%, transparent 70%)',
          filter: 'blur(120px)',
          mixBlendMode: 'screen',
          animation: reducedMotion ? 'none' : 'float1 20s ease-in-out infinite',
        }}
      />
      <div
        className="absolute -bottom-20 -right-20 w-[500px] h-[500px] rounded-full opacity-25"
        style={{
          background: 'radial-gradient(circle, #0ea5e9 0%, transparent 70%)',
          filter: 'blur(140px)',
          mixBlendMode: 'screen',
          animation: reducedMotion ? 'none' : 'float2 25s ease-in-out infinite',
        }}
      />
      <div
        className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] h-[400px] rounded-full opacity-20"
        style={{
          background: 'radial-gradient(circle, #2dd4bf 0%, transparent 70%)',
          filter: 'blur(100px)',
          mixBlendMode: 'screen',
          animation: reducedMotion ? 'none' : 'float3 18s ease-in-out infinite',
        }}
      />

      {/* Floating particles */}
      {!reducedMotion && (
        <div className="absolute inset-0">
          {Array.from({ length: 20 }).map((_, i) => (
            <div
              key={i}
              className="absolute w-1 h-1 bg-white/20 rounded-full"
              style={{
                left: `${Math.random() * 100}%`,
                bottom: '0',
                animation: `particle ${5 + Math.random() * 10}s linear ${Math.random() * 10}s infinite`,
                opacity: 0,
              }}
            />
          ))}
        </div>
      )}

      {/* Geometric shapes */}
      <div
        className="absolute left-[10%] top-[30%] w-32 h-32"
        style={{
          border: '1px solid rgba(45, 212, 191, 0.15)',
          clipPath: 'polygon(50% 0%, 100% 38%, 82% 100%, 18% 100%, 0% 38%)',
          background: 'linear-gradient(135deg, rgba(45,212,191,0.05), transparent)',
          animation: reducedMotion ? 'none' : 'floatShape1 15s ease-in-out infinite',
        }}
      />
      <div
        className="absolute right-[15%] top-[50%] w-40 h-24"
        style={{
          border: '1px solid rgba(255, 140, 107, 0.12)',
          clipPath: 'polygon(25% 0%, 100% 0%, 75% 100%, 0% 100%)',
          background: 'linear-gradient(135deg, rgba(255,140,107,0.04), transparent)',
          animation: reducedMotion ? 'none' : 'floatShape2 20s ease-in-out infinite',
        }}
      />

      {/* Noise texture overlay */}
      <svg className="absolute inset-0 w-full h-full opacity-[0.15]" style={{ mixBlendMode: 'color-dodge' }}>
        <filter id="noise">
          <feTurbulence type="fractalNoise" baseFrequency="0.65" numOctaves="3" stitchTiles="stitch" />
        </filter>
        <rect width="100%" height="100%" filter="url(#noise)" />
      </svg>
    </div>
  )
}
