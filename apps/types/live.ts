export type LiveStatus = 'live' | 'upcoming' | 'ended'

export type LiveCategory =
  | 'anime'
  | 'studios'
  | 'concerts'
  | 'gaming'
  | 'vtuber'
  | 'interviews'
  | 'events'
  | 'news'
  | 'creation'

export interface LiveChannel {
  id: string
  slug: string
  name: string
  avatarUrl: string
  bannerUrl?: string
  description: string
  followerCount: number
  isVerified: boolean
  category: LiveCategory
}

export interface LiveStream {
  id: string
  slug: string
  title: string
  description: string
  thumbnailUrl: string
  bannerUrl?: string
  channel: LiveChannel
  status: LiveStatus
  category: LiveCategory
  viewerCount: number
  peakViewerCount: number
  duration?: string
  startedAt?: string
  scheduledAt?: string
  endedAt?: string
  tags: string[]
  isClip?: boolean
}

export interface LiveEvent {
  id: string
  slug: string
  title: string
  description: string
  thumbnailUrl: string
  channel: LiveChannel
  scheduledAt: string
  category: LiveCategory
  reminderCount: number
  isReminderSet?: boolean
  tags: string[]
}

export interface LiveCategoryInfo {
  id: LiveCategory
  label: string
  icon: string
  streamCount: number
}

export interface LiveRadioChannel {
  id: string
  slug: string
  name: string
  logoUrl: string
  currentShow: string
  isLive: boolean
  category: string
}

export interface LiveHeroItem {
  stream: LiveStream
  headline: string
  subheadline: string
}
