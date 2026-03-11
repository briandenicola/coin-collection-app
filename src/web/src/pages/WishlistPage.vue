<template>
  <div class="container">
    <div class="page-header">
      <h1>Wishlist</h1>
      <router-link to="/add" class="btn btn-primary"><CirclePlus :size="16" /> Add Coin</router-link>
    </div>

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="store.coins.length" class="coins-grid">
      <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" />
    </div>

    <div v-else class="empty-state">
      <h3>Your wishlist is empty</h3>
      <p>Add coins to your wishlist to track what you're looking for</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import CoinCard from '@/components/CoinCard.vue'
import { CirclePlus } from 'lucide-vue-next'

const store = useCoinsStore()

onMounted(() => {
  store.fetchCoins({ wishlist: 'true' })
})
</script>
