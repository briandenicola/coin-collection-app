import axios from 'axios'
import type { Coin, CoinListResponse, CoinImage, AuthResponse, StatsResponse, UserInfo, AppSettings, LogEntry, ApiKey, WebAuthnCredentialInfo, ValueSnapshot } from '@/types'

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

let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = []

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach((p) => {
    if (token) p.resolve(token)
    else p.reject(error)
  })
  failedQueue = []
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      const refreshToken = localStorage.getItem('refreshToken')
      if (!refreshToken) {
        clearAuth()
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({
            resolve: (token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(api(originalRequest))
            },
            reject,
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const res = await axios.post<AuthResponse>(`${API_BASE}/api/auth/refresh`, { refreshToken })
        const { token, refreshToken: newRefresh, user } = res.data
        localStorage.setItem('token', token)
        localStorage.setItem('refreshToken', newRefresh)
        localStorage.setItem('user', JSON.stringify(user))
        processQueue(null, token)
        originalRequest.headers.Authorization = `Bearer ${token}`
        return api(originalRequest)
      } catch (refreshError) {
        processQueue(refreshError, null)
        clearAuth()
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }
    return Promise.reject(error)
  },
)

function clearAuth() {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('user')
  window.location.href = '/login'
}

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
  sort?: string
  order?: string
}) => api.get<CoinListResponse>('/coins', { params })

const NULLABLE_FIELDS = ['weightGrams', 'diameterMm', 'purchasePrice', 'currentValue', 'purchaseDate']

function sanitizeCoin(coin: Partial<Coin>): Partial<Coin> {
  const clean = { ...coin }
  for (const field of NULLABLE_FIELDS) {
    if ((clean as any)[field] === '' || (clean as any)[field] === undefined) {
      ;(clean as any)[field] = null
    }
  }
  // Default currentValue to purchasePrice if not set
  if (!clean.currentValue && clean.purchasePrice) {
    clean.currentValue = clean.purchasePrice
  }
  // Convert date-only strings (YYYY-MM-DD) to RFC3339 for Go
  if (typeof clean.purchaseDate === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(clean.purchaseDate)) {
    clean.purchaseDate = clean.purchaseDate + 'T00:00:00Z'
  }
  return clean
}

export const getCoin = (id: number) => api.get<Coin>(`/coins/${id}`)
export const createCoin = (coin: Partial<Coin>) => api.post<Coin>('/coins', sanitizeCoin(coin))
export const updateCoin = (id: number, coin: Partial<Coin>) => api.put<Coin>(`/coins/${id}`, sanitizeCoin(coin))
export const purchaseCoin = (id: number) => api.post<Coin>(`/coins/${id}/purchase`)
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
export const getValueHistory = () => api.get<ValueSnapshot[]>('/value-history')

// Autocomplete suggestions
export const getSuggestions = (field: string, q: string) =>
  api.get<string[]>('/suggestions', { params: { field, q } })

// User self-service
export const getMe = () => api.get<UserInfo>('/auth/me')
export const changePassword = (currentPassword: string, newPassword: string) =>
  api.post('/auth/change-password', { currentPassword, newPassword })
export const exportCollection = () => api.get('/user/export', { responseType: 'blob' })
export const proxyImage = (url: string) =>
  api.get('/proxy-image', { params: { url }, responseType: 'blob' })
export const importCollection = (coins: Coin[]) => api.post('/user/import', coins)

// API Keys
export const generateApiKey = (name: string) =>
  api.post<{ key: string; apiKey: ApiKey }>('/auth/api-keys', { name })
export const listApiKeys = () => api.get<ApiKey[]>('/auth/api-keys')
export const revokeApiKey = (id: number) => api.delete(`/auth/api-keys/${id}`)

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

// WebAuthn
export const webauthnRegisterBegin = () =>
  api.post('/auth/webauthn/register/begin')
export const webauthnRegisterFinish = (credential: PublicKeyCredential) => {
  const attestation = credential.response as AuthenticatorAttestationResponse
  const body = {
    id: credential.id,
    rawId: bufferToBase64url(credential.rawId),
    type: credential.type,
    authenticatorAttachment: credential.authenticatorAttachment || undefined,
    response: {
      attestationObject: bufferToBase64url(attestation.attestationObject),
      clientDataJSON: bufferToBase64url(attestation.clientDataJSON),
      transports: attestation.getTransports ? attestation.getTransports() : undefined,
    },
  }
  return api.post('/auth/webauthn/register/finish', body)
}
export const webauthnLoginBegin = (username: string) =>
  api.post<{ options: PublicKeyCredentialRequestOptionsJSON; username: string }>('/auth/webauthn/login/begin', { username })
export const webauthnLoginFinish = (username: string, credential: PublicKeyCredential) => {
  const assertion = credential.response as AuthenticatorAssertionResponse
  const body = {
    id: credential.id,
    rawId: bufferToBase64url(credential.rawId),
    type: credential.type,
    response: {
      authenticatorData: bufferToBase64url(assertion.authenticatorData),
      clientDataJSON: bufferToBase64url(assertion.clientDataJSON),
      signature: bufferToBase64url(assertion.signature),
      userHandle: assertion.userHandle ? bufferToBase64url(assertion.userHandle) : null,
    },
  }
  return api.post<AuthResponse>(`/auth/webauthn/login/finish?username=${encodeURIComponent(username)}`, body)
}
export const webauthnCheck = (username: string) =>
  api.get<{ available: boolean }>('/auth/webauthn/check', { params: { username } })
export const webauthnListCredentials = () =>
  api.get<WebAuthnCredentialInfo[]>('/auth/webauthn/credentials')
export const webauthnDeleteCredential = (id: number) =>
  api.delete(`/auth/webauthn/credentials/${id}`)

function bufferToBase64url(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  bytes.forEach((b) => (binary += String.fromCharCode(b)))
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

// Helper types for WebAuthn JSON format
interface PublicKeyCredentialRequestOptionsJSON {
  challenge: string
  timeout?: number
  rpId?: string
  allowCredentials?: Array<{ id: string; type: string; transports?: string[] }>
  userVerification?: string
}

export default api
