<template>
  <section class="compare-card">
    <h2>Compare Sets</h2>
    <div class="compare-controls">
      <select v-model="selected" multiple class="compare-select">
        <option v-for="set in sets" :key="set.id" :value="set.id">{{ set.name }}</option>
      </select>
      <button class="btn btn-secondary" :disabled="selected.length < 2" @click="$emit('compare', selected)">Compare</button>
    </div>
    <div v-if="results.length" class="compare-results">
      <div v-for="result in results" :key="result.setId" class="compare-row">
        <span>{{ result.name }}</span>
        <strong>{{ result.valueChangePercent.toFixed(1) }}%</strong>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { CoinSetComparison, CoinSetSummary } from '@/types'

defineProps<{
  sets: CoinSetSummary[]
  results: CoinSetComparison[]
}>()

defineEmits<{
  compare: [setIds: number[]]
}>()

const selected = ref<number[]>([])
</script>

<style scoped>
.compare-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-md);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.compare-card h2 {
  margin-top: 0;
}

.compare-controls {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
}

.compare-select {
  flex: 1;
  min-height: 5rem;
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  padding: 0.5rem;
}

.compare-results {
  margin-top: 1rem;
}

.compare-row {
  display: flex;
  justify-content: space-between;
  padding: 0.4rem 0;
  border-bottom: 1px solid var(--border-subtle);
}

.compare-row strong {
  color: var(--accent-gold);
}
</style>
