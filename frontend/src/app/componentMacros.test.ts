import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const setupComponents = [
  '../pages/changes/ChangesPage.vue',
  '../pages/systems/SystemsPage.vue',
  '../pages/classification/ClassificationPage.vue',
  '../pages/comparison/ComparisonPage.vue',
  '../pages/settings/SettingsPage.vue',
  '../features/nav-parser/NavParserPanel.vue',
]

describe('Vue compiler macros', () => {
  it.each(setupComponents)('uses defineProps as a direct compiler-macro assignment in %s', (path) => {
    const source = readFileSync(new URL(path, import.meta.url), 'utf8')

    expect(source).not.toMatch(/defineProps<[^>]+>\(\)\./)
  })
})
