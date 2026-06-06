<template>
  <div class="set-creation-wizard">
    <form @submit.prevent="submit">
      <div class="form-group">
        <label for="setType">Set type</label>
        <select id="setType" v-model="form.setType" class="form-input">
          <option value="open">Open</option>
          <option value="defined">Defined</option>
          <option value="goal">Goal</option>
          <option value="smart">Smart</option>
        </select>
      </div>
      <div v-if="form.setType === 'defined' || form.setType === 'goal'" class="form-group">
        <label for="templateId">Template</label>
        <select id="templateId" v-model="form.templateId" class="form-input">
          <option value="">No template</option>
          <option v-for="template in templates" :key="template.id" :value="template.id">
            {{ template.name }}
          </option>
        </select>
      </div>
      <div v-if="form.setType === 'defined' || form.setType === 'goal'" class="form-group">
        <label for="csvTargets">Custom CSV targets</label>
        <textarea
          id="csvTargets"
          v-model="csvTargets"
          rows="4"
          class="form-input"
          placeholder="Label,Year,MintMark,Denomination,Country,Material"
        />
      </div>
      <div v-if="form.setType === 'goal'" class="form-group">
        <label for="targetCompletionDate">Target completion date</label>
        <input id="targetCompletionDate" v-model="form.targetCompletionDate" type="date" class="form-input" />
      </div>
      <SetSmartRuleBuilder
        v-if="form.setType === 'smart'"
        @update="form.smartCriteria = $event"
      />
      <div class="form-group">
        <label for="setName">Name</label>
        <input
          id="setName"
          v-model="form.name"
          type="text"
          required
          maxlength="80"
          class="form-input"
        />
      </div>
      <div class="form-group">
        <label for="setDescription">Description</label>
        <textarea
          id="setDescription"
          v-model="form.description"
          rows="3"
          maxlength="2000"
          class="form-input"
        />
      </div>
      <div class="form-group">
        <label for="setColor">Color</label>
        <input id="setColor" v-model="form.color" type="color" class="color-input" />
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Cancel</button>
        <button type="submit" class="btn btn-primary" :disabled="!form.name.trim()">
          {{ submitLabel }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { getSetTemplates } from '@/api/client'
import type { CoinSetTemplate, CreateCoinSetRequest, SmartCriteriaGroup } from '@/types'
import SetSmartRuleBuilder from '@/components/sets/SetSmartRuleBuilder.vue'

const props = withDefaults(defineProps<{
  initialValue?: Partial<CreateCoinSetRequest>
  submitLabel?: string
}>(), {
  submitLabel: 'Create',
})

const emit = defineEmits<{
  submit: [value: CreateCoinSetRequest, csv?: string]
  cancel: []
}>()

const form = reactive({
  name: props.initialValue?.name ?? '',
  description: props.initialValue?.description ?? '',
  color: props.initialValue?.color ?? '#6b7280',
  setType: props.initialValue?.setType ?? 'open',
  templateId: props.initialValue?.templateId ?? '',
  targetCompletionDate: props.initialValue?.targetCompletionDate ?? '',
  smartCriteria: props.initialValue?.smartCriteria as SmartCriteriaGroup | undefined,
})
const templates = ref<CoinSetTemplate[]>([])
const csvTargets = ref('')

onMounted(async () => {
  try {
    const res = await getSetTemplates()
    templates.value = res.data.templates
  } catch {
    templates.value = []
  }
})

watch(() => props.initialValue, (value) => {
  form.name = value?.name ?? ''
  form.description = value?.description ?? ''
  form.color = value?.color ?? '#6b7280'
  form.setType = value?.setType ?? 'open'
  form.templateId = value?.templateId ?? ''
  form.targetCompletionDate = value?.targetCompletionDate ?? ''
  form.smartCriteria = value?.smartCriteria as SmartCriteriaGroup | undefined
})

function submit() {
  const name = form.name.trim()
  if (!name) return
  emit('submit', {
    name,
    description: form.description.trim(),
    color: form.color,
    setType: form.setType,
    templateId: form.templateId || undefined,
    targetCompletionDate: form.targetCompletionDate || undefined,
    smartCriteria: form.smartCriteria ?? undefined,
  }, csvTargets.value.trim() || undefined)
}
</script>

<style scoped>
.set-creation-wizard {
  width: 100%;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-input,
.color-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  font-family: inherit;
}

.color-input {
  min-height: 2.5rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}
</style>
