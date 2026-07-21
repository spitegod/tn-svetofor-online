// @vitest-environment jsdom

import { nextTick } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { usePersistedChoice } from './preferences'

describe('usePersistedChoice', () => {
  beforeEach(() => window.localStorage.clear())

  it('loads and saves an allowed value', async () => {
    window.localStorage.setItem('preference', 'large')
    const preference = usePersistedChoice('preference', 'standard', ['standard', 'large'] as const)

    expect(preference.value).toBe('large')
    preference.value = 'standard'
    await nextTick()
    expect(window.localStorage.getItem('preference')).toBe('standard')
  })

  it('ignores an unknown stored value', () => {
    window.localStorage.setItem('preference', 'unknown')

    expect(usePersistedChoice('preference', 'standard', ['standard', 'large'] as const).value).toBe('standard')
  })
})
