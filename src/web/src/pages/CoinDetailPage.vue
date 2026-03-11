<template>
  <div class="container">
    <div v-if="store.loading && !coin" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="coin" class="coin-detail">
      <div class="detail-header">
        <button class="btn btn-secondary btn-sm" @click="$router.back()">← Back</button>
        <div class="detail-actions">
          <router-link :to="`/edit/${coin.id}`" class="btn btn-secondary btn-sm">Edit</router-link>
          <button class="btn btn-danger btn-sm" @click="handleDelete">Delete</button>
        </div>
      </div>

      <div class="detail-layout">
        <!-- Images -->
        <div class="detail-images">
          <ImageGallery :images="coin.images || []" />

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
            <p v-if="uploadStatus" class="upload-status" :class="{ error: uploadError }">{{ uploadStatus }}</p>
          </div>

          <!-- AI Analysis -->
          <div class="ai-section">
            <div class="ai-header">
              <h4>AI Analysis</h4>
              <button
                class="btn btn-primary btn-sm"
                :disabled="analyzing || !coin.images?.length"
                @click="handleAnalyze"
              >
                {{ analyzing ? 'Analyzing...' : 'Analyze with AI' }}
              </button>
            </div>
            <div v-if="coin.aiAnalysis" class="ai-content" v-html="renderedAnalysis"></div>
            <p v-else class="ai-empty">Upload images and click "Analyze with AI" to get an expert assessment.</p>
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

          <div v-if="coin.notes" class="notes-section">
            <h3>Notes</h3>
            <p>{{ coin.notes }}</p>
          </div>

          <div v-if="coin.referenceUrl" class="reference-section">
            <a :href="coin.referenceUrl" target="_blank" rel="noopener" class="btn btn-secondary btn-sm">
              🔗 {{ coin.referenceText || 'Reference Link' }}
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCoinsStore } from '@/stores/coins'
import ImageGallery from '@/components/ImageGallery.vue'
import { uploadImage, analyzeCoin, deleteCoin } from '@/api/client'
import MarkdownIt from 'markdown-it'

const route = useRoute()
const router = useRouter()
const store = useCoinsStore()

const uploadType = ref('obverse')
const uploadStatus = ref('')
const uploadError = ref(false)
const analyzing = ref(false)

const md = new MarkdownIt()

const coin = computed(() => store.currentCoin)
const renderedAnalysis = computed(() => (coin.value?.aiAnalysis ? md.render(coin.value.aiAnalysis) : ''))

onMounted(() => {
  const id = Number(route.params['id'])
  store.fetchCoin(id)
})

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

async function handleAnalyze() {
  if (!coin.value) return
  analyzing.value = true
  try {
    await analyzeCoin(coin.value.id)
    store.fetchCoin(coin.value.id)
  } catch {
    alert('AI analysis failed. Ensure Ollama is running.')
  } finally {
    analyzing.value = false
  }
}

async function handleDelete() {
  if (!coin.value || !confirm('Delete this coin from your collection?')) return
  await deleteCoin(coin.value.id)
  router.push('/')
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
.notes-section {
  margin-bottom: 1.5rem;
}

.inscriptions-section h3,
.descriptions-section h3,
.value-section h3,
.notes-section h3 {
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

@media (max-width: 768px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
  .info-grid {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
