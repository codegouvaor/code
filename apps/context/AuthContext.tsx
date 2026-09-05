"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { authApi, refreshAccessToken } from "@/lib/api/auth";
import type { RegisterPayload } from "@/lib/api/auth";
import { LOGIN_ROUTE } from "@/lib/routes";
import { setSharedCookie, deleteSharedCookie } from "@/lib/shared-cookie";
import type { User } from "@/lib/api/types";
import type { PersistedSession } from "@/lib/api/session-persistence";
import {
  initSessionPersistence,
  saveSession,
  loadSession,
  clearSession,
  updateSessionTokens,
  updateSessionUser,
  getSessionPreferences,
  saveSessionPreferences,
  subscribeToSessionChanges,
  startAutoRefresh,
  migrateLegacySession,
} from "@/lib/api/session-persistence";

type AuthStatus = "loading" | "authenticated" | "unauthenticated";

interface AuthContextValue {
  user: User | null;
  accessToken: string | null;
  hasActiveSession: boolean;
  status: AuthStatus;
  isAuthenticated: boolean;
  isLoading: boolean;
  sessionPreferences: {
    rememberMe: boolean;
    autoRefresh: boolean;
    syncAcrossTabs: boolean;
  };
  login: (email: string, password: string, rememberMe?: boolean) => Promise<void>;
  register: (payload: RegisterPayload, rememberMe?: boolean) => Promise<void>;
  logout: (redirectTo?: string) => Promise<void>;
  refresh: () => Promise<string | null>;
  loadCurrentUser: () => Promise<User | null>;
  setSessionPreference: (pref: string, value: boolean) => void;
  clearSession: () => void;
}

