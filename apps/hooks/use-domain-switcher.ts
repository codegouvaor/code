'use client'

import { useMemo } from 'react'
import { usePathname } from 'next/navigation'
import { detectEnvironment, getDomainUrl, type Environment } from '@/lib/domains'

export interface DomainLink {
  label: string
  href: string
  isCurrent: boolean
  environment: Environment
}

export function useDomainSwitcher() {
  const pathname = usePathname()

  const currentEnvironment = useMemo(() => detectEnvironment(), [])

  const links = useMemo<DomainLink[]>(() => {
    const isProd = currentEnvironment === 'production'

    return [
      {
        label: 'Production',
        href: getDomainUrl('main', pathname),
        isCurrent: isProd,
        environment: 'production',
      },
      {
        label: 'Local',
        href: getDomainUrl('main', pathname),
        isCurrent: !isProd,
        environment: 'localhost',
      },
      {
        label: 'Studios',
        href: getDomainUrl('studios', pathname),
        isCurrent: false,
        environment: currentEnvironment,
      },
      {
        label: 'SSO',
        href: getDomainUrl('sso', pathname),
        isCurrent: false,
        environment: currentEnvironment,
      },
    ]
  }, [currentEnvironment, pathname])

  return { currentEnvironment, links }
}
