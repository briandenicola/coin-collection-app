<template>
  <div class="container">
    <div v-if="store.loading && !coin" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="coin" class="coin-detail">
      <div class="detail-header">
        <button class="btn btn-secondary btn-sm" @click="$router.back()">← Back</button>
        <div class="detail-actions">
          <button v-if="coin.isWishlist" class="btn btn-primary btn-sm" @click="handlePurchase">🛒 Mark as Purchased</button>
          <button v-if="!coin.isWishlist && !coin.isSold" class="btn btn-secondary btn-sm" @click="showSellModal = true">Sell</button>
          <router-link :to="`/edit/${coin.id}`" class="btn btn-secondary btn-sm">Edit</router-link>
          <button class="btn btn-danger btn-sm" @click="handleDelete">Delete</button>
        </div>
      </div>

      <div class="detail-layout">
        <!-- Images -->
        <div class="detail-images">
          <ImageGallery :images="coin.images || []" :processing="removingBg" @remove-bg="handleRemoveBackground" @delete-image="handleDeleteImage" />

          <div class="image-upload-section">
            <h4>Upload Images</h4>
            <div class="upload-row">
              <select v-model="uploadType" class="form-select upload-select">
                <option value="obverse">Obverse</option>
                <option value="reverse">Reverse</option>
                <option value="detail">Detail</option>
                <option value="other">Other</option>
              </select>
              <label class="btn btn-secondary btn-sm upload-btn">
                Choose File
                <input type="file" accept="image/*" hidden @change="handleImageUpload" />
              </label>
            </div>

            <div class="url-upload-row">
              <input
                v-model="imageUrl"
                type="url"
                class="form-input url-input"
                placeholder="Or paste an image URL..."
                @keydown.enter="handleUrlUpload"
              />
              <button
                class="btn btn-secondary btn-sm"
                :disabled="!imageUrl || urlLoading"
                @click="handleUrlUpload"
              >
                {{ urlLoading ? 'Fetching...' : 'Fetch' }}
              </button>
            </div>

            <p v-if="uploadStatus" class="upload-status" :class="{ error: uploadError }">{{ uploadStatus }}</p>
          </div>
        </div>

        <!-- Info -->
        <div class="detail-info">
          <div class="detail-title-section">
            <span class="badge" :class="`badge-${coin.category.toLowerCase()}`">{{ coin.category }}</span>
            <h1>{{ coin.name }}</h1>
            <p v-if="coin.ruler" class="detail-ruler">{{ coin.ruler }}</p>
          </div>

          <div class="info-grid">
            <div class="info-card" v-if="coin.denomination">
              <span class="info-label">Denomination</span>
              <span class="info-value">{{ coin.denomination }}</span>
            </div>
            <div class="info-card" v-if="coin.era">
              <span class="info-label">Era</span>
              <span class="info-value">{{ coin.era }}</span>
            </div>
            <div class="info-card" v-if="coin.mint">
              <span class="info-label">Mint</span>
              <span class="info-value">{{ coin.mint }}</span>
            </div>
            <div class="info-card" v-if="coin.material">
              <span class="info-label">Material</span>
              <span class="info-value" :class="`material-${coin.material.toLowerCase()}`">{{ coin.material }}</span>
            </div>
            <div class="info-card" v-if="coin.weightGrams">
              <span class="info-label">Weight</span>
              <span class="info-value">{{ coin.weightGrams }}g</span>
            </div>
            <div class="info-card" v-if="coin.diameterMm">
              <span class="info-label">Diameter</span>
              <span class="info-value">{{ coin.diameterMm }}mm</span>
            </div>
            <div class="info-card" v-if="coin.grade">
              <span class="info-label">Grade</span>
              <span class="info-value gold">{{ coin.grade }}</span>
            </div>
            <div class="info-card" v-if="coin.rarityRating">
              <span class="info-label">Rarity / RIC</span>
              <span class="info-value">{{ coin.rarityRating }}</span>
            </div>
          </div>

          <div v-if="coin.obverseInscription || coin.reverseInscription" class="inscriptions-section">
            <h3>Inscriptions</h3>
            <div v-if="coin.obverseInscription" class="inscription">
              <span class="inscription-label">Obverse:</span>
              <span class="inscription-text">{{ coin.obverseInscription }}</span>
            </div>
            <div v-if="coin.reverseInscription" class="inscription">
              <span class="inscription-label">Reverse:</span>
              <span class="inscription-text">{{ coin.reverseInscription }}</span>
            </div>
          </div>

          <div v-if="coin.obverseDescription || coin.reverseDescription" class="descriptions-section">
            <h3>Design Descriptions</h3>
            <p v-if="coin.obverseDescription"><strong>Obverse:</strong> {{ coin.obverseDescription }}</p>
            <p v-if="coin.reverseDescription"><strong>Reverse:</strong> {{ coin.reverseDescription }}</p>
          </div>

          <div v-if="coin.purchasePrice || coin.currentValue" class="value-section">
            <h3>Value</h3>
            <div class="value-grid">
              <div v-if="coin.purchasePrice" class="value-item">
                <span class="value-label">Purchase Price</span>
                <span class="value-amount">{{ formatCurrency(coin.purchasePrice) }}</span>
              </div>
              <div v-if="coin.currentValue" class="value-item">
                <span class="value-label">Current Value</span>
                <span class="value-amount gold">{{ formatCurrency(coin.currentValue) }}</span>
              </div>
            </div>
            <div v-if="coin.purchaseDate" class="value-meta">
              Purchased {{ new Date(coin.purchaseDate).toLocaleDateString() }}
              <span v-if="coin.purchaseLocation"> from {{ coin.purchaseLocation }}</span>
            </div>
          </div>

          <div class="numista-section">
            <div class="numista-header">
              <h3>Numista Lookup</h3>
              <button
                class="btn btn-secondary btn-sm"
                :disabled="numistaSearching"
                @click="handleNumistaSearch"
              >
                {{ numistaSearching ? 'Searching...' : 'Search' }}
              </button>
            </div>
            <p v-if="numistaError" class="numista-error">{{ numistaError }}</p>
            <div v-if="numistaResults.length" class="numista-results">
              <a
                v-for="item in numistaResults"
                :key="item.id"
                :href="`https://en.numista.com/catalogue/pieces${item.id}.html`"
                target="_blank"
                rel="noopener"
                class="numista-card"
              >
                <img v-if="item.obverse_thumbnail" :src="item.obverse_thumbnail" class="numista-thumb" />
                <div class="numista-card-info">
                  <span class="numista-card-title">{{ item.title }}</span>
                  <span class="numista-card-meta">
                    <template v-if="item.issuer?.name">{{ item.issuer.name }}</template>
                    <template v-if="item.min_year"> · {{ item.min_year }}<template v-if="item.max_year && item.max_year !== item.min_year">–{{ item.max_year }}</template></template>
                  </span>
                </div>
              </a>
            </div>
          </div>

          <div v-if="coin.notes" class="notes-section">
            <h3>Notes</h3>
            <p>{{ coin.notes }}</p>
          </div>

          <!-- Activity Journal -->
          <div class="journal-section">
            <h3>Activity Journal</h3>
            <div class="journal-add">
              <input
                v-model="journalInput"
                type="text"
                class="form-input journal-input"
                placeholder="e.g. Cleaned, sent to grading, displayed at show..."
                @keyup.enter="handleAddJournalEntry"
              />
              <button class="btn btn-primary btn-sm" :disabled="!journalInput.trim()" @click="handleAddJournalEntry">Add</button>
            </div>
            <div v-if="journalEntries.length" class="journal-list">
              <div v-for="entry in journalEntries" :key="entry.id" class="journal-entry">
                <div class="journal-entry-content">
                  <span class="journal-entry-text">{{ entry.entry }}</span>
                  <span class="journal-entry-date">{{ formatJournalDate(entry.createdAt) }}</span>
                </div>
                <button class="btn btn-ghost btn-xs" @click="handleDeleteJournalEntry(entry.id)">✕</button>
              </div>
            </div>
            <p v-else class="journal-empty">No activity recorded yet.</p>
          </div>

          <div v-if="coin.referenceUrl" class="reference-section">
            <a :href="coin.referenceUrl" target="_blank" rel="noopener" class="btn btn-secondary btn-sm">
              🔗 {{ coin.referenceText || 'Reference Link' }}
            </a>
          </div>
        </div>

        <!-- AI Analysis -->
        <div class="detail-ai">
          <div class="ai-section">
            <div class="ai-header">
              <h4>AI Analysis</h4>
              <div class="ai-buttons">
                <button
                  class="btn btn-primary btn-sm"
                  :disabled="analyzing || !hasObverse || !ollamaAvailable"
                  :title="!ollamaAvailable ? ollamaMessage : !hasObverse ? 'No obverse image' : ''"
                  @click="handleAnalyze('obverse')"
                >
                  {{ analyzingSide === 'obverse' ? 'Analyzing...' : 'Analyze Obverse' }}
                </button>
                <button
                  class="btn btn-primary btn-sm"
                  :disabled="analyzing || !hasReverse || !ollamaAvailable"
                  :title="!ollamaAvailable ? ollamaMessage : !hasReverse ? 'No reverse image' : ''"
                  @click="handleAnalyze('reverse')"
                >
                  {{ analyzingSide === 'reverse' ? 'Analyzing...' : 'Analyze Reverse' }}
                </button>
              </div>
            </div>
            <p v-if="!ollamaAvailable" class="ai-unavailable">AI unavailable — configure Ollama in Admin → AI Configuration</p>

            <div v-if="coin.obverseAnalysis" class="ai-result-section">
              <div class="ai-result-header">
                <h5 class="ai-result-heading">Obverse Analysis</h5>
                <button class="btn btn-ghost btn-xs" @click="handleDeleteAnalysis('obverse')">Remove</button>
              </div>
              <div class="ai-content" v-html="renderedObverse"></div>
            </div>

            <div v-if="coin.reverseAnalysis" class="ai-result-section">
              <div class="ai-result-header">
                <h5 class="ai-result-heading">Reverse Analysis</h5>
                <button class="btn btn-ghost btn-xs" @click="handleDeleteAnalysis('reverse')">Remove</button>
              </div>
              <div class="ai-content" v-html="renderedReverse"></div>
            </div>

            <div v-if="coin.aiAnalysis && !coin.obverseAnalysis && !coin.reverseAnalysis" class="ai-result-section">
              <div class="ai-content" v-html="renderedLegacy"></div>
            </div>

            <p v-if="!coin.obverseAnalysis && !coin.reverseAnalysis && !coin.aiAnalysis && ollamaAvailable" class="ai-empty">
              Upload images and click an analyze button to get an expert assessment.
            </p>
          </div>
        </div>
      </div>
    </div>

    <SellModal v-if="showSellModal && coin" :coin="coin" @close="showSellModal = false" @confirm="confirmSell" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCoinsStore } from '@/stores/coins'
