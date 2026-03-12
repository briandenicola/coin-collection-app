import axios from 'axios'
import type { Coin, CoinListResponse, CoinImage, AuthResponse, StatsResponse, UserInfo, AppSettings, LogEntry } from '@/types'

const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

const api = axios.create({
  baseURL: `${API_BASE}/api`,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

// Auth
export const checkSetup = () => api.get<{ needsSetup: boolean }>('/auth/setup')
export const login = (username: string, password: string) =>
  api.post<AuthResponse>('/auth/login', { username, password })
export const register = (username: string, password: string) =>
  api.post<AuthResponse>('/auth/register', { username, password })

// Coins
export const getCoins = (params?: {
  category?: string
  search?: string
  wishlist?: string
  page?: number
  limit?: number
}) => api.get<CoinListResponse>('/coins', { params })

const NULLABLE_FIELDS = ['weightGrams', 'diameterMm', 'purchasePrice', 'currentValue', 'purchaseDate']

function sanitizeCoin(coin: Partial<Coin>): Partial<Coin> {
  const clean = { ...coin }
  for (const field of NULLABLE_FIELDS) {
    if ((clean as any)[field] === '' || (clean as any)[field] === undefined) {
      ;(clean as any)[field] = null
    }
  }
  return clean
}

export const getCoin = (id: number) => api.get<Coin>(`/coins/${id}`)
export const createCoin = (coin: Partial<Coin>) => api.post<Coin>('/coins', sanitizeCoin(coin))
export const updateCoin = (id: number, coin: Partial<Coin>) => api.put<Coin>(`/coins/${id}`, sanitizeCoin(coin))
export const deleteCoin = (id: number) => api.delete(`/coins/${id}`)

// Images
export const uploadImage = (coinId: number, file: File, imageType: string, isPrimary: boolean) => {
  const formData = new FormData()
  formData.append('image', file)
  formData.append('imageType', imageType)
  formData.append('isPrimary', String(isPrimary))
  return api.post<CoinImage>(`/coins/${coinId}/images`, formData)
}
export const deleteImage = (coinId: number, imageId: number) =>
  api.delete(`/coins/${coinId}/images/${imageId}`)

// Analysis
export const analyzeCoin = (coinId: number, side?: 'obverse' | 'reverse') => {
  const params = side ? `?side=${side}` : ''
  return api.post<{ analysis: string; coin: Coin }>(`/coins/${coinId}/analyze${params}`)
}
export const deleteAnalysis = (coinId: number, side: 'obverse' | 'reverse') =>
  api.delete<{ coin: Coin }>(`/coins/${coinId}/analyze?side=${side}`)
export const extractText = (file: File) => {
  const formData = new FormData()
  formData.append('image', file)
  return api.post<{ text: string }>('/extract-text', formData)
}
export const getOllamaStatus = () =>
  api.get<{ available: boolean; model: string; url: string; message: string }>('/ollama-status')

// Stats
export const getStats = () => api.get<StatsResponse>('/stats')

// Autocomplete suggestions
export const getSuggestions = (field: string, q: string) =>
  api.get<string[]>('/suggestions', { params: { field, q } })

// User self-service
export const getMe = () => api.get<UserInfo>('/auth/me')
export const changePassword = (currentPassword: string, newPassword: string) =>
  api.post('/auth/change-password', { currentPassword, newPassword })
export const exportCollection = () => api.get<Coin[]>('/user/export')
export const importCollection = (coins: Coin[]) => api.post('/user/import', coins)

// Admin
export const getUsers = () => api.get<UserInfo[]>('/admin/users')
export const deleteUser = (id: number) => api.delete(`/admin/users/${id}`)
export const resetUserPassword = (id: number, newPassword: string) =>
  api.post(`/admin/users/${id}/reset-password`, { newPassword })
export const getAppSettings = () => api.get<AppSettings>('/admin/settings')
export const getAppSettingDefaults = () => api.get<AppSettings>('/admin/settings/defaults')
export const updateAppSettings = (settings: { key: string; value: string }[]) =>
  api.put('/admin/settings', settings)
export const getAdminLogs = (limit = 500, level?: string) => {
  const params: Record<string, string> = { limit: String(limit) }
  if (level) params.level = level
  return api.get<{ logs: LogEntry[]; count: number; logLevel: string }>('/admin/logs', { params })
}

export default api
