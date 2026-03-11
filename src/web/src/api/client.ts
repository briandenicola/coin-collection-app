import axios from 'axios'
import type { Coin, CoinListResponse, CoinImage, AuthResponse, StatsResponse } from '@/types'

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

export const getCoin = (id: number) => api.get<Coin>(`/coins/${id}`)
export const createCoin = (coin: Partial<Coin>) => api.post<Coin>('/coins', coin)
export const updateCoin = (id: number, coin: Partial<Coin>) => api.put<Coin>(`/coins/${id}`, coin)
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
export const analyzeCoin = (coinId: number) =>
  api.post<{ analysis: string; coin: Coin }>(`/coins/${coinId}/analyze`)

// Stats
export const getStats = () => api.get<StatsResponse>('/stats')

export default api
