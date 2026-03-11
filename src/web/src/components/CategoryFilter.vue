<template>
  <div class="category-filter">
    <button
      class="filter-btn"
      :class="{ active: !modelValue }"
      @click="$emit('update:modelValue', '')"
    >
      All
    </button>
    <button
      v-for="cat in CATEGORIES"
      :key="cat"
      class="filter-btn"
      :class="[`filter-${cat.toLowerCase()}`, { active: modelValue === cat }]"
      @click="$emit('update:modelValue', cat)"
    >
      {{ cat }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { CATEGORIES } from '@/types'

defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<style scoped>
.category-filter {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.filter-btn {
  padding: 0.4rem 0.9rem;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-btn:hover {
  border-color: var(--border-accent);
  color: var(--text-primary);
}

.filter-btn.active {
  background: var(--accent-gold-dim);
  border-color: var(--accent-gold);
  color: var(--accent-gold);
}

.filter-roman.active { background: rgba(155, 89, 182, 0.2); border-color: #9b59b6; color: #b57edc; }
.filter-greek.active { background: rgba(107, 142, 35, 0.2); border-color: #6b8e23; color: #9ab85a; }
.filter-byzantine.active { background: rgba(192, 57, 43, 0.2); border-color: #c0392b; color: #e67e73; }
.filter-modern.active { background: rgba(70, 130, 180, 0.2); border-color: #4682b4; color: #7ab3d4; }
.filter-other.active { background: rgba(136, 136, 136, 0.2); border-color: #888; color: #aaa; }
</style>
