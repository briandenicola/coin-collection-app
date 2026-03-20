<template>
  <div class="coin-card card" @click="$router.push(`/coin/${coin.id}`)">
    <div class="card-image-container">
      <img v-if="primaryImage" :src="primaryImage" :alt="coin.name" class="card-image" />
      <div v-else class="card-image-placeholder"><Coins :size="48" :stroke-width="1" /></div>
    </div>
    <div class="card-body">
      <h3 class="card-title">{{ coin.name }}</h3>
      <template v-if="!wishlist && !sold">
        <div class="card-meta">
          <span v-if="coin.ruler" class="meta-item">{{ coin.ruler }}</span>
          <span v-if="coin.era" class="meta-item">{{ coin.era }}</span>
        </div>
        <div class="card-details">
          <span v-if="coin.category" class="detail" :class="`category-${coin.category.toLowerCase()}`">{{ coin.category }}</span>
          <span v-if="coin.denomination" class="detail">{{ coin.denomination }}</span>
          <span v-if="coin.material" class="detail" :class="`material-${coin.material.toLowerCase()}`">
            {{ coin.material }}
          </span>
        </div>
        <div v-if="coin.grade" class="card-grade">{{ coin.grade }}</div>
      </template>
      <template v-if="sold">
        <div class="card-sold-info">
          <div v-if="coin.soldPrice" class="card-sold-price">Sold: {{ formatCurrency(coin.soldPrice) }}</div>
          <div v-if="coin.purchasePrice" class="card-cost-basis">Paid: {{ formatCurrency(coin.purchasePrice) }}</div>
          <div v-if="coin.soldPrice && coin.purchasePrice" class="card-profit" :class="{ loss: coin.soldPrice < coin.purchasePrice }">
            {{ coin.soldPrice >= coin.purchasePrice ? '+' : '' }}{{ formatCurrency(coin.soldPrice - coin.purchasePrice) }}
          </div>
          <div v-if="coin.soldTo" class="card-sold-to">To: {{ coin.soldTo }}</div>
        </div>
      </template>
      <div v-if="!sold && (coin.currentValue || coin.purchasePrice)" class="card-value">
        {{ formatCurrency(coin.currentValue || coin.purchasePrice || 0) }}
      </div>
      <a
        v-if="wishlist && coin.referenceUrl"
        :href="coin.referenceUrl"
        class="card-reference"
        target="_blank"
        rel="noopener noreferrer"
        @click.stop
      >
        {{ coin.referenceText || coin.referenceUrl }}
      </a>
      <button
        v-if="wishlist"
        class="btn btn-primary btn-sm card-purchase-btn"
        @click.stop="emit('purchase', coin)"
      >
        <ShoppingCart :size="14" /> Purchased
      </button>
      <button
        v-if="!wishlist && !sold && !coin.isSold"
        class="btn btn-secondary btn-sm card-sell-btn"
        @click.stop="emit('sell', coin)"
      >
        <DollarSign :size="14" /> Sell
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Coin, ImageType } from '@/types'
import { computed } from 'vue'
import { Coins, ShoppingCart, DollarSign } from 'lucide-vue-next'

const props = withDefaults(defineProps<{ coin: Coin; imageSide?: ImageType | null; wishlist?: boolean; sold?: boolean }>(), {
  imageSide: null,
  wishlist: false,
  sold: false,
})

const emit = defineEmits<{
  purchase: [coin: Coin]
  sell: [coin: Coin]
}>()

const primaryImage = computed(() => {
  if (props.imageSide) {
    const byType = props.coin.images?.find((img) => img.imageType === props.imageSide)
    if (byType) return `/uploads/${byType.filePath}`
  }
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

.card-reference {
  font-size: 0.8rem;
  color: var(--accent-gold);
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.card-reference:hover {
  text-decoration: underline;
}

.card-purchase-btn {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.5rem;
  width: 100%;
  justify-content: center;
  font-size: 0.8rem;
}

.card-sell-btn {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.5rem;
  width: 100%;
  justify-content: center;
  font-size: 0.8rem;
}

.card-sold-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  margin-top: 0.25rem;
  font-size: 0.82rem;
}

.card-sold-price {
  font-weight: 600;
  color: var(--accent-gold);
}

.card-cost-basis {
  color: var(--text-muted);
  font-size: 0.78rem;
}

.card-profit {
  font-weight: 600;
  color: #4ade80;
  font-size: 0.82rem;
}

.card-profit.loss {
  color: #f87171;
}

.card-sold-to {
  font-size: 0.78rem;
  color: var(--text-secondary);
  margin-top: 0.15rem;
}

.category-roman { color: #b57edc; }
.category-greek { color: #9ab85a; }
.category-byzantine { color: #e67e73; }
.category-modern { color: #7ab3d4; }
.category-other { color: #aaa; }
</style>
