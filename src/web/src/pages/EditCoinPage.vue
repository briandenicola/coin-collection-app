<template>
  <div class="container">
    <div class="form-wrapper">
      <div class="page-header">
        <h1>✏️ Edit Coin</h1>
      </div>
      <div v-if="loading" class="loading-overlay">
        <div class="spinner"></div>
      </div>
      <CoinForm v-else :form="form" submit-label="Save Changes" :loading="saving" @submit="handleSubmit" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CoinForm from '@/components/CoinForm.vue'
import { getCoin, updateCoin } from '@/api/client'
import type { Coin } from '@/types'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const saving = ref(false)

const form = reactive<Partial<Coin>>({})

onMounted(async () => {
  const id = Number(route.params['id'])
  try {
    const res = await getCoin(id)
    Object.assign(form, res.data)
    // Format date for input[type=date]
    if (form.purchaseDate) {
      form.purchaseDate = form.purchaseDate.substring(0, 10)
    }
  } catch {
    alert('Failed to load coin')
    router.push('/')
  } finally {
    loading.value = false
  }
})

async function handleSubmit() {
  saving.value = true
  try {
    await updateCoin(form.id!, form)
    router.push(`/coin/${form.id}`)
  } catch {
    alert('Failed to update coin')
  } finally {
    saving.value = false
  }
}
</script>
