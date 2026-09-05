/**
 * Shared utility for reading/writing cookies that are shared across
 * subdomains (e.g. sso.etheriatimes.localhost ↔ studios.etheriatimes.localhost).
 *
 * The domain is derived from the current hostname:
 *   sso.etheriatimes.localhost → .etheriatimes.localhost
 *   studios.etheriatimes.localhost → .etheriatimes.localhost
 *   etheriatimes.localhost → .etheriatimes.localhost
 */

function getCookieDomain(): string {
  if (typeof window === 'undefined') return '';
  const { hostname } = window.location;
  const parts = hostname.split('.');
  if (parts.length >= 3) {
    return `; domain=.${parts.slice(1).join('.')}`;
  }
  return `; domain=.${hostname}`;
}

export function setSharedCookie(name: string, value: string, maxAge: number): void {
  if (typeof document === 'undefined') return;
  const d = getCookieDomain();
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAge}${d}`;
}

export function getSharedCookie(name: string): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : null;
}

export function deleteSharedCookie(name: string): void {
  if (typeof document === 'undefined') return;
  const d = getCookieDomain();
  document.cookie = `${name}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT${d}`;
}
