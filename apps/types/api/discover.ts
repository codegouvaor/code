/**
 * API response types for the Discover page.
 *
 * These types represent the real data format returned by the backend API.
 * The frontend components currently use the `Anime` type from `./anime.ts`
 * via mock data. Once the API is ready, a mapper will convert from these
 * types to the existing `Anime` type so the UI layer stays untouched.
 *
 * @see apps/app/(public)/[locale]/discover/page.tsx
 */

/* ---------------------------------------------------------------------------
 * Image assets served from CDN
 * ------------------------------------------------------------------------- */

export interface ApiImage {
  url: string
  /** Optional width hint for responsive loading. */
  width?: number
  /** Optional height hint for responsive loading. */
  height?: number
}

export interface ApiImages {
  poster: ApiImage
  backdrop: ApiImage
  logo?: ApiImage
}

/* ---------------------------------------------------------------------------
 * Content format & status enums
 * ------------------------------------------------------------------------- */

/** The type of content — determines the URL path. */
export type ApiContentType = 'anime' | 'movie' | 'ova' | 'ona' | 'special'

/** Broadcast format — TV, theatrical, OVA, etc. */
export type ApiContentFormat = 'tv' | 'movie' | 'ova' | 'ona' | 'special' | 'tv_short'

/** Airing status. */
export type ApiContentStatus = 'airing' | 'finished' | 'upcoming' | 'hiatus' | 'cancelled'

/* ---------------------------------------------------------------------------
 * Metadata attached to every content item
 * ------------------------------------------------------------------------- */

export interface ApiContentMetadata {
  /** Genre names, e.g. ["Fantasy", "Adventure"]. */
  genres: string[]
  /** Studio name, e.g. "Madhouse". */
  studio: string
  /** Community or aggregated rating out of 10. */
  rating: number
  /** Number of ratings that contributed to the score. */
  ratingCount?: number
  /** Age rating label, e.g. "PG-13", "R". */
  ageRating?: string
  /** Year of first release. */
  year: number
  /** Japanese original title. */
  japaneseTitle?: string
  /** Short synopsis. */
  synopsis?: string
}

/* ---------------------------------------------------------------------------
 * Availability / watchability
 * ------------------------------------------------------------------------- */

export interface ApiContentAvailability {
  /** Whether the content can be watched on the platform. */
  watchable: boolean
  /** Total number of episodes available (0 for movies). */
  episodes: number
  /** Total number of seasons (0 for movies). */
  seasons?: number
}

/* ---------------------------------------------------------------------------
 * A single content item inside a section
 * ------------------------------------------------------------------------- */

export interface ApiContentItem {
  id: string
  slug: string
  title: string
  /** Determines URL path: /series/ or /movies/ */
  type: ApiContentType
  /** Broadcast format: tv, movie, ova, etc. */
  format: ApiContentFormat
  status: ApiContentStatus
  year: number
  images: ApiImages
  metadata: ApiContentMetadata
  availability: ApiContentAvailability
}

/* ---------------------------------------------------------------------------
 * Detail page — single item with its seasons & episodes
 * ------------------------------------------------------------------------- */

export interface ApiEpisode {
  id: string
  number: number
  title: string
  synopsis?: string
  thumbnailUrl: string
  /** Duration in seconds. */
  duration: number
  isSubbed: boolean
  isDubbed: boolean
}

export interface ApiSeasonDetail {
  id: string
  number: number
  title: string
  episodeCount: number
  episodes: ApiEpisode[]
}

export interface ApiContentDetailResponse {
  /** The content item header (same shape as catalog/search results). */
  item: ApiContentItem
  /** Real seasons with their episodes (empty for movies). */
  seasons: ApiSeasonDetail[]
}

/* ---------------------------------------------------------------------------
 * Section types inside the page response
 * ------------------------------------------------------------------------- */

/** Display layout hint for the section. */
export type ApiSectionType = 'carousel' | 'grid' | 'hero' | 'top10' | '继续_watching'

export interface ApiSection {
  /** Stable identifier, e.g. "trending", "new-this-week". */
  id: string
  /** Human-readable title for display. */
  title: string
  /** Visual layout hint. */
  type: ApiSectionType
  /** Optional subtitle or description. */
  subtitle?: string
  /** CTA link label, e.g. "View All". */
  ctaLabel?: string
  /** CTA link href. */
  ctaHref?: string
  /** The content items in this section. */
  items: ApiContentItem[]
}

/* ---------------------------------------------------------------------------
 * Top-level page response
 * ------------------------------------------------------------------------- */

export interface DiscoverPageResponse {
  /** Page identifier. */
  page: 'discover'
  /** ISO-8601 timestamp of the last update. */
  updatedAt: string
  /** The sections to render, in order. */
  sections: ApiSection[]
}

/* ---------------------------------------------------------------------------
 * Continue watching — separate endpoint or embedded in page response
 * ------------------------------------------------------------------------- */

export interface ApiContinueWatchingItem {
  /** The content item metadata. */
  content: ApiContentItem
  /** Episode number being watched. */
  episodeNumber: number
  /** Season number. */
  seasonNumber: number
  /** Watch progress as a percentage (0-100). */
  progressPercent: number
  /** Duration in seconds. */
  duration: number
  /** Current position in seconds. */
  currentTime: number
  /** ISO-8601 timestamp of last watch. */
  watchedAt: string
}

export interface ApiContinueWatchingResponse {
  items: ApiContinueWatchingItem[]
}
