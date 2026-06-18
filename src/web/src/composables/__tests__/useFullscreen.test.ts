import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useFullscreen } from '../useFullscreen'

let fullscreenElement: Element | null
const originalRequestFullscreen = HTMLElement.prototype.requestFullscreen
const originalExitFullscreen = document.exitFullscreen

const Harness = defineComponent({
  setup() {
    const target = ref<HTMLElement | null>(null)
    return { api: useFullscreen(target), target }
  },
  template: '<div ref="target" data-fullscreen-target>Target</div>',
})

describe('useFullscreen', () => {
  beforeEach(() => {
    fullscreenElement = null
    Object.defineProperty(document, 'fullscreenElement', {
      get: () => fullscreenElement,
      configurable: true,
    })
    HTMLElement.prototype.requestFullscreen = vi.fn(() => {
      fullscreenElement = document.querySelector('[data-fullscreen-target]')
      document.dispatchEvent(new Event('fullscreenchange'))
      return Promise.resolve()
    })
    Object.defineProperty(document, 'exitFullscreen', {
      value: vi.fn(() => {
        fullscreenElement = null
        document.dispatchEvent(new Event('fullscreenchange'))
        return Promise.resolve()
      }),
      configurable: true,
    })
  })

  afterEach(() => {
    HTMLElement.prototype.requestFullscreen = originalRequestFullscreen
    Object.defineProperty(document, 'exitFullscreen', {
      value: originalExitFullscreen,
      configurable: true,
    })
    document.body.innerHTML = ''
  })

  it('enters and exits fullscreen for the target element', async () => {
    const wrapper = mount(Harness, { attachTo: document.body })

    await wrapper.vm.api.enter()
    await flushPromises()

    expect(wrapper.vm.api.isFullscreen.value).toBe(true)

    await wrapper.vm.api.exit()
    await flushPromises()

    expect(wrapper.vm.api.isFullscreen.value).toBe(false)
    expect(document.exitFullscreen).toHaveBeenCalled()
  })

  it('does not exit fullscreen owned by another element', async () => {
    const wrapper = mount(Harness, { attachTo: document.body })
    fullscreenElement = document.createElement('section')

    await wrapper.vm.api.exit()

    expect(document.exitFullscreen).not.toHaveBeenCalled()
  })
})
