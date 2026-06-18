import { ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PresentModePage from '../PresentModePage.vue'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'

const routerPush = vi.fn()
const fetchCoins = vi.fn()
const store = {
  coins: [
    buildRomanDenariusCore({ id: 1, name: 'First coin' }),
    buildRomanDenariusCore({ id: 2, name: 'Second coin' }),
  ],
  galleryIndex: 0,
  fetchCoins,
}
const fullscreenEnter = vi.fn(async () => true)
const fullscreenExit = vi.fn(async () => true)
const wakeRequest = vi.fn(async () => true)
const wakeRelease = vi.fn(async () => true)

vi.mock('@/stores/coins', () => ({
  useCoinsStore: () => store,
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: { start: '1' } }),
  useRouter: () => ({ push: routerPush }),
}))

vi.mock('@/composables/useReducedMotion', () => ({
  useReducedMotion: () => ({ prefersReducedMotion: ref(false) }),
}))

vi.mock('@/composables/useFullscreen', () => ({
  useFullscreen: () => ({ enter: fullscreenEnter, exit: fullscreenExit }),
}))

vi.mock('@/composables/useWakeLock', () => ({
  useWakeLock: () => ({ request: wakeRequest, release: wakeRelease }),
}))

const presentViewerStub = {
  props: ['coin', 'index', 'total', 'reducedMotion'],
  emits: ['next', 'prev', 'exit'],
  template: `
    <div class="viewer-stub">
      <span class="viewer-name">{{ coin.name }}</span>
      <span class="viewer-index">{{ index }}</span>
      <button class="next" @click="$emit('next')">Next</button>
      <button class="exit" @click="$emit('exit')">Exit</button>
    </div>
  `,
}

describe('PresentModePage', () => {
  beforeEach(() => {
    store.galleryIndex = 0
    fetchCoins.mockReset()
    routerPush.mockReset()
    fullscreenEnter.mockClear()
    fullscreenExit.mockClear()
    wakeRequest.mockClear()
    wakeRelease.mockClear()
  })

  it('starts at the requested index and updates the store gallery index while navigating', async () => {
    const wrapper = mount(PresentModePage, {
      global: { stubs: { PresentCoinViewer: presentViewerStub } },
    })
    await flushPromises()

    expect(wrapper.find('.viewer-name').text()).toBe('Second coin')
    expect(fullscreenEnter).toHaveBeenCalled()
    expect(wakeRequest).toHaveBeenCalled()

    await wrapper.find('.next').trigger('click')

    expect(store.galleryIndex).toBe(0)
    expect(wrapper.find('.viewer-name').text()).toBe('First coin')
  })

  it('releases wake lock and fullscreen before returning to the collection', async () => {
    const wrapper = mount(PresentModePage, {
      global: { stubs: { PresentCoinViewer: presentViewerStub } },
    })
    await flushPromises()

    await wrapper.find('.exit').trigger('click')
    await flushPromises()

    expect(wakeRelease).toHaveBeenCalled()
    expect(fullscreenExit).toHaveBeenCalled()
    expect(routerPush).toHaveBeenCalledWith({ name: 'collection' })
  })
})
