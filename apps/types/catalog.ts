export type CatalogType =
  | 'anime'
  | 'manga'
  | 'manhwa'
  | 'manhua'
  | 'webtoon'
  | 'light-novel'
  | 'novel'
  | 'film'
  | 'ova'
  | 'jeu'

export type CatalogStatus = 'airing' | 'completed' | 'upcoming' | 'hiatus' | 'cancelled'

export type CatalogSort = 'popular' | 'rating' | 'favorites' | 'newest' | 'release' | 'alpha'

export type CatalogView = 'grid' | 'list'

export type CatalogGenre =
  | 'action'
  | 'aventure'
  | 'romance'
  | 'fantasy'
  | 'science-fiction'
  | 'horreur'
  | 'comedie'
  | 'drame'
  | 'mystere'
  | 'thriller'
  | 'sport'
  | 'slice-of-life'
  | 'musique'
  | 'mecha'
  | 'isekai'
  | 'psychologique'
  | 'historique'
  | 'militaire'
  | 'nature'
  | 'policier'

export interface CatalogItem {
  id: string
  slug: string
  title: string
  japaneseTitle?: string
  coverUrl: string
  bannerUrl?: string
  score: number
  type: CatalogType
  genres: CatalogGenre[]
  year: number
  status: CatalogStatus
  studio?: string
  author?: string
  episodes?: number
  chapters?: number
  volumes?: number
  favorites: number
  synopsis: string
  ageRating?: string
  language?: string
}

export interface CatalogFilter {
  type?: CatalogType
  genre?: CatalogGenre
  status?: CatalogStatus
  year?: number
  studio?: string
  sort?: CatalogSort
  search?: string
  view?: CatalogView
  page?: number
  pageSize?: number
}

export interface CatalogQuery {
  type?: string
  genre?: string
  status?: string
  year?: number
  studio?: string
  sort?: string
  search?: string
  page?: number
  pageSize?: number
}

export interface CatalogResponse {
  total: number
  items: CatalogItem[]
  page: number
  pageSize: number
  totalPages: number
}

export interface CatalogGenreInfo {
  id: CatalogGenre
  label: string
  count: number
}

export interface CatalogTypeInfo {
  id: CatalogType
  label: string
  count: number
}
