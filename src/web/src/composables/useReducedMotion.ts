import { onMounted, onUnmounted, readonly, ref } from 'vue'

export function useReducedMotion() {
  const prefersReducedMotion = ref(false)
  let query: MediaQueryList | null = null

  function update(matches: boolean) {
    prefersReducedMotion.value = matches
  }

  function handleChange(event: MediaQueryListEvent) {
    update(event.matches)
  }

  onMounted(() => {
    if (!window.matchMedia) return
    query = window.matchMedia('(prefers-reduced-motion: reduce)')
    update(query.matches)
    query.addEventListener('change', handleChange)
  })

  onUnmounted(() => {
    query?.removeEventListener('change', handleChange)
  })

  return {
    prefersReducedMotion: readonly(prefersReducedMotion),
  }
}
