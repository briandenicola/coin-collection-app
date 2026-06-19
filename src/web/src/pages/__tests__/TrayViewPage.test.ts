import { flushPromises, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TrayViewPage from '@/pages/TrayViewPage.vue'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'

const mockGetCoins = vi.fn()
const mockPush = vi.fn()

vi.mock('@/api/client', () => ({
  getCoins: (params?: Record<string, unknown>) => mockGetCoins(params),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

vi.mock('@/composables/useTrayPreference', () => ({
  useTrayPreference: () => ({
    feltColor: 'red',
  }),
}))

const routerLinkStub = {
  props: ['to'],
  template: '<a :href="to"><slot /></a>',
}

describe('TrayViewPage', () => {
  beforeEach(() => {
    mockGetCoins.mockReset()
    mockPush.mockReset()
  })

  it('fetches every active collection page for tray drawers', async () => {
    const firstPage = Array.from({ length: 100 }, (_, index) =>
      buildRomanDenariusCore({ id: index + 1, name: `Coin ${index + 1}` }),
    )
    const secondPage = Array.from({ length: 20 }, (_, index) =>
      buildRomanDenariusCore({ id: index + 101, name: `Coin ${index + 101}` }),
    )
    mockGetCoins
      .mockResolvedValueOnce({ data: { coins: firstPage, total: 120 } })
      .mockResolvedValueOnce({ data: { coins: secondPage, total: 120 } })

    shallowMount(TrayViewPage, {
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          MuseumTray: true,
          TrayControls: true,
        },
      },
    })
    await flushPromises()

    expect(mockGetCoins).toHaveBeenNthCalledWith(1, {
      wishlist: 'false',
      sold: 'false',
      page: 1,
      limit: 100,
      sort: 'name',
      order: 'asc',
    })
    expect(mockGetCoins).toHaveBeenNthCalledWith(2, {
      wishlist: 'false',
      sold: 'false',
      page: 2,
      limit: 100,
      sort: 'name',
      order: 'asc',
    })
    expect(mockGetCoins).toHaveBeenCalledTimes(2)
  })
})
