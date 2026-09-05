'use client';

import { authApi } from "./auth";
import type { User } from "./types";

/**
 * Enhanced session persistence service with multiple fallback mechanisms
 * and automatic synchronization across tabs.
 */

// Storage keys
const SESSION_STORAGE_KEY = "etheriatimes.session";
const SESSION_PREFERENCES_KEY = "etheriatimes.session.prefs";
const SESSION_VERSION = "v2";

// Session state interface
export interface PersistedSession {
  version: string;
  accessToken: string;
  refreshToken: string;
  user: User;
  sessionId: string;
  expiresAt: number; // timestamp
  createdAt: number; // timestamp
  lastSyncedAt: number; // timestamp
}

// Session preferences interface
export interface SessionPreferences {
  rememberMe: boolean;
  autoRefresh: boolean;
  syncAcrossTabs: boolean;
  lastActivityAt: number;
}

// Session listener type
export type SessionListener = (session: PersistedSession | null) => void;

// Default preferences
const DEFAULT_PREFERENCES: SessionPreferences = {
  rememberMe: true,
  autoRefresh: true,
  syncAcrossTabs: true,
  lastActivityAt: Date.now(),
};

// Storage providers priority (from most to least reliable)
type StorageProvider = {
  name: string;
  getItem: (key: string) => string | null;
  setItem: (key: string, value: string) => void;
  removeItem: (key: string) => void;
  available: () => boolean;
};

// Available storage providers
const storageProviders: StorageProvider[] = [
  {
    name: "localStorage",
    getItem: (key: string) => {
      try {
        return typeof window !== "undefined" ? window.localStorage.getItem(key) : null;
      } catch {
        return null;
      }
    },
    setItem: (key: string, value: string) => {
      try {
        if (typeof window !== "undefined") {
          window.localStorage.setItem(key, value);
        }
      } catch {
        // Ignore quota exceeded errors
      }
    },
    removeItem: (key: string) => {
      try {
        if (typeof window !== "undefined") {
          window.localStorage.removeItem(key);
        }
      } catch {
        // Ignore errors
      }
    },
    available: () => {
      try {
        return typeof window !== "undefined" && window.localStorage !== undefined;
      } catch {
        return false;
      }
    },
  },
  {
    name: "sessionStorage",
    getItem: (key: string) => {
      try {
        return typeof window !== "undefined" ? window.sessionStorage.getItem(key) : null;
      } catch {
        return null;
      }
    },
    setItem: (key: string, value: string) => {
      try {
        if (typeof window !== "undefined") {
          window.sessionStorage.setItem(key, value);
        }
      } catch {
        // Ignore quota exceeded errors
      }
    },
    removeItem: (key: string) => {
      try {
        if (typeof window !== "undefined") {
          window.sessionStorage.removeItem(key);
        }
      } catch {
        // Ignore errors
      }
    },
    available: () => {
      try {
        return typeof window !== "undefined" && window.sessionStorage !== undefined;
      } catch {
        return false;
      }
    },
  },
  {
    name: "inMemory",
    getItem: (key: string) => {
      if (typeof window === "undefined") return null;
      try {
        const store = (window as any).__etheriatimesSessionStore;
        return store?.[key] ?? null;
      } catch {
        return null;
      }
    },
    setItem: (key: string, value: string) => {
      if (typeof window === "undefined") return;
      try {
        if (!(window as any).__etheriatimesSessionStore) {
          (window as any).__etheriatimesSessionStore = {};
        }
        (window as any).__etheriatimesSessionStore[key] = value;
      } catch {
        // Ignore errors
      }
    },
    removeItem: (key: string) => {
      if (typeof window === "undefined") return;
      try {
        const store = (window as any).__etheriatimesSessionStore;
        if (store) {
          delete store[key];
        }
      } catch {
        // Ignore errors
      }
    },
    available: () => typeof window !== "undefined",
  },
];

// Broadcast channel for cross-tab synchronization
let broadcastChannel: BroadcastChannel | null = null;
const SYNC_CHANNEL = "etheriatimes-session-sync";

// Session state listeners
const sessionListeners = new Set<SessionListener>();

// Activity tracking
let lastActivityTime = Date.now();
let inactivityTimeout: ReturnType<typeof setTimeout> | null = null;

