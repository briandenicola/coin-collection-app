<template>
  <div class="coin-card card" @click="$router.push(`/coin/${coin.id}`)">
    <div class="card-image-container">
      <img v-if="primaryImage" :src="primaryImage" :alt="coin.name" class="card-image" />
      <div v-else class="card-image-placeholder"><Coins :size="48" :stroke-width="1" /></div>
      <span class="badge" :class="`badge-${coin.category.toLowerCase()}`">{{ coin.category }}</span>
    </div>
    <div class="card-body">
      <h3 class="card-title">{{ coin.name }}</h3>
      <div class="card-meta">
        <span v-if="coin.ruler" class="meta-item">{{ coin.ruler }}</span>
        <span v-if="coin.era" class="meta-item">{{ coin.era }}</span>
      </div>
      <div class="card-details">
        <span v-if="coin.denomination" class="detail">{{ coin.denomination }}</span>
        <span v-if="coin.material" class="detail" :class="`material-${coin.material.toLowerCase()}`">
          {{ coin.material }}
        </span>
      </div>
      <div v-if="coin.grade" class="card-grade">{{ coin.grade }}</div>
      <div v-if="coin.currentValue || coin.purchasePrice" class="card-value">
        {{ formatCurrency(coin.currentValue || coin.purchasePrice || 0) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Coin } from '@/types'
import { computed } from 'vue'
import { Coins } from 'lucide-vue-next'

const props = defineProps<{ coin: Coin }>()

const primaryImage = computed(() => {
  const primary = props.coin.images?.find((img) => img.isPrimary)
  const first = props.coin.images?.[0]
  const img = primary || first
  return img ? `/uploads/${img.filePath}` : null
})

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)
}
</script>

<style scoped>
.coin-card {
  cursor: pointer;
  overflow: hidden;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.card-image-container {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  overflow: hidden;
  background: var(--bg-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-med);
}

.coin-card:hover .card-image {
  transform: scale(1.05);
}

.card-image-placeholder {
  font-size: 4rem;
  opacity: 0.3;
}

.card-image-container .badge {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
}

.card-body {
  padding: 1rem;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.card-title {
  font-size: 1rem;
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.meta-item {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.card-details {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.detail {
  font-size: 0.75rem;
  padding: 0.15rem 0.5rem;
  background: var(--bg-primary);
  border-radius: var(--radius-full);
  color: var(--text-secondary);
}

.card-grade {
  font-size: 0.8rem;
  color: var(--accent-gold);
  font-weight: 600;
}

.card-value {
  margin-top: auto;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--accent-gold);
}
</style>
