import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { useDeviceOrientation, type DeviceOrientationPermissionState } from '@/composables/useDeviceOrientation'

describe('useDeviceOrientation', () => {
  beforeEach(() => {
    vi.stubGlobal('requestAnimationFrame', vi.fn((callback: FrameRequestCallback) => {
      callback(0)
      return 1
    }))
    vi.stubGlobal('cancelAnimationFrame', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('reports unsupported when DeviceOrientationEvent is unavailable', async () => {
    Object.defineProperty(window, 'DeviceOrientationEvent', { value: undefined, configurable: true })
    const api = mountHarness()

    expect(api.supported).toBe(false)
    await expect(api.requestPermission()).resolves.toBe('unsupported')
  })

  it('requests iOS permission and starts listening when granted', async () => {
    const requestPermission = vi.fn<() => Promise<DeviceOrientationPermissionState>>().mockResolvedValue('granted')
    Object.defineProperty(window, 'DeviceOrientationEvent', {
      value: class {
        static requestPermission = requestPermission
      },
      configurable: true,
    })
    const addListener = vi.spyOn(window, 'addEventListener')
    const api = mountHarness()

    await expect(api.requestPermission()).resolves.toBe('granted')
    api.start()

    expect(requestPermission).toHaveBeenCalled()
    expect(api.active).toBe(true)
    expect(addListener).toHaveBeenCalledWith('deviceorientation', expect.any(Function))
  })

  it('keeps flip-safe state when permission is denied', async () => {
    Object.defineProperty(window, 'DeviceOrientationEvent', {
      value: class {
        static requestPermission = vi.fn().mockResolvedValue('denied')
      },
      configurable: true,
    })
    const api = mountHarness()

    await expect(api.requestPermission()).resolves.toBe('denied')
    api.start()

    expect(api.active).toBe(false)
  })

  it('clamps orientation updates and cleans up listeners', async () => {
    Object.defineProperty(window, 'DeviceOrientationEvent', { value: class {}, configurable: true })
    const removeListener = vi.spyOn(window, 'removeEventListener')
    const wrapper = mount(defineComponent({
      setup() {
        const api = useDeviceOrientation()
        return { api }
      },
      template: '<div />',
    }))
    const api = wrapper.vm.api
    await api.requestPermission()
    api.start()

    window.dispatchEvent(Object.assign(new Event('deviceorientation'), {
      beta: 30,
      gamma: -40,
    }) as DeviceOrientationEvent)
    await nextTick()

    expect(api.tilt.value.rotateX).toBe(15)
    expect(api.tilt.value.rotateY).toBe(-15)
    expect(api.tilt.value.glintX).toBe(20)
    expect(api.tilt.value.glintY).toBe(20)

    wrapper.unmount()
    expect(removeListener).toHaveBeenCalledWith('deviceorientation', expect.any(Function))
  })
})

function mountHarness() {
  const wrapper = mount(defineComponent({
    setup() {
      return useDeviceOrientation()
    },
    template: '<div />',
  }))
  return wrapper.vm
}
