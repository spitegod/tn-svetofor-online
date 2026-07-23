import { describe, expect, it } from 'vitest'
import { pageFromURL, pagePath } from './routes'

describe('application routes', () => {
  it('opens public appearance settings by a direct path', () => {
    expect(pageFromURL({ pathname: '/settings', hash: '' })).toBe('settings')
  })

  it('opens hidden administrator settings by a direct path', () => {
    expect(pageFromURL({ pathname: '/admin-settings', hash: '' })).toBe('admin-settings')
  })

  it('keeps legacy hash links working', () => {
    expect(pageFromURL({ pathname: '/', hash: '#/classification' })).toBe('classification')
  })

  it('uses clean paths for navigation', () => {
    expect(pagePath('changes')).toBe('/')
    expect(pagePath('settings')).toBe('/settings')
    expect(pagePath('admin-settings')).toBe('/admin-settings')
  })
})
