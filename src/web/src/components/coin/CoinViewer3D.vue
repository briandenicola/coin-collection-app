<template>
  <div
    class="coin-viewer-3d"
    :class="[`size-${size}`, { 'reduced-motion': prefersReducedMotion, 'is-static': !canFlip }]"
    :style="stageStyle"
  >
    <div class="coin-stage" @click="emitOpenImage">
      <div class="coin-disc" :class="{ flipped }">
        <div class="coin-face coin-face-front">
          <img v-if="frontSrc" :src="frontSrc" :alt="obverseAlt" />
          <div v-else class="coin-placeholder">No image</div>
        </div>
        <div class="coin-face coin-face-back">
          <img v-if="reverseSrc" :src="reverseSrc" :alt="reverseAlt" />
          <div v-else class="coin-placeholder">No reverse</div>
        </div>
      </div>
      <div class="coin-rim"></div>
      <div class="coin-glint" :style="glintStyle"></div>
    </div>

    <button
      v-if="interactive"
      class="chip coin-flip-button"
      type="button"
      :disabled="!canFlip"
      :aria-label="flipped ? 'Show obverse' : 'Show reverse'"
      @click.stop="handleFlip"
    >
      <RefreshCw :size="14" />
      <span>{{ flipped ? 'Obverse' : 'Reverse' }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useDeviceOrientation } from '@/composables/useDeviceOrientation'
import { useReducedMotion } from '@/composables/useReducedMotion'

const props = withDefaults(defineProps<{
  obverseSrc?: string | null
  reverseSrc?: string | null
  obverseAlt?: string
  reverseAlt?: string
  size?: 'card' | 'hero' | 'tray'
  interactive?: boolean
  enableTilt?: boolean
}>(), {
  obverseSrc: null,
  reverseSrc: null,
  obverseAlt: 'Obverse',
  reverseAlt: 'Reverse',
  size: 'card',
  interactive: true,
  enableTilt: false,
})

const emit = defineEmits<{
  flip: [side: 'obverse' | 'reverse']
  'open-image': [side: 'obverse' | 'reverse']
}>()

const flipped = ref(false)
const { prefersReducedMotion } = useReducedMotion()
const orientation = useDeviceOrientation()

const frontSrc = computed(() => props.obverseSrc ?? props.reverseSrc ?? null)
const canFlip = computed(() => Boolean(props.obverseSrc && props.reverseSrc))
const currentSide = computed<'obverse' | 'reverse'>(() => {
  if (flipped.value) return 'reverse'
  return props.obverseSrc ? 'obverse' : 'reverse'
})

const stageStyle = computed(() => {
  if (!props.enableTilt || prefersReducedMotion.value || !orientation.active.value) return {}
  return {
    '--coin-tilt-x': `${orientation.tilt.value.rotateX}deg`,
    '--coin-tilt-y': `${orientation.tilt.value.rotateY}deg`,
  }
})

const glintStyle = computed(() => ({
  '--glint-x': `${orientation.tilt.value.glintX}%`,
  '--glint-y': `${orientation.tilt.value.glintY}%`,
}))

async function enableTiltIfAllowed() {
  if (!props.enableTilt || prefersReducedMotion.value || !orientation.supported.value) return
  const permission = orientation.permissionState.value === 'granted'
    ? 'granted'
    : await orientation.requestPermission()
  if (permission === 'granted') {
    orientation.start()
  }
}

async function handleFlip() {
  if (!canFlip.value) return
  await enableTiltIfAllowed()
  flipped.value = !flipped.value
  emit('flip', currentSide.value)
}

function emitOpenImage() {
  if (!frontSrc.value) return
  emit('open-image', currentSide.value)
}
</script>

<style scoped>
.coin-viewer-3d {
  --coin-tilt-x: 0deg;
  --coin-tilt-y: 0deg;
  --glint-x: 50%;
  --glint-y: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
}

.coin-stage {
  position: relative;
  width: min(100%, 18rem);
  aspect-ratio: 1;
  perspective: 900px;
  transform: rotateX(var(--coin-tilt-x)) rotateY(var(--coin-tilt-y));
  transform-style: preserve-3d;
  transition: transform var(--transition-fast);
  cursor: pointer;
}

.size-hero .coin-stage {
  width: min(100%, 26rem);
}

.size-tray .coin-stage {
  width: min(100%, 8rem);
}

.coin-disc {
  position: absolute;
  inset: 0;
  transform-style: preserve-3d;
  transition: transform var(--transition-med);
}

.coin-disc.flipped {
  transform: rotateY(180deg);
}

.reduced-motion .coin-disc {
  transition: none;
}

.coin-face {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 2px solid var(--border-accent);
  background: radial-gradient(circle, var(--bg-card-hover), var(--bg-input));
  clip-path: circle(50% at 50% 50%);
  box-shadow: inset 0 0 1.25rem rgba(255, 255, 255, 0.08), var(--shadow-card);
  backface-visibility: hidden;
}

.coin-face-back {
  transform: rotateY(180deg);
}

.coin-face img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  clip-path: circle(50% at 50% 50%);
  pointer-events: none;
}

.coin-placeholder {
  color: var(--text-muted);
  font-size: 0.85rem;
  text-align: center;
}

.coin-rim {
  position: absolute;
  inset: -0.2rem;
  border: 0.35rem solid rgba(201, 168, 76, 0.28);
  clip-path: circle(50% at 50% 50%);
  box-shadow: inset 0 0 0.5rem rgba(0, 0, 0, 0.45), 0 0.5rem 1.4rem rgba(0, 0, 0, 0.35);
  pointer-events: none;
}

.coin-glint {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at var(--glint-x) var(--glint-y), rgba(255, 255, 255, 0.32), rgba(255, 255, 255, 0.08) 18%, transparent 42%);
  clip-path: circle(50% at 50% 50%);
  mix-blend-mode: screen;
  opacity: 0.55;
  pointer-events: none;
  transition: background-position var(--transition-fast);
}

.is-static .coin-glint {
  opacity: 0.3;
}

.coin-flip-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  font-weight: 500;
  transition: all var(--transition-fast);
  touch-action: manipulation;
}

.coin-flip-button:hover:not(:disabled) {
  background: var(--accent-gold-dim);
}

.coin-flip-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (prefers-reduced-motion: reduce) {
  .coin-stage,
  .coin-disc,
  .coin-flip-button {
    transition: none;
  }
}
</style>
