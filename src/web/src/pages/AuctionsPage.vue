<template>
  <PullToRefresh :on-refresh="handleRefresh">
  <div class="container">
    <div class="page-header">
      <h1>Auctions</h1>
      <div class="header-actions">
        <button class="btn btn-secondary" :disabled="syncing" @click="syncWatchlist">
          <RefreshCw :size="16" :class="{ spinning: syncing }" />
          {{ syncing ? 'Syncing...' : 'Sync Watchlist' }}
        </button>
        <button class="btn btn-primary" @click="showImport = true"><Import :size="16" /> Add Lot</button>
      </div>
    </div>

    <div v-if="syncMessage" class="sync-toast">{{ syncMessage }}</div>

    <div class="filter-bar">
      <div class="status-filters">
        <button
          v-for="s in statuses"
          :key="s.value"
          class="filter-btn"
          :class="{ active: activeStatus === s.value }"
          @click="activeStatus = s.value"
        >
          {{ s.label }}
          <span v-if="statusCounts[s.value]" class="count-badge">{{ statusCounts[s.value] }}</span>
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="lots.length" class="lots-grid">
      <AuctionLotCard v-for="lot in lots" :key="lot.id" :lot="lot" @select="openLot" />
    </div>

    <div v-else class="empty-state">
      <h3>No auction lots{{ activeStatus ? ` with status "${activeStatus}"` : '' }}</h3>
      <p>Import lots from NumisBids to start tracking auctions</p>
      <button class="btn btn-primary" @click="showImport = true" style="margin-top: 0.75rem">
        <Import :size="16" /> Import Your First Lot
      </button>
    </div>

    <ImportLotModal v-if="showImport" @close="showImport = false" @imported="handleImported" />

    <!-- Lot detail drawer -->
    <div v-if="selectedLot" class="modal-overlay" @click.self="selectedLot = null">
      <div class="lot-detail card">
        <div class="detail-header">
          <h2>{{ selectedLot.title }}</h2>
          <button class="btn-close" @click="selectedLot = null"><X :size="18" /></button>
        </div>

        <div v-if="selectedLot.imageUrl" class="detail-image-container">
          <img :src="proxiedDetailImageUrl" :alt="selectedLot.title" class="detail-image" />
        </div>

        <div class="detail-body">
          <div class="detail-row" v-if="selectedLot.auctionHouse">
            <span class="detail-label">Auction House</span>
            <span>{{ selectedLot.auctionHouse }}</span>
          </div>
          <div class="detail-row" v-if="selectedLot.saleName">
            <span class="detail-label">Sale</span>
            <span>{{ selectedLot.saleName }}</span>
          </div>
          <div class="detail-row" v-if="selectedLot.saleDate">
            <span class="detail-label">Sale Date</span>
            <span>{{ formatDate(selectedLot.saleDate) }}</span>
          </div>
          <div class="detail-row" v-if="selectedLot.estimate">
            <span class="detail-label">Estimate</span>
            <span>{{ formatCurrency(selectedLot.estimate, selectedLot.currency) }}</span>
          </div>
          <div class="detail-row" v-if="selectedLot.currentBid">
            <span class="detail-label">Current Bid</span>
            <span class="bid-value">{{ formatCurrency(selectedLot.currentBid, selectedLot.currency) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-tag" :class="`status-${selectedLot.status}`">{{ selectedLot.status }}</span>
          </div>
          <div v-if="selectedLot.description" class="detail-description">
            <span class="detail-label">Description</span>
            <p>{{ selectedLot.description }}</p>
          </div>
        </div>

        <div class="detail-actions">
          <div class="action-row">
            <select v-model="newStatus" class="form-input status-select">
              <option value="watching">Watching</option>
              <option value="bidding">Bidding</option>
              <option value="won">Won</option>
              <option value="lost">Lost</option>
              <option value="passed">Passed</option>
            </select>
            <button class="btn btn-secondary" @click="changeStatus" :disabled="newStatus === selectedLot.status">
              Update Status
            </button>
          </div>
          <div class="action-row">
            <a :href="selectedLot.numisBidsUrl" class="btn btn-primary" target="_blank" rel="noopener noreferrer">
              <ExternalLink :size="14" /> View on NumisBids
            </a>
            <button v-if="selectedLot.status === 'won'" class="btn btn-primary" @click="convertToCoin">
              <ArrowRightCircle :size="14" /> Add to Collection
            </button>
            <button class="btn btn-danger" @click="removeLot">
              <Trash2 :size="14" /> Remove
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getAuctionLots, updateAuctionLotStatus, convertAuctionLotToCoin, deleteAuctionLot, syncNumisBidsWatchlist } from '@/api/client'
import type { AuctionLot, AuctionLotStatus } from '@/types'
import AuctionLotCard from '@/components/AuctionLotCard.vue'
import ImportLotModal from '@/components/ImportLotModal.vue'
import PullToRefresh from '@/components/PullToRefresh.vue'
import { Import, X, ExternalLink, ArrowRightCircle, Trash2, RefreshCw } from 'lucide-vue-next'

const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

const lots = ref<AuctionLot[]>([])
const allLots = ref<AuctionLot[]>([])
const loading = ref(true)
const showImport = ref(false)
const selectedLot = ref<AuctionLot | null>(null)
const newStatus = ref<AuctionLotStatus>('watching')
const activeStatus = ref('')
const syncing = ref(false)
const syncMessage = ref('')

const proxiedDetailImageUrl = computed(() => {
  if (!selectedLot.value?.imageUrl) return ''
  const token = localStorage.getItem('token') ?? ''
  return `${API_BASE}/api/proxy-image?url=${encodeURIComponent(selectedLot.value.imageUrl)}&token=${encodeURIComponent(token)}`
})

const statuses = [
  { value: '', label: 'All' },
  { value: 'watching', label: 'Watching' },
  { value: 'bidding', label: 'Bidding' },
  { value: 'won', label: 'Won' },
  { value: 'lost', label: 'Lost' },
  { value: 'passed', label: 'Passed' },
]

const statusCounts = computed(() => {
  const counts: Record<string, number> = {}
  for (const lot of allLots.value) {
    counts[lot.status] = (counts[lot.status] ?? 0) + 1
  }
  return counts
})

watch(activeStatus, () => fetchLots())

async function fetchLots() {
  loading.value = true
  try {
    const params: Record<string, string> = { sort: 'updated_at', order: 'desc' }
    if (activeStatus.value) params.status = activeStatus.value
    const res = await getAuctionLots(params)
    lots.value = res.data?.lots ?? []
    if (!activeStatus.value) allLots.value = lots.value
  } catch {
    lots.value = []
  } finally {
    loading.value = false
  }
}

async function fetchAllCounts() {
  try {
    const res = await getAuctionLots({ limit: 999 })
    allLots.value = res.data?.lots ?? []
  } catch { /* ignore */ }
}

async function handleRefresh() {
  await fetchLots()
  await fetchAllCounts()
}

function openLot(lot: AuctionLot) {
  selectedLot.value = lot
  newStatus.value = lot.status
}

function handleImported() {
  showImport.value = false
  fetchLots()
  fetchAllCounts()
}

async function changeStatus() {
  if (!selectedLot.value) return
  try {
    const res = await updateAuctionLotStatus(selectedLot.value.id, newStatus.value)
    selectedLot.value = res.data
    fetchLots()
    fetchAllCounts()
  } catch { /* ignore */ }
}

async function convertToCoin() {
  if (!selectedLot.value) return
  try {
    await convertAuctionLotToCoin(selectedLot.value.id)
    selectedLot.value = null
    fetchLots()
    fetchAllCounts()
  } catch { /* ignore */ }
}

async function removeLot() {
  if (!selectedLot.value) return
  try {
    await deleteAuctionLot(selectedLot.value.id)
    selectedLot.value = null
    fetchLots()
    fetchAllCounts()
  } catch { /* ignore */ }
}

async function syncWatchlist() {
  syncing.value = true
  syncMessage.value = ''
  try {
    const res = await syncNumisBidsWatchlist()
    const count = res.data?.synced ?? 0
    syncMessage.value = `Synced ${count} lot${count !== 1 ? 's' : ''} from NumisBids`
    fetchLots()
    fetchAllCounts()
    setTimeout(() => { syncMessage.value = '' }, 4000)
  } catch (err: unknown) {
    const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Sync failed'
    syncMessage.value = msg
    setTimeout(() => { syncMessage.value = '' }, 5000)
  } finally {
    syncing.value = false
  }
}

function formatCurrency(value: number, currency?: string) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: currency || 'USD' }).format(value)
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

