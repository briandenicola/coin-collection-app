import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import PwaCollectionHeader from '../collection/PwaCollectionHeader.vue'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const headerPath = path.resolve(__dirname, '../collection/PwaCollectionHeader.vue')

describe('PwaCollectionHeader', () => {
  it('keeps select mode in the menu instead of the top row', () => {
    const source = fs.readFileSync(headerPath, 'utf8')

    expect(source).not.toContain('class="pwa-icon-btn"')
    expect(source).toContain('<span class="pwa-menu-label">Selection</span>')
    expect(source).toContain("{{ selectMode ? 'Exit Selection Mode' : 'Enable Selection Mode' }}")
  })

  it('emits present from the PWA menu and closes the menu', async () => {
    const wrapper = mount(PwaCollectionHeader, {
      props: {
        search: '',
        selectMode: false,
        menuOpen: true,
        selectedCategory: '',
        selectedEra: '',
        selectedTag: '',
        userTags: [],
        sortKey: 'updated_at_desc',
        viewMode: 'swipe',
        gridSide: null,
      },
      global: {
        stubs: {
          SearchBar: true,
          CategoryFilter: true,
          EraFilter: true,
          SortSelect: true,
          Layers: true,
          LayoutGrid: true,
          MonitorPlay: true,
          SlidersHorizontal: true,
        },
      },
    })

    await wrapper.findAll('button').find((button) => button.text().includes('Present Collection'))!.trigger('click')

    expect(wrapper.emitted('present')).toHaveLength(1)
    expect(wrapper.emitted('update:menuOpen')).toEqual([[false]])
  })
})