/**
 * Get the best available storage provider
 */
function getBestStorageProvider(): StorageProvider {
  for (const provider of storageProviders) {
    if (provider.available()) {
      return provider;
    }
  }
  return storageProviders[storageProviders.length - 1]; // fallback to inMemory
}

/**
 * Try to get value from all providers in priority order
 */
function getFromAllProviders(key: string): string | null {
  for (const provider of storageProviders) {
    if (provider.available()) {
      const value = provider.getItem(key);
      if (value !== null) {
        return value;
      }
    }
  }
  return null;
}

/**
 * Write value to all available providers
 */
function writeToAllProviders(key: string, value: string): void {
  for (const provider of storageProviders) {
    if (provider.available()) {
      try {
        provider.setItem(key, value);
      } catch {
        // Continue with next provider
      }
    }
  }
}

/**
 * Remove value from all available providers
 */
function removeFromAllProviders(key: string): void {
  for (const provider of storageProviders) {
    if (provider.available()) {
      try {
        provider.removeItem(key);
      } catch {
        // Continue with next provider
      }
    }
  }
}

/**
 * Initialize cross-tab synchronization
 */
function initCrossTabSync(): void {
  if (typeof window === "undefined" || !window.BroadcastChannel) return;
  
  try {
    broadcastChannel = new BroadcastChannel(SYNC_CHANNEL);
    
    broadcastChannel.addEventListener("message", (event) => {
      if (event.data?.type === "SESSION_SYNC") {
        const session: PersistedSession | null = event.data.session;
        
        // Update local state
        if (session) {
          writeToAllProviders(SESSION_STORAGE_KEY, JSON.stringify(session));
        } else {
          removeFromAllProviders(SESSION_STORAGE_KEY);
        }
        
        // Notify listeners
        notifySessionListeners(session);
      }
    });
  } catch {
    // BroadcastChannel not supported
  }
}

/**
 * Broadcast session state to other tabs
 */
function broadcastSessionState(session: PersistedSession | null): void {
  if (!broadcastChannel) return;
  
  try {
    broadcastChannel.postMessage({
      type: "SESSION_SYNC",
      session,
      timestamp: Date.now(),
    });
  } catch {
    // Ignore broadcast errors
  }
}

/**
 * Notify all session listeners
 */
function notifySessionListeners(session: PersistedSession | null): void {
  sessionListeners.forEach((listener) => {
    try {
      listener(session);
    } catch {
      // Ignore listener errors
    }
  });
}

/**
 * Track user activity for session timeout
 */
function trackActivity(): void {
  lastActivityTime = Date.now();
  
  // Reset inactivity timeout if it exists
  if (inactivityTimeout) {
    clearTimeout(inactivityTimeout);
  }
}

/**
 * Start inactivity monitoring
 */
function startInactivityMonitoring(inactivityDuration: number = 30 * 60 * 1000): void {
  // 30 minutes default inactivity duration
  if (inactivityTimeout) {
    clearTimeout(inactivityTimeout);
  }
  
  inactivityTimeout = setTimeout(() => {
    const prefs = getSessionPreferences();
    if (!prefs.rememberMe) {
      // Clear session on inactivity if rememberMe is false
      clearSession();
    }
  }, inactivityDuration);
}

/**
 * Stop inactivity monitoring
 */
function stopInactivityMonitoring(): void {
  if (inactivityTimeout) {
    clearTimeout(inactivityTimeout);
    inactivityTimeout = null;
  }
}

// ============================================================================
// Public API
// ============================================================================

/**
 * Initialize session persistence service
 */
export function initSessionPersistence(): void {
  initCrossTabSync();
  
  // Setup activity tracking
  if (typeof window !== "undefined") {
    const events = ["mousedown", "keydown", "scroll", "touchstart", "click"];
    events.forEach((event) => {
      window.addEventListener(event, trackActivity, { passive: true });
    });
  }
}

/**
 * Save session with all metadata
 */
