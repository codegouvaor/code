'use client'

import {
  createContext,
  useCallback,
  useRef,
  useState,
  type ReactNode,
} from 'react'

/* -------------------------------------------------------------------------- *
 * Public types
 * -------------------------------------------------------------------------- */

export interface CastDevice {
  id: string
  name: string
}

export interface CastMediaInfo {
  url: string
  title?: string
  subtitle?: string
  poster?: string
  contentType?: string
}

export interface CastContextValue {
  isInitialized: boolean
  isScanning: boolean
  isCasting: boolean
  castState: string
  devices: CastDevice[]
  connect: () => Promise<void>
  disconnect: () => void
  sendMedia: (media: CastMediaInfo) => void
}

export const CastContext = createContext<CastContextValue | null>(null)

/* -------------------------------------------------------------------------- *
 * Constants
 * -------------------------------------------------------------------------- */

const RECEIVER_APP_ID = 'CC1AD845'
const SDK_SRC =
  'https://www.gstatic.com/cv/js/sender/v1/cast_sender.js?loadCastFramework=1'

/* -------------------------------------------------------------------------- *
 * Provider — SDK loaded on-demand when user clicks Cast
 * -------------------------------------------------------------------------- */

export function CastProvider({ children }: { children: ReactNode }) {
  const [isInitialized, setIsInitialized] = useState(false)
  const [isScanning, setIsScanning] = useState(false)
  const [isCasting, setIsCasting] = useState(false)
  const [castState, setCastState] = useState<string>('NOT_CONNECTED')
  const [devices, setDevices] = useState<CastDevice[]>([])

  const ctxRef = useRef<cast.framework.CastContext | null>(null)
  const sessionRef = useRef<cast.framework.CastSession | null>(null)

  /* ------------------------------------------------------------------------ *
   * Load the Cast SDK and resolve when framework is ready
   * ------------------------------------------------------------------------ */
  function ensureSDK(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (typeof window === 'undefined') return reject(new Error('SSR'))

      // Already initialised
      if (ctxRef.current) return resolve()

      // Framework already in memory
      if (window.cast?.framework) {
        initCast()
        return resolve()
      }

      // Script already loading — just poll
      if (document.querySelector(`script[src="${SDK_SRC}"]`)) {
        pollFramework(resolve, reject)
        return
      }

      // Set callback BEFORE injecting script
      ;(window as any).__onGCastApiAvailable = (isAvailable: boolean) => {
        if (isAvailable) {
          initCast()
          resolve()
        } else {
          reject(new Error('Cast API not available'))
        }
      }

      const script = document.createElement('script')
      script.src = SDK_SRC
      script.async = true
      script.onerror = () => reject(new Error('Failed to load Cast SDK'))
      document.head.appendChild(script)

      // Safety timeout
      setTimeout(() => {
        if (!ctxRef.current) reject(new Error('Cast SDK timed out'))
      }, 15_000)
    })
  }

  function pollFramework(
    resolve: () => void,
    reject: (err: Error) => void,
  ) {
    let ticks = 0
    const max = 60
    const id = setInterval(() => {
      ticks++
      if (window.cast?.framework) {
        clearInterval(id)
        initCast()
        resolve()
      } else if (ticks >= max) {
        clearInterval(id)
        reject(new Error('Cast framework never loaded'))
      }
    }, 250)
  }

  /* ------------------------------------------------------------------------ *
   * Initialise CastContext singleton
   * ------------------------------------------------------------------------ */
  function initCast() {
    if (ctxRef.current) return

    const castCtx = window.cast.framework.CastContext.getInstance()

    castCtx.setOptions({
      receiverApplicationId: RECEIVER_APP_ID,
      autoJoinPolicy: (window as any).chrome?.cast?.AutoJoinPolicy?.ORIGIN_SCOPED ?? 'origin_scoped',
    })

    castCtx.addEventListener(
      window.cast.framework.CastContextEventType.CAST_STATE_CHANGED,
      (e: any) => setCastState(e.castState),
    )

    castCtx.addEventListener(
      window.cast.framework.CastContextEventType.SESSION_STATE_CHANGED,
      (e: any) => {
        const s = castCtx.getCurrentSession()
        sessionRef.current = s

        setIsCasting(
          s !== null &&
            e.sessionState ===
              window.cast.framework.SessionState.SESSION_STARTED,
        )

        if (
          e.sessionState ===
            window.cast.framework.SessionState.SESSION_ENDED ||
          e.sessionState === 'SESSION_ENDED'
        ) {
          setDevices([])
          setIsCasting(false)
        }
      },
    )

    setCastState(castCtx.getCastState())

    const existing = castCtx.getCurrentSession()
    if (existing) {
      sessionRef.current = existing
      setIsCasting(true)
    }

    ctxRef.current = castCtx
    setIsInitialized(true)
  }

  /* ------------------------------------------------------------------------ *
   * Connect
   * ------------------------------------------------------------------------ */
  const connect = useCallback(async () => {
    setIsScanning(true)
    try {
      await ensureSDK()

      const castCtx = ctxRef.current
      if (!castCtx) return

      const prev = castCtx.getCurrentSession()
      if (prev) castCtx.endCurrentSession(true)

      await castCtx.requestSession()

      const session = castCtx.getCurrentSession()
      sessionRef.current = session
      setIsCasting(session !== null)
    } catch (err: any) {
      const msg = typeof err === 'string' ? err : err?.message ?? ''
      if (err?.code !== 2001 && !msg.includes('cancel')) {
        console.error('[cast]', err)
      }
    } finally {
      setIsScanning(false)
    }
  }, [])

  /* ------------------------------------------------------------------------ *
   * Disconnect
   * ------------------------------------------------------------------------ */
  const disconnect = useCallback(() => {
    ctxRef.current?.endCurrentSession(true)
    sessionRef.current = null
    setIsCasting(false)
    setDevices([])
  }, [])

  /* ------------------------------------------------------------------------ *
   * Send media
   * ------------------------------------------------------------------------ */
  const sendMedia = useCallback((media: CastMediaInfo) => {
    const session = sessionRef.current
    if (!session) return

    const mime =
      media.contentType ??
      (media.url.includes('.m3u8')
        ? 'application/x-mpegURL'
        : 'video/mp4')

    const info = new window.cast.framework.MediaInformation(media.url, mime)
    info.streamType = 'BUFFERED'
    info.metadata = {
      title: media.title ?? '',
      subtitle: media.subtitle ?? '',
      images: media.poster ? [{ url: media.poster }] : [],
    }

    const request = new window.cast.framework.LoadRequestData(info)
    request.autoplay = true
    request.currentTime = 0

    session.loadMedia(request).catch((e: any) => console.error('[cast]', e))
  }, [])

  /* ------------------------------------------------------------------------ *
   * Value
   * ------------------------------------------------------------------------ */
  const value: CastContextValue = {
    isInitialized,
    isScanning,
    isCasting,
    castState,
    devices,
    connect,
    disconnect,
    sendMedia,
  }

  return <CastContext.Provider value={value}>{children}</CastContext.Provider>
}
