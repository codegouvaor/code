import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import createIntlMiddleware from "next-intl/middleware";
import { routing } from "./i18n/routing";

/**
 * Locale-aware middleware of the Next.js app (`apps/` is the Next project
 * root, so this file MUST live here — a middleware at the repository root is
 * never picked up by `next dev`/`next build`).
 *
 * Routing principle:
 *  - each product lives on its own domain (SSO, studios, public portal);
 *  - on the public portal domain, locale routing is delegated to the
 *    next-intl middleware: `/` is negotiated to `/fr` (or `/en`), locale
 *    prefixed paths like `/fr/…` pass through untouched and the request
 *    locale is announced via the `x-next-intl-locale` header so every
 *    server component resolves the right messages.
 */
const IS_DEVELOPMENT = process.env.NODE_ENV === "development";

type Locale = (typeof routing.locales)[number];

const handleI18nRouting = createIntlMiddleware(routing);

/* -------------------------------------------------------------------------- *
 * Domain ↔ route group mapping
 * -------------------------------------------------------------------------- */

type DomainGroup = "sso" | "studios" | "main";

const SSO_HOSTS = ["sso.gouv.localhost", "sso.gouv.lan"];
const STUDIOS_HOSTS = ["studios.gouv.localhost", "studios.gouv.lan"];
const MAIN_HOSTS = [
  "code.gouv.localhost",
  "code.gouv.lan",
  "code.gouv.aor",
  "www.code.gouv.aor",
];

const AUTH_PATHS = [
  "/login",
  "/register",
  "/profile-change",
  "/mfa-validate",
  "/mfa-setup",
  "/mfa-verify",
  "/mfa-recovery",
  "/mfa-recovery-setup",
  "/mfa-recovery-verify",
];
const PLATFORM_PATHS = ["/dash"];

function detectGroup(host: string): DomainGroup {
  const hostname = host.split(":")[0];
  if (SSO_HOSTS.includes(hostname)) return "sso";
  if (STUDIOS_HOSTS.includes(hostname)) return "studios";
  return "main";
}

function getDomainForGroup(group: DomainGroup, currentUrl: URL): string {
  const hostname = currentUrl.hostname;
  const protocol = currentUrl.protocol;

  if (IS_DEVELOPMENT) {
    switch (group) {
      case "sso":
        return `${protocol}//sso.gouv.localhost`;
      case "studios":
        return `${protocol}//studios.gouv.localhost`;
      case "main":
        return `${protocol}//code.gouv.localhost`;
    }
  }

  switch (group) {
    case "sso":
      return `${protocol}//sso.gouv.localhost`;
    case "studios":
      return `${protocol}//studios.gouv.localhost`;
    case "main":
      return `${protocol}//${hostname}`;
  }
}

function getTargetGroup(pathname: string): DomainGroup | null {
  if (AUTH_PATHS.some((p) => pathname === p || pathname.startsWith(p + "/"))) {
    return "sso";
  }
  if (PLATFORM_PATHS.some((p) => pathname === p || pathname.startsWith(p + "/"))) {
    return "studios";
  }
  return null;
}

/* -------------------------------------------------------------------------- *
 * Auth helpers
 * -------------------------------------------------------------------------- */

const REFRESH_COOKIE = "kami_sama_refresh";
const ACCESS_TOKEN_COOKIE = "kami_sama_access_token";

function isAuthCookiePresent(request: NextRequest): boolean {
  const refresh = request.cookies.get(REFRESH_COOKIE);
  const access = request.cookies.get(ACCESS_TOKEN_COOKIE);
  return Boolean(
    (refresh?.value && refresh.value.length > 0) ||
      (access?.value && access.value.length > 0)
  );
}

function getAccessToken(request: NextRequest): string | null {
  return request.cookies.get(ACCESS_TOKEN_COOKIE)?.value || null;
}

function hasAdminAccess(request: NextRequest): boolean {
  try {
    const token = getAccessToken(request);
    if (!token) return false;
    const payload = JSON.parse(Buffer.from(token.split(".")[1], "base64").toString());
    const roles: string[] = payload.roles || [];
    return roles.includes("admin") || roles.includes("superadmin") || roles.includes("owner");
  } catch {
    return false;
  }
}

/* -------------------------------------------------------------------------- *
 * Locale helpers
 * -------------------------------------------------------------------------- */

function isValidLocale(segment: string): segment is Locale {
  return routing.locales.includes(segment as Locale);
}

/** Strips a leading locale segment, e.g. `/fr/login` → `/login`. */
function stripLocalePrefix(pathname: string, localePath: string): string {
  const cleanPath = pathname.replace(localePath, "");
  return cleanPath === "" ? "/" : cleanPath;
}

/* -------------------------------------------------------------------------- *
 * Middleware entry
 * -------------------------------------------------------------------------- */

export default function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const host = request.headers.get("host") || request.nextUrl.hostname;
  const currentGroup = detectGroup(host);

  /* ---- Root path (no locale prefix yet) ---- */
  if (pathname === "/" || pathname === "") {
    switch (currentGroup) {
      case "sso":
        return NextResponse.redirect(new URL("/login", request.url));
      case "studios":
        return NextResponse.redirect(new URL("/dash", request.url));
      case "main":
      default:
        // Logged-in users of the platform are sent to their profile area.
        if (isAuthCookiePresent(request)) {
          return NextResponse.redirect(new URL("/profile-change", request.url));
        }
        // Locale negotiation and redirect to /fr or /en.
        return handleI18nRouting(request);
    }
  }

  /* ---- Cross-domain routing (unprefixed auth/platform paths) ---- */
  const targetGroup = getTargetGroup(pathname);
  if (targetGroup && targetGroup !== currentGroup) {
    return NextResponse.redirect(new URL(pathname, getDomainForGroup(targetGroup, request.nextUrl)));
  }

  /* ---- SSO domain: auth routes only ---- */
  if (currentGroup === "sso") {
    return NextResponse.next();
  }

  /* ---- Studios domain: platform routes only ---- */
  if (currentGroup === "studios") {
    if (PLATFORM_PATHS.some((p) => pathname === p || pathname.startsWith(p + "/"))) {
      if (IS_DEVELOPMENT) return NextResponse.next();
      if (isAuthCookiePresent(request) && hasAdminAccess(request)) return NextResponse.next();
      return NextResponse.redirect(new URL("/dash", request.url));
    }
    return NextResponse.redirect(new URL("/dash", request.url));
  }

  /* ---- Main (portal) domain ---- */

  const firstSegment = pathname.split("/").filter(Boolean)[0];

  // Locale-prefixed auth/platform routes belong to the SSO/studios domains:
  // strip the locale and let the cross-domain rules above route them.
  if (firstSegment && isValidLocale(firstSegment)) {
    const localePath = `/${firstSegment}`;

    if (AUTH_PATHS.some((p) => pathname.startsWith(localePath + p) || pathname === localePath + p)) {
      return NextResponse.redirect(
        new URL(stripLocalePrefix(pathname, localePath), request.url)
      );
    }
    if (PLATFORM_PATHS.some((p) => pathname.startsWith(localePath + p) || pathname === localePath + p)) {
      return NextResponse.redirect(
        new URL(stripLocalePrefix(pathname, localePath), request.url)
      );
    }
  }

  // Everything else on the portal domain is handled by next-intl: it keeps
  // `/fr/…`/`/en/…` requests untouched, redirects unprefixed public paths to
  // the negotiated locale and sets the request-locale header used by the
  // server components.
  return handleI18nRouting(request);
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|.*\\..*|health).*)"],
};
