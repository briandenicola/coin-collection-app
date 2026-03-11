<template>
  <div class="container">
    <div class="page-header">
      <h1>🏛️ My Collection</h1>
      <div class="header-actions">
        <SearchBar v-model="search" />
        <router-link to="/add" class="btn btn-primary">➕ Add Coin</router-link>
      </div>
    </div>

    <CategoryFilter v-model="selectedCategory" />

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>Loading collection...</p>
    </div>

    <div v-else-if="store.coins.length" class="coins-grid" style="margin-top: 1.5rem">
      <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" />
    </div>

    <div v-else class="empty-state">
      <h3>{{ search || selectedCategory ? 'No coins match your search' : 'Your collection is empty' }}</h3>
      <p>{{ search || selectedCategory ? 'Try different filters' : 'Add your first coin to get started' }}</p>
      <router-link v-if="!search && !selectedCategory" to="/add" class="btn btn-primary" style="margin-top: 1rem">
        Add Your First Coin
      </router-link>
    </div>

    <div v-if="store.total > 50" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="page--">← Previous</button>
      <span class="page-info">Page {{ page }} of {{ Math.ceil(store.total / 50) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="page * 50 >= store.total" @click="page++">Next →</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import CoinCard from '@/components/CoinCard.vue'
import CategoryFilter from '@/components/CategoryFilter.vue'
import SearchBar from '@/components/SearchBar.vue'

const store = useCoinsStore()
const selectedCategory = ref('')
const search = ref('')
const page = ref(1)

let debounceTimer: ReturnType<typeof setTimeout>

function loadCoins() {
  store.fetchCoins({
    category: selectedCategory.value || undefined,
    search: search.value || undefined,
    wishlist: 'false',
    page: page.value,
  })
}

watch(selectedCategory, () => {
  page.value = 1
  loadCoins()
})

watch(search, () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    page.value = 1
    loadCoins()
  }, 300)
})

watch(page, loadCoins)

loadCoins()
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-subtle);
}

.page-info {
  color: var(--text-secondary);
  font-size: 0.85rem;
}
</style>
