<template>
  <div class="container">
    <div class="page-header">
      <h1>📊 Collection Stats</h1>
    </div>

    <div v-if="!stats" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else class="stats-layout">
      <!-- Summary Cards -->
      <div class="stats-summary">
        <div class="stat-card">
          <span class="stat-number">{{ stats.totalCoins }}</span>
          <span class="stat-label">Coins Owned</span>
        </div>
        <div class="stat-card">
          <span class="stat-number">{{ stats.totalWishlist }}</span>
          <span class="stat-label">On Wishlist</span>
        </div>
        <div class="stat-card">
          <span class="stat-number gold">{{ formatCurrency(stats.values.totalCurrentValue) }}</span>
          <span class="stat-label">Total Value</span>
        </div>
        <div class="stat-card">
          <span class="stat-number">{{ formatCurrency(stats.values.totalPurchasePrice) }}</span>
          <span class="stat-label">Total Invested</span>
        </div>
      </div>

      <!-- Category Breakdown -->
      <div class="stats-section card">
        <h2>By Category</h2>
        <div class="bar-chart">
          <div v-for="item in stats.byCategory" :key="item.category" class="bar-row">
            <span class="bar-label">
              <span class="badge" :class="`badge-${item.category.toLowerCase()}`">{{ item.category }}</span>
            </span>
            <div class="bar-track">
              <div
                class="bar-fill"
                :class="`fill-${item.category.toLowerCase()}`"
                :style="{ width: `${(item.count / maxCategoryCount) * 100}%` }"
              ></div>
            </div>
            <span class="bar-value">{{ item.count }}</span>
          </div>
        </div>
      </div>

      <!-- Material Breakdown -->
      <div class="stats-section card">
        <h2>By Material</h2>
        <div class="bar-chart">
          <div v-for="item in stats.byMaterial" :key="item.material" class="bar-row">
            <span class="bar-label">
              <span :class="`material-${item.material.toLowerCase()}`">{{ item.material }}</span>
            </span>
            <div class="bar-track">
              <div
                class="bar-fill fill-material"
                :style="{ width: `${(item.count / maxMaterialCount) * 100}%` }"
              ></div>
            </div>
            <span class="bar-value">{{ item.count }}</span>
          </div>
        </div>
      </div>

      <!-- Value Summary -->
      <div class="stats-section card">
        <h2>Value Summary</h2>
        <div class="value-stats">
          <div class="value-stat">
            <span class="value-stat-label">Average Purchase Price</span>
            <span class="value-stat-amount">{{ formatCurrency(stats.values.avgPurchasePrice) }}</span>
          </div>
          <div class="value-stat">
            <span class="value-stat-label">Average Current Value</span>
            <span class="value-stat-amount gold">{{ formatCurrency(stats.values.avgCurrentValue) }}</span>
          </div>
          <div class="value-stat" v-if="stats.values.totalCurrentValue && stats.values.totalPurchasePrice">
            <span class="value-stat-label">Return on Investment</span>
            <span
              class="value-stat-amount"
              :class="roi >= 0 ? 'positive' : 'negative'"
            >
              {{ roi >= 0 ? '+' : '' }}{{ roi.toFixed(1) }}%
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useCoinsStore } from '@/stores/coins'

const store = useCoinsStore()
const stats = computed(() => store.stats)

const maxCategoryCount = computed(() =>
  Math.max(...(stats.value?.byCategory.map((c) => c.count) || [1])),
)
const maxMaterialCount = computed(() =>
  Math.max(...(stats.value?.byMaterial.map((m) => m.count) || [1])),
)

const roi = computed(() => {
  if (!stats.value?.values.totalPurchasePrice) return 0
  return (
    ((stats.value.values.totalCurrentValue - stats.value.values.totalPurchasePrice) /
      stats.value.values.totalPurchasePrice) *
    100
  )
})

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)
}

onMounted(() => store.fetchStats())
</script>

<style scoped>
.stats-layout {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.stats-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 1.5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.stat-number {
  font-family: 'Cinzel', serif;
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
}

.stat-number.gold {
  color: var(--accent-gold);
}

.stat-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stats-section h2 {
  margin-bottom: 1.25rem;
  font-size: 1.1rem;
}

.bar-chart {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.bar-row {
  display: grid;
  grid-template-columns: 100px 1fr 40px;
  gap: 0.75rem;
  align-items: center;
}

.bar-label {
  font-size: 0.85rem;
}

.bar-track {
  height: 24px;
  background: var(--bg-primary);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: var(--radius-sm);
  transition: width 0.5s ease;
  min-width: 4px;
}

.fill-roman { background: linear-gradient(90deg, #7b2d8e, #9b59b6); }
.fill-greek { background: linear-gradient(90deg, #4a6e18, #6b8e23); }
.fill-byzantine { background: linear-gradient(90deg, #8b1a1a, #c0392b); }
.fill-modern { background: linear-gradient(90deg, #2c5f8a, #4682b4); }
.fill-other { background: linear-gradient(90deg, #555, #888); }
.fill-material { background: linear-gradient(90deg, var(--accent-bronze), var(--accent-gold)); }

.bar-value {
  font-size: 0.85rem;
  font-weight: 600;
  text-align: right;
  color: var(--text-secondary);
}

.value-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.value-stat {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.value-stat-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.value-stat-amount {
  font-size: 1.3rem;
  font-weight: 600;
}

.value-stat-amount.gold { color: var(--accent-gold); }
.value-stat-amount.positive { color: #2ecc71; }
.value-stat-amount.negative { color: #e74c3c; }
</style>
