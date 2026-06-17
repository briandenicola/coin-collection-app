import { onUnmounted, readonly, ref } from 'vue'

const MAX_TILT_DEGREES = 15

export type DeviceOrientationPermissionState = 'unsupported' | 'prompt' | 'granted' | 'denied'

export interface CoinTilt {
  rotateX: number
  rotateY: number
  glintX: number
  glintY: number
}

interface DeviceOrientationEventConstructorWithPermission {
  requestPermission?: () => Promise<'granted' | 'denied'>
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

function getOrientationConstructor(): DeviceOrientationEventConstructorWithPermission | null {
  if (typeof window === 'undefined' || !('DeviceOrientationEvent' in window)) return null
  const ctor = window.DeviceOrientationEvent
  return ctor ? ctor as unknown as DeviceOrientationEventConstructorWithPermission : null
}

export function useDeviceOrientation() {
  const supported = ref(getOrientationConstructor() !== null)
  const permissionState = ref<DeviceOrientationPermissionState>(supported.value ? 'prompt' : 'unsupported')
  const active = ref(false)
  const tilt = ref<CoinTilt>({ rotateX: 0, rotateY: 0, glintX: 50, glintY: 50 })
  let frame: number | null = null
  let latest: DeviceOrientationEvent | null = null

  function applyOrientation(event: DeviceOrientationEvent) {
    const beta = event.beta ?? 0
    const gamma = event.gamma ?? 0
    const rotateX = clamp(beta, -MAX_TILT_DEGREES, MAX_TILT_DEGREES)
    const rotateY = clamp(gamma, -MAX_TILT_DEGREES, MAX_TILT_DEGREES)

    tilt.value = {
      rotateX,
      rotateY,
      glintX: clamp(50 + rotateY * 2, 15, 85),
      glintY: clamp(50 - rotateX * 2, 15, 85),
    }
  }

  function handleOrientation(event: DeviceOrientationEvent) {
    latest = event
    if (frame !== null) return
    frame = window.requestAnimationFrame(() => {
      frame = null
      if (latest) {
        applyOrientation(latest)
      }
    })
  }

  async function requestPermission(): Promise<DeviceOrientationPermissionState> {
    const ctor = getOrientationConstructor()
    if (!ctor) {
      supported.value = false
      permissionState.value = 'unsupported'
      return permissionState.value
    }

    if (ctor.requestPermission) {
      permissionState.value = await ctor.requestPermission()
    } else {
      permissionState.value = 'granted'
    }
    return permissionState.value
  }

  function start() {
    if (!supported.value || permissionState.value !== 'granted' || active.value) return
    window.addEventListener('deviceorientation', handleOrientation)
    active.value = true
  }

  function stop() {
    if (active.value) {
      window.removeEventListener('deviceorientation', handleOrientation)
      active.value = false
    }
    if (frame !== null) {
      window.cancelAnimationFrame(frame)
      frame = null
    }
  }

  onUnmounted(stop)

  return {
    supported: readonly(supported),
    permissionState: readonly(permissionState),
    active: readonly(active),
    tilt: readonly(tilt),
    requestPermission,
    start,
    stop,
  }
}
