<template>
  <div class="container">
    <div class="form-wrapper">
      <div class="page-header">
        <h1>Add Coin</h1>
      </div>
      <CoinForm ref="coinFormRef" :form="form" submit-label="Add to Collection" :loading="saving" @submit="handleSubmit" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useCoinsStore } from '@/stores/coins'
import { uploadImage, deleteImage } from '@/api/client'
import CoinForm from '@/components/CoinForm.vue'
import type { Coin } from '@/types'

const router = useRouter()
const store = useCoinsStore()
const saving = ref(false)
const coinFormRef = ref<InstanceType<typeof CoinForm> | null>(null)

const form = reactive<Partial<Coin>>({
  name: '',
  category: 'Roman',
  material: 'Silver',
  denomination: '',
  mint: '',
  weightGrams: null,
  diameterMm: null,
  grade: '',
  obverseInscription: '',
  reverseInscription: '',
  obverseDescription: '',
  reverseDescription: '',
  rarityRating: '',
  purchasePrice: null,
  currentValue: null,
  purchaseDate: null,
  purchaseLocation: '',
  notes: '',
  referenceUrl: '',
  referenceText: 'Store Link',
  isWishlist: false,
})

async function handleSubmit() {
  saving.value = true
  try {
    const coin = await store.addCoin(form)
    // Upload images if selected
    const formComp = coinFormRef.value
    if (formComp?.obverseFile) {
      await uploadImage(coin.id, formComp.obverseFile, 'obverse', true)
    }
    if (formComp?.reverseFile) {
      await uploadImage(coin.id, formComp.reverseFile, 'reverse', false)
    }
    router.push(`/coin/${coin.id}`)
  } catch {
    alert('Failed to add coin')
  } finally {
    saving.value = false
  }
}
</script>
