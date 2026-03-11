<template>
  <form class="coin-form" @submit.prevent="$emit('submit')">
    <div class="form-grid">
      <!-- Basic Info -->
      <fieldset class="form-section">
        <legend>Basic Information</legend>
        <div class="form-group">
          <label class="form-label">Name *</label>
          <AutocompleteInput v-model="form.name!" field="name" required placeholder="e.g. Augustus Denarius" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Category</label>
            <select v-model="form.category" class="form-select">
              <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">Material</label>
            <select v-model="form.material" class="form-select">
              <option v-for="m in MATERIALS" :key="m" :value="m">{{ m }}</option>
            </select>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Denomination</label>
            <AutocompleteInput v-model="form.denomination!" field="denomination" placeholder="e.g. Denarius" />
          </div>
          <div class="form-group">
            <label class="form-label">Mint</label>
            <input v-model="form.mint" class="form-input" placeholder="e.g. Rome" />
          </div>
        </div>
      </fieldset>

      <!-- Physical Details -->
      <fieldset class="form-section">
        <legend>Physical Details</legend>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Weight (grams)</label>
            <input v-model.number="form.weightGrams" class="form-input" type="number" step="0.01" />
          </div>
          <div class="form-group">
            <label class="form-label">Diameter (mm)</label>
            <input v-model.number="form.diameterMm" class="form-input" type="number" step="0.1" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Grade</label>
            <input v-model="form.grade" class="form-input" placeholder="e.g. VF, EF, MS-65" />
          </div>
          <div class="form-group">
            <label class="form-label">Rarity Rating (RIC)</label>
            <input v-model="form.rarityRating" class="form-input" placeholder="e.g. RIC 207" />
          </div>
        </div>
      </fieldset>

      <!-- Inscriptions -->
      <fieldset class="form-section full-width">
        <legend>Inscriptions & Descriptions</legend>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Obverse Inscription</label>
            <input v-model="form.obverseInscription" class="form-input" placeholder="Obverse legend text" />
          </div>
          <div class="form-group">
            <label class="form-label">Reverse Inscription</label>
            <input v-model="form.reverseInscription" class="form-input" placeholder="Reverse legend text" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Obverse Description</label>
            <textarea v-model="form.obverseDescription" class="form-textarea" placeholder="Describe the obverse design" />
          </div>
          <div class="form-group">
            <label class="form-label">Reverse Description</label>
            <textarea v-model="form.reverseDescription" class="form-textarea" placeholder="Describe the reverse design" />
          </div>
        </div>
      </fieldset>

      <!-- Images -->
      <fieldset class="form-section full-width">
        <legend>Images</legend>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Obverse Image</label>
            <div v-if="obversePreview || existingObverse" class="image-preview-box">
              <img :src="obversePreview || existingObverse!" alt="Obverse" class="image-preview" />
              <button type="button" class="image-remove-btn" @click="clearObverse" title="Remove">✕</button>
            </div>
            <input type="file" accept=".jpg,.jpeg,.png" class="form-input file-input" @change="onObverseFile" ref="obverseInput" />
          </div>
          <div class="form-group">
            <label class="form-label">Reverse Image</label>
            <div v-if="reversePreview || existingReverse" class="image-preview-box">
              <img :src="reversePreview || existingReverse!" alt="Reverse" class="image-preview" />
              <button type="button" class="image-remove-btn" @click="clearReverse" title="Remove">✕</button>
            </div>
            <input type="file" accept=".jpg,.jpeg,.png" class="form-input file-input" @change="onReverseFile" ref="reverseInput" />
          </div>
        </div>
      </fieldset>

      <!-- Purchase Info -->
      <fieldset class="form-section">
        <legend>Purchase & Value</legend>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Purchase Price ($)</label>
            <input v-model.number="form.purchasePrice" class="form-input" type="number" step="0.01" />
          </div>
          <div class="form-group">
            <label class="form-label">Current Value ($)</label>
            <input v-model.number="form.currentValue" class="form-input" type="number" step="0.01" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Purchase Date</label>
            <input v-model="form.purchaseDate" class="form-input" type="date" />
          </div>
          <div class="form-group">
            <label class="form-label">Purchase Location</label>
            <AutocompleteInput v-model="form.purchaseLocation!" field="purchaseLocation" placeholder="e.g. Heritage Auctions" />
          </div>
        </div>
      </fieldset>

      <!-- Links & Notes -->
      <fieldset class="form-section">
        <legend>Reference & Notes</legend>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Reference URL</label>
            <input v-model="form.referenceUrl" class="form-input" type="url" placeholder="https://..." />
          </div>
          <div class="form-group">
            <label class="form-label">Reference Text</label>
            <input v-model="form.referenceText" class="form-input" placeholder="Link display text" />
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">Notes</label>
          <textarea v-model="form.notes" class="form-textarea" rows="3" placeholder="Any additional notes..." />
        </div>
        <div class="form-group">
          <label class="form-check">
            <input v-model="form.isWishlist" type="checkbox" />
            <span>Add to wishlist (not yet owned)</span>
          </label>
        </div>
      </fieldset>
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$router.back()">Cancel</button>
      <button type="submit" class="btn btn-primary" :disabled="loading">
        {{ loading ? 'Saving...' : submitLabel }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { CATEGORIES, MATERIALS } from '@/types'
import type { Coin } from '@/types'
import AutocompleteInput from '@/components/AutocompleteInput.vue'

const props = defineProps<{
  form: Partial<Coin>
  submitLabel: string
  loading?: boolean
  coinId?: number
}>()

defineEmits<{ submit: [] }>()

const obverseFile = ref<File | null>(null)
const reverseFile = ref<File | null>(null)
const obversePreview = ref<string | null>(null)
const reversePreview = ref<string | null>(null)
const obverseInput = ref<HTMLInputElement | null>(null)
const reverseInput = ref<HTMLInputElement | null>(null)
const removedObverseId = ref<number | null>(null)
const removedReverseId = ref<number | null>(null)

const existingObverse = computed(() => {
  if (removedObverseId.value) return null
  const img = props.form.images?.find((i) => i.imageType === 'obverse')
  return img ? `/uploads/${img.filePath}` : null
})

const existingReverse = computed(() => {
  if (removedReverseId.value) return null
  const img = props.form.images?.find((i) => i.imageType === 'reverse')
  return img ? `/uploads/${img.filePath}` : null
})

function onObverseFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  obverseFile.value = file
  obversePreview.value = URL.createObjectURL(file)
}

function onReverseFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  reverseFile.value = file
  reversePreview.value = URL.createObjectURL(file)
}

function clearObverse() {
  const existing = props.form.images?.find((i) => i.imageType === 'obverse')
  if (existing) removedObverseId.value = existing.id
  obverseFile.value = null
  obversePreview.value = null
  if (obverseInput.value) obverseInput.value.value = ''
}

function clearReverse() {
  const existing = props.form.images?.find((i) => i.imageType === 'reverse')
  if (existing) removedReverseId.value = existing.id
  reverseFile.value = null
  reversePreview.value = null
  if (reverseInput.value) reverseInput.value.value = ''
}

// Expose pending images for parent to upload after save
defineExpose({
  obverseFile,
  reverseFile,
  removedObverseId,
  removedReverseId,
})
</script>

<style scoped>
.coin-form {
  max-width: 900px;
  margin-left: auto;
  margin-right: auto;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.form-section {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  background: var(--bg-card);
}

.form-section.full-width {
  grid-column: 1 / -1;
}

legend {
  font-family: 'Cinzel', serif;
  color: var(--accent-gold);
  font-size: 0.95rem;
  font-weight: 500;
  padding: 0 0.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.form-check {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.form-check input[type='checkbox'] {
  accent-color: var(--accent-gold);
  width: 16px;
  height: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-subtle);
}

.file-input {
  font-size: 0.85rem;
  padding: 0.4rem;
}

.image-preview-box {
  position: relative;
  display: inline-block;
  margin-bottom: 0.5rem;
}

.image-preview {
  width: 140px;
  height: 140px;
  object-fit: cover;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.image-remove-btn {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: none;
  background: #c0392b;
  color: white;
  font-size: 0.7rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>
