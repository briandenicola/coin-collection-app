<template>
  <PullToRefresh :on-refresh="handleRefresh">
  <div class="container">
    <div class="page-header">
      <h1>Sold Coins</h1>
    </div>

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="store.coins.length" class="coins-grid">
      <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" sold />
    </div>

    <div v-else class="empty-state">
      <h3>No sold coins yet</h3>
      <p>When you sell coins from your collection, they'll appear here with their sale history.</p>
    </div>
  </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { useCoinsStore } from '@/stores/coins'
import CoinCard from '@/components/CoinCard.vue'
import PullToRefresh from '@/components/PullToRefresh.vue'

const store = useCoinsStore()

store.fetchCoins({ sold: 'true', sort: 'updated_at', order: 'desc' })

async function handleRefresh() {
  await store.fetchCoins({ sold: 'true', sort: 'updated_at', order: 'desc' })
}
</script>
