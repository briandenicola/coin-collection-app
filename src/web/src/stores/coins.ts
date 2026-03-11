import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Coin, StatsResponse } from '@/types'
import * as api from '@/api/client'

export const useCoinsStore = defineStore('coins', () => {
  const coins = ref<Coin[]>([])
  const currentCoin = ref<Coin | null>(null)
  const total = ref(0)
  const loading = ref(false)
  const stats = ref<StatsResponse | null>(null)

  async function fetchCoins(params?: {
    category?: string
    search?: string
    wishlist?: string
    page?: number
  }) {
    loading.value = true
    try {
      const res = await api.getCoins(params)
      coins.value = res.data.coins || []
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  async function fetchCoin(id: number) {
    loading.value = true
    try {
      const res = await api.getCoin(id)
      currentCoin.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function addCoin(coin: Partial<Coin>) {
    const res = await api.createCoin(coin)
    return res.data
  }

  async function editCoin(id: number, coin: Partial<Coin>) {
    const res = await api.updateCoin(id, coin)
    return res.data
  }

  async function removeCoin(id: number) {
    await api.deleteCoin(id)
    coins.value = coins.value.filter((c) => c.id !== id)
  }

  async function fetchStats() {
    const res = await api.getStats()
    stats.value = res.data
  }

  return {
    coins,
    currentCoin,
    total,
    loading,
    stats,
    fetchCoins,
    fetchCoin,
    addCoin,
    editCoin,
    removeCoin,
    fetchStats,
  }
})