const AuthContext = React.createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = React.useState<User | null>(null);
  const [accessToken, setAccessToken] = React.useState<string | null>(null);
  const [status, setStatus] = React.useState<AuthStatus>("loading");
  const [sessionPreferences, setSessionPreferences] = React.useState({
    rememberMe: true,
    autoRefresh: true,
    syncAcrossTabs: true,
  });

  // Initialize session persistence on mount
  React.useEffect(() => {
    initSessionPersistence();
    migrateLegacySession();
    
    // Load preferences
    const prefs = getSessionPreferences();
    setSessionPreferences({
      rememberMe: prefs.rememberMe,
      autoRefresh: prefs.autoRefresh,
      syncAcrossTabs: prefs.syncAcrossTabs,
    });
  }, []);

  // Subscribe to session changes from other tabs.
  // IMPORTANT: skip the initial synchronous callback when status is still
  // "loading" — the bootstrap effect handles the first auth decision.
  // Without this guard, navigating to a new subdomain (e.g. studios)
  // would immediately set "unauthenticated" (localStorage is per-origin)
  // before the cookie-based refresh has a chance to succeed.
  React.useEffect(() => {
    const unsubscribe = subscribeToSessionChanges((session: PersistedSession | null) => {
      if (session) {
        setAccessToken(session.accessToken);
        setUser(session.user);
        setStatus("authenticated");
      } else {
        setAccessToken(null);
        setUser(null);
        setStatus((prev) => (prev === "loading" ? prev : "unauthenticated"));
      }
    });
    
    return unsubscribe;
  }, []);

  // Setup auto-refresh
  React.useEffect(() => {
    if (!sessionPreferences.autoRefresh) return;
    
    const unsubscribe = startAutoRefresh(async () => {
      try {
        const newToken = await refreshAccessToken();
        if (newToken) {
          const session = loadSession();
          if (session) {
            updateSessionTokens(
              newToken,
              session.refreshToken,
              3600
            );
            setAccessToken(newToken);
            return true;
          }
        }
        return false;
      } catch {
        return false;
      }
    }, 60 * 1000); // Check every minute
    
    return unsubscribe;
  }, [sessionPreferences.autoRefresh]);

  const loadCurrentUser = React.useCallback(async () => {
    try {
      const nextUser = await authApi.getCurrentUser();
      setUser(nextUser);
      updateSessionUser(nextUser);
      return nextUser;
    } catch {
      setUser(null);
      return null;
    }
  }, []);

  const refresh = React.useCallback(async () => {
    try {
      const newToken = await refreshAccessToken();
      if (newToken) {
        const session = loadSession();
        if (session) {
          updateSessionTokens(
            newToken,
            session.refreshToken,
            3600
          );
          setAccessToken(newToken);
          setUser(session.user);
          updateSessionUser(session.user);
          setStatus("authenticated");
          return newToken;
        }
      }
    } catch {
      // Refresh failed
    }
    
    const session = loadSession();
    if (session) {
      setAccessToken(session.accessToken);
      setUser(session.user);
      setStatus("authenticated");
      return session.accessToken;
    }
    
    setAccessToken(null);
    setUser(null);
    setStatus("unauthenticated");
    return null;
  }, []);

  // Initial bootstrap
  React.useEffect(() => {
    let cancelled = false;

    async function bootstrap() {
      let nextUser: User | null = null;
      try {
        nextUser = await authApi.bootstrap();
      } catch {
        // Bootstrap failed (network error, invalid session, etc.)
      }
      if (cancelled) {
        return;
      }

      // authApi.bootstrap() validates the stored tokens. If it can't refresh,
      // the session is invalid and we must not render the account UI as logged-in.
      if (nextUser) {
        setAccessToken(authApi.getStoredToken());
        setUser(nextUser);
        setStatus("authenticated");
      } else {
        clearSession();
        setAccessToken(null);
        setUser(null);
        setStatus("unauthenticated");
        deleteSharedCookie('kami_sama_access_token');
        deleteSharedCookie('kami_sama_refresh');
      }
    }

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, []);

  const setSessionPreference = React.useCallback((pref: string, value: boolean) => {
    const currentPrefs = getSessionPreferences();
    const updatedPrefs = { ...currentPrefs, [pref]: value };
    saveSessionPreferences(updatedPrefs);
    setSessionPreferences({
      rememberMe: updatedPrefs.rememberMe,
      autoRefresh: updatedPrefs.autoRefresh,
      syncAcrossTabs: updatedPrefs.syncAcrossTabs,
    });
  }, []);

  const handleClearSession = React.useCallback(() => {
    clearSession();
    setAccessToken(null);
    setUser(null);
    setStatus("unauthenticated");
    
    // Clear cookies
    deleteSharedCookie('kami_sama_access_token');
    deleteSharedCookie('kami_sama_refresh');
  }, []);

  function getRedirectTarget(): string {
    if (typeof window === "undefined") return "/profile-change";
    const params = new URLSearchParams(window.location.search);
    return params.get("redirect") || "/profile-change";
  }

  const login = React.useCallback(
    async (email: string, password: string, rememberMe: boolean = true) => {
      // Save preference
      saveSessionPreferences({ rememberMe });
      setSessionPreferences(prev => ({ ...prev, rememberMe }));
      
      const response = await authApi.login({ email, password });
      
      // Save session with persistence
      saveSession(
        response.accessToken,
        response.refreshToken ?? "",
        response.user,
        response.sessionId ?? `session-${Date.now()}`,
        response.expiresIn ?? 3600
      );
      
      // Set cookies for server-side authentication
      const maxAgeAccess = response.expiresIn ?? 3600;
      if (rememberMe) {
        setSharedCookie('kami_sama_access_token', response.accessToken, maxAgeAccess);
      }
      if (response.refreshToken) {
        // 7 days for refresh token
        setSharedCookie('kami_sama_refresh', response.refreshToken, 7 * 24 * 60 * 60);
      }
      
      setAccessToken(response.accessToken);
      setUser(response.user);
      setStatus("authenticated");
      
      const target = getRedirectTarget();
      if (typeof window !== "undefined") {
        window.location.assign(target);
        return;
      }
      router.replace(target);
    },
    [router]
  );

  const register = React.useCallback(
    async (payload: RegisterPayload, rememberMe: boolean = true) => {
      // Save preference
      saveSessionPreferences({ rememberMe });
      setSessionPreferences(prev => ({ ...prev, rememberMe }));
      
      const response = await authApi.register(payload);
      
      // Save session with persistence
      saveSession(
        response.accessToken,
        response.refreshToken ?? "",
        response.user,
        response.sessionId ?? `session-${Date.now()}`,
        response.expiresIn ?? 3600
      );
      
      // Set cookies for server-side authentication
      const maxAgeAccess = response.expiresIn ?? 3600;
      if (rememberMe) {
        setSharedCookie('kami_sama_access_token', response.accessToken, maxAgeAccess);
      }
      if (response.refreshToken) {
        setSharedCookie('kami_sama_refresh', response.refreshToken, 7 * 24 * 60 * 60);
      }
      
      setAccessToken(response.accessToken);
      setUser(response.user);
      setStatus("authenticated");
      
      const target = getRedirectTarget();
      if (typeof window !== "undefined") {
        window.location.assign(target);
        return;
      }
      router.replace(target);
    },
    [router]
  );

  const logout = React.useCallback(async (redirectTo?: string) => {
    await authApi.logout();
    handleClearSession();
    
    // Clear cookies
    deleteSharedCookie('kami_sama_access_token');
    deleteSharedCookie('kami_sama_refresh');
    
    const target = redirectTo || LOGIN_ROUTE;
    router.push(target);
    router.refresh();
  }, [router, handleClearSession]);

  const value = React.useMemo<AuthContextValue>(
    () => ({
      user,
      accessToken,
      hasActiveSession: Boolean(accessToken),
      status,
      isAuthenticated: status === "authenticated",
      isLoading: status === "loading",
      sessionPreferences,
      login,
      register,
      logout,
      refresh,
      loadCurrentUser,
      setSessionPreference,
      clearSession: handleClearSession,
    }),
    [
      accessToken,
      loadCurrentUser,
      login,
      logout,
      refresh,
      register,
      status,
      user,
      sessionPreferences,
      setSessionPreference,
      handleClearSession,
    ]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

const defaultAuthValue: AuthContextValue = {
  user: null,
  accessToken: null,
  hasActiveSession: false,
  status: "unauthenticated",
  isAuthenticated: false,
  isLoading: false,
  sessionPreferences: {
    rememberMe: true,
    autoRefresh: true,
    syncAcrossTabs: true,
  },
  login: async () => {},
  register: async () => {},
  logout: async () => {},
  refresh: async () => null,
  loadCurrentUser: async () => null,
  setSessionPreference: () => {},
  clearSession: () => {},
};

export function useAuth(): AuthContextValue {
  const context = React.useContext(AuthContext);
  // When no AuthProvider is present (e.g. components rendered outside the
  // auth tree such as the GovernmentHeader), return a safe default that
  // reports the user as unauthenticated instead of throwing.
  return context ?? defaultAuthValue;
}
