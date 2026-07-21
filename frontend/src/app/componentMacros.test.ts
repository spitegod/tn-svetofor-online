import { describe, expect, it } from 'vitest'

const setupComponents = import.meta.glob([
  '../pages/*/*.vue',
  '../features/nav-parser/NavParserPanel.vue',
], {
  eager: true,
  import: 'default',
  query: '?raw',
}) as Record<string, string>

describe('Vue compiler macros', () => {
  it.each(Object.entries(setupComponents))('uses defineProps as a direct compiler-macro assignment in %s', (_path, source) => {
    expect(source).not.toMatch(/defineProps<[^>]+>\(\)\./)
  })
})
