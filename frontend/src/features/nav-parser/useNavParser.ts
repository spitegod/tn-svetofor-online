import { computed, getCurrentScope, onScopeDispose, ref } from 'vue'
import { apiFetch } from '@/shared/api/client'
import type { NavParserProgress, NavParserRun, NavParserSettings } from '@/shared/api/types'

type NavParserOptions = {
  onCatalogsInvalidated?: () => void | Promise<void>
}

const initialProgress = (): NavParserProgress => ({
  running: false,
  source: 'manual',
  stage: 'Ожидание',
  message: 'Парсер готов к запуску',
  percent: 0,
  processed: 0,
  total: 0,
  found: 0,
  updated: 0,
  failed: 0,
  notFound: 0,
  startedAt: null,
  finishedAt: null,
  logs: [],
})

export function useNavParser(options: NavParserOptions = {}) {
  const isNavParsing = ref(false)
  const isNavParserCancelling = ref(false)
  const navParseMessage = ref('')
  const navParseError = ref('')
  const navParseNotFound = ref<string[]>([])
  const navParseFailedSystems = ref<string[]>([])
  const navParserIntervalDays = ref(7)
  const navParserProgress = ref<NavParserProgress>(initialProgress())
  const navParserLogsNewestFirst = computed(() => [...navParserProgress.value.logs].reverse())
  const navParserRuns = ref<NavParserRun[]>([])
  const isNavParserLogOpen = ref(false)
  const isNavParserHistoryOpen = ref(false)
  const openedNavParserRunId = ref<number | null>(null)
  const navParserWorkerCount = ref(4)
  const navParserRequestTimeout = ref(35)
  const navParserRetryAttempts = ref(3)
  const navParserRetryDelay = ref(2)
  const navParserFallbackSearch = ref(true)
  const navParserNextRunAt = ref<string | null>(null)
  const isNavParserSettingsOpen = ref(false)
  const isNavSettingsSaving = ref(false)
  const navSettingsMessage = ref('')
  const navSettingsError = ref('')

  let pollTimer: ReturnType<typeof window.setInterval> | null = null
  let requestPending = false
  let cancelRequested = false

  function applySettings(settings: NavParserSettings) {
    navParserIntervalDays.value = settings.updateIntervalDays ?? 1
    navParserWorkerCount.value = settings.workerCount ?? 4
    navParserRequestTimeout.value = settings.requestTimeoutSeconds ?? 35
    navParserRetryAttempts.value = settings.retryAttempts ?? 3
    navParserRetryDelay.value = settings.retryDelaySeconds ?? 2
    navParserFallbackSearch.value = settings.fallbackSearch ?? true
    navParserNextRunAt.value = settings.nextRunAt ?? null
  }

  async function loadNavParserSettings() {
    try {
      const response = await apiFetch('/nav-parser/settings')
      if (!response.ok) throw new Error('Не удалось загрузить настройки парсера')
      applySettings(await response.json() as NavParserSettings)
      navSettingsError.value = ''
    } catch (error) {
      navSettingsError.value = error instanceof Error ? error.message : 'Не удалось загрузить настройки парсера'
    }
  }

  async function loadNavParserRuns() {
    try {
      const response = await apiFetch('/nav-parser/runs?limit=5')
      if (!response.ok) throw new Error('Не удалось загрузить историю запусков')
      const runs = await response.json() as NavParserRun[] | null
      navParserRuns.value = (runs ?? []).map((run) => ({ ...run, logs: run.logs ?? [] }))
    } catch (error) {
      navSettingsError.value = error instanceof Error ? error.message : 'Не удалось загрузить историю запусков'
    }
  }

  async function loadNavParserProgress() {
    try {
      const wasRunning = navParserProgress.value.running
      const response = await apiFetch('/nav-parser/status')
      if (!response.ok) throw new Error('Не удалось загрузить прогресс парсера')
      const progress = await response.json() as NavParserProgress
      navParserProgress.value = { ...progress, logs: progress.logs ?? [] }
      if (progress.running) {
        isNavParserLogOpen.value = true
      } else if (navParserProgress.value.logs.length === 0) {
        isNavParserLogOpen.value = false
      }
      isNavParsing.value = progress.running || requestPending
      if (progress.running && !pollTimer) startNavParserPolling()
      if (wasRunning && !progress.running) {
        cancelRequested = false
        void Promise.all([
          loadNavParserRuns(),
          loadNavParserSettings(),
          Promise.resolve(options.onCatalogsInvalidated?.()),
        ]).catch((error) => {
          navParseError.value = error instanceof Error ? error.message : 'Не удалось обновить данные после парсинга'
        })
      }
    } catch (error) {
      if (requestPending) {
        navParseError.value = error instanceof Error ? error.message : 'Не удалось загрузить прогресс парсера'
      }
    }
  }

  function startNavParserPolling() {
    if (pollTimer) return
    void loadNavParserProgress()
    pollTimer = window.setInterval(() => {
      void loadNavParserProgress().then(() => {
        if (!navParserProgress.value.running && !requestPending) stopNavParserPolling()
      })
    }, 750)
  }

  function stopNavParserPolling() {
    if (!pollTimer) return
    window.clearInterval(pollTimer)
    pollTimer = null
  }

  async function runNavParser() {
    if (isNavParsing.value) return
    requestPending = true
    cancelRequested = false
    isNavParsing.value = true
    navParseMessage.value = ''
    navParseError.value = ''
    navParseNotFound.value = []
    navParseFailedSystems.value = []
    isNavParserLogOpen.value = true
    startNavParserPolling()
    try {
      const response = await apiFetch('/nav-parser/runs', { method: 'POST' })
      if (!response.ok) {
        const payload = await response.json().catch(() => null)
        if (response.status === 409) {
          await loadNavParserProgress()
          return
        }
        throw new Error(payload?.error ?? 'Не удалось выполнить парсинг nav.tn.ru')
      }
      navParseMessage.value = 'Парсер запущен в фоновом режиме'
      await loadNavParserProgress()
    } catch (error) {
      if (cancelRequested) navParseMessage.value = 'Парсинг отменён'
      else navParseError.value = error instanceof Error ? error.message : 'Не удалось выполнить парсинг nav.tn.ru'
    } finally {
      requestPending = false
      await Promise.all([loadNavParserProgress(), loadNavParserRuns()])
      isNavParsing.value = navParserProgress.value.running
      cancelRequested = false
      if (!navParserProgress.value.running) stopNavParserPolling()
    }
  }

  async function cancelNavParser() {
    if (!isNavParsing.value || isNavParserCancelling.value) return
    isNavParserCancelling.value = true
    cancelRequested = true
    navParseError.value = ''
    try {
      const response = await apiFetch('/nav-parser/cancel', { method: 'POST' })
      if (!response.ok) {
        const payload = await response.json().catch(() => null)
        if (response.status !== 409) throw new Error(payload?.error ?? 'Не удалось остановить парсинг')
      }
      await Promise.all([loadNavParserProgress(), loadNavParserRuns()])
    } catch (error) {
      cancelRequested = false
      navParseError.value = error instanceof Error ? error.message : 'Не удалось остановить парсинг'
    } finally {
      isNavParserCancelling.value = false
    }
  }

  async function saveNavParserSettings() {
    const days = Math.min(365, Math.max(1, Math.round(Number(navParserIntervalDays.value) || 1)))
    navParserIntervalDays.value = days
    navParserWorkerCount.value = Math.min(10, Math.max(1, Math.round(Number(navParserWorkerCount.value) || 4)))
    navParserRequestTimeout.value = Math.min(120, Math.max(5, Math.round(Number(navParserRequestTimeout.value) || 35)))
    navParserRetryAttempts.value = Math.min(5, Math.max(1, Math.round(Number(navParserRetryAttempts.value) || 3)))
    navParserRetryDelay.value = Math.min(30, Math.max(1, Math.round(Number(navParserRetryDelay.value) || 2)))
    isNavSettingsSaving.value = true
    navSettingsMessage.value = ''
    navSettingsError.value = ''
    try {
      const response = await apiFetch('/nav-parser/settings', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          updateIntervalDays: days,
          workerCount: navParserWorkerCount.value,
          requestTimeoutSeconds: navParserRequestTimeout.value,
          retryAttempts: navParserRetryAttempts.value,
          retryDelaySeconds: navParserRetryDelay.value,
          fallbackSearch: navParserFallbackSearch.value,
        }),
      })
      if (!response.ok) {
        const payload = await response.json().catch(() => null)
        throw new Error(payload?.error ?? 'Не удалось сохранить частоту обновления')
      }
      applySettings(await response.json() as NavParserSettings)
      navSettingsMessage.value = 'Частота обновления сохранена'
    } catch (error) {
      navSettingsError.value = error instanceof Error ? error.message : 'Не удалось сохранить частоту обновления'
    } finally {
      isNavSettingsSaving.value = false
    }
  }

  function formatNavParserLogTime(value: string) {
    if (!value) return '—'
    return new Intl.DateTimeFormat('ru-RU', {
      hour: '2-digit', minute: '2-digit', second: '2-digit',
    }).format(new Date(value))
  }

  function formatNavParserRunDate(value: string) {
    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
    }).format(new Date(value))
  }

  function formatNavParserRunDuration(run: NavParserRun) {
    const seconds = Math.max(0, Math.round((new Date(run.finishedAt).getTime() - new Date(run.startedAt).getTime()) / 1000))
    if (seconds < 60) return `${seconds} сек.`
    const minutes = Math.floor(seconds / 60)
    const remainder = seconds % 60
    return remainder ? `${minutes} мин. ${remainder} сек.` : `${minutes} мин.`
  }

  function navParserSourceLabel(source: string) {
    return source === 'scheduled' ? 'По расписанию' : 'Ручной запуск'
  }

  function navParserRunLogsNewestFirst(run: NavParserRun) {
    return [...run.logs].reverse()
  }

  function navParserNextRunLabel() {
    if (!navParserNextRunAt.value) return 'Следующий запуск — после первого успешного запуска'
    const value = new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
    }).format(new Date(navParserNextRunAt.value))
    return `Следующий запуск: ${value}`
  }

  if (getCurrentScope()) {
    onScopeDispose(stopNavParserPolling)
  }

  return {
    cancelNavParser,
    formatNavParserLogTime,
    formatNavParserRunDate,
    formatNavParserRunDuration,
    isNavParserCancelling,
    isNavParserHistoryOpen,
    isNavParserLogOpen,
    isNavParserSettingsOpen,
    isNavParsing,
    isNavSettingsSaving,
    loadNavParserProgress,
    loadNavParserRuns,
    loadNavParserSettings,
    navParseError,
    navParseFailedSystems,
    navParseMessage,
    navParseNotFound,
    navParserFallbackSearch,
    navParserIntervalDays,
    navParserLogsNewestFirst,
    navParserNextRunLabel,
    navParserProgress,
    navParserRequestTimeout,
    navParserRetryAttempts,
    navParserRetryDelay,
    navParserRunLogsNewestFirst,
    navParserRuns,
    navParserSourceLabel,
    navParserWorkerCount,
    navSettingsError,
    navSettingsMessage,
    openedNavParserRunId,
    runNavParser,
    saveNavParserSettings,
    stopNavParserPolling,
  }
}
