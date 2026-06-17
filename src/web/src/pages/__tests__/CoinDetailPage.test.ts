import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CoinDetailPage from '../CoinDetailPage.vue'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'

const coin = buildRomanDenariusCore()
const fetchCoin = vi.fn()

vi.mock('@/stores/coins', () => ({
  useCoinsStore: () => ({
    loading: false,
    currentCoin: coin,
    fetchCoin,
  }),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: String(coin.id) } }),
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/api/client', () => ({
  deleteCoin: vi.fn(),
  purchaseCoin: vi.fn(),
  sellCoin: vi.fn(),
}))

vi.mock('@/composables/useDialog', () => ({
  useDialog: () => ({
    showConfirm: vi.fn(),
    showAlert: vi.fn(),
  }),
}))

describe('CoinDetailPage', () => {
  beforeEach(() => {
    fetchCoin.mockReset()
    Object.defineProperty(window, 'matchMedia', {
      value: vi.fn(() => ({
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      })),
      configurable: true,
    })
    Object.defineProperty(window, 'DeviceOrientationEvent', { value: undefined, configurable: true })
  })

  it('renders the shared 3D viewer in the detail hero', () => {
    const wrapper = mount(CoinDetailPage, {
      global: {
        stubs: pageStubs(),
      },
    })

    expect(wrapper.findComponent({ name: 'CoinViewer3D' }).exists()).toBe(true)
    expect(fetchCoin).toHaveBeenCalledWith(coin.id)
  })

  it('opens the existing image lightbox for the current viewer face', async () => {
    const wrapper = mount(CoinDetailPage, {
      global: {
        stubs: pageStubs(),
      },
    })

    await wrapper.find('.coin-stage').trigger('click')
    await flushPromises()

    expect(wrapper.findComponent({ name: 'ImageLightbox' }).exists()).toBe(true)
  })
})

function pageStubs() {
  return {
    RouterLink: {
      props: ['to'],
      template: '<a :href="to"><slot /></a>',
    },
    SellModal: true,
    PurchaseModal: true,
    ImageLightbox: true,
    CoinTagsSection: true,
    CoinDetailMetadataTable: true,
    CoinDetailSectionLinks: true,
    CoinListingStatus: true,
    CoinReferencesSection: true,
    CoinDetailHeaderActions: true,
    RefreshCw: true,
  }
}
