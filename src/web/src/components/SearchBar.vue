<template>
  <div class="search-bar">
    <Search class="search-icon" :size="16" />
    <input
      type="text"
      class="search-input"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      placeholder="Search coins by name, ruler, inscription..."
    />
    <button v-if="modelValue" class="search-clear" @click="$emit('update:modelValue', '')"><X :size="14" /></button>
  </div>
</template>

<script setup lang="ts">
import { Search, X } from 'lucide-vue-next'
defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<style scoped>
.search-bar {
  position: relative;
  display: flex;
  align-items: center;
  max-width: 400px;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 0.8rem;
  color: var(--text-muted);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.6rem 2.5rem 0.6rem 2.5rem;
  background: var(--bg-input);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  color: var(--text-primary);
  font-size: 0.85rem;
  transition: border-color var(--transition-fast);
}

.search-input:focus {
  outline: none;
  border-color: var(--accent-gold);
  box-shadow: 0 0 0 2px var(--accent-gold-glow);
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-clear {
  position: absolute;
  right: 0.5rem;
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.3rem;
  font-size: 0.8rem;
}

.search-clear:hover {
  color: var(--text-primary);
}

@media (max-width: 768px) {
  .search-bar {
    max-width: 100%;
  }
}
</style>
