export interface Coin {
  id: number
  name: string
  category: Category
  denomination: string
  ruler: string
  era: string
  mint: string
  material: Material
  weightGrams: number | null
  diameterMm: number | null
  grade: string
  obverseInscription: string
  reverseInscription: string
  obverseDescription: string
  reverseDescription: string
  rarityRating: string
  purchasePrice: number | null
  currentValue: number | null
  purchaseDate: string | null
  purchaseLocation: string
  notes: string
  aiAnalysis: string
  obverseAnalysis: string
  reverseAnalysis: string
  referenceUrl: string
  referenceText: string
  isWishlist: boolean
  isSold: boolean
  soldPrice: number | null
  soldDate: string | null
  soldTo: string
  isPrivate: boolean
  userId: number
  images: CoinImage[]
  createdAt: string
  updatedAt: string
}

export interface CoinImage {
  id: number
  coinId: number
  filePath: string
  imageType: ImageType
  isPrimary: boolean
  createdAt: string
}

export type Category = 'Roman' | 'Greek' | 'Byzantine' | 'Modern' | 'Other'
export type Material = 'Gold' | 'Silver' | 'Bronze' | 'Copper' | 'Electrum' | 'Other'
export type ImageType = 'obverse' | 'reverse' | 'detail' | 'other'

export const CATEGORIES: Category[] = ['Roman', 'Greek', 'Byzantine', 'Modern', 'Other']
export const MATERIALS: Material[] = ['Gold', 'Silver', 'Bronze', 'Copper', 'Electrum', 'Other']
export const IMAGE_TYPES: ImageType[] = ['obverse', 'reverse', 'detail', 'other']

export const CATEGORY_COLORS: Record<Category, string> = {
  Roman: '#7b2d8e',
  Greek: '#6b8e23',
  Byzantine: '#8b1a1a',
  Modern: '#4682b4',
  Other: '#888888',
}

export interface User {
  id: number
  username: string
  role: 'admin' | 'user'
  email: string
  avatarPath: string
  isPublic: boolean
  bio: string
}

export interface AuthResponse {
  token: string
  refreshToken: string
  user: User
}

export interface WebAuthnCredentialInfo {
  id: number
  credentialId: string
  name: string
  createdAt: string
}

export interface CoinListResponse {
  coins: Coin[]
  total: number
  page: number
  limit: number
}

export interface StatsResponse {
  totalCoins: number
  totalWishlist: number
  byCategory: { category: string; count: number }[]
  byMaterial: { material: string; count: number }[]
  byGrade: { grade: string; count: number }[]
  values: {
    totalPurchasePrice: number
    totalCurrentValue: number
    avgPurchasePrice: number
    avgCurrentValue: number
  }
}

export interface ValueSnapshot {
  id: number
  userId: number
  totalValue: number
  totalInvested: number
  coinCount: number
  recordedAt: string
}

export interface CoinJournal {
  id: number
  coinId: number
  userId: number
  entry: string
  createdAt: string
}

export interface NumistaType {
  id: number
  title: string
  category: string
  issuer?: { name: string }
  min_year?: number
  max_year?: number
  obverse_thumbnail?: string
  reverse_thumbnail?: string
}

export interface NumistaSearchResponse {
  count: number
  types: NumistaType[]
}

export interface UserInfo {
  id: number
  username: string
  role: 'admin' | 'user'
  email: string
  avatarPath: string
  isPublic: boolean
  bio: string
  emailMissing: boolean
  createdAt: string
}

export interface AppSettings {
  OllamaURL: string
  OllamaModel: string
  ObversePrompt: string
  ReversePrompt: string
  TextExtractionPrompt: string
  OllamaTimeout: string
  LogLevel: string
  [key: string]: string
}

export type Theme = 'dark' | 'light'

export const LOG_LEVELS = ['trace', 'debug', 'info', 'warn', 'error'] as const

export interface LogEntry {
  timestamp: string
  level: string
  message: string
}

export interface ApiKey {
  id: number
  userId: number
  keyPrefix: string
  name: string
  createdAt: string
  lastUsedAt: string | null
  revokedAt: string | null
}

export interface AgentChatMessage {
  role: 'user' | 'assistant'
  content: string
}

export interface CoinSuggestion {
  name: string
  description: string
  category: string
  era: string
  ruler: string
  material: string
  denomination: string
  estPrice: string
  imageUrl: string
  sourceUrl: string
  sourceName: string
}

export interface AgentChatResponse {
  message: string
  suggestions: CoinSuggestion[]
}

export interface FollowUser {
  id: number
  username: string
  avatarPath: string
  isPublic: boolean
  bio: string
  isFollowing: boolean
  followStatus: string // '', 'pending', 'accepted', 'blocked'
  coinCount: number
  status?: string // used in followers list: 'pending' | 'accepted'
}

export interface PublicProfile extends FollowUser {
  followerCount: number
  followingCount: number
}

export interface CoinComment {
  id: number
  coinId: number
  userId: number
  username: string
  avatarPath: string
  comment: string
  rating: number
  createdAt: string
}

export interface CoinRating {
  average: number
  count: number
  userRating: number
}

export interface LimitedCoin {
  id: number
  name: string
  category: Category
  denomination: string
  ruler: string
  era: string
  material: Material
  grade: string
  images: CoinImage[]
}
