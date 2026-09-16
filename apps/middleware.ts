import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { routing, type Locale } from "./i18n/routing";

/**
 * Locale-aware middleware of the Next.js app (`apps/` is the Next project
 * root, so this file MUST live here — a middleware at the repository root is
 * never picked up by `next dev`/`next build`).
 *
 * Routing principle:
 *  - the whole application is served from a single root hostname
 *    (`code.gouv.localhost` in development, `code.gouv.aor` in production):
 *    public portal, platform and auth all live on the same app;
 *  - the public routes live under a top-level `[locale]` segment and the
 *    platform is nested under that same segment, so there is exactly one
 *    root-level dynamic segment and no route-group collision;
 *  - unprefixed pathnames (`/`, `/contact`, `/owner/repo`, `/dashboard`, …)
 *    are REWRITTEN to the negotiated locale (`/{locale}/…`) — the URL stays
 *    untouched while Next.js serves the localized route internally;
 *  - locale is announced through the `x-next-intl-locale` header so every
 *    server component resolves the right messages, and persisted in the
 *    `NEXT_LOCALE` cookie.
 */
const LOCALE_COOKIE = "NEXT_LOCALE";

// Auth routes live at the root (separate `(auth)` route group) and must NOT
// be locale-prefixed, otherwise they would be rewritten to `/fr/login` which
// does not exist. Health is excluded via the matcher below.
const AUTH_PATHS = [
  "/login",
  "/profile-change",
  "/mfa-validate",
  "/mfa-setup",
  "/mfa-verify",
  "/mfa-recovery",
  "/mfa-recovery-setup",
  "/mfa-recovery-verify",
  "/callback",
  "/verify-email",
];

function isAuthPath(pathname: string): boolean {
  return AUTH_PATHS.some((p) => pathname === p || pathname.startsWith(p + "/"));
}

/** Negotiates a locale from the cookie, then the Accept-Language header. */
function detectLocale(request: NextRequest): Locale {
  const cookie = request.cookies.get(LOCALE_COOKIE)?.value;
  if (cookie && (routing.locales as string[]).includes(cookie)) {
    return cookie as Locale;
  }
  const acceptLanguage = request.headers.get("accept-language") || "";
  for (const locale of routing.locales) {
    if (acceptLanguage.toLowerCase().includes(locale)) return locale;
  }
  return routing.defaultLocale;
}

function persistLocale(response: NextResponse, locale: string): void {
  response.cookies.set(LOCALE_COOKIE, locale, {
    path: "/",
    httpOnly: false,
    sameSite: "lax",
    maxAge: 60 * 60 * 24 * 365,
  });
}

export default function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const firstSegment = pathname.split("/").filter(Boolean)[0];

  // Auth routes: serve directly at the root, just negotiate the locale.
  if (isAuthPath(pathname)) {
    const response = NextResponse.next();
    response.headers.set("x-next-intl-locale", detectLocale(request));
    persistLocale(response, detectLocale(request));
    return response;
  }

  // Already locale-prefixed (`/fr/…`, `/en/…`): pass through untouched.
  if (firstSegment && (routing.locales as string[]).includes(firstSegment)) {
    const response = NextResponse.next();
    response.headers.set("x-next-intl-locale", firstSegment);
    persistLocale(response, firstSegment);
    return response;
  }

  // Unprefixed path: rewrite to the negotiated locale while keeping the URL.
  const locale = detectLocale(request);
  const url = request.nextUrl.clone();
  url.pathname = `/${locale}${pathname === "/" ? "" : pathname}`;
  if (search) url.search = search;
  const response = NextResponse.rewrite(url);
  response.headers.set("x-next-intl-locale", locale);
  persistLocale(response, locale);
  return response;
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|.*\\..*|health).*)"],
};