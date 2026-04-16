<template>
  <div class="coin-card card" @click="$router.push(`/coin/${coin.id}`)">
    <div class="card-image-container">
      <img v-if="primaryImage" :src="primaryImage" :alt="coin.name" class="card-image" />
      <div v-else class="card-image-placeholder"><Coins :size="48" :stroke-width="1" /></div>
      <div v-if="wishlist && coin.listingStatus === 'unavailable'" class="listing-overlay"></div>
      <span v-if="wishlist && coin.listingStatus === 'unavailable'" class="listing-badge listing-badge-unavailable">Unavailable</span>
      <button
        v-if="wishlist && coin.listingStatus === 'unavailable'"
        class="listing-dismiss-btn"
        @click.stop="emit('dismiss-status', coin.id)"
      >Dismiss</button>
    </div>
    <div class="card-body">
      <h3 class="card-title">
        <span
          v-if="wishlist && coin.listingStatus === 'available'"
          class="status-dot status-dot-available"
          title="Available"
        ></span>
        <span
          v-if="wishlist && coin.listingStatus === 'unknown'"
          class="status-dot status-dot-unknown"
          title="Unknown"
        ></span>
        {{ coin.name }}
      </h3>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Coin, ImageType } from '@/types'
import { computed } from 'vue'
import { Coins, ShoppingCart } from 'lucide-vue-next'

const props = withDefaults(defineProps<{ coin: Coin; imageSide?: ImageType | null; wishlist?: boolean; sold?: boolean }>(), {
  imageSide: null,
  wishlist: false,
  sold: false,
})

const emit = defineEmits<{
  purchase: [coin: Coin]
  'dismiss-status': [coinId: number]
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
  background: radial-gradient(ellipse at center, var(--bg-secondary) 0%, var(--bg-primary) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Vignette overlay + gold accent border */
.card-image-container::after {
  content: '';
  position: absolute;
  inset: 0;
  box-shadow: inset 0 0 40px rgba(0, 0, 0, 0.35);
  border-bottom: 1px solid var(--accent-gold-dim);
  pointer-events: none;
  z-index: 1;
  transition: box-shadow var(--transition-med);
}

.coin-card:hover .card-image-container::after {
  box-shadow: inset 0 0 25px rgba(0, 0, 0, 0.2),
              0 0 20px var(--accent-gold-glow);
}

.card-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-med), filter var(--transition-med);
}

.coin-card:hover .card-image {
  transform: scale(1.05);
  filter: brightness(1.1);
}

/* PWA: taller image area for more prominence */
@media (display-mode: standalone) {
  .card-image-container {
    aspect-ratio: 5 / 6;
  }
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

/* Listing status overlay & badge */
.listing-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 2;
  pointer-events: none;
}

.listing-badge {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-full);
  font-size: 0.7rem;
  font-weight: 600;
  z-index: 3;
}

.listing-badge-unavailable {
  background: rgba(231, 76, 60, 0.85);
  color: #fff;
}

.listing-dismiss-btn {
  position: absolute;
  bottom: 0.5rem;
  right: 0.5rem;
  padding: 0.15rem 0.5rem;
  font-size: 0.65rem;
  background: rgba(0, 0, 0, 0.7);
  color: var(--text-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: pointer;
  z-index: 3;
}

.listing-dismiss-btn:hover {
  color: var(--text-primary);
  background: rgba(0, 0, 0, 0.85);
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 0.35rem;
  vertical-align: middle;
  flex-shrink: 0;
}

.status-dot-available {
  background: #2ecc71;
}

.status-dot-unknown {
  background: #f1c40f;
}
</style>
