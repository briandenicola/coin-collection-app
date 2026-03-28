<template>
  <div class="container">
    <div class="page-header">
      <h1>Wishlist</h1>
      <div class="header-actions">
        <button class="btn btn-primary" @click="showChat = true"><Bot :size="16" /> Find Coins</button>
        <router-link to="/add?wishlist=true" class="btn btn-secondary"><CirclePlus :size="16" /> Add Coin</router-link>
      </div>
    </div>

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="store.coins.length" class="coins-grid">
      <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" wishlist @purchase="openPurchaseModal" />
    </div>

    <div v-else class="empty-state">
      <h3>Your wishlist is empty</h3>
      <p>Add coins to your wishlist to track what you're looking for</p>
      <button class="btn btn-primary" @click="showChat = true" style="margin-top: 0.75rem">
        <Bot :size="16" /> Search for Coins with AI
      </button>
    </div>

    <PurchaseModal
      v-if="purchaseTarget"
      :coin="purchaseTarget"
      @close="purchaseTarget = null"
      @confirm="handlePurchaseConfirm"
    />

    <CoinSearchChat v-if="showChat" @close="showChat = false" @added="loadCoins" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import CoinCard from '@/components/CoinCard.vue'
import CoinSearchChat from '@/components/CoinSearchChat.vue'
import PurchaseModal from '@/components/PurchaseModal.vue'
import { purchaseCoin } from '@/api/client'
import type { Coin } from '@/types'
import { CirclePlus, Bot } from 'lucide-vue-next'

const store = useCoinsStore()
const showChat = ref(false)
const purchaseTarget = ref<Coin | null>(null)

function loadCoins() {
  store.fetchCoins({ wishlist: 'true', sort: 'updated_at', order: 'desc' })
}

function openPurchaseModal(coin: Coin) {
  purchaseTarget.value = coin
}

async function handlePurchaseConfirm(data: { purchasePrice?: number; purchaseDate?: string; purchaseLocation?: string }) {
  if (!purchaseTarget.value) return
  try {
    await purchaseCoin(purchaseTarget.value.id, data)
    purchaseTarget.value = null
    loadCoins()
  } catch {
    purchaseTarget.value = null
  }
}

loadCoins()
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
</style>