export function saveSession(
  accessToken: string,
  refreshToken: string,
  user: User,
  sessionId: string,
  expiresIn: number
): void {
  const session: PersistedSession = {
    version: SESSION_VERSION,
    accessToken,
    refreshToken,
    user,
    sessionId,
    expiresAt: Date.now() + expiresIn * 1000,
    createdAt: Date.now(),
    lastSyncedAt: Date.now(),
  };
  
  // Save to all available storage providers
  writeToAllProviders(SESSION_STORAGE_KEY, JSON.stringify(session));
  
  // Broadcast to other tabs
  broadcastSessionState(session);
  
  // Notify listeners
  notifySessionListeners(session);
  
  // Start inactivity monitoring
  startInactivityMonitoring();
}

/**
 * Load persisted session
 */
export function loadSession(): PersistedSession | null {
  const raw = getFromAllProviders(SESSION_STORAGE_KEY);
  
  if (!raw) {
    return null;
  }
  
  try {
    const session: PersistedSession = JSON.parse(raw);
    
    // Validate session version
    if (session.version !== SESSION_VERSION) {
      // Migrate or clear old session
      clearSession();
      return null;
    }
    
    // Check if session is expired
    if (session.expiresAt && session.expiresAt < Date.now()) {
      clearSession();
      return null;
    }
    
    // Update last synced time
    session.lastSyncedAt = Date.now();
    writeToAllProviders(SESSION_STORAGE_KEY, JSON.stringify(session));
    
    return session;
  } catch {
    return null;
  }
}

/**
 * Clear session from all storage providers
 */
export function clearSession(): void {
  removeFromAllProviders(SESSION_STORAGE_KEY);
  broadcastSessionState(null);
  notifySessionListeners(null);
  stopInactivityMonitoring();
}

/**
 * Update session tokens (e.g., after refresh)
 */
export function updateSessionTokens(
  accessToken: string,
  refreshToken: string,
  expiresIn: number
): void {
  const session = loadSession();
  
  if (!session) {
    return;
  }
  
  session.accessToken = accessToken;
  session.refreshToken = refreshToken;
  session.expiresAt = Date.now() + expiresIn * 1000;
  session.lastSyncedAt = Date.now();
  
  writeToAllProviders(SESSION_STORAGE_KEY, JSON.stringify(session));
  broadcastSessionState(session);
  notifySessionListeners(session);
}

/**
 * Update user data in session
 */
export function updateSessionUser(user: User): void {
  const session = loadSession();
  
  if (!session) {
    return;
  }
  
  session.user = user;
  session.lastSyncedAt = Date.now();
  
  writeToAllProviders(SESSION_STORAGE_KEY, JSON.stringify(session));
  broadcastSessionState(session);
  notifySessionListeners(session);
}

// ============================================================================
// Session Preferences
// ============================================================================

/**
 * Save session preferences
 */
export function saveSessionPreferences(prefs: Partial<SessionPreferences>): void {
  const current = getSessionPreferences();
  const updated: SessionPreferences = {
    ...current,
    ...prefs,
    lastActivityAt: Date.now(),
  };
  
  writeToAllProviders(SESSION_PREFERENCES_KEY, JSON.stringify(updated));
}

/**
 * Get session preferences
 */
export function getSessionPreferences(): SessionPreferences {
  const raw = getFromAllProviders(SESSION_PREFERENCES_KEY);
  
  if (!raw) {
    return { ...DEFAULT_PREFERENCES };
  }
  
  try {
    return { ...DEFAULT_PREFERENCES, ...JSON.parse(raw) };
  } catch {
    return { ...DEFAULT_PREFERENCES };
  }
}

/**
 * Clear session preferences
 */
export function clearSessionPreferences(): void {
  removeFromAllProviders(SESSION_PREFERENCES_KEY);
}

// ============================================================================
// Session Listeners
// ============================================================================

/**
 * Subscribe to session changes
 */
export function subscribeToSessionChanges(listener: SessionListener): () => void {
  sessionListeners.add(listener);
  
  // Send current session to new listener
  const session = loadSession();
  listener(session);
  
  return () => {
    sessionListeners.delete(listener);
  };
}

// ============================================================================
// Session Validation
// ============================================================================

/**
 * Check if session is valid and not expired
 */
export function isSessionValid(): boolean {
  const session = loadSession();
  
  if (!session) {
    return false;
  }
  
  // Check expiration
  if (session.expiresAt && session.expiresAt < Date.now()) {
    return false;
  }
  
  return true;
}

/**
 * Get remaining session time in milliseconds
 */
