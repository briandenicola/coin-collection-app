<template>
  <div class="set-dashboard-card" @click="$emit('click')">
    <div class="set-card-header">
      <div class="set-icon-wrapper" :style="{ backgroundColor: set.color }">
        <component v-if="set.icon" :is="getIcon(set.icon)" :size="20" />
        <FolderOpen v-else :size="20" />
      </div>
      <div class="set-info">
        <h3 class="set-name">{{ set.name }}</h3>
        <span class="set-type-badge">{{ formatSetType(set.setType) }}</span>
      </div>
    </div>

    <div class="set-stats">
      <div class="stat-item">
        <span class="stat-label">Coins</span>
        <span class="stat-value">{{ set.coinCount }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Total Value</span>
        <span class="stat-value">${{ formatNumber(set.totalValue) }}</span>
      </div>
      <div v-if="set.completionPercentage != null" class="stat-item">
        <span class="stat-label">Completion</span>
        <span class="stat-value">{{ set.completionPercentage }}%</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { FolderOpen } from 'lucide-vue-next'
import type { CoinSetSummary } from '@/types'

defineProps<{
  set: CoinSetSummary
}>()

defineEmits<{
  (e: 'click'): void
}>()

function getIcon(iconName: string) {
  void iconName
  return FolderOpen
}

function formatSetType(type: string): string {
  const types: Record<string, string> = {
    open: 'Open',
    defined: 'Defined',
    smart: 'Smart',
    goal: 'Goal'
  }
  return types[type] || type
}

function formatNumber(value: number): string {
  return value.toFixed(2)
}
</script>

<style scoped>
.set-dashboard-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-sm);
  padding: 1rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.set-dashboard-card:hover {
  border-color: var(--accent-gold);
  box-shadow: var(--shadow-card);
}

.set-card-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.set-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius:var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--bg-primary);
}

.set-info {
  flex: 1;
  min-width: 0;
}

.set-name {
  font-size: 0.9rem;
  font-weight: 600;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.set-type-badge {
  display: inline-block;
  font-size: 0.75rem;
  padding: 0.125rem 0.5rem;
  background: var(--bg-input);
  border-radius:var(--radius-full);
  margin-top: 0.25rem;
}

.set-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(80px, 1fr));
  gap: 0.75rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.stat-value {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--accent-gold);
}
</style>
