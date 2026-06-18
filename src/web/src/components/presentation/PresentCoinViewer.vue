<template>
  <section
    class="present-viewer"
    :class="{ 'reduced-motion': reducedMotion, 'metadata-visible': metadataVisible }"
    tabindex="0"
    aria-label="Coin presentation viewer"
    @keydown.left.prevent="$emit('prev')"
    @keydown.right.prevent="$emit('next')"
    @keydown.esc.prevent="$emit('exit')"
  >
    <div class="present-topbar">
      <button class="btn btn-ghost btn-sm" type="button" @click="$emit('exit')">
        <X :size="16" />
        Exit
      </button>
      <span class="section-label">{{ index + 1 }} / {{ total }}</span>
    </div>

    <button class="present-nav present-nav-prev btn btn-ghost" type="button" aria-label="Previous coin" @click="$emit('prev')">
      <ChevronLeft :size="24" />
    </button>

    <div class="present-stage" @pointerdown="handlePointerDown" @click="toggleMetadata">
      <div class="present-spotlight"></div>
      <img v-if="imageSrc" class="present-image" :src="imageSrc" :alt="imageAlt" draggable="false" />
      <div v-else class="present-empty">
        <Coins :size="80" :stroke-width="1" />
        <p>No image available</p>
      </div>
      <Transition name="metadata-fade">
        <div v-if="metadataVisible && coin" class="metadata-overlay" @click.stop>
          <h1>{{ coin.name }}</h1>
          <div class="metadata-grid">
            <div v-for="item in metadataItems" :key="item.label" class="metadata-item">
              <span class="info-label">{{ item.label }}</span>
              <span>{{ item.value }}</span>
            </div>
          </div>
        </div>
      </Transition>
    </div>

    <button class="present-nav present-nav-next btn btn-ghost" type="button" aria-label="Next coin" @click="$emit('next')">
      <ChevronRight :size="24" />
    </button>

    <div class="present-controls" @click.stop>
      <button
        class="chip"
        type="button"
        :class="{ active: activeSide === 'obverse' }"
        :disabled="!hasObverse"
        @click="activeSide = 'obverse'"
      >
        Obverse
      </button>
      <button
        class="chip"
        type="button"
        :class="{ active: activeSide === 'reverse' }"
        :disabled="!hasReverse"
        @click="activeSide = 'reverse'"
      >
        Reverse
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Coins, X } from 'lucide-vue-next'
import type { Coin, CoinImage, ImageType } from '@/types'

type PresentCoin = Pick<Coin, 'id' | 'name' | 'ruler' | 'denomination' | 'era' | 'material' | 'grade' | 'images'>

const props = defineProps<{
  coin: PresentCoin | null
  index: number
  total: number
  reducedMotion?: boolean
}>()

const emit = defineEmits<{
  next: []
  prev: []
  exit: []
}>()

const metadataVisible = ref(false)
const activeSide = ref<ImageType>('obverse')
let pointerStartX = 0
let pointerStartY = 0
let dragged = false

const hasObverse = computed(() => Boolean(findImage('obverse')))
const hasReverse = computed(() => Boolean(findImage('reverse')))

const activeImage = computed(() => {
  if (!props.coin) return null
  return findImage(activeSide.value)
    ?? props.coin.images?.find((image) => image.isPrimary)
    ?? props.coin.images?.[0]
    ?? null
})

const imageSrc = computed(() => activeImage.value ? `/uploads/${activeImage.value.filePath}` : null)
const imageAlt = computed(() => props.coin ? `${props.coin.name} ${activeSide.value}` : 'Presented coin')

const metadataItems = computed(() => {
  if (!props.coin) return []
  return [
    { label: 'Ruler', value: props.coin.ruler },
    { label: 'Denomination', value: props.coin.denomination },
    { label: 'Era', value: props.coin.era },
    { label: 'Material', value: props.coin.material },
    { label: 'Grade', value: props.coin.grade },
  ].filter((item): item is { label: string; value: string } => Boolean(item.value))
})

watch(() => props.coin?.id, () => {
  metadataVisible.value = false
  activeSide.value = hasObverse.value || !hasReverse.value ? 'obverse' : 'reverse'
})

