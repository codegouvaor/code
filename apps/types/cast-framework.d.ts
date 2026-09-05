/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * Minimal ambient declarations for the Google Cast Sender SDK.
 * The SDK is loaded at runtime via `cast_sender.js` and exposes itself
 * on `window.cast` (and optionally `window.chrome.cast`).
 */

// ── cast.framework ──────────────────────────────────────────────────────────
declare namespace cast {
  namespace framework {
    /** Singleton entry-point for the Cast framework. */
    class CastContext {
      static getInstance(): CastContext
      setOptions(options: Record<string, any>): void
      getCastState(): CastState
      getCurrentSession(): CastSession | null
      requestSession(): Promise<void>
      endCurrentSession(stopCasting?: boolean): void
      addEventListener(
        event: string,
        handler: (evt: any) => void,
      ): void
      removeEventListener(
        event: string,
        handler: (evt: any) => void,
      ): void
    }

    class CastSession {
      getSessionState(): SessionState
      isMediaSessionActive(): boolean
      getCastDevice(): CastDevice
      loadMedia(request: LoadRequestData): Promise<void>
      sendMessage(
        ns: string,
        data: any,
        onSuccess?: () => void,
        onError?: (err: any) => void,
      ): void
      addEventListener(
        event: string,
        handler: (evt: any) => void,
      ): void
      removeEventListener(
        event: string,
        handler: (evt: any) => void,
      ): void
    }

    class CastDevice {
      friendlyName: string
      deviceId: string
    }

    class RemotePlayerController {
      player: RemotePlayer
      addEventListener(event: string, handler: (evt: any) => void): void
      removeEventListener(event: string, handler: (evt: any) => void): void
      playOrPause(): void
      seek(): void
      setVolumeLevel(): void
      stop(): void
    }

    class RemotePlayer {
      currentTime: number
      duration: number
      isPaused: boolean
      isConnected: boolean
      mediaInfo: any
      addEventListener(event: string, handler: (evt: any) => void): void
      removeEventListener(event: string, handler: (evt: any) => void): void
    }

    /** Data object passed to `CastSession.loadMedia()`. */
    class LoadRequestData {
      media: cast.framework.MediaInformation
      autoplay: boolean
      currentTime: number
      constructor(mediaInfo: cast.framework.MediaInformation)
    }

    /** Describes the media to be played on the receiver. */
    class MediaInformation {
      contentId: string
      contentType: string
      metadata: any
      streamType: string
      constructor(contentId: string, contentType: string)
    }

    // ── event-name constants ────────────────────────────────────────────────
    const CastContextEventType: {
      CAST_STATE_CHANGED: string
      SESSION_STATE_CHANGED: string
    }

    const SessionState: {
      SESSION_ENDED: SessionState
      SESSION_STARTING: SessionState
      SESSION_STARTED: SessionState
      SESSION_RESUMED: SessionState
      SESSION_ENDING: SessionState
    }

    // ── enums ───────────────────────────────────────────────────────────────
    enum AutoJoinPolicy {
      TAB_AND_SCAN_SCOPE = 'tab_and_scan_scope',
      PAGE_SCOPED = 'page_scoped',
      ORIGIN_SCOPED = 'origin_scoped',
    }

    // ── union types for state values ────────────────────────────────────────
    type CastState =
      | 'NO_DEVICES_AVAILABLE'
      | 'NOT_CONNECTED'
      | 'CONNECTING'
      | 'CONNECTED'

    type SessionState =
      | 'SESSION_ENDED'
      | 'SESSION_STARTING'
      | 'SESSION_STARTED'
      | 'SESSION_RESUMED'
      | 'SESSION_ENDING'
  }
}

/**
 * Augment the global Window so `window.cast.framework.*` is typed.
 */
interface Window {
  cast: {
    framework: typeof cast.framework
  }
}
