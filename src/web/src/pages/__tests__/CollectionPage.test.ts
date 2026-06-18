import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const collectionPagePath = path.resolve(__dirname, '../CollectionPage.vue')

describe('CollectionPage', () => {
  it('does not include a floating add button in PWA mode', () => {
    const source = fs.readFileSync(collectionPagePath, 'utf8')
    expect(source).not.toContain('class="add-fab"')
    expect(source).not.toMatch(/\.add-fab\s*\{/)
  })

  it('wires collection headers to Present mode without adding a floating action button', () => {
    const source = fs.readFileSync(collectionPagePath, 'utf8')

    expect(source).toContain('@present="openPresentMode"')
    expect(source).toContain("router.push({ name: 'present', query: { start: String(store.galleryIndex) } })")
  })
})