fetchLots()
fetchAllCounts()
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.filter-bar {
  margin-bottom: 1.25rem;
}

.status-filters {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.filter-btn {
  padding: 0.4rem 0.9rem;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.82rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.filter-btn:hover {
  border-color: var(--accent-gold-dim);
  color: var(--text-primary);
}

.filter-btn.active {
  background: var(--accent-gold-glow);
  border-color: var(--accent-gold-dim);
  color: var(--accent-gold);
}

.count-badge {
  background: var(--bg-primary);
  padding: 0.05rem 0.4rem;
  border-radius: var(--radius-full);
  font-size: 0.7rem;
  font-weight: 600;
}

.lots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1.25rem;
}

/* Lot detail modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.lot-detail {
  max-width: 560px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  padding: 0;
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-subtle);
  gap: 1rem;
}

.detail-header h2 {
  font-size: 1.1rem;
  line-height: 1.35;
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.btn-close:hover {
  color: var(--text-primary);
}

.detail-image-container {
  width: 100%;
  max-height: 300px;
  overflow: hidden;
  background: var(--bg-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-image {
  width: 100%;
  height: 300px;
  object-fit: contain;
}

.detail-body {
  padding: 1.25rem 1.5rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border-subtle);
  font-size: 0.88rem;
}

.detail-label {
  color: var(--text-secondary);
  font-size: 0.82rem;
}

.bid-value {
  font-weight: 600;
  color: var(--accent-gold);
}

.status-tag {
  padding: 0.15rem 0.55rem;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-watching { background: rgba(100, 150, 255, 0.2); color: #6496ff; }
.status-bidding { background: var(--accent-gold-glow); color: var(--accent-gold); }
.status-won { background: rgba(74, 222, 128, 0.15); color: #4ade80; }
.status-lost { background: rgba(248, 113, 113, 0.15); color: #f87171; }
.status-passed { background: rgba(120, 120, 120, 0.15); color: #999; }

.detail-description {
  margin-top: 0.75rem;
}

.detail-description p {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-top: 0.4rem;
  line-height: 1.5;
  max-height: 120px;
  overflow-y: auto;
}

.detail-actions {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.action-row {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.status-select {
  flex: 1;
  min-width: 120px;
  max-width: 160px;
}

.btn-danger {
  background: transparent;
  border: 1px solid rgba(248, 113, 113, 0.4);
  color: #f87171;
  padding: 0.5rem 0.9rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 0.82rem;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  transition: all var(--transition-fast);
}

.btn-danger:hover {
  background: rgba(248, 113, 113, 0.1);
  border-color: rgba(248, 113, 113, 0.6);
}

.sync-toast {
  padding: 0.6rem 1rem;
  margin-bottom: 1rem;
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  color: var(--text-primary);
  font-size: 0.85rem;
  text-align: center;
  animation: fadeIn 0.2s ease;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