import ImageGallery from '@/components/ImageGallery.vue'
import SellModal from '@/components/SellModal.vue'
import { uploadImage, proxyImage, analyzeCoin, deleteAnalysis, deleteCoin, deleteImage, purchaseCoin, sellCoin, getOllamaStatus, getJournalEntries, addJournalEntry, deleteJournalEntry, searchNumista } from '@/api/client'
import { removeBackground as removeBg } from '@imgly/background-removal'
import type { CoinImage, CoinJournal, NumistaType } from '@/types'
import MarkdownIt from 'markdown-it'

const route = useRoute()
const router = useRouter()
const store = useCoinsStore()

const uploadType = ref('obverse')
const uploadStatus = ref('')
const uploadError = ref(false)
const imageUrl = ref('')
const urlLoading = ref(false)
const analyzing = ref(false)
const analyzingSide = ref<string | null>(null)
const ollamaAvailable = ref(true)
const ollamaMessage = ref('')
const removingBg = ref(false)
const showSellModal = ref(false)

// Journal
const journalEntries = ref<CoinJournal[]>([])
const journalInput = ref('')

// Numista
const numistaResults = ref<NumistaType[]>([])
const numistaSearching = ref(false)
const numistaError = ref('')

const md = new MarkdownIt()

const coin = computed(() => store.currentCoin)
const renderedObverse = computed(() => (coin.value?.obverseAnalysis ? md.render(coin.value.obverseAnalysis) : ''))
const renderedReverse = computed(() => (coin.value?.reverseAnalysis ? md.render(coin.value.reverseAnalysis) : ''))
const renderedLegacy = computed(() => (coin.value?.aiAnalysis ? md.render(coin.value.aiAnalysis) : ''))
const hasObverse = computed(() => coin.value?.images?.some(i => i.imageType === 'obverse'))
const hasReverse = computed(() => coin.value?.images?.some(i => i.imageType === 'reverse'))

