import { ref, onMounted, onUnmounted, type Ref } from 'vue'

const THRESHOLD = 80
const MAX_PULL = 130
const RESISTANCE = 0.45

export function usePullToRefresh(
  containerRef: Ref<HTMLElement | null>,
  onRefresh: () => Promise<void>,
) {
  const pullDistance = ref(0)
  const refreshing = ref(false)
  let startY = 0
  let pulling = false

  function isAtTop(): boolean {
    return window.scrollY <= 0
  }

  function onTouchStart(e: TouchEvent) {
    if (refreshing.value || !isAtTop()) return
    startY = e.touches[0].clientY
    pulling = true
  }

  function onTouchMove(e: TouchEvent) {
    if (!pulling || refreshing.value) return
    if (!isAtTop()) {
      pullDistance.value = 0
      return
    }

    const dy = e.touches[0].clientY - startY
    if (dy < 0) {
      pullDistance.value = 0
      return
    }

    // Prevent native scroll while pulling
    e.preventDefault()
    pullDistance.value = Math.min(dy * RESISTANCE, MAX_PULL)
  }

  async function onTouchEnd() {
    if (!pulling) return
    pulling = false

    if (pullDistance.value >= THRESHOLD) {
      refreshing.value = true
      pullDistance.value = THRESHOLD * 0.6
      try {
        await onRefresh()
      } finally {
        refreshing.value = false
        pullDistance.value = 0
      }
    } else {
      pullDistance.value = 0
    }
  }

  onMounted(() => {
    const el = containerRef.value
    if (!el) return
    el.addEventListener('touchstart', onTouchStart, { passive: true })
    el.addEventListener('touchmove', onTouchMove, { passive: false })
    el.addEventListener('touchend', onTouchEnd, { passive: true })
  })

  onUnmounted(() => {
    const el = containerRef.value
    if (!el) return
    el.removeEventListener('touchstart', onTouchStart)
    el.removeEventListener('touchmove', onTouchMove)
    el.removeEventListener('touchend', onTouchEnd)
  })

  return { pullDistance, refreshing }
}
