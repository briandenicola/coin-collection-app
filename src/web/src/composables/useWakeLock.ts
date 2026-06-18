import { onMounted, onUnmounted, readonly, ref } from 'vue'

interface WakeLockSentinelLike extends EventTarget {
  readonly released: boolean
  readonly type: 'screen'
  release: () => Promise<void>
}

interface WakeLockLike {
  request: (type: 'screen') => Promise<WakeLockSentinelLike>
}

type NavigatorWithWakeLock = Navigator & {
  wakeLock?: WakeLockLike
}

function getWakeLock(): WakeLockLike | null {
  if (typeof navigator === 'undefined') return null
  return (navigator as NavigatorWithWakeLock).wakeLock ?? null
}

export function useWakeLock() {
  const active = ref(false)
  const supported = ref(false)
  const error = ref<unknown>(null)
  let sentinel: WakeLockSentinelLike | null = null
  let requested = false

  function handleRelease() {
    active.value = false
    sentinel?.removeEventListener('release', handleRelease)
    sentinel = null
  }

  async function request(): Promise<boolean> {
    error.value = null
    requested = true
    const wakeLock = getWakeLock()
    supported.value = Boolean(wakeLock)
    if (!wakeLock) return false

    try {
      sentinel?.removeEventListener('release', handleRelease)
      sentinel = await wakeLock.request('screen')
      sentinel.addEventListener('release', handleRelease)
      active.value = !sentinel.released
      return active.value
    } catch (err) {
      error.value = err
      active.value = false
      sentinel = null
      return false
    }
  }

  async function release(): Promise<boolean> {
    error.value = null
    requested = false
    if (!sentinel) {
      active.value = false
      return false
    }

    try {
      const current = sentinel
      current.removeEventListener('release', handleRelease)
      sentinel = null
      await current.release()
      active.value = false
      return true
    } catch (err) {
      error.value = err
      return false
    }
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible' && requested && !active.value) {
      void request()
    }
  }

  onMounted(() => {
    supported.value = Boolean(getWakeLock())
    document.addEventListener('visibilitychange', handleVisibilityChange)
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    void release()
  })

  return {
    active: readonly(active),
    supported: readonly(supported),
    error: readonly(error),
    request,
    release,
  }
}
