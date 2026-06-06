<template>
  <div class="smart-builder">
    <span class="section-label">Smart rule</span>
    <div class="rule-row">
      <select v-model="rule.field" class="rule-input">
        <option value="material">Material</option>
        <option value="category">Category</option>
        <option value="mint">Mint</option>
        <option value="grade">Grade</option>
        <option value="currentValue">Current value</option>
        <option value="purchaseDate">Purchase date</option>
      </select>
      <select v-model="rule.op" class="rule-input">
        <option value="eq">Equals</option>
        <option value="contains">Contains</option>
        <option value="gte">At least</option>
        <option value="lte">At most</option>
      </select>
      <input v-model="rule.value" class="rule-input" placeholder="Value" @input="emitCriteria" />
      <button type="button" class="btn btn-secondary btn-sm" @click="preview">Preview</button>
    </div>
    <p v-if="previewResult" class="preview-result">
      {{ previewResult.coinCount }} matching coins, ${{ previewResult.totalValue.toFixed(2) }} total value
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { previewSmartSet } from '@/api/client'
import type { SmartCriteriaGroup, SmartCriteriaRule, SmartSetPreview } from '@/types'

const emit = defineEmits<{
  update: [criteria: SmartCriteriaGroup]
}>()

const rule = reactive<SmartCriteriaRule>({
  field: 'material',
  op: 'eq',
  value: 'Silver',
})
const previewResult = ref<SmartSetPreview | null>(null)

watch(rule, emitCriteria, { deep: true, immediate: true })

function buildCriteria(): SmartCriteriaGroup {
  return {
    operator: 'and',
    rules: [{ ...rule }],
  }
}

function emitCriteria() {
  emit('update', buildCriteria())
}

async function preview() {
  const res = await previewSmartSet(buildCriteria())
  previewResult.value = res.data
}
</script>

<style scoped>
.smart-builder {
  margin-top: 1rem;
}

.rule-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.rule-input {
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  padding: 0.45rem 0.6rem;
}

.preview-result {
  color: var(--text-secondary);
  font-size: 0.85rem;
}
</style>