onMounted(async () => {
  const id = Number(route.params['id'])
  store.fetchCoin(id)
  loadJournal(id)
  try {
    const res = await getOllamaStatus()
    ollamaAvailable.value = res.data.available
    ollamaMessage.value = res.data.message
  } catch {
    ollamaAvailable.value = false
    ollamaMessage.value = 'Unable to check Ollama status'
  }
})

async function loadJournal(coinId: number) {
  try {
    const res = await getJournalEntries(coinId)
    journalEntries.value = res.data || []
  } catch {
    journalEntries.value = []
  }
}

async function handleAddJournalEntry() {
  if (!coin.value || !journalInput.value.trim()) return
  try {
    await addJournalEntry(coin.value.id, journalInput.value.trim())
    journalInput.value = ''
    loadJournal(coin.value.id)
  } catch {
    alert('Failed to add journal entry')
  }
}

async function handleDeleteJournalEntry(entryId: number) {
  if (!coin.value) return
  try {
    await deleteJournalEntry(coin.value.id, entryId)
    loadJournal(coin.value.id)
  } catch {
    alert('Failed to delete journal entry')
  }
}

function formatJournalDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

async function handleNumistaSearch() {
  if (!coin.value) return
  numistaSearching.value = true
  numistaError.value = ''
  numistaResults.value = []
  try {
    const q = [coin.value.name, coin.value.denomination, coin.value.ruler].filter(Boolean).join(' ')
    const res = await searchNumista(q)
    numistaResults.value = res.data.types || []
    if (!numistaResults.value.length) {
      numistaError.value = 'No results found on Numista'
    }
  } catch (err: any) {
    numistaError.value = err.response?.data?.error || 'Numista search failed'
  } finally {
    numistaSearching.value = false
  }
}

