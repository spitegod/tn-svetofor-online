export const pageKeys = [
  'changes',
  'systems',
  'classification',
  'comparison',
  'settings',
  'admin-settings',
] as const

export type PageKey = typeof pageKeys[number]

const pageKeySet = new Set<string>(pageKeys)

export function isPageKey(value: string): value is PageKey {
  return pageKeySet.has(value)
}

export function pagePath(page: PageKey): string {
  return page === 'changes' ? '/' : `/${page}`
}

export function pageFromURL(location: Pick<Location, 'pathname' | 'hash'>): PageKey {
  const pathPage = location.pathname.replace(/^\/+|\/+$/g, '')
  if (isPageKey(pathPage)) {
    return pathPage
  }

  const hashPage = location.hash.replace(/^#\/?/, '').replace(/\/+$/g, '')
  return isPageKey(hashPage) ? hashPage : 'changes'
}
