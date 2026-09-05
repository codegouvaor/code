export type WorkType =
  | 'manga'
  | 'manhwa'
  | 'manhua'
  | 'webtoon'
  | 'light-novel'
  | 'novel'
  | 'comics'
  | 'bande-dessinee'
  | 'artbook'
  | 'databook'
  | 'one-shot'

export type WorkFormat =
  | 'tankobon'
  | 'digital'
  | 'hardcover'
  | 'paperback'
  | 'deluxe'
  | 'collector'
  | 'omnibus'
  | 'bunko'

export type WorkStatus = 'ongoing' | 'completed' | 'hiatus' | 'cancelled'

export type ReadingStatus = 'reading' | 'plan-to-read' | 'completed' | 'on-hold' | 'dropped'

export type DownloadStatus = 'downloaded' | 'downloading' | 'pending' | 'error'

export type SortOption =
  | 'recently-read'
  | 'title-asc'
  | 'title-desc'
  | 'rating-desc'
  | 'date-desc'
  | 'progress-desc'

export interface Genre {
  id: string
  name: string
}

export interface Author {
  id: string
  name: string
  avatarUrl?: string
}

export interface Publisher {
  id: string
  name: string
}

export interface Volume {
  id: string
  number: number
  title?: string
  coverUrl?: string
  pageCount: number
  releaseDate: string
}

export interface Chapter {
  id: string
  volumeId?: string
  number: number
  title?: string
  pageCount: number
  releaseDate: string
}

export interface Work {
  id: string
  slug: string
  type: WorkType
  title: string
  alternativeTitles: string[]
  synopsis: string
  coverUrl: string
  bannerUrl?: string
  genres: Genre[]
  authors: Author[]
  illustrators?: Author[]
  publishers: Publisher[]
  status: WorkStatus
  year: number
  rating: number
  ratingCount: number
  volumes: Volume[]
  chapters: Chapter[]
  totalChapters: number
  languages: string[]
  formats: WorkFormat[]
  tags: string[]
  themes: string[]
  collections?: CollectionRef[]
  universe?: string
  mainCharacters?: Character[]
  relatedWorks?: RelatedWork[]
}

export interface WorkSummary {
  id: string
  slug: string
  type: WorkType
  title: string
  coverUrl: string
  rating: number
  status: WorkStatus
  genres: Genre[]
}

export interface LibraryEntry {
  work: Work
  readingStatus: ReadingStatus
  progress: ReadingProgress
  isFavorite: boolean
  isDownloaded: boolean
  downloadStatus?: DownloadStatus
  rating?: number
  dateAdded: string
  lastReadAt?: string
  lastReadChapter?: number
}

export interface ReadingProgress {
  currentChapter: number
  currentPage: number
  totalPagesRead: number
  totalChaptersRead: number
  percentComplete: number
  lastReadAt: string
  estimatedTimeLeft?: string
}

export interface CollectionRef {
  id: string
  name: string
  slug: string
  type: 'universe' | 'saga' | 'franchise' | 'official' | 'deluxe' | 'collector'
}

export interface Collection {
  id: string
  slug: string
  name: string
  description: string
  coverUrl: string
  bannerUrl?: string
  type: CollectionRef['type']
  works: WorkSummary[]
  workCount: number
}

export interface RelatedWork {
  work: WorkSummary
  relation: 'sequel' | 'prequel' | 'spin-off' | 'adaptation' | 'same-universe'
}

export interface Character {
  id: string
  name: string
  imageUrl?: string
  role: 'main' | 'supporting'
}

export interface DownloadItem {
  work: WorkSummary
  chapters: Chapter[]
  status: DownloadStatus
  progress: number
  fileSize: string
  downloadedAt: string
}

export interface ReadingHistoryItem {
  work: WorkSummary
  lastChapterRead: number
  lastPageRead: number
  progressPercent: number
  readAt: string
  duration?: string
}

export interface LibrarySearchFilters {
  query: string
  types: WorkType[]
  genres: string[]
  languages: string[]
  authors: string[]
  publishers: string[]
  formats: WorkFormat[]
  status: WorkStatus[]
  readingStatus: ReadingStatus[]
  yearFrom?: number
  yearTo?: number
  sortBy: SortOption
}