async function handleImageUpload(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file || !coin.value) return

  uploadStatus.value = 'Uploading...'
  uploadError.value = false

  try {
    await uploadImage(coin.value.id, file, uploadType.value, coin.value.images?.length === 0)
    uploadStatus.value = 'Upload complete!'
    store.fetchCoin(coin.value.id)
  } catch {
    uploadStatus.value = 'Upload failed'
    uploadError.value = true
  }
}

async function handleUrlUpload() {
  if (!imageUrl.value || !coin.value) return

  urlLoading.value = true
  uploadStatus.value = 'Fetching image...'
  uploadError.value = false

  try {
    const imgRes = await proxyImage(imageUrl.value)
    const blob = imgRes.data as Blob
    if (blob.size === 0) {
      uploadStatus.value = 'No image data received from URL'
      uploadError.value = true
      return
    }
    const ext = blob.type.includes('png') ? '.png' : '.jpg'
    const file = new File([blob], `${uploadType.value}${ext}`, { type: blob.type || 'image/jpeg' })
    await uploadImage(coin.value!.id, file, uploadType.value, coin.value!.images?.length === 0)
    uploadStatus.value = 'Image saved from URL!'
    imageUrl.value = ''
    store.fetchCoin(coin.value!.id)
  } catch {
    uploadStatus.value = 'Failed to fetch image from URL'
    uploadError.value = true
  } finally {
    urlLoading.value = false
  }
}

