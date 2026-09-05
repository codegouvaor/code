/**
 * API models for the The Etheria Times platform.
 * All data is expected to come from the external API at api.etheriatimes.com.
 * These interfaces describe the shape of that data.
 */

export type AnimeStatus = 'airing' | 'completed' | 'upcoming' | 'hiatus'
export type ContentType = 'series' | 'movies'

export interface Genre {
  id: string
  name: string
}

export interface Studio {
  id: string
  name: string
}

export interface Character {
  id: string
  name: string
  japaneseName?: string
  role: 'main' | 'supporting'
  image: string
  voiceActor?: string
}

export interface SubtitleTrack {
  /** BCP-47 language code, e.g. `fr`, `en`, `pt-BR`. */
  lang: string
  /** Human-readable label shown in the captions menu, e.g. `Français`. */
  label: string
  /** URL to a WebVTT (.vtt) sidecar file. */
  src: string
  /** Mark this track as the default selection. */
  default?: boolean
}

export interface Episode {
  id: string
  animeId: string
  season: number
  number: number
  title: string
  description?: string
  thumbnail: string
  cover: string
  videoUrl: string
  /** Caption / subtitle sidecar tracks, in display order. */
  tracks: SubtitleTrack[]
  /** Duration in seconds. */
  duration: number
  releaseDate: string
  /** Watch progress in seconds for the current user, if any. */
  progress?: number
}

export interface Season {
  number: number
  title: string
  episodeCount: number
  year: number
}

export interface Review {
  id: string
  user: UserSummary
  rating: number
  title: string
  body: string
  createdAt: string
  likes: number
}

export interface UserSummary {
  id: string
  username: string
  displayName: string
  avatar: string
}

export interface Anime {
  id: string
  slug: string
  type: ContentType
  title: string
  japaneseTitle: string
  synopsis: string
  cover: string
  banner: string
  genres: Genre[]
  studio: Studio
  year: number
  status: AnimeStatus
  /** Community rating out of 10. */
  rating: number
  /** Number of ratings that contributed to the score. */
  ratingCount: number
  ageRating: string
  seasons: Season[]
  totalEpisodes: number
}

export interface ContinueWatchingItem {
  anime: Anime
  episode: Episode
  /** 0-100 progress percentage. */
  progressPercent: number
  remainingLabel: string
}

export interface TrendingItem {
  anime: Anime
  rank: number
  /** Weekly change in rank. */
  trend: number
}

export interface RecentlyAddedItem {
  anime: Anime
  episode: Episode
}

export interface CommunityPick {
  anime: Anime
  votes: number
  communityRating: number
}

export interface Recommendation {
  anime: Anime
  reason: string
  detail: string
}

export interface SimulcastItem {
  anime: Anime
  nextEpisode?: Episode
  airDay: string
  nextAirDate: string
}

export interface SeasonalPick {
  anime: Anime
  label: string
}

export interface CommunityActivity {
  id: string
  type: 'review' | 'discussion' | 'rating' | 'request'
  user: UserSummary
  anime?: Pick<Anime, 'id' | 'slug' | 'title' | 'cover'>
  content: string
  createdAt: string
  meta?: string
}
