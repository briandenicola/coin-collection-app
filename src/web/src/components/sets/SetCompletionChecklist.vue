<template>
  <section class="completion-card">
    <div class="completion-header">
      <div>
        <h2>Completion</h2>
        <p>{{ completion.completedTargets }} of {{ completion.totalTargets }} targets complete</p>
      </div>
      <span class="completion-percent">{{ completion.completionPercentage.toFixed(1) }}%</span>
    </div>
    <div class="completion-bar">
      <div class="completion-fill" :style="{ width: `${Math.min(completion.completionPercentage, 100)}%` }"></div>
    </div>
    <div v-if="completion.missingTargets.length" class="missing-targets">
      <span class="section-label">Missing targets</span>
      <div class="target-list">
        <span v-for="target in completion.missingTargets" :key="target.id || target.label" class="chip-sm">
          {{ target.label }}
        </span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { CoinSetCompletion } from '@/types'

defineProps<{
  completion: CoinSetCompletion
}>()
</script>

<style scoped>
.completion-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-md);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.completion-header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.completion-header h2,
.completion-header p {
  margin: 0;
}

.completion-header p {
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.completion-percent {
  color: var(--accent-gold);
  font-size: 1.5rem;
  font-weight: 600;
}

.completion-bar {
  height: 0.6rem;
  border-radius:var(--radius-full);
  background: var(--bg-input);
  overflow: hidden;
}

.completion-fill {
  height: 100%;
  background: var(--accent-gold);
  transition: width var(--transition-med);
}

.missing-targets {
  margin-top: 1rem;
}

.target-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.5rem;
}
</style>
