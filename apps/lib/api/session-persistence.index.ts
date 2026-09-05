/**
 * Session Persistence API
 * 
 * This module provides enhanced session persistence capabilities including:
 * - Multiple storage fallback mechanisms (localStorage, sessionStorage, inMemory)
 * - Cross-tab synchronization using BroadcastChannel
 * - Session preferences management (remember me, auto-refresh, etc.)
 * - Automatic session refresh monitoring
 * - Activity tracking and inactivity timeout
 * - Session validation and expiration handling
 * - Legacy session migration
 */

export {
  // Core session persistence functions
  initSessionPersistence,
  saveSession,
  loadSession,
  clearSession,
  updateSessionTokens,
  updateSessionUser,
  
  // Session preferences
  saveSessionPreferences,
  getSessionPreferences,
  clearSessionPreferences,
  
  // Session listeners and synchronization
  subscribeToSessionChanges,
  
  // Session validation
  isSessionValid,
  getSessionRemainingTime,
  shouldRefreshSession,
  
  // Automatic session management
  startAutoRefresh,
  
  // Cleanup utilities
  cleanupExpiredSessions,
  migrateLegacySession,
  
  // Enhanced token provider
  enhancedTokenProvider,
  
  // Types
  type PersistedSession,
  type SessionPreferences,
  type SessionListener,
} from "./session-persistence";

// Re-export default
export { default } from "./session-persistence";
