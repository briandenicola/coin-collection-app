<template>
  <div class="swipe-gallery">
    <!-- Obverse / Reverse toggle -->
    <div class="side-toggle">
      <button
        class="toggle-btn"
        :class="{ active: activeSide === 'obverse' }"
        @click="activeSide = 'obverse'"
      >
        Obverse
      </button>
      <button
        class="toggle-btn"
        :class="{ active: activeSide === 'reverse' }"
        @click="activeSide = 'reverse'"
      >
        Reverse
      </button>
    </div>

    <!-- Card counter -->
    <div v-if="coins.length" class="card-counter">
      {{ currentIndex + 1 }} / {{ coins.length }}
    </div>

    <!-- Card stack -->
    <div v-if="coins.length" class="card-stack" ref="stackRef">
      <!-- Next card (underneath) -->
      <div v-if="nextCoin" class="swipe-card next-card">
        <div class="swipe-card-image">
          <img v-if="getImage(nextCoin)" :src="getImage(nextCoin)!" :alt="nextCoin.name" />
          <div v-else class="swipe-card-placeholder"><Coins :size="64" :stroke-width="1" /></div>
        </div>
        <div class="swipe-card-name">{{ nextCoin.name }}</div>
      </div>

      <!-- Active card (on top, draggable) -->
      <div
        v-if="currentCoin"
        class="swipe-card active-card"
        :class="{ animating: isAnimating }"
        :style="cardStyle"
        @pointerdown="onPointerDown"
        @click="onCardTap"
      >
        <div class="swipe-card-image">
          <img v-if="getImage(currentCoin)" :src="getImage(currentCoin)!" :alt="currentCoin.name" />
          <div v-else class="swipe-card-placeholder"><Coins :size="64" :stroke-width="1" /></div>
        </div>
        <div class="swipe-card-name">{{ currentCoin.name }}</div>
        <div class="swipe-hint left-hint" :style="{ opacity: leftHintOpacity }"><ChevronLeft :size="32" /></div>
        <div class="swipe-hint right-hint" :style="{ opacity: rightHintOpacity }"><ChevronRight :size="32" /></div>
      </div>
    </div>

    <div v-else class="swipe-empty">
      <p>No coins to display</p>
    </div>

    <!-- Arrow navigation -->
    <div v-if="coins.length > 1" class="swipe-nav">
      <button class="nav-btn" @click="goPrev">
        <ChevronLeft :size="16" /> Prev
      </button>
      <button class="nav-btn" @click="goNext">
        Next <ChevronRight :size="16" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Coin, ImageType } from '@/types'
