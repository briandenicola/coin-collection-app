<template>
  <main ref="presentRoot" class="present-mode-page">
    <PresentCoinViewer
      v-if="currentCoin"
      :coin="currentCoin"
      :index="currentIndex"
      :total="coins.length"
      :reduced-motion="prefersReducedMotion"
      @next="goNext"
      @prev="goPrev"
      @exit="exitPresentMode"
    />
    <section v-else class="present-empty-page">
      <h1>No coins to present</h1>
      <p>Return to the collection and load a gallery before starting Present mode.</p>
      <button class="btn btn-primary" type="button" @click="exitPresentMode">Back to Collection</button>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PresentCoinViewer from '@/components/presentation/PresentCoinViewer.vue'
import { useFullscreen } from '@/composables/useFullscreen'
import { useReducedMotion } from '@/composables/useReducedMotion'
import { useWakeLock } from '@/composables/useWakeLock'
import { useCoinsStore } from '@/stores/coins'

const store = useCoinsStore()
const route = useRoute()
const router = useRouter()
const presentRoot = ref<HTMLElement | null>(null)
const { prefersReducedMotion } = useReducedMotion()
const fullscreen = useFullscreen(presentRoot)
const wakeLock = useWakeLock()

const requestedStart = Number(route.query.start ?? store.galleryIndex ?? 0)
const currentIndex = ref(Number.isFinite(requestedStart) && requestedStart >= 0 ? requestedStart : 0)
const coins = computed(() => store.coins)
const currentCoin = computed(() => coins.value[currentIndex.value] ?? null)

watch(currentIndex, (index) => {
  store.galleryIndex = index
})

watch(() => coins.value.length, (length) => {
  if (length === 0) return
  currentIndex.value = Math.min(currentIndex.value, length - 1)
}, { immediate: true })

onMounted(async () => {
  if (store.coins.length === 0) {
    await store.fetchCoins({ page: 1 })
  }
  await nextTick()
  await fullscreen.enter()
  await wakeLock.request()
})

onUnmounted(() => {
  void wakeLock.release()
  void fullscreen.exit()
})

function goNext() {
  if (coins.value.length <= 1) return
  currentIndex.value = (currentIndex.value + 1) % coins.value.length
}

function goPrev() {
  if (coins.value.length <= 1) return
  currentIndex.value = (currentIndex.value - 1 + coins.value.length) % coins.value.length
}

async function exitPresentMode() {
  await wakeLock.release()
  await fullscreen.exit()
  await router.push({ name: 'collection' })
}
</script>

<style scoped>
.present-mode-page,
.present-empty-page {
  min-height: 100vh;
  background: var(--bg-primary);
}

.present-empty-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 1.5rem;
  text-align: center;
  color: var(--text-secondary);
}
</style>
