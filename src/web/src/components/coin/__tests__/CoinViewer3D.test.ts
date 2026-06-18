import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CoinViewer3D from '../CoinViewer3D.vue'

function stubMotion(matches = false) {
  Object.defineProperty(window, 'matchMedia', {
    value: vi.fn(() => ({
      matches,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })),
    configurable: true,
  })
  Object.defineProperty(window, 'DeviceOrientationEvent', { value: undefined, configurable: true })
}

describe('CoinViewer3D', () => {
  beforeEach(() => {
    stubMotion()
  })

  it('flips between obverse and reverse when both faces exist', async () => {
    const wrapper = mount(CoinViewer3D, {
      props: {
        obverseSrc: '/uploads/obverse.webp',
        reverseSrc: '/uploads/reverse.webp',
      },
      global: {
        stubs: { RefreshCw: true },
      },
    })

    await wrapper.find('.coin-flip-button').trigger('click')
    await flushPromises()

    expect(wrapper.find('.coin-disc').classes()).toContain('flipped')
    expect(wrapper.emitted('flip')).toEqual([['reverse']])
  })

  it('disables flip for a single-image coin', () => {
    const wrapper = mount(CoinViewer3D, {
      props: { obverseSrc: '/uploads/obverse.webp' },
      global: {
        stubs: { RefreshCw: true },
      },
    })

    expect(wrapper.find('.coin-flip-button').attributes('disabled')).toBeDefined()
    expect(wrapper.find('.coin-disc').classes()).not.toContain('flipped')
  })

  it('renders a placeholder when no image exists', () => {
    const wrapper = mount(CoinViewer3D, {
      global: {
        stubs: { RefreshCw: true },
      },
    })

    expect(wrapper.text()).toContain('No image')
  })

  it('emits open-image with the current face', async () => {
    const wrapper = mount(CoinViewer3D, {
      props: {
        obverseSrc: '/uploads/obverse.webp',
        reverseSrc: '/uploads/reverse.webp',
      },
      global: {
        stubs: { RefreshCw: true },
      },
    })

    await wrapper.find('.coin-stage').trigger('click')
    await wrapper.find('.coin-flip-button').trigger('click')
    await flushPromises()
    await wrapper.find('.coin-stage').trigger('click')

    expect(wrapper.emitted('open-image')).toEqual([['obverse'], ['reverse']])
  })

  it('emits reverse for a reverse-only static coin', async () => {
    const wrapper = mount(CoinViewer3D, {
      props: {
        reverseSrc: '/uploads/reverse.webp',
      },
      global: {
        stubs: { RefreshCw: true },
      },
    })

    await wrapper.find('.coin-stage').trigger('click')

    expect(wrapper.emitted('open-image')).toEqual([['reverse']])
  })

  it('marks reduced-motion mode from the operating system preference', async () => {
    stubMotion(true)

    const wrapper = mount(CoinViewer3D, {
      props: {
        obverseSrc: '/uploads/obverse.webp',
        reverseSrc: '/uploads/reverse.webp',
      },
      global: {
        stubs: { RefreshCw: true },
      },
    })
    await flushPromises()

    expect(wrapper.find('.coin-viewer-3d').classes()).toContain('reduced-motion')
  })
})
