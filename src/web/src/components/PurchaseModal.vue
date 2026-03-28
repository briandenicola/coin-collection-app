<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal card">
      <h3>Move to Collection</h3>
      <p class="purchase-coin-name">{{ coin.name }}</p>

      <div class="form-group">
        <label class="form-label">Purchase Price</label>
        <div class="price-input-wrapper">
          <span class="price-prefix">$</span>
          <input
            ref="priceInput"
            v-model="priceStr"
            type="number"
            step="0.01"
            min="0"
            class="form-input price-input"
            placeholder="0.00"
          />
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">Purchase Date</label>
        <input
          v-model="purchaseDate"
          type="date"
          class="form-input"
        />
      </div>

      <div class="form-group">
        <label class="form-label">Purchased From</label>
        <input
          v-model="purchaseLocation"
          type="text"
          class="form-input"
          placeholder="e.g. VCoins, Heritage Auctions"
        />
      </div>

      <p class="optional-note">All fields are optional. You can update these later.</p>

      <div v-if="error" class="purchase-error">{{ error }}</div>

      <div class="modal-actions">
        <button class="btn btn-secondary" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? 'Saving...' : 'Add to Collection' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Coin } from '@/types'

const props = defineProps<{
  coin: Coin
}>()

const emit = defineEmits<{
  close: []
  confirm: [data: { purchasePrice?: number; purchaseDate?: string; purchaseLocation?: string }]
}>()

const priceStr = ref('')
const purchaseDate = ref(new Date().toISOString().slice(0, 10))
const purchaseLocation = ref('')
const error = ref('')
const submitting = ref(false)
const priceInput = ref<HTMLInputElement>()

onMounted(() => {
  if (props.coin.purchasePrice) {
    priceStr.value = String(props.coin.purchasePrice)
  }
  if (props.coin.purchaseLocation) {
    purchaseLocation.value = props.coin.purchaseLocation
  }
  priceInput.value?.focus()
})

function handleSubmit() {
  error.value = ''

  const data: { purchasePrice?: number; purchaseDate?: string; purchaseLocation?: string } = {}

  if (priceStr.value) {
    const price = parseFloat(priceStr.value)
    if (isNaN(price) || price < 0) {
      error.value = 'Please enter a valid price'
      return
    }
    data.purchasePrice = price
  }

  if (purchaseDate.value) {
    data.purchaseDate = purchaseDate.value
  }

  if (purchaseLocation.value.trim()) {
    data.purchaseLocation = purchaseLocation.value.trim()
  }

  submitting.value = true
  emit('confirm', data)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  max-width: 420px;
  width: 90%;
  padding: 2rem;
}

.modal h3 {
  margin-bottom: 0.25rem;
  font-size: 1.1rem;
}

.purchase-coin-name {
  color: var(--accent-gold);
  font-size: 0.9rem;
  margin-bottom: 1.25rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  font-size: 0.82rem;
  font-weight: 600;
  margin-bottom: 0.35rem;
  color: var(--text-secondary);
}

.price-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.price-prefix {
  position: absolute;
  left: 0.75rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
  pointer-events: none;
}

.price-input {
  padding-left: 1.5rem;
}

.price-input::-webkit-inner-spin-button,
.price-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.price-input[type='number'] {
  -moz-appearance: textfield;
}

.optional-note {
  font-size: 0.78rem;
  color: var(--text-secondary);
  margin-bottom: 0.75rem;
}

.purchase-error {
  color: #e74c3c;
  font-size: 0.82rem;
  margin-bottom: 0.75rem;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}
</style>
