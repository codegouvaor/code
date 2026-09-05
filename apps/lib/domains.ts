export type Environment = 'production' | 'localhost'

export interface DomainConfig {
  main: string
  studios: string
  sso: string
  protocol: string
}

const DOMAINS: Record<Environment, DomainConfig> = {
  production: {
    main: 'code.gouv.aor',
    studios: 'studios.gouv.aor',
    sso: 'sso.gouv.aor',
    protocol: 'https',
  },
  localhost: {
    main: 'code.gouv.localhost',
    studios: 'studios.gouv.localhost',
    sso: 'sso.gouv.localhost',
    protocol: 'http',
  },
}

export function detectEnvironment(): Environment {
  if (typeof window === 'undefined') {
    return process.env.NODE_ENV === 'production' ? 'production' : 'localhost'
  }

  return window.location.hostname.includes('localhost') ? 'localhost' : 'production'
}

export function getDomainConfig(): DomainConfig {
  return DOMAINS[detectEnvironment()]
}

export function getDomainUrl(service: 'main' | 'studios' | 'sso', path: string = ''): string {
  const config = getDomainConfig()
  return `${config.protocol}://${config[service]}${path}`
}

export function switchDomain(target: 'main' | 'studios' | 'sso', path: string): string {
  return getDomainUrl(target, path)
}