function findImage(side: ImageType): CoinImage | undefined {
  return props.coin?.images?.find((image) => image.imageType === side)
}

function toggleMetadata() {
  if (dragged) {
    dragged = false
    return
  }
  metadataVisible.value = !metadataVisible.value
}

function handlePointerDown(event: PointerEvent) {
  pointerStartX = event.clientX
  pointerStartY = event.clientY
  dragged = false
  window.addEventListener('pointerup', handlePointerUp, { once: true })
}

function handlePointerUp(event: PointerEvent) {
  const deltaX = event.clientX - pointerStartX
  const deltaY = event.clientY - pointerStartY
  if (Math.abs(deltaX) < 55 || Math.abs(deltaX) < Math.abs(deltaY)) return
  dragged = true
  if (deltaX < 0) {
    emit('next')
  } else {
    emit('prev')
  }
}
</script>

<style scoped>
.present-viewer {
  position: fixed;
  inset: 0;
  display: grid;
  grid-template: auto 1fr auto / auto 1fr auto;
  gap: 1rem;
  padding: calc(1rem + env(safe-area-inset-top)) calc(1rem + env(safe-area-inset-right)) calc(1rem + env(safe-area-inset-bottom)) calc(1rem + env(safe-area-inset-left));
  background:
    radial-gradient(circle at center, var(--bg-card-hover) 0%, var(--bg-secondary) 36%, var(--bg-primary) 72%),
    var(--bg-primary);
  color: var(--text-primary);
  z-index: 1000;
  overflow: hidden;
}

.present-topbar {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  opacity: 0.82;
}

.present-stage {
  position: relative;
  grid-column: 2;
  grid-row: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  min-height: 0;
  touch-action: none;
  cursor: pointer;
}

.present-spotlight {
  position: absolute;
  width: min(82vw, 82vh);
  aspect-ratio: 1;
  background: radial-gradient(circle, var(--accent-gold-glow) 0%, transparent 66%);
  filter: blur(18px);
  pointer-events: none;
}

.present-image {
  position: relative;
  max-width: min(92vw, 58rem);
  max-height: 76vh;
  object-fit: contain;
  filter: drop-shadow(0 1.5rem 2.5rem var(--overlay-dark));
  user-select: none;
  transition: opacity var(--transition-med), transform var(--transition-med);
}

.present-empty {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  color: var(--text-muted);
}

.metadata-overlay {
  position: absolute;
  left: 50%;
  bottom: 2rem;
  width: min(40rem, calc(100vw - 2rem));
  transform: translateX(-50%);
  padding: 1.5rem;
  background: var(--overlay-full);
  border: 1px solid var(--border-accent);
  box-shadow: var(--shadow-card);
}

.metadata-overlay h1 {
  margin-bottom: 1rem;
  text-align: center;
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
  gap: 0.75rem;
}

.metadata-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.metadata-item span:last-child {
  color: var(--text-primary);
  font-size: 0.9rem;
}

.present-nav {
  align-self: center;
  justify-self: center;
  padding: 0.6rem;
  opacity: 0.72;
}

.present-nav-prev {
  grid-column: 1;
  grid-row: 2;
}

.present-nav-next {
  grid-column: 3;
  grid-row: 2;
}

.present-controls {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  gap: 0.35rem;
  opacity: 0.86;
}

.present-controls .chip:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.metadata-fade-enter-active,
.metadata-fade-leave-active {
  transition: opacity var(--transition-med), transform var(--transition-med);
}

.metadata-fade-enter-from,
.metadata-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, 0.75rem);
}

.reduced-motion .present-image,
.reduced-motion .metadata-fade-enter-active,
.reduced-motion .metadata-fade-leave-active {
  transition: none;
}

@media (max-width: 768px) {
  .present-viewer {
    grid-template: auto 1fr auto / 1fr 1fr;
    gap: 0.75rem;
  }

  .present-stage {
    grid-column: 1 / -1;
  }

  .present-nav {
    grid-row: 3;
  }

  .present-nav-prev {
    grid-column: 1;
  }

  .present-nav-next {
    grid-column: 2;
  }

  .present-controls {
    grid-row: 4;
  }

  .metadata-overlay {
    bottom: 1rem;
    padding: 1rem;
  }
}
</style>
