<template>
  <div class="container">
    <div class="page-header">
      <h1>Image Processor</h1>
      <p class="page-desc">Remove backgrounds and crop coin images</p>
    </div>

    <div v-if="coinId" class="target-coin-banner card">
      <span>Saving to coin #{{ coinId }}</span>
      <router-link :to="`/coin/${coinId}`" class="btn btn-secondary btn-sm">View Coin</router-link>
    </div>

    <ImageProcessor :coin-id="coinId" @saved="handleSaved" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ImageProcessor from '@/components/ImageProcessor.vue'

const route = useRoute()
const router = useRouter()

const coinId = computed(() => {
  const id = route.query.coinId
  return id ? Number(id) : undefined
})

function handleSaved() {
  if (coinId.value) {
    router.push(`/coin/${coinId.value}`)
  }
}
</script>

<style scoped>
.page-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.page-desc {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.target-coin-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
  padding: 0.75rem 1rem;
  font-size: 0.9rem;
  color: var(--text-secondary);
}
</style>
