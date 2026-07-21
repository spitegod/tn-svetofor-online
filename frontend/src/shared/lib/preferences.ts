import { ref, watch, type Ref } from 'vue'

export function usePersistedChoice<T extends string>(
  storageKey: string,
  fallback: T,
  allowedValues: readonly T[],
): Ref<T> {
  let initialValue = fallback
  try {
    const savedValue = window.localStorage.getItem(storageKey) as T | null
    if (savedValue && allowedValues.includes(savedValue)) {
      initialValue = savedValue
    }
  } catch {
    // The fallback remains active when browser storage is unavailable.
  }

  const value = ref(initialValue) as Ref<T>
  watch(value, (nextValue) => {
    if (!allowedValues.includes(nextValue)) return
    try {
      window.localStorage.setItem(storageKey, nextValue)
    } catch {
      // The selection still works for the current session.
    }
  })
  return value
}

export function usePersistedPageSize(key: string, fallback: string, allowedValues: readonly string[]) {
  return usePersistedChoice(`tn-svetofor:page-size:${key}`, fallback, allowedValues)
}