async function handleRemoveBackground(image: CoinImage) {
  if (!coin.value) return
  removingBg.value = true
  uploadStatus.value = ''
  uploadError.value = false

  try {
    // Fetch the original image
    const response = await fetch(`/uploads/${image.filePath}`)
    const srcBlob = await response.blob()

    // Remove background using @imgly/background-removal
    const resultBlob = await removeBg(srcBlob, {
      output: { format: 'image/png', quality: 1 },
    })

    // Upload the processed image with the same type and primary status
    const file = new File([resultBlob], `${image.imageType}-processed.png`, { type: 'image/png' })
    await uploadImage(coin.value.id, file, image.imageType, image.isPrimary)

    // Delete the old image
    await deleteImage(coin.value.id, image.id)

    uploadStatus.value = 'Background removed!'
    store.fetchCoin(coin.value.id)
  } catch (err) {
    console.error('Background removal failed:', err)
    uploadStatus.value = 'Background removal failed'
    uploadError.value = true
  } finally {
    removingBg.value = false
  }
}

async function handleDeleteImage(image: CoinImage) {
  if (!coin.value || !confirm(`Delete this ${image.imageType} image?`)) return
  try {
    await deleteImage(coin.value.id, image.id)
    store.fetchCoin(coin.value.id)
  } catch {
    alert('Failed to delete image')
  }
}

async function handleAnalyze(side: 'obverse' | 'reverse') {
  if (!coin.value) return
  analyzing.value = true
  analyzingSide.value = side
  try {
    await analyzeCoin(coin.value.id, side)
    store.fetchCoin(coin.value.id)
  } catch {
    alert(`AI analysis failed for ${side}. Ensure Ollama is running.`)
  } finally {
    analyzing.value = false
    analyzingSide.value = null
  }
}

async function handlePurchase() {
  if (!coin.value || !confirm(`Move "${coin.value.name}" to your collection?`)) return
  try {
    await purchaseCoin(coin.value.id)
    store.fetchCoin(coin.value.id)
  } catch {
    alert('Failed to mark as purchased')
  }
}

async function handleDelete() {
  if (!coin.value || !confirm('Delete this coin from your collection?')) return
  await deleteCoin(coin.value.id)
  router.push('/')
}

async function confirmSell(soldPrice: number | null, soldTo: string) {
  if (!coin.value) return
  try {
    await sellCoin(coin.value.id, soldPrice, soldTo)
    showSellModal.value = false
    router.push('/sold')
  } catch {
    alert('Failed to mark as sold')
    showSellModal.value = false
  }
}

async function handleDeleteAnalysis(side: 'obverse' | 'reverse') {
  if (!coin.value || !confirm(`Delete the ${side} analysis?`)) return
  try {
    await deleteAnalysis(coin.value.id, side)
    store.fetchCoin(coin.value.id)
  } catch {
    alert(`Failed to delete ${side} analysis`)
  }
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)
}
</script>

<style scoped>
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.detail-actions {
  display: flex;
  gap: 0.5rem;
}

.detail-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: start;
  max-width: 1000px;
  margin-left: auto;
  margin-right: auto;
}

.detail-ai {
  grid-column: 1 / -1;
}

.detail-title-section {
  margin-bottom: 1.5rem;
}

.detail-title-section h1 {
  margin-top: 0.5rem;
}

.detail-ruler {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin-top: 0.25rem;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.info-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.75rem;
}

.info-label {
  display: block;
  font-size: 0.7rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.2rem;
}

.info-value {
  font-size: 0.95rem;
  font-weight: 500;
}

.info-value.gold {
  color: var(--accent-gold);
}

.inscriptions-section,
.descriptions-section,
.value-section,
.notes-section,
.journal-section {
  margin-bottom: 1.5rem;
}

.inscriptions-section h3,
.descriptions-section h3,
.value-section h3,
.notes-section h3,
.journal-section h3 {
  margin-bottom: 0.75rem;
  font-size: 1rem;
}

.inscription {
  margin-bottom: 0.4rem;
}

.inscription-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-right: 0.4rem;
}

