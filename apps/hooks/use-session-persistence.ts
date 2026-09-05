'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  loadSession,
  saveSession,
  clearSession,
  updateSessionTokens,
  updateSessionUser,
  getSessionPreferences,
  saveSessionPreferences,
  subscribeToSessionChanges,
  isSessionValid,
  getSessionRemainingTime,
  shouldRefreshSession,
  startAutoRefresh,
  cleanupExpiredSessions,
  type PersistedSession,
  type SessionPreferences,
} from '@/lib/api/session-persistence';
import type { User } from '@/lib/api/types';

/**
 * Custom hook for managing session persistence with enhanced features
 */
export function useSessionPersistence() {
  const [session, setSession] = useState<PersistedSession | null>(null);
  const [preferences, setPreferences] = useState<SessionPreferences | null>(null);
  const [isValid, setIsValid] = useState<boolean>(false);
  const [remainingTime, setRemainingTime] = useState<number>(0);

  // Load initial state
  useEffect(() => {
    const currentSession = loadSession();
    const currentPrefs = getSessionPreferences();
    
    setSession(currentSession);
    setPreferences(currentPrefs);
    setIsValid(isSessionValid());
    setRemainingTime(getSessionRemainingTime());
  }, []);

  // Subscribe to session changes
  useEffect(() => {
    const unsubscribe = subscribeToSessionChanges((newSession: PersistedSession | null) => {
      setSession(newSession);
      setIsValid(isSessionValid());
      setRemainingTime(getSessionRemainingTime());
    });
    
    return unsubscribe;
  }, []);

  // Setup auto-refresh for remaining time
  useEffect(() => {
    const interval = setInterval(() => {
      setRemainingTime(getSessionRemainingTime());
      setIsValid(isSessionValid());
    }, 1000); // Update every second
    
    return () => clearInterval(interval);
  }, []);

  // Save session
  const save = useCallback((
    accessToken: string,
    refreshToken: string,
    user: User,
    sessionId: string,
    expiresIn: number
  ) => {
    saveSession(accessToken, refreshToken, user, sessionId, expiresIn);
    setSession(loadSession());
  }, []);

  // Clear session
  const clear = useCallback(() => {
    clearSession();
    setSession(null);
    setIsValid(false);
    setRemainingTime(0);
  }, []);

  // Update tokens
  const updateTokens = useCallback((
    accessToken: string,
    refreshToken: string,
    expiresIn: number
  ) => {
    updateSessionTokens(accessToken, refreshToken, expiresIn);
    setSession(loadSession());
  }, []);

  // Update user
  const updateUser = useCallback((user: User) => {
    updateSessionUser(user);
    setSession(loadSession());
  }, []);

  // Update preferences
  const updatePreferences = useCallback((prefs: Partial<SessionPreferences>) => {
    saveSessionPreferences(prefs);
    setPreferences(getSessionPreferences());
  }, []);

  // Start auto-refresh
  const startRefresh = useCallback((
    refreshFn: () => Promise<boolean>,
    intervalMs?: number
  ) => {
    return startAutoRefresh(refreshFn, intervalMs);
  }, []);

  // Cleanup expired sessions
  const cleanup = useCallback(() => {
    cleanupExpiredSessions();
    setSession(loadSession());
  }, []);

  return {
    // State
    session,
    preferences,
    isValid,
    remainingTime,
    shouldRefresh: shouldRefreshSession(),
    
    // Actions
    save,
    clear,
    updateTokens,
    updateUser,
    updatePreferences,
    startRefresh,
    cleanup,
    
    // Utilities
    getSession: loadSession,
    getPreferences: getSessionPreferences,
  };
}

/**
 * Hook for managing session timeout warnings
 */
export function useSessionTimeoutWarning(
  warningThreshold: number = 5 * 60 * 1000, // 5 minutes
  criticalThreshold: number = 2 * 60 * 1000, // 2 minutes
  onWarning?: () => void,
  onCritical?: () => void,
  onExpired?: () => void
) {
  const [warningType, setWarningType] = useState<'none' | 'warning' | 'critical' | 'expired'>('none');

  useEffect(() => {
    const checkSession = () => {
      const remaining = getSessionRemainingTime();
      const isValid = isSessionValid();
      
      if (!isValid) {
        setWarningType('expired');
        onExpired?.();
      } else if (remaining < criticalThreshold) {
        setWarningType('critical');
        onCritical?.();
      } else if (remaining < warningThreshold) {
        setWarningType('warning');
        onWarning?.();
      } else {
        setWarningType('none');
      }
    };

    // Check immediately
    checkSession();
    
    // Setup interval
    const interval = setInterval(checkSession, 1000);
    
    // Subscribe to session changes
    const unsubscribe = subscribeToSessionChanges(() => {
      checkSession();
    });
    
    return () => {
      clearInterval(interval);
      unsubscribe();
    };
  }, [warningThreshold, criticalThreshold, onWarning, onCritical, onExpired]);

  return warningType;
}

/**
 * Hook for session activity tracking
 */
export function useSessionActivity(
  inactivityDuration: number = 30 * 60 * 1000, // 30 minutes
  onInactive?: () => void
) {
  useEffect(() => {
    let timeout: ReturnType<typeof setTimeout> | null = null;
    
    const resetTimeout = () => {
      if (timeout) clearTimeout(timeout);
      
      timeout = setTimeout(() => {
        const prefs = getSessionPreferences();
        if (!prefs.rememberMe) {
          onInactive?.();
        }
      }, inactivityDuration);
    };
    
    // Setup event listeners
    const events = ['mousedown', 'keydown', 'scroll', 'touchstart', 'click'];
    events.forEach((event) => {
      window.addEventListener(event, resetTimeout, { passive: true });
    });
    
    // Initial setup
    resetTimeout();
    
    return () => {
      if (timeout) clearTimeout(timeout);
      events.forEach((event) => {
        window.removeEventListener(event, resetTimeout);
      });
    };
  }, [inactivityDuration, onInactive]);
}

export default useSessionPersistence;
