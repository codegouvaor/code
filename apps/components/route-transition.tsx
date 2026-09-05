'use client'

import * as React from 'react'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'

interface RouteTransitionProps {
  children: React.ReactNode
  className?: string
}

export function RouteTransition({ children, className }: RouteTransitionProps) {
  const pathname = usePathname()
  const [progress, setProgress] = React.useState(0)
  const [isVisible, setIsVisible] = React.useState(false)
  const prevPathname = React.useRef(pathname)
  const intervalRef = React.useRef<ReturnType<typeof setInterval> | null>(null)

  React.useEffect(() => {
    if (prevPathname.current === pathname) return

    prevPathname.current = pathname
    setIsVisible(true)
    setProgress(0)

    let current = 0
    intervalRef.current = setInterval(() => {
      current += Math.random() * 12 + 4
      if (current >= 90) {
        if (intervalRef.current) clearInterval(intervalRef.current)
        setProgress(90)
        return
      }
      setProgress(current)
    }, 100)

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [pathname])

  React.useEffect(() => {
    if (!isVisible) return

    if (intervalRef.current) clearInterval(intervalRef.current)

    setProgress(100)
    const timer = setTimeout(() => {
      setIsVisible(false)
      setProgress(0)
    }, 400)

    return () => clearTimeout(timer)
  }, [children, isVisible])

  return (
    <div className={cn('relative min-h-full', className)}>
      {/* Progress bar - Crunchyroll style */}
      <div
        className={cn(
          'fixed top-0 left-0 right-0 z-100 transition-opacity duration-300',
          isVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'
        )}
      >
        {/* Background track */}
        <div className="h-1 w-full bg-line/50" />
        
        {/* Animated bar */}
        <div
          className="absolute top-0 left-0 h-1 transition-all duration-150 ease-out"
          style={{
            width: `${progress}%`,
            background: 'linear-gradient(90deg, #F47521 0%, #ff8c42 50%, #F47521 100%)',
            boxShadow: '0 0 12px rgba(244, 117, 33, 0.8), 0 0 24px rgba(244, 117, 33, 0.4), 0 2px 8px rgba(244, 117, 33, 0.6)',
          }}
        />
        
        {/* Shine effect */}
        <div
          className="absolute top-0 left-0 h-1 w-32 opacity-60"
          style={{
            width: `${progress}%`,
            background: 'linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.4) 50%, transparent 100%)',
            animation: 'shimmer 1.5s infinite',
          }}
        />
      </div>

      {/* Content */}
      {children}
    </div>
  )
}
