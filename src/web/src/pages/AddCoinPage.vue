<template>
  <div class="container">
    <div class="page-header">
      <h1>➕ Add Coin</h1>
    </div>
    <CoinForm :form="form" submit-label="Add to Collection" :loading="saving" @submit="handleSubmit" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useCoinsStore } from '@/stores/coins'
import CoinForm from '@/components/CoinForm.vue'
import type { Coin } from '@/types'

const router = useRouter()
const store = useCoinsStore()
const saving = ref(false)

const form = reactive<Partial<Coin>>({
  name: '',
  category: 'Roman',
  material: 'Silver',
  denomination: '',
  ruler: '',
  era: '',
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
  referenceText: '',
  isWishlist: false,
})

async function handleSubmit() {
  saving.value = true
  try {
    const coin = await store.addCoin(form)
    router.push(`/coin/${coin.id}`)
  } catch {
    alert('Failed to add coin')
  } finally {
    saving.value = false
  }
}
</script>
