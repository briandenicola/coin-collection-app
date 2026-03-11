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
  referenceUrl: string
  referenceText: string
  isWishlist: boolean
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
}

export interface AuthResponse {
  token: string
  user: User
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
  values: {
    totalPurchasePrice: number
    totalCurrentValue: number
    avgPurchasePrice: number
    avgCurrentValue: number
  }
}

export interface UserInfo {
  id: number
  username: string
  role: 'admin' | 'user'
  createdAt: string
}

export interface AppSettings {
  OllamaURL: string
  OllamaModel: string
  AiAnalysisPrompt: string
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