import { Coins, ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps<{ coins: Coin[] }>()
const router = useRouter()

const activeSide = ref<'obverse' | 'reverse'>('obverse')
const currentIndex = ref(0)
const isAnimating = ref(false)
const stackRef = ref<HTMLElement | null>(null)

// Drag state
const dragX = ref(0)
const dragY = ref(0)
const isDragging = ref(false)
let startX = 0
let startY = 0
let hasDragged = false
let pointerId: number | null = null

const SWIPE_THRESHOLD = 100
const FLY_DISTANCE = 600

const currentCoin = computed(() => props.coins[currentIndex.value] ?? null)
const nextCoin = computed(() => {
  if (!props.coins.length) return null
  const nextIdx = (currentIndex.value + 1) % props.coins.length
  return props.coins[nextIdx] ?? null
})

const leftHintOpacity = computed(() => Math.min(1, Math.max(0, -dragX.value / SWIPE_THRESHOLD)))
const rightHintOpacity = computed(() => Math.min(1, Math.max(0, dragX.value / SWIPE_THRESHOLD)))

const cardStyle = computed(() => {
  if (!isDragging.value && !isAnimating.value) return {}
  const rotate = dragX.value * 0.05
  return {
    transform: `translate(${dragX.value}px, ${dragY.value}px) rotate(${rotate}deg)`,
    transition: isAnimating.value ? 'transform 0.3s ease' : 'none',
  }
})

function getImage(coin: Coin): string | null {
  const targetType: ImageType = activeSide.value
  const byType = coin.images?.find((img) => img.imageType === targetType)
  if (byType) return `/uploads/${byType.filePath}`
  const primary = coin.images?.find((img) => img.isPrimary)
  const first = coin.images?.[0]
  const img = primary || first
  return img ? `/uploads/${img.filePath}` : null
}

function onPointerDown(e: PointerEvent) {
  if (isAnimating.value) return
  const target = e.target as HTMLElement
  target.setPointerCapture(e.pointerId)
  pointerId = e.pointerId
  startX = e.clientX
  startY = e.clientY
  hasDragged = false
  isDragging.value = true
  dragX.value = 0
  dragY.value = 0

  target.addEventListener('pointermove', onPointerMove)
  target.addEventListener('pointerup', onPointerUp)
  target.addEventListener('pointercancel', onPointerUp)
}

function onPointerMove(e: PointerEvent) {
  if (!isDragging.value) return
  dragX.value = e.clientX - startX
  dragY.value = (e.clientY - startY) * 0.3
  if (Math.abs(dragX.value) > 5) hasDragged = true
}

function onPointerUp(e: PointerEvent) {
  if (!isDragging.value) return
  const target = e.target as HTMLElement
  target.removeEventListener('pointermove', onPointerMove)
  target.removeEventListener('pointerup', onPointerUp)
  target.removeEventListener('pointercancel', onPointerUp)
  if (pointerId !== null) {
    target.releasePointerCapture(pointerId)
    pointerId = null
  }
  isDragging.value = false

  if (Math.abs(dragX.value) > SWIPE_THRESHOLD) {
    flyAway(dragX.value > 0 ? 1 : -1)
  } else {
    // Spring back
    isAnimating.value = true
    dragX.value = 0
    dragY.value = 0
    setTimeout(() => {
      isAnimating.value = false
    }, 300)
  }
}

function flyAway(direction: 1 | -1) {
  isAnimating.value = true
  dragX.value = direction * FLY_DISTANCE
  dragY.value = direction * -50

  setTimeout(() => {
    isAnimating.value = false
    dragX.value = 0
    dragY.value = 0

    const len = props.coins.length
    if (len === 0) return
    if (direction > 0) {
      currentIndex.value = (currentIndex.value + 1) % len
    } else {
      currentIndex.value = (currentIndex.value - 1 + len) % len
    }
  }, 300)
}

function onCardTap() {
  if (hasDragged || !currentCoin.value) return
  router.push(`/coin/${currentCoin.value.id}`)
}

function goNext() {
  if (props.coins.length > 1 && !isAnimating.value) {
    flyAway(1)
  }
}

function goPrev() {
  if (props.coins.length > 1 && !isAnimating.value) {
    flyAway(-1)
  }
}

onUnmounted(() => {
  // Cleanup handled by pointer event listeners on targets
})
</script>

<style scoped>
.swipe-gallery {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 1rem 0;
  user-select: none;
  -webkit-user-select: none;
}

.side-toggle {
  display: flex;
  gap: 0;
  border-radius: var(--radius-full);
  overflow: hidden;
  border: 1px solid var(--border-accent);
}

.toggle-btn {
  padding: 0.5rem 1.25rem;
  border: none;
  background: var(--bg-card);
  color: var(--text-secondary);
  font-family: 'Cinzel', serif;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toggle-btn.active {
  background: var(--accent-gold);
  color: var(--bg-primary);
  font-weight: 600;
}

.card-counter {
  font-size: 0.8rem;
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.card-stack {
  position: relative;
  width: 315px;
  height: 399px;
  touch-action: none;
}

@media (min-width: 480px) {
  .card-stack {
    width: 340px;
    height: 420px;
  }
}

.swipe-card {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  box-shadow: var(--shadow-card);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.next-card {
  transform: scale(0.95);
  opacity: 0.6;
}

.active-card {
  z-index: 2;
  cursor: grab;
}

.active-card:active {
  cursor: grabbing;
}

.swipe-card-image {
  flex: 1;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
}

.swipe-card-image img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
}

.swipe-card-placeholder {
  font-size: 5rem;
  opacity: 0.2;
}

.swipe-card-name {
  padding: 0.75rem 1rem;
  text-align: center;
  font-family: 'Cinzel', serif;
  font-size: 1rem;
  color: var(--text-heading);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-top: 1px solid var(--border-subtle);
}

.swipe-hint {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  font-size: 2rem;
  color: var(--accent-gold);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.1s;
}

.left-hint {
  left: 1rem;
}

.right-hint {
  right: 1rem;
}

.swipe-nav {
  display: flex;
  gap: 1rem;
}

.nav-btn {
  padding: 0.5rem 1.25rem;
  border: 1px solid var(--border-accent);
  border-radius: var(--radius-full);
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.nav-btn:hover:not(:disabled) {
  background: var(--accent-gold-dim);
  color: var(--text-primary);
}

.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.swipe-empty {
  padding: 3rem;
  color: var(--text-muted);
  text-align: center;
}
</style>
