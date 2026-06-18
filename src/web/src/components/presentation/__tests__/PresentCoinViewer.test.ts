import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PresentCoinViewer from '../PresentCoinViewer.vue'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'

const coin = buildRomanDenariusCore({
  purchasePrice: 200,
  currentValue: 250,
  notes: 'Private note',
  aiAnalysis: 'Private analysis',
  images: [
    { id: 1, coinId: 1, filePath: 'obverse.webp', imageType: 'obverse', isPrimary: true, createdAt: '2026-01-01T00:00:00Z' },
    { id: 2, coinId: 1, filePath: 'reverse.webp', imageType: 'reverse', isPrimary: false, createdAt: '2026-01-01T00:00:00Z' },
  ],
})

function mountViewer() {
  return mount(PresentCoinViewer, {
    props: { coin, index: 0, total: 2 },
    global: {
      stubs: {
        ChevronLeft: true,
        ChevronRight: true,
        Coins: true,
        X: true,
      },
    },
  })
}

describe('PresentCoinViewer', () => {
  it('toggles a metadata overlay without price, value, notes, or analysis fields', async () => {
    const wrapper = mountViewer()

    await wrapper.find('.present-stage').trigger('click')

    expect(wrapper.text()).toContain(coin.name)
    expect(wrapper.text()).toContain(coin.ruler)
    expect(wrapper.text()).toContain(coin.denomination)
    expect(wrapper.text()).not.toContain('200')
    expect(wrapper.text()).not.toContain('250')
    expect(wrapper.text()).not.toContain('Private note')
    expect(wrapper.text()).not.toContain('Private analysis')

    await wrapper.find('.present-stage').trigger('click')

    expect(wrapper.find('.metadata-overlay').exists()).toBe(false)
  })

  it('switches between obverse and reverse images', async () => {
    const wrapper = mountViewer()

    expect(wrapper.find('.present-image').attributes('src')).toBe('/uploads/obverse.webp')

    await wrapper.findAll('.chip')[1]!.trigger('click')

    expect(wrapper.find('.present-image').attributes('src')).toBe('/uploads/reverse.webp')
  })

  it('emits keyboard and swipe navigation events', async () => {
    const wrapper = mountViewer()

    await wrapper.find('.present-viewer').trigger('keydown', { key: 'ArrowRight' })
    await wrapper.find('.present-viewer').trigger('keydown', { key: 'ArrowLeft' })
    wrapper.find('.present-stage').element.dispatchEvent(new MouseEvent('pointerdown', { clientX: 200, clientY: 20 }))
    window.dispatchEvent(new MouseEvent('pointerup', { clientX: 100, clientY: 24 }))

    expect(wrapper.emitted('next')).toHaveLength(2)
    expect(wrapper.emitted('prev')).toHaveLength(1)
  })

  it('marks reduced-motion mode when requested', () => {
    const wrapper = mount(PresentCoinViewer, {
      props: { coin, index: 0, total: 1, reducedMotion: true },
      global: {
        stubs: { ChevronLeft: true, ChevronRight: true, Coins: true, X: true },
      },
    })

    expect(wrapper.find('.present-viewer').classes()).toContain('reduced-motion')
  })
})
