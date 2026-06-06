<template>
  <section class="trend-card">
    <div class="trend-header">
      <h2>Value Trend</h2>
      <select :value="range" class="range-select" @change="$emit('update:range', ($event.target as HTMLSelectElement).value)">
        <option value="1m">1 month</option>
        <option value="3m">3 months</option>
        <option value="1y">1 year</option>
        <option value="all">All</option>
      </select>
    </div>
    <div v-if="snapshots.length" class="trend-list">
      <div v-for="snapshot in snapshots" :key="snapshot.snapshotDate" class="trend-row">
        <span>{{ formatDate(snapshot.snapshotDate) }}</span>
        <strong>${{ snapshot.totalValue.toFixed(2) }}</strong>
      </div>
    </div>
    <p v-else class="empty-trend">No snapshots yet. Capture one to start tracking trends.</p>
  </section>
</template>

<script setup lang="ts">
import type { CoinSetSnapshot } from '@/types'

defineProps<{
  snapshots: CoinSetSnapshot[]
  range: string
}>()

defineEmits<{
  'update:range': [value: string]
}>()

function formatDate(value: string) {
  return new Date(value).toLocaleDateString()
}
</script>

<style scoped>
.trend-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-md);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.trend-header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.trend-header h2 {
  margin: 0;
}

.range-select {
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  padding: 0.4rem 0.6rem;
}

.trend-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.trend-row {
  display: flex;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 0.4rem;
  color: var(--text-secondary);
}

.trend-row strong {
  color: var(--accent-gold);
}

.empty-trend {
  color: var(--text-secondary);
  margin: 0;
}
</style>