export function getSessionRemainingTime(): number {
  const session = loadSession();
  
  if (!session || !session.expiresAt) {
    return 0;
  }
  
  const remaining = session.expiresAt - Date.now();
  return Math.max(0, remaining);
}

/**
 * Check if session needs refresh (less than 5 minutes remaining)
 */
export function shouldRefreshSession(): boolean {
  const remaining = getSessionRemainingTime();
  return remaining > 0 && remaining < 5 * 60 * 1000; // 5 minutes
}

// ============================================================================
// Automatic Session Management
// ============================================================================

/**
 * Start automatic session refresh monitoring
 */
export function startAutoRefresh(
  refreshFn: () => Promise<boolean>,
  intervalMs: number = 60 * 1000
): () => void {
  let refreshInterval: ReturnType<typeof setInterval> | null = null;
  
  const checkAndRefresh = async () => {
    const prefs = getSessionPreferences();
    
    if (!prefs.autoRefresh) {
      return;
    }
    
    if (shouldRefreshSession()) {
      try {
        const success = await refreshFn();
        if (!success) {
          // Refresh failed, clear session
          clearSession();
        }
      } catch {
        // Refresh error, clear session
        clearSession();
      }
    }
  };
  
  // Initial check
  void checkAndRefresh();
  
  // Setup interval
  refreshInterval = setInterval(checkAndRefresh, intervalMs);
  
  return () => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  };
}

// ============================================================================
// Cleanup Utilities
// ============================================================================

/**
 * Clean up expired sessions from storage
 */
export function cleanupExpiredSessions(): void {
  const session = loadSession();
  
  if (session && session.expiresAt < Date.now()) {
    clearSession();
  }
}

/**
 * Migrate old session format to new format
 */
export function migrateLegacySession(): void {
  // Check for old session format
  const oldAccessToken = getFromAllProviders("aether.account.accessToken");
  const oldRefreshToken = getFromAllProviders("aether.account.refreshToken");
  const oldUser = getFromAllProviders("aether.account.user");
  
  if (oldAccessToken && oldRefreshToken && oldUser) {
    try {
      const user: User = JSON.parse(oldUser);
      const session: PersistedSession = {
        version: SESSION_VERSION,
        accessToken: oldAccessToken,
        refreshToken: oldRefreshToken,
        user,
        sessionId: "legacy-" + Date.now(),
        expiresAt: Date.now() + 24 * 60 * 60 * 1000, // 24 hours
        createdAt: Date.now(),
        lastSyncedAt: Date.now(),
      };
      
      // Save new format
      writeToAllProviders(SESSION_STORAGE_KEY, JSON.stringify(session));
      
      // Remove old format
      removeFromAllProviders("aether.account.accessToken");
      removeFromAllProviders("aether.account.refreshToken");
      removeFromAllProviders("aether.account.user");
      
    } catch {
      // Migration failed, keep old format
    }
  }
}

// ============================================================================
// Integration with existing auth system
// ============================================================================

/**
 * Enhanced token provider that integrates with session persistence
 */
export const enhancedTokenProvider = {
  async getAccessToken(): Promise<string | null> {
    const session = loadSession();
    return session?.accessToken ?? null;
  },
  
  async refreshAccessToken(): Promise<string | null> {
    try {
      const session = loadSession();
      if (!session?.refreshToken) {
        return null;
      }
      
      const response = await authApi.login({ 
        email: session.user.email, 
        password: "" 
      });
      
      if (response?.accessToken) {
        updateSessionTokens(
          response.accessToken,
          response.refreshToken ?? session.refreshToken,
          response.expiresIn ?? 3600
        );
        return response.accessToken;
      }
      
      return null;
    } catch {
      clearSession();
      return null;
    }
  },
};

export default {
  initSessionPersistence,
  saveSession,
  loadSession,
  clearSession,
  updateSessionTokens,
  updateSessionUser,
  saveSessionPreferences,
  getSessionPreferences,
  clearSessionPreferences,
  subscribeToSessionChanges,
  isSessionValid,
  getSessionRemainingTime,
  shouldRefreshSession,
  startAutoRefresh,
  cleanupExpiredSessions,
  migrateLegacySession,
  enhancedTokenProvider,
};
