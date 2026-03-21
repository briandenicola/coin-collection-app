import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, AuthResponse } from '@/types'
import * as api from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function setTokens(data: AuthResponse) {
    token.value = data.token
    user.value = data.user
    localStorage.setItem('token', data.token)
    localStorage.setItem('refreshToken', data.refreshToken)
    localStorage.setItem('user', JSON.stringify(data.user))
  }

  async function doLogin(username: string, password: string) {
    const res = await api.login(username, password)
    setTokens(res.data)
  }

  async function doRegister(username: string, password: string, email?: string) {
    const res = await api.register(username, password, email)
    setTokens(res.data)
  }

  async function doWebAuthnLogin(username: string) {
    // Begin ceremony — get challenge from server
    const beginRes = await api.webauthnLoginBegin(username)
    const { options } = beginRes.data

    // Convert base64url challenge to ArrayBuffer
    const challenge = base64urlToBuffer(options.challenge)
    const allowCredentials = options.allowCredentials?.map((c: { id: string; type: string; transports?: string[] }) => ({
      id: base64urlToBuffer(c.id),
      type: c.type as PublicKeyCredentialType,
      transports: c.transports as AuthenticatorTransport[] | undefined,
    }))

    // Call browser WebAuthn API (triggers Face ID / fingerprint)
    const credential = await navigator.credentials.get({
      publicKey: {
        challenge,
        rpId: options.rpId,
        allowCredentials,
        userVerification: (options.userVerification as UserVerificationRequirement) || 'preferred',
        timeout: options.timeout || 60000,
      },
    }) as PublicKeyCredential

    // Finish ceremony — send assertion to server, get tokens
    const finishRes = await api.webauthnLoginFinish(username, credential)
    setTokens(finishRes.data)
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
  }

  return { token, user, isAuthenticated, isAdmin, doLogin, doRegister, doWebAuthnLogin, logout }
})

function base64urlToBuffer(base64url: string): ArrayBuffer {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4 === 0 ? '' : '='.repeat(4 - (base64.length % 4))
  const binary = atob(base64 + pad)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes.buffer
}
