<template>
  <div class="container">
    <div class="form-wrapper">
      <div class="page-header">
        <h1>✏️ Edit Coin</h1>
      </div>
      <div v-if="loading" class="loading-overlay">
        <div class="spinner"></div>
      </div>
      <CoinForm v-else ref="coinFormRef" :form="form" :coin-id="form.id" submit-label="Save Changes" :loading="saving" @submit="handleSubmit" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CoinForm from '@/components/CoinForm.vue'
import { getCoin, updateCoin, uploadImage, deleteImage } from '@/api/client'
import type { Coin } from '@/types'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const coinFormRef = ref<InstanceType<typeof CoinForm> | null>(null)

const form = reactive<Partial<Coin>>({})

onMounted(async () => {
  const id = Number(route.params['id'])
  try {
    const res = await getCoin(id)
    Object.assign(form, res.data)
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

    const formComp = coinFormRef.value
    const coinId = form.id!

    // Delete removed images
    if (formComp?.removedObverseId) {
      await deleteImage(coinId, formComp.removedObverseId)
    }
    if (formComp?.removedReverseId) {
      await deleteImage(coinId, formComp.removedReverseId)
    }

    // Upload new/replacement images
    if (formComp?.obverseFile) {
      // If replacing, delete the old one first (if not already removed)
      const existingObverse = form.images?.find((i) => i.imageType === 'obverse')
      if (existingObverse && !formComp.removedObverseId) {
        await deleteImage(coinId, existingObverse.id)
      }
      await uploadImage(coinId, formComp.obverseFile, 'obverse', true)
    }
    if (formComp?.reverseFile) {
      const existingReverse = form.images?.find((i) => i.imageType === 'reverse')
      if (existingReverse && !formComp.removedReverseId) {
        await deleteImage(coinId, existingReverse.id)
      }
      await uploadImage(coinId, formComp.reverseFile, 'reverse', false)
    }

    router.push(`/coin/${coinId}`)
  } catch {
    alert('Failed to update coin')
  } finally {
    saving.value = false
  }
}
</script>
