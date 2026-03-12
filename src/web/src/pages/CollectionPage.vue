<template>
  <div class="container">
    <div class="page-header">
      <h1>My Collection</h1>
      <div class="header-actions">
        <SearchBar v-model="search" />
        <div class="view-toggle">
          <button class="view-btn" :class="{ active: viewMode === 'swipe' }" @click="viewMode = 'swipe'" title="Swipe view">
            <Layers :size="18" />
          </button>
          <button class="view-btn" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'" title="Grid view">
            <LayoutGrid :size="18" />
          </button>
        </div>
        <router-link to="/add" class="btn btn-primary"><CirclePlus :size="16" /> Add Coin</router-link>
      </div>
    </div>

    <CategoryFilter v-model="selectedCategory" />

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>Loading collection...</p>
    </div>

    <template v-else-if="store.coins.length">
      <!-- Swipe view -->
      <SwipeGallery v-if="viewMode === 'swipe'" :coins="store.coins" />

      <!-- Grid view -->
      <div v-else>
        <div class="grid-toolbar">
          <div class="side-toggle">
            <button class="toggle-btn" :class="{ active: gridSide === null }" @click="gridSide = null">
              Primary
            </button>
            <button class="toggle-btn" :class="{ active: gridSide === 'obverse' }" @click="gridSide = 'obverse'">
              Obverse
            </button>
            <button class="toggle-btn" :class="{ active: gridSide === 'reverse' }" @click="gridSide = 'reverse'">
              Reverse
            </button>
          </div>
        </div>
        <div class="coins-grid">
          <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" :image-side="gridSide" />
        </div>
      </div>
    </template>

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
import { ref, watch, onMounted } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import type { ImageType } from '@/types'
import CoinCard from '@/components/CoinCard.vue'
import SwipeGallery from '@/components/SwipeGallery.vue'
import CategoryFilter from '@/components/CategoryFilter.vue'
import SearchBar from '@/components/SearchBar.vue'

import { Layers, LayoutGrid, CirclePlus } from 'lucide-vue-next'

const store = useCoinsStore()
const selectedCategory = store.selectedCategory !== undefined ? ref(store.selectedCategory) : ref('')
const search = ref(store.searchQuery)
const page = ref(1)

// Default to swipe in standalone PWA mode, grid otherwise
const isPwa = window.matchMedia('(display-mode: standalone)').matches
  || (window.navigator as any).standalone === true
const viewMode = ref<'grid' | 'swipe'>(isPwa ? 'swipe' : 'grid')
const gridSide = ref<ImageType | null>(null)

let debounceTimer: ReturnType<typeof setTimeout>

function loadCoins() {
  store.selectedCategory = selectedCategory.value
  store.searchQuery = search.value
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

.view-toggle {
  display: flex;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.view-btn {
  padding: 0.4rem 0.6rem;
  border: none;
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 1rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  line-height: 1;
}

.view-btn.active {
  background: var(--accent-gold-dim);
  color: var(--accent-gold);
}

.view-btn:hover:not(.active) {
  background: var(--bg-card-hover);
}

.grid-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
  margin-bottom: 1rem;
}

.side-toggle {
  display: flex;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.toggle-btn {
  padding: 0.35rem 0.75rem;
  border: none;
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toggle-btn.active {
  background: var(--accent-gold-dim);
  color: var(--accent-gold);
}

.toggle-btn:hover:not(.active) {
  background: var(--bg-card-hover);
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
