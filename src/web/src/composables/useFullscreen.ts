import type { Ref } from 'vue'
import { onMounted, onUnmounted, readonly, ref } from 'vue'

export function useFullscreen(target: Ref<HTMLElement | null>) {
  const isFullscreen = ref(false)
  const error = ref<unknown>(null)

  function syncState() {
    isFullscreen.value = document.fullscreenElement === target.value
  }

  async function enter(): Promise<boolean> {
    error.value = null
    const element = target.value
    if (!element || typeof element.requestFullscreen !== 'function') return false

    try {
      await element.requestFullscreen()
      syncState()
      return true
    } catch (err) {
      error.value = err
      return false
    }
  }

  async function exit(): Promise<boolean> {
    error.value = null
    if (document.fullscreenElement !== target.value || typeof document.exitFullscreen !== 'function') {
      syncState()
      return false
    }

    try {
      await document.exitFullscreen()
      syncState()
      return true
    } catch (err) {
      error.value = err
      return false
    }
  }

  onMounted(() => {
    document.addEventListener('fullscreenchange', syncState)
    syncState()
  })

  onUnmounted(() => {
    document.removeEventListener('fullscreenchange', syncState)
  })

  return {
    isFullscreen: readonly(isFullscreen),
    error: readonly(error),
    enter,
    exit,
  }
}