.inscription-text {
  font-style: italic;
  color: var(--text-secondary);
}

.descriptions-section p {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 0.4rem;
}

.value-grid {
  display: flex;
  gap: 1.5rem;
}

.value-item {
  display: flex;
  flex-direction: column;
}

.value-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
}

.value-amount {
  font-size: 1.2rem;
  font-weight: 600;
}

.value-amount.gold {
  color: var(--accent-gold);
}

.value-meta {
  margin-top: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.notes-section p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  white-space: pre-wrap;
}

.reference-section {
  margin-top: 1rem;
}

/* Journal */
.journal-add {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.journal-input {
  flex: 1;
}

.journal-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.journal-entry {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
}

.journal-entry-content {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}

.journal-entry-text {
  font-size: 0.85rem;
}

.journal-entry-date {
  font-size: 0.7rem;
  color: var(--text-muted);
}

.journal-empty {
  font-size: 0.85rem;
  color: var(--text-muted);
  font-style: italic;
}

/* Numista */
.numista-section {
  margin-bottom: 1.5rem;
}

.numista-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.numista-header h3 {
  font-size: 1rem;
  margin: 0;
}

.numista-error {
  font-size: 0.85rem;
  color: #e67e22;
  font-style: italic;
}

.numista-results {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 0.75rem;
}

.numista-card {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  text-decoration: none;
  color: inherit;
  transition: border-color var(--transition-fast);
}

.numista-card:hover {
  border-color: var(--accent-gold);
}

.numista-thumb {
  width: 48px;
  height: 48px;
  object-fit: contain;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.numista-card-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.numista-card-title {
  font-size: 0.85rem;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.numista-card-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* Image upload */
.image-upload-section {
  margin-top: 1.25rem;
  padding: 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.image-upload-section h4 {
  font-size: 0.9rem;
  margin-bottom: 0.75rem;
}

.upload-row {
  display: flex;
  gap: 0.5rem;
}

.upload-select {
  flex: 1;
}

.upload-btn {
  white-space: nowrap;
  cursor: pointer;
}

.upload-status {
  margin-top: 0.5rem;
  font-size: 0.8rem;
  color: var(--accent-gold);
}

.upload-status.error {
  color: #e74c3c;
}

.url-upload-row {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.url-input {
  flex: 1;
  min-width: 0;
  font-size: 0.82rem;
}

/* AI section */
.ai-section {
  margin-top: 1.25rem;
  padding: 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.ai-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.ai-header h4 {
  font-size: 0.9rem;
}

.ai-buttons {
  display: flex;
  gap: 0.4rem;
}

.ai-result-section {
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-subtle);
}

.ai-result-section:last-of-type {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.ai-result-heading {
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--accent-gold);
  margin-bottom: 0.5rem;
}

.ai-result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ai-result-header .ai-result-heading {
  margin-bottom: 0;
}

.btn-ghost {
  background: transparent;
  border: 1px solid var(--border-subtle);
  color: var(--text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-ghost:hover {
  color: #e74c3c;
  border-color: #e74c3c;
}

.btn-xs {
  padding: 0.15rem 0.45rem;
  font-size: 0.7rem;
  border-radius: var(--radius-sm);
}

.ai-content {
  font-size: 0.85rem;
  line-height: 1.7;
  color: var(--text-secondary);
}

.ai-content :deep(h1),
.ai-content :deep(h2),
.ai-content :deep(h3) {
  color: var(--accent-gold);
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}

.ai-content :deep(strong) {
  color: var(--text-primary);
}

.ai-content :deep(ul),
.ai-content :deep(ol) {
  padding-left: 1.25rem;
}

.ai-empty {
  font-size: 0.85rem;
  color: var(--text-muted);
  font-style: italic;
}

.ai-unavailable {
  font-size: 0.85rem;
  color: #e67e22;
  font-style: italic;
  margin-bottom: 0.5rem;
}

@media (max-width: 768px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
  .detail-images { order: 1; }
  .detail-info { order: 2; }
  .detail-ai { order: 3; }
  .info-grid {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
