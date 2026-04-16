<template>
  <div ref="pullContainer" class="container" :class="{ 'pwa-mode': isPwa }" :style="pullDistance > 0 ? `transform: translateY(${pullDistance}px); transition: none` : ''">
    <div class="pull-indicator" :class="{ visible: pullDistance > 0 || refreshing, refreshing }" :style="`top: ${-50 + pullDistance * 0.6}px; opacity: ${Math.min(pullDistance / 60, 1)}`">
      <div class="pull-spinner" :style="refreshing ? '' : `transform: rotate(${pullDistance * 3}deg)`"></div>
      <span class="pull-text">{{ refreshing ? 'Refreshing...' : pullDistance >= 60 ? 'Release to refresh' : 'Pull to refresh' }}</span>
    </div>

    <!-- PWA compact header: search + filter/sort -->
    <div v-if="isPwa" class="pwa-header">
      <SearchBar v-model="search" />
      <div class="hamburger-wrapper">
        <button class="hamburger-btn" @click="menuOpen = !menuOpen" :class="{ active: menuOpen }">
          <SlidersHorizontal :size="22" />
        </button>
        <Transition name="menu-slide">
          <div v-if="menuOpen" class="pwa-menu">
            <div class="pwa-menu-section">
              <span class="pwa-menu-label">Category</span>
              <CategoryFilter v-model="selectedCategory" />
            </div>
            <div class="pwa-menu-section">
              <span class="pwa-menu-label">Sort</span>
              <SortSelect v-model="sortKey" />
            </div>
            <div class="pwa-menu-section">
              <span class="pwa-menu-label">View</span>
              <div class="pwa-menu-row">
                <div class="view-toggle">
                  <button class="view-btn" :class="{ active: viewMode === 'swipe' }" @click="viewMode = 'swipe'" title="Swipe view">
                    <Layers :size="18" />
                  </button>
                  <button class="view-btn" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'" title="Grid view">
                    <LayoutGrid :size="18" />
                  </button>
                </div>
                <div v-if="viewMode === 'grid'" class="side-toggle">
                  <button class="toggle-btn" :class="{ active: gridSide === null }" @click="gridSide = null">Primary</button>
                  <button class="toggle-btn" :class="{ active: gridSide === 'obverse' }" @click="gridSide = 'obverse'">Obverse</button>
                  <button class="toggle-btn" :class="{ active: gridSide === 'reverse' }" @click="gridSide = 'reverse'">Reverse</button>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>
    <div v-if="isPwa && menuOpen" class="pwa-menu-backdrop" @click="menuOpen = false"></div>

    <!-- Desktop header (hidden in PWA) -->
    <div v-if="!isPwa" class="page-header collection-header">
      <h1>My Collection</h1>
      <SearchBar v-model="search" />
      <SortSelect v-model="sortKey" />
    </div>

    <div v-if="!isPwa" class="collection-toolbar">
      <CategoryFilter v-model="selectedCategory" />
      <div class="toolbar-right">
        <div v-if="viewMode === 'grid'" class="side-toggle">
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

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>Loading collection...</p>
    </div>

    <template v-else-if="store.coins.length">
      <SwipeGallery v-if="viewMode === 'swipe'" :coins="store.coins" />
      <div v-else class="coins-grid">
        <CoinCard v-for="coin in store.coins" :key="coin.id" :coin="coin" :image-side="gridSide" />
      </div>
    </template>

    <div v-else class="empty-state">
      <h3>{{ search || selectedCategory ? 'No coins match your search' : 'Your collection is empty' }}</h3>
      <p>{{ search || selectedCategory ? 'Try different filters' : 'Add your first coin to get started' }}</p>
      <router-link v-if="!search && !selectedCategory" to="/add" class="btn btn-primary" style="margin-top: 1rem">
        Add Your First Coin
      </router-link>
    </div>

    <div v-if="store.total > 50 && viewMode === 'grid'" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="page--">← Previous</button>
      <span class="page-info">Page {{ page }} of {{ Math.ceil(store.total / 50) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="page * 50 >= store.total" @click="page++">Next →</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import type { ImageType } from '@/types'
import { usePullToRefresh } from '@/composables/usePullToRefresh'
import CoinCard from '@/components/CoinCard.vue'
import SwipeGallery from '@/components/SwipeGallery.vue'
import CategoryFilter from '@/components/CategoryFilter.vue'
import SearchBar from '@/components/SearchBar.vue'
import SortSelect from '@/components/SortSelect.vue'

import { Layers, LayoutGrid, CirclePlus, SlidersHorizontal } from 'lucide-vue-next'

const store = useCoinsStore()
const auth = useAuthStore()
const router = useRouter()
const selectedCategory = store.selectedCategory !== undefined ? ref(store.selectedCategory) : ref('')
const search = ref(store.searchQuery)
const page = ref(1)
const sortKey = ref(localStorage.getItem('defaultSort') || 'updated_at_desc')
const menuOpen = ref(false)

// Use saved preference if set, otherwise default to swipe in PWA mode
const savedView = localStorage.getItem('defaultView') as 'grid' | 'swipe' | null
const isPwa = window.matchMedia('(display-mode: standalone)').matches
  || (window.navigator as any).standalone === true
const viewMode = ref<'grid' | 'swipe'>(savedView || (isPwa ? 'swipe' : 'grid'))
const gridSide = ref<ImageType | null>(null)

const pullContainer = ref<HTMLElement | null>(null)
const { pullDistance, refreshing } = usePullToRefresh(pullContainer, async () => {
  await new Promise<void>((resolve) => {
    loadCoins()
    const unwatch = watch(() => store.loading, (loading) => {
      if (!loading) { unwatch(); resolve() }
    })
    if (!store.loading) { unwatch(); resolve() }
  })
})

let debounceTimer: ReturnType<typeof setTimeout>

function loadCoins() {
  const [sort, order] = sortKey.value.split('_').length === 3
    ? [sortKey.value.split('_').slice(0, 2).join('_'), sortKey.value.split('_')[2]]
    : [sortKey.value.split('_')[0], sortKey.value.split('_')[1]]
  store.selectedCategory = selectedCategory.value
  store.searchQuery = search.value
  store.fetchCoins({
    category: selectedCategory.value || undefined,
    search: search.value || undefined,
    wishlist: 'false',
    sold: 'false',
    page: page.value,
    sort,
    order,
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
watch(sortKey, () => {
  page.value = 1
  loadCoins()
})

loadCoins()

</script>

<style scoped>
/* --- PWA compact header --- */
.pwa-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.pwa-header :deep(.search-bar) {
  flex: 1;
  max-width: none;
}

.hamburger-wrapper {
  position: relative;
  flex-shrink: 0;
}

.hamburger-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.hamburger-btn.active,
.hamburger-btn:hover {
  border-color: var(--accent-gold);
  color: var(--accent-gold);
  background: var(--accent-gold-dim);
}

/* --- PWA popover menu --- */
.pwa-menu {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  z-index: 100;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 1rem;
  min-width: 260px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
}

.pwa-menu-backdrop {
  position: fixed;
  inset: 0;
  z-index: 90;
}

.pwa-menu-section {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.pwa-menu-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  font-weight: 600;
}

.pwa-menu-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.pwa-add-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  text-decoration: none;
}

/* Menu slide transition */
.menu-slide-enter-active,
.menu-slide-leave-active {
  transition: all 0.2s ease;
}
.menu-slide-enter-from,
.menu-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* --- Desktop header (unchanged) --- */
.collection-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.collection-header h1 {
  white-space: nowrap;
}

.collection-header :deep(.search-bar) {
  flex: 1;
  max-width: 400px;
  margin: 0 auto;
}

.collection-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
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

/* --- Pull to refresh --- */
.pull-indicator {
  position: fixed;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  z-index: 100;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s;
}

.pull-indicator.visible {
  pointer-events: auto;
}

.pull-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--accent-gold);
  border-radius: 50%;
}

.pull-indicator.refreshing .pull-spinner {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pull-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

@media (max-width: 768px) {
  .collection-header {
    flex-direction: column;
    align-items: stretch;
  }

  .collection-header :deep(.search-bar) {
    max-width: 100%;
  }

  .header-filters {
    justify-content: flex-start;
  }
}
</style>
