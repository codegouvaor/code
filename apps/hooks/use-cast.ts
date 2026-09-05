'use client'

import { useContext } from 'react'
import { CastContext, type CastContextValue } from '@/context/CastContext'

export function useCast(): CastContextValue {
  const ctx = useContext(CastContext)
  if (!ctx) {
    throw new Error('useCast must be used within a <CastProvider>')
  }
  return ctx
}
