import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useWakeLock } from '../useWakeLock'

class TestWakeLockSentinel extends EventTarget {
  released = false
  readonly type = 'screen'
  release = vi.fn(async () => {
    this.released = true
    this.dispatchEvent(new Event('release'))
  })
}

const Harness = defineComponent({
  setup() {
    return { api: useWakeLock() }
  },
  template: '<div />',
})

describe('useWakeLock', () => {
  let sentinel: TestWakeLockSentinel
  let request: ReturnType<typeof vi.fn>

  beforeEach(() => {
    sentinel = new TestWakeLockSentinel()
    request = vi.fn(async () => sentinel)
    Object.defineProperty(navigator, 'wakeLock', {
      value: { request },
      configurable: true,
    })
    Object.defineProperty(document, 'visibilityState', {
      value: 'visible',
      configurable: true,
    })
  })

  afterEach(() => {
    delete (navigator as Navigator & { wakeLock?: unknown }).wakeLock
  })

  it('requests and releases a screen wake lock', async () => {
    const wrapper = mount(Harness)

    await wrapper.vm.api.request()
    await flushPromises()

    expect(request).toHaveBeenCalledWith('screen')
    expect(wrapper.vm.api.active.value).toBe(true)

    await wrapper.vm.api.release()
    await flushPromises()

    expect(sentinel.release).toHaveBeenCalled()
    expect(wrapper.vm.api.active.value).toBe(false)
  })

  it('releases an active wake lock on unmount', async () => {
    const wrapper = mount(Harness)

    await wrapper.vm.api.request()
    wrapper.unmount()
    await flushPromises()

    expect(sentinel.release).toHaveBeenCalled()
  })

  it('treats unsupported wake lock as a non-blocking false result', async () => {
    delete (navigator as Navigator & { wakeLock?: unknown }).wakeLock
    const wrapper = mount(Harness)

    await expect(wrapper.vm.api.request()).resolves.toBe(false)

    expect(wrapper.vm.api.supported.value).toBe(false)
  })
})
