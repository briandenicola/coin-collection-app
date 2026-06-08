<template>
  <section class="admin-section card">
    <h2>Coin Properties</h2>
    <p class="section-description">Configure category and era dropdown values for coin forms. Enter one value per line.</p>
    
    <form @submit.prevent="$emit('save')">
      <div class="form-group">
        <label class="form-label">Category Options</label>
        <textarea 
          v-model="localCategoryOptions" 
          class="form-textarea property-textarea" 
          rows="6"
          placeholder="Roman&#10;Greek&#10;Byzantine&#10;Modern&#10;Other"
        />
        <span class="form-hint">One category per line. Empty lines will be ignored.</span>
      </div>

      <div class="form-group">
        <label class="form-label">Era Options</label>
        <textarea 
          v-model="localEraOptions" 
          class="form-textarea property-textarea" 
          rows="4"
          placeholder="ancient&#10;medieval&#10;modern"
        />
        <span class="form-hint">One era per line. An "Unspecified" option will be added automatically.</span>
      </div>

      <p v-if="msg" class="msg" :class="{ error }">{{ msg }}</p>
      <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">
        {{ saving ? 'Saving...' : 'Save Coin Properties' }}
      </button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  categoryOptions: string
  eraOptions: string
  saving: boolean
  msg: string
  error: boolean
}>()

const emit = defineEmits<{
  save: []
  'update:categoryOptions': [value: string]
  'update:eraOptions': [value: string]
}>()

const localCategoryOptions = ref(props.categoryOptions)
const localEraOptions = ref(props.eraOptions)

watch(() => props.categoryOptions, (v) => { localCategoryOptions.value = v })
watch(() => props.eraOptions, (v) => { localEraOptions.value = v })

watch(localCategoryOptions, (v) => emit('update:categoryOptions', v))
watch(localEraOptions, (v) => emit('update:eraOptions', v))
</script>

<style scoped>
.section-description {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: -0.5rem 0 1.5rem;
}

.property-textarea {
  font-family: 'Courier New', Courier, monospace;
  font-size: 0.85rem;
}

.msg {
  font-size: 0.85rem;
  color: var(--accent-gold);
  margin: 0.5rem 0;
}

.msg.error {
  color: var(--cat-byzantine);
}
</style>
