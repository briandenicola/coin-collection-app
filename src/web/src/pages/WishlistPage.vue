<template>
  <div class="container">
    <div class="page-header">
      <h1>Wishlist</h1>
      <div class="header-actions">
        <SortSelect v-model="sortKey" />
        <router-link to="/add?wishlist=true" class="btn btn-primary"><CirclePlus :size="16" /> Add Coin</router-link>
      </div>
    </div>

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="store.coins.length" class="coins-grid">
      <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" wishlist @purchase="handlePurchase" />
    </div>

    <div v-else class="empty-state">
      <h3>Your wishlist is empty</h3>
      <p>Add coins to your wishlist to track what you're looking for</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import CoinCard from '@/components/CoinCard.vue'
import SortSelect from '@/components/SortSelect.vue'
import { purchaseCoin } from '@/api/client'
import type { Coin } from '@/types'
import { CirclePlus } from 'lucide-vue-next'

const store = useCoinsStore()
const sortKey = ref('updated_at_desc')

function loadCoins() {
  const [sort, order] = sortKey.value.split('_').length === 3
    ? [sortKey.value.split('_').slice(0, 2).join('_'), sortKey.value.split('_')[2]]
    : [sortKey.value.split('_')[0], sortKey.value.split('_')[1]]
  store.fetchCoins({ wishlist: 'true', sort, order })
}

async function handlePurchase(coin: Coin) {
  if (!confirm(`Move "${coin.name}" to your collection?`)) return
  try {
    await purchaseCoin(coin.id)
    loadCoins()
  } catch {
    alert('Failed to mark as purchased')
  }
}

watch(sortKey, loadCoins)
loadCoins()
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
}
</style>
