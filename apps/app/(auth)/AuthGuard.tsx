"use client";

import * as React from "react";
import { usePathname } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { RouteTransition } from "@/components/route-transition";
import { getDomainUrl } from "@/lib/domains";

// Routes accessible to authenticated users within (auth)
const AUTHENTICATED_ALLOWED_ROUTES = ["/mfa-validate", "/mfa-setup", "/callback", "/verify-email"];
// Routes accessible without authentication within (auth)
const PUBLIC_AUTH_ROUTES = ["/login", "/register"];

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { isAuthenticated, isLoading } = useAuth();

  const isAllowedAuthenticatedRoute = AUTHENTICATED_ALLOWED_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  const isPublicAuthRoute = PUBLIC_AUTH_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  React.useEffect(() => {
    if (isLoading) return;

    // Unauthenticated users: only allow public auth routes (login)
    if (!isAuthenticated && !isPublicAuthRoute) {
      window.location.href = getDomainUrl("sso", "/login");
      return;
    }

    // Authenticated users on non-auth routes go back to the calendar.
    if (isAuthenticated && !isAllowedAuthenticatedRoute) {
      window.location.href = getDomainUrl("info", "/");
    }
  }, [isAuthenticated, isLoading, isAllowedAuthenticatedRoute, isPublicAuthRoute]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-sm text-muted-foreground">
        Vérification de la session…
      </div>
    );
  }

  // Block unauthenticated users from non-public routes
  if (!isAuthenticated && !isPublicAuthRoute) {
    return null;
  }

  // Block authenticated users from non-allowed routes
  if (isAuthenticated && !isAllowedAuthenticatedRoute) {
    return null;
  }

  return <RouteTransition>{children}</RouteTransition>;
}
