<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import genericFileIcon from 'bootstrap-icons/icons/file-earmark.svg'
import docFileIcon from 'bootstrap-icons/icons/filetype-doc.svg'
import docxFileIcon from 'bootstrap-icons/icons/filetype-docx.svg'
import jpgFileIcon from 'bootstrap-icons/icons/filetype-jpg.svg'
import pdfFileIcon from 'bootstrap-icons/icons/filetype-pdf.svg'
import pngFileIcon from 'bootstrap-icons/icons/filetype-png.svg'
import xlsFileIcon from 'bootstrap-icons/icons/filetype-xls.svg'
import xlsxFileIcon from 'bootstrap-icons/icons/filetype-xlsx.svg'
import { apiFetch, apiURL } from '@/shared/api/client'
import type {
  ClassificationChange,
  ClassificationResponse,
  ClassificationStats,
  ComparisonRow,
  Order,
  SystemCatalogResponse,
  SystemCatalogRow,
  SystemCatalogStats,
  SystemCharacteristic,
  SystemDocumentResponse,
  SystemDocumentRow,
  SystemTypeOption,
} from '@/shared/api/types'
import { usePersistedChoice, usePersistedPageSize } from '@/shared/lib/preferences'
import AppHeader from '@/widgets/AppHeader.vue'
import ClassLegendFooter from '@/widgets/ClassLegendFooter.vue'
import SystemHistoryModal from '@/features/system-history/SystemHistoryModal.vue'
import { useNavParser } from '@/features/nav-parser/useNavParser'
import SettingsPage, { type SettingsPageViewModel } from '@/pages/settings/SettingsPage.vue'
import ComparisonPage, { type ComparisonPageViewModel } from '@/pages/comparison/ComparisonPage.vue'
import ClassificationPage, { type ClassificationPageViewModel } from '@/pages/classification/ClassificationPage.vue'
import SystemsPage, { type SystemsPageViewModel } from '@/pages/systems/SystemsPage.vue'
import ChangesPage, { type ChangesPageViewModel } from '@/pages/changes/ChangesPage.vue'

type FontSizePreset = 'small' | 'standard' | 'large'

const fontSizePresets: Array<{ key: FontSizePreset; label: string; size: number }> = [
  { key: 'small', label: 'Маленький', size: 12 },
  { key: 'standard', label: 'Стандартный', size: 15 },
  { key: 'large', label: 'Большой', size: 18 },
]

const fontSizePreset = usePersistedChoice<FontSizePreset>(
  'tn-svetofor:font-size',
  'standard',
  fontSizePresets.map((preset) => preset.key),
)

const minimumAppFontSize = computed(() => (
  fontSizePresets.find((preset) => preset.key === fontSizePreset.value)?.size ?? 15
))


const pageKeys = new Set(['changes', 'systems', 'classification', 'comparison', 'settings'])

function pageFromLocation() {
  const page = window.location.hash.replace(/^#\/?/, '')
  return pageKeys.has(page) ? page : 'changes'
}

const activePage = ref(pageFromLocation())
const isScrollTopVisible = ref(false)

const navItems = [
  { key: 'changes', label: 'Изменения' },
  { key: 'systems', label: 'Список систем' },
  { key: 'classification', label: 'Классификация' },
  { key: 'comparison', label: 'Сравнение' },
  { key: 'settings', label: 'Настройки' },
]

const constructionTypes = [
  'Все',
  'Промышленное и гражданское строительство',
  'Индивидуальное жилищное строительство',
  'Транспортное и дорожное строительство',
  'Специальные сооружения',
]
const selectedConstructionType = ref(constructionTypes[0])

const classOptions = ['Разрешенная', 'Рекомендованная', 'Запрещенная']
const curatorOptions = ['Все кураторы', 'Сендецкий В.', 'Уртенков А.', 'Золотарев М.', 'Кузнецова Н.']

const orders = ref<Order[]>([])
const ordersPageSize = usePersistedPageSize('settings-orders', '10', ['5', '10', '20', '50', '100', 'all'])
const ordersPage = ref(1)
const selectedOrderId = ref<number | null>(null)
const MAX_COMPARISON_ORDERS = 6
const comparisonOrderIds = ref<number[]>([])
const comparisonCatalogByOrder = ref<Record<number, SystemDocumentRow[]>>({})
const comparisonPendingIds = ref<number[]>([])
const attachmentPendingIds = ref<number[]>([])
const isBulkComparisonUpdating = ref(false)
const hiddenComparisonRows = ref<string[]>([])
const comparisonOnlyDifferences = ref(false)
const comparisonSort = ref<'differences-first' | 'name-asc' | 'name-desc'>('differences-first')
const comparisonPageSize = usePersistedPageSize('comparison', '50', ['50', '100', 'all'])
const comparisonPage = ref(1)
const isComparisonLoading = ref(true)
const comparisonError = ref('')
const isOrdersLoading = ref(true)
const isOrderWorkbookImporting = ref(false)
const ordersError = ref('')
const settingsOrderMenuId = ref<number | null>(null)
const orderRenameTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const classificationEditTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const systemCatalogEditTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const documentCommentTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const orderRenameControllers = new Map<number, AbortController>()
const classificationEditControllers = new Map<number, AbortController>()
const systemCatalogEditControllers = new Map<number, AbortController>()
const documentCommentControllers = new Map<number, AbortController>()
let systemDocumentSearchTimer: ReturnType<typeof window.setTimeout> | null = null
let classificationSearchTimer: ReturnType<typeof window.setTimeout> | null = null
let classificationFilterFeedbackTimer: ReturnType<typeof window.setTimeout> | null = null
let classificationRequestController: AbortController | null = null
let systemCatalogRequestController: AbortController | null = null
let classificationCatalogRequestController: AbortController | null = null
let systemDocumentRequestController: AbortController | null = null
let documentTableRequestController: AbortController | null = null
const selectedSystemTypeSlug = ref('')
const isSystemTypesOpen = ref(false)
const selectedHistorySystem = ref<SystemDocumentRow | null>(null)
const systemHistoryRows = ref<SystemDocumentRow[]>([])
const isSystemHistoryLoading = ref(false)
const systemHistoryError = ref('')
const isHistoryOpen = ref(false)
const openedSelect = ref<string | null>(null)
const draggedComparisonOrderId = ref<number | null>(null)
const comparisonDropIndex = ref<number | null>(null)
const importFileInput = ref<HTMLInputElement | null>(null)
const systemCatalogFileInput = ref<HTMLInputElement | null>(null)
const orderWorkbookInput = ref<HTMLInputElement | null>(null)
const classificationRows = ref<ClassificationChange[]>([])
const classificationPageSize = usePersistedPageSize('changes', '50', ['50', '100', 'all'])
const classificationPage = ref(1)
const settingsClassificationPageSize = usePersistedPageSize('settings-table-1', '10', ['10', '20', '50', '100', 'all'])
const settingsClassificationPage = ref(1)
const isSettingsClassificationUnlocked = ref(false)
const classificationStats = ref<ClassificationStats>({
  addedSystems: 0,
  recommended: 0,
  allowed: 0,
  classificationChanges: 0,
})
const beforeOptions = ref(['Все', 'Новая система', ...classOptions])
const afterOptions = ref(['Все', ...classOptions])
const selectedBeforeFilter = ref('Все')
const selectedAfterFilter = ref('Все')
const tableSearch = ref('')
const isClassificationLoading = ref(true)
const isClassificationFiltering = ref(false)
const isChangesRefreshing = ref(false)
const isChangesRefreshDone = ref(false)
const changesLastRefreshedAt = ref('')
const classificationLoadingMessage = ref('Загрузка таблицы...')
const classificationError = ref('')
const classificationConstructionTypes = computed(() => [...constructionTypes, 'Тип не присвоен'].map((name) => ({
  name,
  label: name === 'Тип не присвоен' ? 'Тип не найден' : name,
  count: name === 'Все'
    ? classificationRows.value.length
    : classificationRows.value.filter((row) => row.constructionType === name).length,
})))
const activeChangeFilterCount = computed(() => [
  selectedConstructionType.value !== 'Все',
  selectedBeforeFilter.value !== 'Все',
  selectedAfterFilter.value !== 'Все',
].filter(Boolean).length)
const hasActiveChangeFilters = computed(() => activeChangeFilterCount.value > 0)
const systemCatalogRows = ref<SystemCatalogRow[]>([])
const settingsSystemCatalogPageSize = usePersistedPageSize('settings-table-2', '10', ['10', '20', '50', '100', 'all'])
const settingsSystemCatalogPage = ref(1)
const isSettingsSystemCatalogUnlocked = ref(false)
const systemDocumentRows = ref<SystemDocumentRow[]>([])
const systemDocumentPageSize = usePersistedPageSize('systems', '50', ['50', '100', 'all'])
const systemDocumentPage = ref(1)
const systemsConstructionTypes = computed(() => constructionTypes.map((name) => ({
  name,
  count: systemDocumentRows.value.filter((system) => systemMatchesConstructionType(system, name)).length,
})))
const documentRows = ref<SystemDocumentRow[]>([])
const documentSearch = ref('')
const settingsDocumentsPageSize = usePersistedPageSize('settings-table-3', '10', ['10', '20', '50', '100', 'all'])
const settingsDocumentsPage = ref(1)
const documentError = ref('')
const isDocumentTableLoading = ref(true)
const classificationCatalogRows = ref<SystemCatalogRow[]>([])
const classificationCatalogSearch = ref('')
const classificationCatalogSearchInput = ref('')
const isClassificationSearchPending = ref(false)
const classificationView = ref<'grid' | 'list'>('grid')
const classificationCatalogPageSize = usePersistedPageSize('classification', '50', ['50', '100', '200', 'all'])
const classificationCatalogPage = ref(1)
const isClassificationCatalogLoading = ref(true)
const classificationCatalogError = ref('')
const classificationCatalogConstructionTypes = computed(() => constructionTypes.map((name) => ({
  name,
  count: classificationCatalogRows.value.filter((system) => systemMatchesConstructionType(system, name)).length,
})))
const parsedSystemTypes = ref<SystemTypeOption[]>([])
const openedClassificationSystemId = ref<number | null>(null)
const classificationCardColumns = ref(3)
const openedClassificationFilter = ref<string | null>(null)
const selectedClassificationFilters = ref<Record<string, string>>({})
const classificationFilterSearch = ref('')
const isClassificationLegendOpen = ref(false)
const systemTypeSourceRows = computed(() => activePage.value === 'systems' ? systemDocumentRows.value : classificationCatalogRows.value)
const systemTypes = computed(() => [{ slug: '', name: 'Все системы', imageUrl: '', position: 0 }, ...parsedSystemTypes.value].map((type) => ({
  ...type,
  count: systemTypeSourceRows.value.filter((system) =>
    matchesConstructionType(system) && matchesSystemType(system, type),
  ).length,
})))
const visibleSystemTypes = computed(() => systemTypes.value.filter((type) => !type.slug || type.count > 0))
const selectedSystemType = computed(() =>
  systemTypes.value.find((type) => type.slug === selectedSystemTypeSlug.value) ?? systemTypes.value[0],
)
const filteredSystemDocumentRows = computed(() => systemDocumentRows.value.filter((system) =>
  matchesConstructionType(system) && matchesSystemType(system, selectedSystemType.value),
))
const allVisibleSystemsSelected = computed(() =>
  filteredSystemDocumentRows.value.length > 0 && filteredSystemDocumentRows.value.every((row) => row.comparisonSelected),
)
const someVisibleSystemsSelected = computed(() =>
  filteredSystemDocumentRows.value.some((row) => row.comparisonSelected) && !allVisibleSystemsSelected.value,
)
const classificationBaseSystems = computed(() => classificationCatalogRows.value.filter((system) =>
  matchesConstructionType(system) && matchesSystemType(system, selectedSystemType.value),
))
const classificationFilterGroups = computed(() => {
  const names = new Set<string>()
  for (const system of classificationBaseSystems.value) {
    for (const characteristic of system.characteristics ?? []) {
      if (characteristic.name !== 'Тип системы' && characteristic.name !== 'Сегмент строительства' && characteristic.value) {
        names.add(characteristic.name)
      }
    }
  }
  return [...names]
})
const visibleClassificationFilterGroups = computed(() => {
  const query = classificationFilterSearch.value.trim().toLocaleLowerCase('ru-RU')
  const groups = query
    ? classificationFilterGroups.value.filter((name) => name.toLocaleLowerCase('ru-RU').includes(query))
    : classificationFilterGroups.value

  return [...groups].sort((left, right) => {
    const leftSelected = Boolean(selectedClassificationFilters.value[left])
    const rightSelected = Boolean(selectedClassificationFilters.value[right])
    return Number(rightSelected) - Number(leftSelected)
  })
})
const selectedClassificationFilterCount = computed(() => Object.keys(selectedClassificationFilters.value).length)
const activeClassificationPageFilterCount = computed(() => [
  classificationCatalogSearchInput.value.trim() !== '',
  selectedConstructionType.value !== 'Все',
  selectedSystemTypeSlug.value !== '',
].filter(Boolean).length)
const hasActiveClassificationPageFilters = computed(() => activeClassificationPageFilterCount.value > 0)
const classificationClassPriority: Record<string, number> = {
  'Рекомендованная': 0,
  'Разрешенная': 1,
  'Запрещенная': 2,
}
const classificationSystems = computed(() => {
  const query = classificationCatalogSearch.value.trim().toLocaleLowerCase('ru-RU')
  const selectedFilters = Object.entries(selectedClassificationFilters.value)
  const systems = classificationBaseSystems.value.filter((system) => {
    const matchesSearch = !query ||
      system.systemName.toLocaleLowerCase('ru-RU').includes(query) ||
      system.code.toLocaleLowerCase('ru-RU').includes(query)
    const matchesFilters = selectedFilters.every(([name, value]) =>
      system.characteristics?.some((characteristic) => characteristic.name === name && characteristic.value === value),
    )
    return matchesSearch && matchesFilters
  })
  return [...systems].sort((left, right) => {
    const classDifference =
      (classificationClassPriority[left.systemClass] ?? Number.MAX_SAFE_INTEGER) -
      (classificationClassPriority[right.systemClass] ?? Number.MAX_SAFE_INTEGER)

    if (classDifference !== 0) return classDifference
    return left.systemName.localeCompare(right.systemName, 'ru')
  })
})
const classificationEmptyMessage = computed(() => {
  if (classificationCatalogRows.value.length === 0) {
    return 'Импортируйте таблицу 2 для выбранного распоряжения'
  }

  if (classificationCatalogSearch.value) {
    return 'Системы не найдены'
  }

  return 'Системы не найдены по выбранным фильтрам'
})
const paginatedClassificationSystems = computed(() => {
  if (classificationCatalogPageSize.value === 'all') {
    return classificationSystems.value
  }
  const pageSize = Number(classificationCatalogPageSize.value)
  const start = (classificationCatalogPage.value - 1) * pageSize
  return classificationSystems.value.slice(start, start + pageSize)
})
const classificationSystemRows = computed(() => {
  const rows: SystemCatalogRow[][] = []
  const columns = classificationView.value === 'list' ? 1 : classificationCardColumns.value
  for (let index = 0; index < paginatedClassificationSystems.value.length; index += columns) {
    rows.push(paginatedClassificationSystems.value.slice(index, index + columns))
  }
  return rows
})
const openedClassificationSystem = computed(() =>
  classificationSystems.value.find((system) => system.id === openedClassificationSystemId.value) ?? null,
)
const systemCatalogStats = ref<SystemCatalogStats>({
  total: 0,
  recommended: 0,
  allowed: 0,
  forbidden: 0,
  curators: 0,
})
const systemCatalogClassOptions = ref(['Все', ...classOptions])
const systemCatalogCuratorOptions = ref(['Все кураторы', ...curatorOptions.slice(1)])
const systemCatalogSearch = ref('')
const selectedSystemCatalogClass = ref('Все')
const selectedSystemCatalogCurator = ref('Все кураторы')
const isSystemCatalogLoading = ref(true)
const isSystemDocumentLoading = ref(true)
const systemFilterRequestCount = ref(0)
const isSystemFiltering = computed(() => systemFilterRequestCount.value > 0)
const isSystemsRefreshing = ref(false)
const isSystemsRefreshDone = ref(false)
const systemsLastRefreshedAt = ref('')
const systemCatalogError = ref('')
const activeSystemFilterCount = computed(() => [
  systemCatalogSearch.value.trim() !== '',
  selectedSystemCatalogClass.value !== 'Все',
  selectedSystemCatalogCurator.value !== 'Все кураторы',
  selectedConstructionType.value !== 'Все',
  selectedSystemTypeSlug.value !== '',
].filter(Boolean).length)
const hasActiveSystemFilters = computed(() => activeSystemFilterCount.value > 0)
const {
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
} = useNavParser({
  onCatalogsInvalidated: () => {
    selectedClassificationFilters.value = {}
    return Promise.all([
      loadSystemCatalog(),
      loadClassificationCatalog(),
      loadSystemDocuments(),
      loadDocumentTable(),
    ]).then(() => undefined)
  },
})

function selectedOrderName() {
  return orders.value.find((order) => order.id === selectedOrderId.value)?.name ?? 'Распоряжение не выбрано'
}

function comparisonOrderName(orderId: number) {
  return orders.value.find((order) => order.id === orderId)?.name ?? 'Распоряжение удалено'
}

function formatOrderDateTime(value: string) {
  if (!value) {
    return '—'
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

async function refreshChangesPage() {
  if (isChangesRefreshing.value) {
    return
  }
  const startedAt = performance.now()
  isChangesRefreshDone.value = false
  isChangesRefreshing.value = true
  try {
    await loadOrders()
    await loadClassificationChanges()
    changesLastRefreshedAt.value = new Date().toISOString()
  } finally {
    const remainingAnimationTime = Math.max(0, 1000 - (performance.now() - startedAt))
    if (remainingAnimationTime > 0) {
      await new Promise((resolve) => window.setTimeout(resolve, remainingAnimationTime))
    }
    isChangesRefreshing.value = false
    if (!ordersError.value && !classificationError.value) {
      isChangesRefreshDone.value = true
      window.setTimeout(() => { isChangesRefreshDone.value = false }, 1600)
    }
  }
}

const visibleOrders = computed(() => {
  if (ordersPageSize.value === 'all') return orders.value
  const pageSize = Number(ordersPageSize.value)
  const start = (ordersPage.value - 1) * pageSize
  return orders.value.slice(start, start + pageSize)
})

function ordersPageCount() {
  if (ordersPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(orders.value.length / Number(ordersPageSize.value)))
}

function ordersRangeStart() {
  if (orders.value.length === 0) return 0
  if (ordersPageSize.value === 'all') return 1
  return (ordersPage.value - 1) * Number(ordersPageSize.value) + 1
}

function ordersRangeEnd() {
  if (ordersPageSize.value === 'all') return orders.value.length
  return Math.min(ordersPage.value * Number(ordersPageSize.value), orders.value.length)
}

function changeOrdersPageSize() {
  ordersPage.value = 1
  settingsOrderMenuId.value = null
}

async function changeOrdersPage(nextPage: number) {
  ordersPage.value = Math.min(Math.max(nextPage, 1), ordersPageCount())
  settingsOrderMenuId.value = null
  await nextTick()
  document.querySelector<HTMLElement>('.orders-settings')?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
}

async function refreshSystemsPage() {
  if (isSystemsRefreshing.value) {
    return
  }
  const startedAt = performance.now()
  isSystemsRefreshDone.value = false
  isSystemsRefreshing.value = true
  try {
    await Promise.all([loadSystemCatalog(true), loadSystemDocuments(true)])
    if (!systemCatalogError.value) {
      systemsLastRefreshedAt.value = new Date().toISOString()
    }
  } finally {
    const remainingAnimationTime = Math.max(0, 1000 - (performance.now() - startedAt))
    if (remainingAnimationTime > 0) {
      await new Promise((resolve) => window.setTimeout(resolve, remainingAnimationTime))
    }
    isSystemsRefreshing.value = false
    if (!systemCatalogError.value) {
      isSystemsRefreshDone.value = true
      window.setTimeout(() => { isSystemsRefreshDone.value = false }, 1600)
    }
  }
}

async function filterSystemsByClass(value: string) {
  selectedSystemCatalogClass.value = selectedSystemCatalogClass.value === value ? 'Все' : value
  openedSelect.value = null
  await loadSystemDocuments(true)
}

async function resetSystemFilters() {
  systemCatalogSearch.value = ''
  selectedSystemCatalogClass.value = 'Все'
  selectedSystemCatalogCurator.value = 'Все кураторы'
  selectedConstructionType.value = 'Все'
  selectedSystemTypeSlug.value = ''
  systemDocumentPage.value = 1
  openedSelect.value = null
  await loadSystemDocuments(true)
}

function filterChangesByClass(value: string) {
  selectedAfterFilter.value = value
  openedSelect.value = null
  classificationPage.value = 1
  showClassificationFilterFeedback()
}

function resetChangesFilters() {
  selectedConstructionType.value = 'Все'
  selectedBeforeFilter.value = 'Все'
  selectedAfterFilter.value = 'Все'
  classificationPage.value = 1
  openedSelect.value = null
  showClassificationFilterFeedback()
}

function selectBeforeChangeFilter(value: string) {
  selectedBeforeFilter.value = value
  classificationPage.value = 1
  openedSelect.value = null
  showClassificationFilterFeedback()
}

function selectAfterChangeFilter(value: string) {
  selectedAfterFilter.value = value
  classificationPage.value = 1
  openedSelect.value = null
  showClassificationFilterFeedback()
}

function showClassificationFilterFeedback() {
  if (classificationFilterFeedbackTimer) {
    window.clearTimeout(classificationFilterFeedbackTimer)
  }
  isClassificationFiltering.value = true
  classificationFilterFeedbackTimer = window.setTimeout(() => {
    isClassificationFiltering.value = false
    classificationFilterFeedbackTimer = null
  }, 260)
}

async function loadOrders() {
  isOrdersLoading.value = true
  ordersError.value = ''

  try {
    const response = await apiFetch(`/orders`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить распоряжения')
    }

    orders.value = await response.json()
    ordersPage.value = Math.min(ordersPage.value, ordersPageCount())
    if (!selectedOrderId.value && orders.value.length > 0) {
      selectedOrderId.value = orders.value[0].id
    }
    if (comparisonOrderIds.value.length === 0) {
      comparisonOrderIds.value = orders.value.slice(0, 2).map((order) => order.id)
    } else {
      comparisonOrderIds.value = comparisonOrderIds.value.filter((id) => orders.value.some((order) => order.id === id))
    }
    await loadComparisonCatalogs()
  } catch (error) {
    ordersError.value = error instanceof Error ? error.message : 'Не удалось загрузить распоряжения'
    isComparisonLoading.value = false
  } finally {
    isOrdersLoading.value = false
  }
}

async function selectOrder(order: Order) {
  selectedOrderId.value = order.id
  isSettingsClassificationUnlocked.value = false
  isSettingsSystemCatalogUnlocked.value = false
  settingsSystemCatalogPage.value = 1
  changesLastRefreshedAt.value = ''
  systemsLastRefreshedAt.value = ''
  classificationCatalogPage.value = 1
  settingsClassificationPage.value = 1
  openedSelect.value = null
  classificationCatalogSearch.value = ''
  classificationCatalogSearchInput.value = ''
  isClassificationSearchPending.value = false
  systemCatalogSearch.value = ''
  selectedSystemTypeSlug.value = ''
  clearClassificationFilters()
  await Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable()])
}

function comparisonRowKey(row: Pick<SystemDocumentRow, 'code' | 'systemName'>) {
  const code = row.code.trim()
  if (code) {
    return `code:${code.toLowerCase()}`
  }

  return `name:${row.systemName.trim().toLowerCase()}`
}

async function loadComparisonCatalog(orderId: number) {
  comparisonError.value = ''

  const query = new URLSearchParams({ orderId: String(orderId), comparison: 'true' })
  const response = await apiFetch(`/system-documents?${query.toString()}`)
  if (!response.ok) {
    throw new Error('Не удалось загрузить данные сравнения')
  }

  const payload: SystemDocumentResponse = await response.json()
  comparisonCatalogByOrder.value = {
    ...comparisonCatalogByOrder.value,
    [orderId]: payload.rows,
  }
}

async function loadComparisonCatalogs() {
  isComparisonLoading.value = true
  comparisonError.value = ''

  try {
    await Promise.all(comparisonOrderIds.value.map((orderId) => loadComparisonCatalog(orderId)))
  } catch (error) {
    comparisonError.value = error instanceof Error ? error.message : 'Не удалось загрузить данные сравнения'
  } finally {
    isComparisonLoading.value = false
  }
}

async function loadComparisonOrder(orderId: number) {
  isComparisonLoading.value = true
  comparisonError.value = ''

  try {
    await loadComparisonCatalog(orderId)
  } catch (error) {
    comparisonError.value = error instanceof Error ? error.message : 'Не удалось загрузить данные сравнения'
  } finally {
    isComparisonLoading.value = false
  }
}

async function createOrder() {
  const name = window.prompt('Название распоряжения')
  if (!name?.trim()) {
    return
  }

  const response = await apiFetch(`/orders`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: name.trim() }),
  })
  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    ordersError.value = payload?.error ?? 'Не удалось создать распоряжение'
    return
  }

  const order: Order = await response.json()
  orders.value = [order, ...orders.value]
  selectedOrderId.value = order.id
  await Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable()])
}

function openOrderWorkbookImport() {
  orderWorkbookInput.value?.click()
}

async function importOrderWorkbook(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }

  const formData = new FormData()
  formData.append('file', file)
  isOrderWorkbookImporting.value = true
  ordersError.value = ''

  try {
    const response = await apiFetch(`/orders/import`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось импортировать распоряжение')
    }

    const order: Order = await response.json()
    selectedOrderId.value = order.id
    await loadOrders()
    await selectOrder(orders.value.find((item) => item.id === order.id) ?? order)
  } catch (error) {
    ordersError.value = error instanceof Error ? error.message : 'Не удалось импортировать распоряжение'
  } finally {
    isOrderWorkbookImporting.value = false
    input.value = ''
  }
}

async function deleteOrder(order: Order) {
  if (!window.confirm(`Удалить ${order.name}? Данные таблиц этого распоряжения тоже будут удалены.`)) {
    return
  }

  const response = await apiFetch(`/orders/${order.id}`, { method: 'DELETE' })
  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    ordersError.value = payload?.error ?? 'Не удалось удалить распоряжение'
    return
  }

  orders.value = orders.value.filter((item) => item.id !== order.id)
  ordersPage.value = Math.min(ordersPage.value, ordersPageCount())
  if (selectedOrderId.value === order.id) {
    selectedOrderId.value = orders.value[0]?.id ?? null
  }
  comparisonOrderIds.value = comparisonOrderIds.value.filter((id) => id !== order.id)
  const nextCatalogs = { ...comparisonCatalogByOrder.value }
  delete nextCatalogs[order.id]
  comparisonCatalogByOrder.value = nextCatalogs
  await Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable()])
}

function scheduleOrderRename(order: Order) {
  const previousTimer = orderRenameTimers.get(order.id)
  if (previousTimer) {
    window.clearTimeout(previousTimer)
  }

  orderRenameTimers.set(
    order.id,
    window.setTimeout(() => {
      void saveOrderName(order)
    }, 350),
  )
}

async function saveOrderName(order: Order) {
  const previousTimer = orderRenameTimers.get(order.id)
  if (previousTimer) {
    window.clearTimeout(previousTimer)
    orderRenameTimers.delete(order.id)
  }

  const nextName = order.name.trim()
  if (!nextName) {
    ordersError.value = 'Название распоряжения не может быть пустым'
    await loadOrders()
    return
  }

  orderRenameControllers.get(order.id)?.abort()
  const controller = new AbortController()
  orderRenameControllers.set(order.id, controller)

  try {
    const response = await apiFetch(`/orders/${order.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: nextName }),
      signal: controller.signal,
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось сохранить название распоряжения')
    }

    const updatedOrder: Order = await response.json()
    const targetOrder = orders.value.find((item) => item.id === updatedOrder.id)
    if (targetOrder) {
      Object.assign(targetOrder, updatedOrder)
    }
    ordersError.value = ''
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    ordersError.value = error instanceof Error ? error.message : 'Не удалось сохранить название распоряжения'
  } finally {
    if (orderRenameControllers.get(order.id) === controller) {
      orderRenameControllers.delete(order.id)
    }
  }
}

async function selectComparisonOrder(index: number, order: Order) {
  comparisonOrderIds.value[index] = order.id
  openedSelect.value = null
  await loadComparisonOrder(order.id)
}

function comparisonOrderOptions(currentOrderId: number) {
  return orders.value.filter((order) => order.id === currentOrderId || !comparisonOrderIds.value.includes(order.id))
}

function availableComparisonOrders() {
  return orders.value.filter((order) => !comparisonOrderIds.value.includes(order.id))
}

async function addComparisonOrder(order?: Order) {
  if (comparisonOrderIds.value.length >= MAX_COMPARISON_ORDERS) {
    openedSelect.value = null
    return
  }

  const nextOrder = order ?? availableComparisonOrders()[0]
  if (nextOrder) {
    comparisonOrderIds.value.push(nextOrder.id)
    await loadComparisonOrder(nextOrder.id)
  }
  openedSelect.value = null
}

function removeComparisonOrder(orderId: number) {
  comparisonOrderIds.value = comparisonOrderIds.value.filter((id) => id !== orderId)
  openedSelect.value = null
}

function startComparisonOrderDrag(event: DragEvent, orderId: number) {
  const dragHandle = event.currentTarget as HTMLElement | null
  const orderCard = dragHandle?.closest<HTMLElement>('.comparison-order')
  draggedComparisonOrderId.value = orderId
  comparisonDropIndex.value = comparisonOrderIds.value.indexOf(orderId)
  openedSelect.value = null

  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', String(orderId))
    if (orderCard) {
      const bounds = orderCard.getBoundingClientRect()
      event.dataTransfer.setDragImage(
        orderCard,
        Math.max(0, Math.min(event.clientX - bounds.left, bounds.width)),
        Math.max(0, Math.min(event.clientY - bounds.top, bounds.height)),
      )
    }
  }
}

function enterComparisonOrderDrop(index: number) {
  if (draggedComparisonOrderId.value !== null) {
    comparisonDropIndex.value = index
  }
}

function dropComparisonOrder(targetIndex: number) {
  const orderId = draggedComparisonOrderId.value
  if (orderId === null) {
    return
  }

  const sourceIndex = comparisonOrderIds.value.indexOf(orderId)
  if (sourceIndex !== -1 && sourceIndex !== targetIndex) {
    const reorderedIds = [...comparisonOrderIds.value]
    reorderedIds.splice(sourceIndex, 1)
    reorderedIds.splice(targetIndex, 0, orderId)
    comparisonOrderIds.value = reorderedIds
  }

  endComparisonOrderDrag()
}

function endComparisonOrderDrag() {
  draggedComparisonOrderId.value = null
  comparisonDropIndex.value = null
}

function comparisonRows() {
  const orderedKeys: string[] = []
  const namesByKey = new Map<string, string>()
  const valuesByOrder = new Map<number, Map<string, string>>()

  for (const orderId of comparisonOrderIds.value) {
    const rows = comparisonCatalogByOrder.value[orderId] ?? []
    const values = new Map<string, string>()

    for (const row of rows) {
      const key = comparisonRowKey(row)
      values.set(key, row.systemClass || 'н/д')

      if (!namesByKey.has(key)) {
        namesByKey.set(key, row.systemName)
        orderedKeys.push(key)
      }
    }

    valuesByOrder.set(orderId, values)
  }

  const completeKeys = orderedKeys.filter((key) =>
    comparisonOrderIds.value.every((orderId) => valuesByOrder.get(orderId)?.has(key)),
  )
  const partialKeys = orderedKeys.filter((key) =>
    !comparisonOrderIds.value.every((orderId) => valuesByOrder.get(orderId)?.has(key)),
  )

  const rows = [...completeKeys, ...partialKeys]
    .filter((key) => !hiddenComparisonRows.value.includes(key))
    .map((key) => ({
      key,
      name: namesByKey.get(key) ?? key,
      values: comparisonOrderIds.value.map((orderId) => valuesByOrder.get(orderId)?.get(key) ?? 'н/д'),
    }))

  return rows
    .filter((row) => !comparisonOnlyDifferences.value || new Set(row.values).size > 1)
    .sort((left, right) => {
      if (comparisonSort.value === 'differences-first') {
        const differenceOrder = Number(new Set(right.values).size > 1) - Number(new Set(left.values).size > 1)
        if (differenceOrder !== 0) return differenceOrder
      }
      const nameOrder = normalizedComparisonName(left.name).localeCompare(normalizedComparisonName(right.name), 'ru', { numeric: true })
      return comparisonSort.value === 'name-desc' ? -nameOrder : nameOrder
    })
}

function visibleComparisonRows() {
  const rows = comparisonRows()
  if (comparisonPageSize.value === 'all') return rows
  const pageSize = Number(comparisonPageSize.value)
  const start = (comparisonPage.value - 1) * pageSize
  return rows.slice(start, start + pageSize)
}

function comparisonPageCount() {
  if (comparisonPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(comparisonRows().length / Number(comparisonPageSize.value)))
}

function comparisonRangeStart() {
  if (comparisonRows().length === 0) return 0
  if (comparisonPageSize.value === 'all') return 1
  return (comparisonPage.value - 1) * Number(comparisonPageSize.value) + 1
}

function comparisonRangeEnd() {
  if (comparisonPageSize.value === 'all') return comparisonRows().length
  return Math.min(comparisonPage.value * Number(comparisonPageSize.value), comparisonRows().length)
}

function changeComparisonPageSize() {
  comparisonPage.value = 1
}

async function changeComparisonPage(nextPage: number) {
  comparisonPage.value = Math.min(Math.max(nextPage, 1), comparisonPageCount())
  await nextTick()
  document.querySelector<HTMLElement>('.comparison-table')?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
}

function normalizedComparisonName(value: string) {
  return value
    .trim()
    .replace(/^ремонтная\s+система\s+/i, '')
    .replace(/^[а-яёa-z]{1,5}\s*[-–—:]\s*/i, '')
}

function comparisonSortLabel() {
  switch (comparisonSort.value) {
    case 'name-asc': return 'Без префикса (А–Я)'
    case 'name-desc': return 'Без префикса (Я–А)'
    default: return 'Различия сначала'
  }
}

function selectComparisonSort(value: 'differences-first' | 'name-asc' | 'name-desc') {
  comparisonSort.value = value
  comparisonPage.value = 1
  openedSelect.value = null
}

function selectComparisonDifferenceFilter(onlyDifferences: boolean) {
  comparisonOnlyDifferences.value = onlyDifferences
  comparisonPage.value = 1
  openedSelect.value = null
}

function hideComparisonRow(row: ComparisonRow) {
  if (!hiddenComparisonRows.value.includes(row.key)) {
    hiddenComparisonRows.value = [...hiddenComparisonRows.value, row.key]
    comparisonPage.value = Math.min(comparisonPage.value, comparisonPageCount())
  }
}

function comparisonValue(row: ComparisonRow, index: number) {
  return row.values[index] ?? 'н/д'
}

function comparisonRowHasDifference(row: ComparisonRow) {
  return new Set(row.values).size > 1
}

async function exportComparisonTable() {
  const headers = ['Название системы', ...comparisonOrderIds.value.map(comparisonOrderName)]
  const rows = comparisonRows().map((row) => [row.name, ...row.values])

  const response = await apiFetch(`/comparison/export`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ headers, rows }),
  })
  if (!response.ok) {
    comparisonError.value = 'Не удалось экспортировать сравнение'
    return
  }

  const url = URL.createObjectURL(await response.blob())
  const link = document.createElement('a')
  link.href = url
  link.download = 'comparison.xlsx'
  link.click()
  URL.revokeObjectURL(url)
}

function buildClassificationQuery() {
  const params = new URLSearchParams()
  if (selectedOrderId.value) {
    params.set('orderId', String(selectedOrderId.value))
  }
  if (tableSearch.value.trim()) {
    params.set('q', tableSearch.value.trim())
  }
  if (selectedBeforeFilter.value && selectedBeforeFilter.value !== 'Все') {
    params.set('before', selectedBeforeFilter.value)
  }
  if (selectedAfterFilter.value && selectedAfterFilter.value !== 'Все') {
    params.set('after', selectedAfterFilter.value)
  }

  return params
}

function applyClassificationPayload(payload: ClassificationResponse) {
  classificationRows.value = payload.rows
  classificationPage.value = 1
  classificationStats.value = payload.stats
  beforeOptions.value = payload.beforeOptions.length > 1 ? payload.beforeOptions : beforeOptions.value
  afterOptions.value = payload.afterOptions.length > 1 ? payload.afterOptions : afterOptions.value
}

async function loadClassificationChanges() {
  classificationRequestController?.abort()
  const controller = new AbortController()
  classificationRequestController = controller
  isClassificationLoading.value = true
  classificationLoadingMessage.value = 'Загрузка таблицы...'
  classificationError.value = ''

  try {
    const query = new URLSearchParams()
    if (selectedOrderId.value) {
      query.set('orderId', String(selectedOrderId.value))
    }
    const response = await apiFetch(`/classification-changes?${query.toString()}`, { signal: controller.signal })
    if (!response.ok) {
      throw new Error('Не удалось загрузить таблицу 1')
    }

    applyClassificationPayload(await response.json())
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    classificationError.value = error instanceof Error ? error.message : 'Не удалось загрузить таблицу 1'
  } finally {
    if (classificationRequestController === controller) {
      classificationRequestController = null
      isClassificationLoading.value = false
    }
  }
}

function openTableImport() {
  importFileInput.value?.click()
}

async function importTableFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }

  const formData = new FormData()
  formData.append('file', file)
  if (selectedOrderId.value) {
    formData.append('orderId', String(selectedOrderId.value))
  }
  isClassificationLoading.value = true
  classificationLoadingMessage.value = 'Определяем ссылки и типы систем на Навигаторе...'
  classificationError.value = ''

  try {
    const response = await apiFetch(`/classification-changes/import`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const errorPayload = await response.json().catch(() => null)
      throw new Error(errorPayload?.error ?? 'Не удалось импортировать таблицу')
    }

    selectedBeforeFilter.value = 'Все'
    selectedAfterFilter.value = 'Все'
    tableSearch.value = ''
    applyClassificationPayload(await response.json())
  } catch (error) {
    classificationError.value = error instanceof Error ? error.message : 'Не удалось импортировать таблицу'
  } finally {
    isClassificationLoading.value = false
    input.value = ''
  }
}

async function exportClassificationTable() {
  const query = buildClassificationQuery()
  if (selectedConstructionType.value && selectedConstructionType.value !== 'Все') {
    query.set('constructionType', selectedConstructionType.value)
  }
  const response = await apiFetch(`/classification-changes/export?${query.toString()}`)
  if (!response.ok) {
    classificationError.value = 'Не удалось экспортировать таблицу'
    return
  }

  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'classification-changes.xlsx'
  link.click()
  URL.revokeObjectURL(url)
}

function currentClassificationRows() {
  const query = activePage.value === 'settings' ? tableSearch.value.trim().toLocaleLowerCase('ru-RU') : ''
  return classificationRows.value.filter((row) =>
    (selectedConstructionType.value === 'Все' || row.constructionType === selectedConstructionType.value) &&
    (selectedBeforeFilter.value === 'Все' || row.classBefore === selectedBeforeFilter.value) &&
    (selectedAfterFilter.value === 'Все' || row.classAfter === selectedAfterFilter.value) &&
    (!query || row.systemName.toLocaleLowerCase('ru-RU').includes(query)),
  )
}

function visibleClassificationRows() {
  const rows = currentClassificationRows()
  if (classificationPageSize.value === 'all') return rows
  const pageSize = Number(classificationPageSize.value)
  const start = (classificationPage.value - 1) * pageSize
  return rows.slice(start, start + pageSize)
}

function classificationPageCount() {
  if (classificationPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(currentClassificationRows().length / Number(classificationPageSize.value)))
}

function classificationRangeStart() {
  if (currentClassificationRows().length === 0) return 0
  if (classificationPageSize.value === 'all') return 1
  return (classificationPage.value - 1) * Number(classificationPageSize.value) + 1
}

function classificationRangeEnd() {
  if (classificationPageSize.value === 'all') return currentClassificationRows().length
  return Math.min(classificationPage.value * Number(classificationPageSize.value), currentClassificationRows().length)
}

function changeClassificationPageSize() {
  classificationPage.value = 1
}

function visibleSettingsClassificationRows() {
  const rows = currentSettingsClassificationRows()
  if (settingsClassificationPageSize.value === 'all') return rows
  const pageSize = Number(settingsClassificationPageSize.value)
  const start = (settingsClassificationPage.value - 1) * pageSize
  return rows.slice(start, start + pageSize)
}

function currentSettingsClassificationRows() {
  const query = tableSearch.value.trim().toLocaleLowerCase('ru-RU')
  return classificationRows.value.filter((row) => !query || row.systemName.toLocaleLowerCase('ru-RU').includes(query))
}

function settingsClassificationPageCount() {
  if (settingsClassificationPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(currentSettingsClassificationRows().length / Number(settingsClassificationPageSize.value)))
}

function settingsClassificationRangeStart() {
  if (currentSettingsClassificationRows().length === 0) return 0
  if (settingsClassificationPageSize.value === 'all') return 1
  return (settingsClassificationPage.value - 1) * Number(settingsClassificationPageSize.value) + 1
}

function settingsClassificationRangeEnd() {
  if (settingsClassificationPageSize.value === 'all') return currentSettingsClassificationRows().length
  return Math.min(settingsClassificationPage.value * Number(settingsClassificationPageSize.value), currentSettingsClassificationRows().length)
}

function changeSettingsClassificationPageSize() {
  settingsClassificationPage.value = 1
}

async function changeSettingsClassificationPage(nextPage: number) {
  settingsClassificationPage.value = Math.min(Math.max(nextPage, 1), settingsClassificationPageCount())
  await nextTick()
  document.querySelector<HTMLElement>('.settings-classification-block')?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
}

async function changeClassificationPage(nextPage: number) {
  classificationPage.value = Math.min(Math.max(nextPage, 1), classificationPageCount())
  await nextTick()
  const table = document.querySelector<HTMLElement>('.changes-table-card')
  table?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
  table?.querySelector<HTMLElement>('tbody tr')?.focus({ preventScroll: true })
}

function classificationChangesEmptyMessage() {
  if (classificationRows.value.length === 0) {
    return 'В этом распоряжении пока нет данных таблицы 1'
  }

  return `Для типа «${selectedConstructionType.value}» системы не найдены`
}

function scheduleClassificationRowSave(row: ClassificationChange) {
  const currentTimer = classificationEditTimers.get(row.id)
  if (currentTimer) {
    window.clearTimeout(currentTimer)
  }
  classificationEditTimers.set(row.id, window.setTimeout(() => saveClassificationRow(row), 350))
}

async function saveClassificationRow(row: ClassificationChange) {
  const currentTimer = classificationEditTimers.get(row.id)
  if (currentTimer) {
    window.clearTimeout(currentTimer)
    classificationEditTimers.delete(row.id)
  }
  const submitted = {
    systemName: row.systemName,
    classBefore: row.classBefore,
    classAfter: row.classAfter,
  }
  classificationEditControllers.get(row.id)?.abort()
  const controller = new AbortController()
  classificationEditControllers.set(row.id, controller)
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const response = await apiFetch(`/classification-changes/${row.id}?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(submitted),
      signal: controller.signal,
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось сохранить строку таблицы 1')
    }
    const saved: ClassificationChange = await response.json()
    if (row.systemName === submitted.systemName && row.classBefore === submitted.classBefore && row.classAfter === submitted.classAfter) {
      row.systemName = saved.systemName
      row.classBefore = saved.classBefore
      row.classAfter = saved.classAfter
    }
    classificationError.value = ''
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    classificationError.value = error instanceof Error ? error.message : 'Не удалось сохранить строку таблицы 1'
  } finally {
    if (classificationEditControllers.get(row.id) === controller) {
      classificationEditControllers.delete(row.id)
    }
  }
}

function buildSystemCatalogQuery() {
  const params = new URLSearchParams()
  if (selectedOrderId.value) {
    params.set('orderId', String(selectedOrderId.value))
  }
  if (systemCatalogSearch.value.trim()) {
    params.set('q', systemCatalogSearch.value.trim())
  }
  if (selectedSystemCatalogClass.value && selectedSystemCatalogClass.value !== 'Все') {
    params.set('class', selectedSystemCatalogClass.value)
  }
  if (selectedSystemCatalogCurator.value && selectedSystemCatalogCurator.value !== 'Все кураторы') {
    params.set('curator', selectedSystemCatalogCurator.value)
  }

  return params
}

function applySystemCatalogPayload(payload: SystemCatalogResponse) {
  systemCatalogRows.value = payload.rows
  parsedSystemTypes.value = payload.systemTypes ?? parsedSystemTypes.value
  if (selectedSystemTypeSlug.value && !parsedSystemTypes.value.some((type) => type.slug === selectedSystemTypeSlug.value)) {
    selectedSystemTypeSlug.value = ''
  }
}

async function loadClassificationCatalog() {
  classificationCatalogRequestController?.abort()
  const controller = new AbortController()
  classificationCatalogRequestController = controller
  isClassificationCatalogLoading.value = true
  classificationCatalogError.value = ''

  try {
    const query = new URLSearchParams()
    if (selectedOrderId.value) {
      query.set('orderId', String(selectedOrderId.value))
    }

    const response = await apiFetch(`/system-catalog?${query.toString()}`, { signal: controller.signal })
    if (!response.ok) {
      throw new Error('Не удалось загрузить системы из таблицы 2')
    }

    const payload: SystemCatalogResponse = await response.json()
    classificationCatalogRows.value = payload.rows
    parsedSystemTypes.value = payload.systemTypes ?? parsedSystemTypes.value
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    classificationCatalogError.value = error instanceof Error ? error.message : 'Не удалось загрузить системы из таблицы 2'
  } finally {
    if (classificationCatalogRequestController === controller) {
      classificationCatalogRequestController = null
      isClassificationCatalogLoading.value = false
    }
  }
}

async function loadSystemCatalog(silent = false) {
  systemCatalogRequestController?.abort()
  const controller = new AbortController()
  systemCatalogRequestController = controller
  if (!silent) {
    isSystemCatalogLoading.value = true
  }
  systemCatalogError.value = ''

  try {
    const query = buildSystemCatalogQuery()
    const response = await apiFetch(`/system-catalog?${query.toString()}`, { signal: controller.signal })
    if (!response.ok) {
      throw new Error('Не удалось загрузить таблицу 2')
    }

    const payload: SystemCatalogResponse = await response.json()
    applySystemCatalogPayload(payload)
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось загрузить таблицу 2'
  } finally {
    if (systemCatalogRequestController === controller) {
      systemCatalogRequestController = null
      isSystemCatalogLoading.value = false
    }
  }
}

function openSystemCatalogImport() {
  systemCatalogFileInput.value?.click()
}

async function importSystemCatalogFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }

  const formData = new FormData()
  formData.append('file', file)
  if (selectedOrderId.value) {
    formData.append('orderId', String(selectedOrderId.value))
  }
  isSystemCatalogLoading.value = true
  systemCatalogError.value = ''

  try {
    const response = await apiFetch(`/system-catalog/import`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const errorPayload = await response.json().catch(() => null)
      throw new Error(errorPayload?.error ?? 'Не удалось импортировать таблицу 2')
    }

    selectedSystemCatalogClass.value = 'Все'
    selectedSystemCatalogCurator.value = 'Все кураторы'
    systemCatalogSearch.value = ''
    settingsSystemCatalogPage.value = 1
    const payload: SystemCatalogResponse = await response.json()
    applySystemCatalogPayload(payload)
    classificationCatalogRows.value = payload.rows
    await Promise.all([loadSystemDocuments(), loadDocumentTable()])
    if (selectedOrderId.value && comparisonOrderIds.value.includes(selectedOrderId.value)) {
      await loadComparisonCatalog(selectedOrderId.value)
    }
  } catch (error) {
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось импортировать таблицу 2'
  } finally {
    isSystemCatalogLoading.value = false
    input.value = ''
  }
}

async function exportSystemCatalog() {
  const query = buildSystemCatalogQuery()
  if (selectedConstructionType.value !== 'Все') {
    query.set('constructionType', selectedConstructionType.value)
  }
  if (selectedSystemTypeSlug.value) {
    query.set('systemType', selectedSystemType.value.name)
  }
  const response = await apiFetch(`/system-documents/export?${query.toString()}`)
  if (!response.ok) {
    systemCatalogError.value = 'Не удалось экспортировать таблицу 2'
    return
  }

  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'system-documents.xlsx'
  link.click()
  URL.revokeObjectURL(url)
}

function currentSettingsSystemCatalogRows() {
  const query = systemCatalogSearch.value.trim().toLocaleLowerCase('ru-RU')
  return systemCatalogRows.value.filter((row) => !query || [row.code, row.systemName, row.curator]
    .some((value) => value.toLocaleLowerCase('ru-RU').includes(query)))
}

function visibleSettingsSystemCatalogRows() {
  const rows = currentSettingsSystemCatalogRows()
  if (settingsSystemCatalogPageSize.value === 'all') return rows
  const pageSize = Number(settingsSystemCatalogPageSize.value)
  const start = (settingsSystemCatalogPage.value - 1) * pageSize
  return rows.slice(start, start + pageSize)
}

function settingsSystemCatalogPageCount() {
  if (settingsSystemCatalogPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(currentSettingsSystemCatalogRows().length / Number(settingsSystemCatalogPageSize.value)))
}

function settingsSystemCatalogRangeStart() {
  if (currentSettingsSystemCatalogRows().length === 0) return 0
  if (settingsSystemCatalogPageSize.value === 'all') return 1
  return (settingsSystemCatalogPage.value - 1) * Number(settingsSystemCatalogPageSize.value) + 1
}

function settingsSystemCatalogRangeEnd() {
  if (settingsSystemCatalogPageSize.value === 'all') return currentSettingsSystemCatalogRows().length
  return Math.min(settingsSystemCatalogPage.value * Number(settingsSystemCatalogPageSize.value), currentSettingsSystemCatalogRows().length)
}

function changeSettingsSystemCatalogPageSize() {
  settingsSystemCatalogPage.value = 1
}

async function changeSettingsSystemCatalogPage(nextPage: number) {
  settingsSystemCatalogPage.value = Math.min(Math.max(nextPage, 1), settingsSystemCatalogPageCount())
  await nextTick()
  document.querySelector<HTMLElement>('.settings-system-catalog-block')?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
}

function scheduleSystemCatalogRowSave(row: SystemCatalogRow) {
  const currentTimer = systemCatalogEditTimers.get(row.id)
  if (currentTimer) {
    window.clearTimeout(currentTimer)
  }
  systemCatalogEditTimers.set(row.id, window.setTimeout(() => saveSystemCatalogRow(row), 350))
}

function syncSystemCatalogRow(saved: SystemCatalogRow) {
  for (const collection of [systemCatalogRows.value, classificationCatalogRows.value]) {
    const target = collection.find((item) => item.id === saved.id)
    if (target) {
      Object.assign(target, saved, { characteristics: target.characteristics })
    }
  }
  for (const collection of [systemDocumentRows.value, documentRows.value]) {
    const target = collection.find((item) => item.systemCatalogId === saved.id)
    if (target) {
      target.code = saved.code
      target.systemName = saved.systemName
      target.systemUrl = saved.systemUrl
      target.systemClass = saved.systemClass
      target.curator = saved.curator
    }
  }
}

async function saveSystemCatalogRow(row: SystemCatalogRow) {
  const currentTimer = systemCatalogEditTimers.get(row.id)
  if (currentTimer) {
    window.clearTimeout(currentTimer)
    systemCatalogEditTimers.delete(row.id)
  }
  const submitted = {
    code: row.code,
    systemName: row.systemName,
    systemClass: row.systemClass,
    curator: row.curator,
  }
  systemCatalogEditControllers.get(row.id)?.abort()
  const controller = new AbortController()
  systemCatalogEditControllers.set(row.id, controller)
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const response = await apiFetch(`/system-catalog/${row.id}?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(submitted),
      signal: controller.signal,
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось сохранить строку таблицы 2')
    }
    const saved: SystemCatalogRow = await response.json()
    if (row.code === submitted.code && row.systemName === submitted.systemName && row.systemClass === submitted.systemClass && row.curator === submitted.curator) {
      syncSystemCatalogRow(saved)
    }
    systemCatalogError.value = ''
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось сохранить строку таблицы 2'
  } finally {
    if (systemCatalogEditControllers.get(row.id) === controller) {
      systemCatalogEditControllers.delete(row.id)
    }
  }
}

function currentSystemDocumentRows() {
  return filteredSystemDocumentRows.value
}

function visibleSystemDocumentRows() {
  const rows = currentSystemDocumentRows()
  if (systemDocumentPageSize.value === 'all') {
    return rows
  }
  const pageSize = Number(systemDocumentPageSize.value)
  const start = (systemDocumentPage.value - 1) * pageSize
  return rows.slice(start, start + pageSize)
}

function systemDocumentPageCount() {
  if (systemDocumentPageSize.value === 'all') {
    return 1
  }
  return Math.max(1, Math.ceil(currentSystemDocumentRows().length / Number(systemDocumentPageSize.value)))
}

function systemDocumentRangeStart() {
  if (currentSystemDocumentRows().length === 0) return 0
  if (systemDocumentPageSize.value === 'all') return 1
  return (systemDocumentPage.value - 1) * Number(systemDocumentPageSize.value) + 1
}

function systemDocumentRangeEnd() {
  if (systemDocumentPageSize.value === 'all') return currentSystemDocumentRows().length
  return Math.min(systemDocumentPage.value * Number(systemDocumentPageSize.value), currentSystemDocumentRows().length)
}

function changeSystemDocumentPageSize() {
  systemDocumentPage.value = 1
}

async function changeSystemDocumentPage(nextPage: number) {
  systemDocumentPage.value = Math.min(Math.max(nextPage, 1), systemDocumentPageCount())
  await nextTick()

  const table = document.querySelector<HTMLElement>('.systems-table--catalog')
  table?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
  table?.querySelector<HTMLElement>('tbody tr')?.focus({ preventScroll: true })
}

function setPage(page: string, updateLocation = true) {
  if (!pageKeys.has(page)) return
  activePage.value = page
  if (updateLocation) {
    const nextHash = `#/${page}`
    if (window.location.hash !== nextHash) {
      window.location.hash = nextHash
    }
  }
  openedSelect.value = null
  if (page === 'changes') {
    void loadClassificationChanges()
  } else if (page === 'systems') {
    void loadSystemDocuments()
  } else if (page === 'classification') {
    void loadClassificationCatalog()
  } else if (page === 'settings') {
    void Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadDocumentTable(), loadNavParserSettings(), loadNavParserProgress(), loadNavParserRuns()])
  }
}

function syncPageFromLocation() {
  const page = pageFromLocation()
  if (page !== activePage.value) {
    setPage(page, false)
  }
}

function toggleSelect(name: string) {
  openedSelect.value = openedSelect.value === name ? null : name
}

function systemMatchesConstructionType(system: { characteristics?: SystemCharacteristic[] }, constructionType: string) {
  return constructionType === 'Все' ||
    system.characteristics?.some((characteristic) =>
      characteristic.name === 'Сегмент строительства' && characteristic.value.includes(constructionType),
    )
}

function matchesConstructionType(system: { characteristics?: SystemCharacteristic[] }) {
  return systemMatchesConstructionType(system, selectedConstructionType.value)
}

function matchesSystemType(system: { characteristics?: SystemCharacteristic[] }, type: SystemTypeOption) {
  return type.slug === '' || system.characteristics?.some((characteristic) =>
    characteristic.name === 'Тип системы' && characteristic.value === type.name,
  )
}

function selectSystemType(type: SystemTypeOption) {
  selectedSystemTypeSlug.value = type.slug
  systemDocumentPage.value = 1
  classificationCatalogPage.value = 1
  openedClassificationSystemId.value = null
  clearClassificationFilters()
}

function hideBrokenSystemTypeImage(event: Event) {
  const image = event.currentTarget as HTMLImageElement
  image.hidden = true
}

function systemTypeImageSource(type: SystemTypeOption) {
  return type.imageUrl && type.slug
    ? apiURL(`/nav-system-types/${encodeURIComponent(type.slug)}/image`)
    : ''
}

function selectConstructionType(type: string) {
  selectedConstructionType.value = type
  classificationPage.value = 1
  systemDocumentPage.value = 1
  classificationCatalogPage.value = 1
  selectedSystemTypeSlug.value = ''
  openedClassificationSystemId.value = null
  clearClassificationFilters()
  if (activePage.value === 'changes') {
    showClassificationFilterFeedback()
  }
}

function classificationFilterOptions(name: string) {
  const values = new Set<string>()
  for (const system of classificationBaseSystems.value) {
    if (!matchesSelectedClassificationFilters(system, name)) {
      continue
    }
    for (const characteristic of system.characteristics ?? []) {
      if (characteristic.name === name && characteristic.value) {
        values.add(characteristic.value)
      }
    }
  }
  return [...values].sort((left, right) => left.localeCompare(right, 'ru-RU'))
}

function classificationFilterOptionCount(name: string, value: string) {
  return classificationBaseSystems.value.filter((system) =>
    matchesSelectedClassificationFilters(system, name) &&
    system.characteristics?.some((characteristic) => characteristic.name === name && characteristic.value === value),
  ).length
}

function classificationFilterAvailableCount(name: string) {
  return classificationBaseSystems.value.filter((system) => matchesSelectedClassificationFilters(system, name)).length
}

function matchesSelectedClassificationFilters(system: SystemCatalogRow, excludedName = '') {
  return Object.entries(selectedClassificationFilters.value).every(([name, value]) =>
    name === excludedName || system.characteristics?.some((characteristic) => characteristic.name === name && characteristic.value === value),
  )
}

function toggleClassificationFilter(name: string) {
  openedClassificationFilter.value = openedClassificationFilter.value === name ? null : name
}

function selectClassificationFilter(name: string, value: string) {
  const next = { ...selectedClassificationFilters.value }
  if (value) {
    next[name] = value
  } else {
    delete next[name]
  }
  selectedClassificationFilters.value = next
  classificationCatalogPage.value = 1
  openedClassificationFilter.value = null
  openedClassificationSystemId.value = null
}

function clearClassificationFilters() {
  selectedClassificationFilters.value = {}
  classificationCatalogPage.value = 1
  openedClassificationFilter.value = null
  openedClassificationSystemId.value = null
}

function resetClassificationPageFilters() {
  if (classificationSearchTimer) {
    window.clearTimeout(classificationSearchTimer)
    classificationSearchTimer = null
  }
  classificationCatalogSearchInput.value = ''
  classificationCatalogSearch.value = ''
  classificationFilterSearch.value = ''
  isClassificationSearchPending.value = false
  selectedConstructionType.value = 'Все'
  selectedSystemTypeSlug.value = ''
  classificationCatalogPage.value = 1
  openedClassificationSystemId.value = null
}

function collapseClassificationFilters() {
  openedClassificationFilter.value = null
}

function prepareClassificationFilterEnter(element: Element) {
  const shell = element as HTMLElement
  shell.style.height = '0px'
  shell.style.opacity = '0'
}

function animateClassificationFilterEnter(element: Element) {
  const shell = element as HTMLElement
  window.requestAnimationFrame(() => {
    shell.style.height = `${shell.scrollHeight}px`
    shell.style.opacity = '1'
  })
}

function finishClassificationFilterMotion(element: Element) {
  const shell = element as HTMLElement
  shell.style.height = ''
  shell.style.opacity = ''
}

function prepareClassificationFilterLeave(element: Element) {
  const shell = element as HTMLElement
  shell.style.height = `${shell.scrollHeight}px`
  shell.style.opacity = '1'
}

function animateClassificationFilterLeave(element: Element) {
  const shell = element as HTMLElement
  void shell.offsetHeight
  window.requestAnimationFrame(() => {
    shell.style.height = '0px'
    shell.style.opacity = '0'
  })
}

function classificationRowPositions() {
  return [...document.querySelectorAll<HTMLElement>('.classification-card-row')].map((element) => ({
    element,
    top: element.getBoundingClientRect().top,
  }))
}

function applySystemDocumentPayload(payload: SystemDocumentResponse) {
  systemDocumentRows.value = payload.rows
  systemDocumentPage.value = 1
  systemCatalogStats.value = payload.stats
  systemCatalogClassOptions.value = payload.classOptions.length > 1 ? payload.classOptions : ['Все', ...classOptions]
  systemCatalogCuratorOptions.value = payload.curatorOptions.length > 1
    ? ['Все кураторы', ...payload.curatorOptions.filter((option) => option !== 'Все')]
    : ['Все кураторы']
}

async function loadSystemDocuments(silent = false) {
  systemDocumentRequestController?.abort()
  const controller = new AbortController()
  const tracksFilterRequest = silent
  systemDocumentRequestController = controller
  if (!silent) {
    isSystemDocumentLoading.value = true
  } else {
    systemFilterRequestCount.value += 1
  }
  systemCatalogError.value = ''
  try {
    const query = buildSystemCatalogQuery()
    const response = await apiFetch(`/system-documents?${query.toString()}`, { signal: controller.signal })
    if (!response.ok) {
      throw new Error('Не удалось загрузить список систем')
    }
    applySystemDocumentPayload(await response.json())
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось загрузить список систем'
  } finally {
    const isLatestRequest = systemDocumentRequestController === controller
    if (isLatestRequest) {
      systemDocumentRequestController = null
      isSystemDocumentLoading.value = false
    }
    if (tracksFilterRequest) {
      systemFilterRequestCount.value = Math.max(0, systemFilterRequestCount.value - 1)
    }
  }
}

function scheduleSystemDocumentSearch() {
  if (systemDocumentSearchTimer) {
    window.clearTimeout(systemDocumentSearchTimer)
  }
  if (classificationFilterFeedbackTimer) {
    window.clearTimeout(classificationFilterFeedbackTimer)
  }
  systemDocumentSearchTimer = window.setTimeout(() => {
    systemDocumentSearchTimer = null
    loadSystemDocuments(true)
  }, 250)
}

function scheduleClassificationCatalogSearch() {
  if (classificationSearchTimer) {
    window.clearTimeout(classificationSearchTimer)
  }
  if (classificationFilterFeedbackTimer) {
    window.clearTimeout(classificationFilterFeedbackTimer)
  }
  isClassificationSearchPending.value = true
  classificationSearchTimer = window.setTimeout(() => {
    classificationSearchTimer = null
    classificationCatalogSearch.value = classificationCatalogSearchInput.value
    classificationCatalogPage.value = 1
    openedClassificationSystemId.value = null
    isClassificationSearchPending.value = false
  }, 180)
}

function classificationCatalogPageCount() {
  if (classificationCatalogPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(classificationSystems.value.length / Number(classificationCatalogPageSize.value)))
}

function classificationCatalogRangeStart() {
  if (classificationSystems.value.length === 0) return 0
  if (classificationCatalogPageSize.value === 'all') return 1
  return (classificationCatalogPage.value - 1) * Number(classificationCatalogPageSize.value) + 1
}

function classificationCatalogRangeEnd() {
  if (classificationCatalogPageSize.value === 'all') return classificationSystems.value.length
  return Math.min(classificationCatalogPage.value * Number(classificationCatalogPageSize.value), classificationSystems.value.length)
}

function changeClassificationCatalogPageSize() {
  classificationCatalogPage.value = 1
  openedClassificationSystemId.value = null
}

async function changeClassificationCatalogPage(nextPage: number) {
  classificationCatalogPage.value = Math.min(Math.max(nextPage, 1), classificationCatalogPageCount())
  openedClassificationSystemId.value = null
  await nextTick()
  const resultsToolbar = document.querySelector<HTMLElement>('.classification-results-toolbar')
  resultsToolbar?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
  resultsToolbar?.focus({ preventScroll: true })
}

async function loadDocumentTable() {
  documentTableRequestController?.abort()
  const controller = new AbortController()
  documentTableRequestController = controller
  isDocumentTableLoading.value = true
  documentError.value = ''
  try {
    const query = new URLSearchParams()
    if (selectedOrderId.value) {
      query.set('orderId', String(selectedOrderId.value))
    }
    const response = await apiFetch(`/system-documents?${query.toString()}`, { signal: controller.signal })
    if (!response.ok) {
      throw new Error('Не удалось загрузить таблицу 3')
    }
    const payload: SystemDocumentResponse = await response.json()
    documentRows.value = payload.rows
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    documentError.value = error instanceof Error ? error.message : 'Не удалось загрузить таблицу 3'
  } finally {
    if (documentTableRequestController === controller) {
      documentTableRequestController = null
      isDocumentTableLoading.value = false
    }
  }
}

const filteredDocumentRows = computed(() => {
  const query = documentSearch.value.trim().toLocaleLowerCase('ru-RU')
  if (!query) {
    return documentRows.value
  }
  return documentRows.value.filter((row) =>
    row.systemName.toLocaleLowerCase('ru-RU').includes(query) || row.code.toLocaleLowerCase('ru-RU').includes(query),
  )
})

const visibleDocumentRows = computed(() => {
  if (settingsDocumentsPageSize.value === 'all') return filteredDocumentRows.value
  const pageSize = Number(settingsDocumentsPageSize.value)
  const start = (settingsDocumentsPage.value - 1) * pageSize
  return filteredDocumentRows.value.slice(start, start + pageSize)
})

function settingsDocumentsPageCount() {
  if (settingsDocumentsPageSize.value === 'all') return 1
  return Math.max(1, Math.ceil(filteredDocumentRows.value.length / Number(settingsDocumentsPageSize.value)))
}

function settingsDocumentsRangeStart() {
  if (filteredDocumentRows.value.length === 0) return 0
  if (settingsDocumentsPageSize.value === 'all') return 1
  return (settingsDocumentsPage.value - 1) * Number(settingsDocumentsPageSize.value) + 1
}

function settingsDocumentsRangeEnd() {
  if (settingsDocumentsPageSize.value === 'all') return filteredDocumentRows.value.length
  return Math.min(settingsDocumentsPage.value * Number(settingsDocumentsPageSize.value), filteredDocumentRows.value.length)
}

function changeSettingsDocumentsPageSize() {
  settingsDocumentsPage.value = 1
}

async function changeSettingsDocumentsPage(nextPage: number) {
  settingsDocumentsPage.value = Math.min(Math.max(nextPage, 1), settingsDocumentsPageCount())
  await nextTick()
  document.querySelector<HTMLElement>('.settings-documents-block')?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'start',
  })
}

function scheduleDocumentCommentSave(row: SystemDocumentRow) {
  const currentTimer = documentCommentTimers.get(row.id)
  if (currentTimer) {
    window.clearTimeout(currentTimer)
  }
  documentCommentTimers.set(row.id, window.setTimeout(() => saveDocumentComment(row), 500))
}

async function saveDocumentComment(row: SystemDocumentRow) {
  const currentTimer = documentCommentTimers.get(row.id)
  if (currentTimer) {
    window.clearTimeout(currentTimer)
    documentCommentTimers.delete(row.id)
  }
  if (!selectedOrderId.value) {
    return
  }
  documentCommentControllers.get(row.id)?.abort()
  const controller = new AbortController()
  documentCommentControllers.set(row.id, controller)
  try {
    const query = new URLSearchParams({ orderId: String(selectedOrderId.value) })
    const response = await apiFetch(`/system-documents/${row.id}?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ comment: row.comment }),
      signal: controller.signal,
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось сохранить комментарий')
    }
    const saved: SystemDocumentRow = await response.json()
    const listRow = systemDocumentRows.value.find((item) => item.id === saved.id)
    if (listRow) {
      listRow.comment = saved.comment
      listRow.updatedAt = saved.updatedAt
    }
    documentError.value = ''
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
    documentError.value = error instanceof Error ? error.message : 'Не удалось сохранить комментарий'
  } finally {
    if (documentCommentControllers.get(row.id) === controller) {
      documentCommentControllers.delete(row.id)
    }
  }
}

function systemDocumentAttachmentUrl(row: SystemDocumentRow) {
  const query = new URLSearchParams({ orderId: String(row.orderId) })
  return apiURL(`/system-documents/${row.id}/attachment?${query.toString()}`)
}

function openAttachmentPicker(row: SystemDocumentRow) {
  document.getElementById(`document-attachment-${row.id}`)?.click()
}

async function uploadSystemDocumentAttachment(row: SystemDocumentRow, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || attachmentPendingIds.value.includes(row.id)) {
    return
  }
  const extension = file.name.split('.').pop()?.toLocaleLowerCase('ru-RU') ?? ''
  if (!['pdf', 'doc', 'docx'].includes(extension)) {
    documentError.value = 'Можно прикрепить только документы PDF, DOC или DOCX'
    input.value = ''
    return
  }
  if (file.size > 25 * 1024 * 1024) {
    documentError.value = 'Размер документа не должен превышать 25 МБ'
    input.value = ''
    return
  }

  attachmentPendingIds.value = [...attachmentPendingIds.value, row.id]
  documentError.value = ''
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const form = new FormData()
    form.append('file', file)
    const response = await apiFetch(`/system-documents/${row.id}/attachment?${query.toString()}`, {
      method: 'POST',
      body: form,
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось загрузить документ')
    }
    await Promise.all([loadDocumentTable(), loadSystemDocuments()])
  } catch (error) {
    documentError.value = error instanceof Error ? error.message : 'Не удалось загрузить документ'
  } finally {
    attachmentPendingIds.value = attachmentPendingIds.value.filter((id) => id !== row.id)
    input.value = ''
  }
}

async function deleteSystemDocumentAttachment(row: SystemDocumentRow) {
  if (!row.attachmentName || attachmentPendingIds.value.includes(row.id) ||
      !window.confirm(`Удалить документ «${row.attachmentName}»?`)) {
    return
  }

  attachmentPendingIds.value = [...attachmentPendingIds.value, row.id]
  documentError.value = ''
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const response = await apiFetch(`/system-documents/${row.id}/attachment?${query.toString()}`, {
      method: 'DELETE',
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось удалить документ')
    }
    await Promise.all([loadDocumentTable(), loadSystemDocuments()])
  } catch (error) {
    documentError.value = error instanceof Error ? error.message : 'Не удалось удалить документ'
  } finally {
    attachmentPendingIds.value = attachmentPendingIds.value.filter((id) => id !== row.id)
  }
}

async function toggleSystemComparison(row: SystemDocumentRow, event: Event) {
  if (!selectedOrderId.value || comparisonPendingIds.value.includes(row.id)) {
    return
  }
  const selected = (event.target as HTMLInputElement).checked
  const previous = row.comparisonSelected
  row.comparisonSelected = selected
  comparisonPendingIds.value = [...comparisonPendingIds.value, row.id]
  systemCatalogError.value = ''
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const response = await apiFetch(`/system-documents/comparison?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        selected,
        allOrders: true,
        systems: [{ code: row.code, systemName: row.systemName }],
      }),
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось сохранить выбор для сравнения')
    }
    const documentRow = documentRows.value.find((item) => item.id === row.id)
    if (documentRow) {
      documentRow.comparisonSelected = selected
    }
    await loadComparisonCatalogs()
  } catch (error) {
    row.comparisonSelected = previous
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось сохранить выбор для сравнения'
  } finally {
    comparisonPendingIds.value = comparisonPendingIds.value.filter((id) => id !== row.id)
  }
}

async function toggleAllSystemComparisons(event: Event) {
  if (!selectedOrderId.value || isBulkComparisonUpdating.value || filteredSystemDocumentRows.value.length === 0) {
    return
  }
  const selected = (event.target as HTMLInputElement).checked
  const rows = [...filteredSystemDocumentRows.value]
  const previous = new Map(rows.map((row) => [row.id, row.comparisonSelected]))
  rows.forEach((row) => { row.comparisonSelected = selected })
  isBulkComparisonUpdating.value = true
  systemCatalogError.value = ''
  try {
    const query = new URLSearchParams({ orderId: String(selectedOrderId.value) })
    const response = await apiFetch(`/system-documents/comparison?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        selected,
        allOrders: true,
        systems: rows.map((row) => ({ code: row.code, systemName: row.systemName })),
      }),
    })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось сохранить массовый выбор')
    }
    documentRows.value.forEach((row) => {
      if (rows.some((selectedRow) => selectedRow.id === row.id)) {
        row.comparisonSelected = selected
      }
    })
    await loadComparisonCatalogs()
  } catch (error) {
    rows.forEach((row) => { row.comparisonSelected = previous.get(row.id) ?? false })
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось сохранить массовый выбор'
  } finally {
    isBulkComparisonUpdating.value = false
  }
}

async function toggleClassificationSystem(systemId: number) {
  const shouldOpen = openedClassificationSystemId.value !== systemId
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

  if (!shouldOpen && !reduceMotion) {
    const details = document.querySelector<HTMLElement>('.classification-details-shell')
    if (details) {
      await details.animate(
        [
          { opacity: 1, transform: 'translateY(0)' },
          { opacity: 0, transform: 'translateY(-4px)' },
        ],
        { duration: 110, easing: 'ease-out', fill: 'forwards' },
      ).finished.catch(() => undefined)
    }
  }

  const previousPositions = classificationRowPositions()
  openedClassificationSystemId.value = shouldOpen ? systemId : null
  await nextTick()

  if (reduceMotion) {
    return
  }

  for (const { element, top } of previousPositions) {
    if (!element.isConnected) {
      continue
    }

    const offset = top - element.getBoundingClientRect().top
    if (Math.abs(offset) < 1) {
      continue
    }

    element.getAnimations().forEach((animation) => animation.cancel())
    element.animate(
      [
        { transform: `translateY(${offset}px)` },
        { transform: 'translateY(0)' },
      ],
      { duration: 300, easing: 'cubic-bezier(0.22, 1, 0.36, 1)' },
    )
  }

  if (shouldOpen) {
    document.querySelector<HTMLElement>('.classification-details-shell')?.animate(
      [
        { opacity: 0, transform: 'translateY(-10px) scaleY(0.96)', transformOrigin: 'top center' },
        { opacity: 1, transform: 'translateY(0) scaleY(1)', transformOrigin: 'top center' },
      ],
      { duration: 300, easing: 'cubic-bezier(0.22, 1, 0.36, 1)' },
    )
  }
}

function updateClassificationCardColumns() {
  if (window.innerWidth <= 640) {
    classificationCardColumns.value = 1
  } else if (window.innerWidth <= 980) {
    classificationCardColumns.value = 2
  } else {
    classificationCardColumns.value = 3
  }
}

async function openSystemHistory(system: SystemDocumentRow) {
  selectedHistorySystem.value = system
  isHistoryOpen.value = true
  isSystemHistoryLoading.value = true
  systemHistoryError.value = ''
  systemHistoryRows.value = []
  try {
    const query = new URLSearchParams({ code: system.code, systemName: system.systemName })
    const response = await apiFetch(`/system-documents/history?${query.toString()}`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить историю системы')
    }
    systemHistoryRows.value = await response.json()
  } catch (error) {
    systemHistoryError.value = error instanceof Error ? error.message : 'Не удалось загрузить историю системы'
  } finally {
    isSystemHistoryLoading.value = false
  }
}

function closeSystemHistory() {
  selectedHistorySystem.value = null
  isHistoryOpen.value = false
  systemHistoryRows.value = []
  systemHistoryError.value = ''
}

function attachmentFileExtension(fileName: string) {
  return fileName.trim().toLocaleLowerCase('ru-RU').split('.').pop() ?? ''
}

function attachmentFileIcon(fileName: string) {
  switch (attachmentFileExtension(fileName)) {
    case 'doc': return docFileIcon
    case 'docx': return docxFileIcon
    case 'xls': return xlsFileIcon
    case 'xlsx': return xlsxFileIcon
    case 'pdf': return pdfFileIcon
    case 'png': return pngFileIcon
    case 'jpg':
    case 'jpeg': return jpgFileIcon
    default: return genericFileIcon
  }
}

function formatFileSize(size: number) {
  if (!size) return ''
  if (size < 1024) return `${size} Б`
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} КБ`
  return `${(size / (1024 * 1024)).toFixed(1).replace('.', ',')} МБ`
}

function attachmentFileKind(fileName: string) {
  switch (attachmentFileExtension(fileName)) {
    case 'doc':
    case 'docx': return 'word'
    case 'xls':
    case 'xlsx': return 'excel'
    case 'pdf': return 'pdf'
    case 'png': return 'png'
    case 'jpg':
    case 'jpeg': return 'jpeg'
    default: return 'generic'
  }
}

function classModifier(value: string) {
  if (value === 'Рекомендованная') {
    return 'recommended'
  }

  if (value === 'Разрешенная') {
    return 'allowed'
  }

  if (value === 'Запрещенная') {
    return 'forbidden'
  }

  return ''
}

function afterFilterAccentModifier() {
  return classModifier(selectedAfterFilter.value)
}

function statusAccentColor(value: string) {
  switch (classModifier(value)) {
    case 'recommended':
      return '#2f9b5c'
    case 'allowed':
      return '#d4a900'
    case 'forbidden':
      return '#ed1c24'
    default:
      return '#8b97a8'
  }
}

function pageTitle() {
  return navItems.find((item) => item.key === activePage.value)?.label ?? ''
}

const settingsPageModel: SettingsPageViewModel = {
  attachmentFileIcon,
  attachmentPendingIds,
  cancelNavParser,
  changeOrdersPage,
  changeOrdersPageSize,
  changeSettingsClassificationPage,
  changeSettingsClassificationPageSize,
  changeSettingsDocumentsPage,
  changeSettingsDocumentsPageSize,
  changeSettingsSystemCatalogPage,
  changeSettingsSystemCatalogPageSize,
  classModifier,
  classOptions,
  classificationError,
  classificationLoadingMessage,
  createOrder,
  currentSettingsClassificationRows,
  currentSettingsSystemCatalogRows,
  deleteOrder,
  deleteSystemDocumentAttachment,
  documentError,
  documentSearch,
  filteredDocumentRows,
  fontSizePreset,
  fontSizePresets,
  formatFileSize,
  formatNavParserLogTime,
  formatNavParserRunDate,
  formatNavParserRunDuration,
  formatOrderDateTime,
  importOrderWorkbook,
  importSystemCatalogFile,
  importTableFile,
  isClassificationLoading,
  isDocumentTableLoading,
  isNavParserCancelling,
  isNavParserHistoryOpen,
  isNavParserLogOpen,
  isNavParserSettingsOpen,
  isNavParsing,
  isNavSettingsSaving,
  isOrderWorkbookImporting,
  isOrdersLoading,
  isSettingsClassificationUnlocked,
  isSettingsSystemCatalogUnlocked,
  isSystemCatalogLoading,
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
  openAttachmentPicker,
  openOrderWorkbookImport,
  openSystemCatalogImport,
  openTableImport,
  openedNavParserRunId,
  openedSelect,
  orders,
  ordersError,
  ordersPage,
  ordersPageCount,
  ordersPageSize,
  ordersRangeEnd,
  ordersRangeStart,
  runNavParser,
  saveClassificationRow,
  saveDocumentComment,
  saveNavParserSettings,
  saveOrderName,
  saveSystemCatalogRow,
  scheduleClassificationRowSave,
  scheduleDocumentCommentSave,
  scheduleOrderRename,
  scheduleSystemCatalogRowSave,
  selectOrder,
  selectedOrderId,
  selectedOrderName,
  settingsClassificationPage,
  settingsClassificationPageCount,
  settingsClassificationPageSize,
  settingsClassificationRangeEnd,
  settingsClassificationRangeStart,
  settingsDocumentsPage,
  settingsDocumentsPageCount,
  settingsDocumentsPageSize,
  settingsDocumentsRangeEnd,
  settingsDocumentsRangeStart,
  settingsOrderMenuId,
  settingsSystemCatalogPage,
  settingsSystemCatalogPageCount,
  settingsSystemCatalogPageSize,
  settingsSystemCatalogRangeEnd,
  settingsSystemCatalogRangeStart,
  systemCatalogError,
  systemCatalogSearch,
  systemDocumentAttachmentUrl,
  tableSearch,
  toggleSelect,
  uploadSystemDocumentAttachment,
  visibleDocumentRows,
  visibleOrders,
  visibleSettingsClassificationRows,
  visibleSettingsSystemCatalogRows,
}

const comparisonPageModel: ComparisonPageViewModel = {
  MAX_COMPARISON_ORDERS,
  addComparisonOrder,
  availableComparisonOrders,
  changeComparisonPage,
  changeComparisonPageSize,
  classModifier,
  comparisonDropIndex,
  comparisonError,
  comparisonOnlyDifferences,
  comparisonOrderIds,
  comparisonOrderName,
  comparisonOrderOptions,
  comparisonPage,
  comparisonPageCount,
  comparisonPageSize,
  comparisonRangeEnd,
  comparisonRangeStart,
  comparisonRowHasDifference,
  comparisonRows,
  comparisonSort,
  comparisonSortLabel,
  comparisonValue,
  draggedComparisonOrderId,
  dropComparisonOrder,
  endComparisonOrderDrag,
  enterComparisonOrderDrop,
  exportComparisonTable,
  hideComparisonRow,
  isComparisonLoading,
  openedSelect,
  removeComparisonOrder,
  selectComparisonDifferenceFilter,
  selectComparisonOrder,
  selectComparisonSort,
  startComparisonOrderDrag,
  toggleSelect,
  visibleComparisonRows,
}

const classificationPageModel: ClassificationPageViewModel = {
  activeClassificationPageFilterCount,
  animateClassificationFilterEnter,
  animateClassificationFilterLeave,
  changeClassificationCatalogPage,
  changeClassificationCatalogPageSize,
  classModifier,
  classificationBaseSystems,
  classificationCatalogConstructionTypes,
  classificationCatalogError,
  classificationCatalogPage,
  classificationCatalogPageCount,
  classificationCatalogPageSize,
  classificationCatalogRangeEnd,
  classificationCatalogRangeStart,
  classificationCatalogSearchInput,
  classificationEmptyMessage,
  classificationFilterAvailableCount,
  classificationFilterGroups,
  classificationFilterOptionCount,
  classificationFilterOptions,
  classificationFilterSearch,
  classificationSystemRows,
  classificationSystems,
  classificationView,
  clearClassificationFilters,
  collapseClassificationFilters,
  finishClassificationFilterMotion,
  hasActiveClassificationPageFilters,
  hideBrokenSystemTypeImage,
  isClassificationCatalogLoading,
  isClassificationSearchPending,
  isSystemTypesOpen,
  openedClassificationFilter,
  openedClassificationSystem,
  openedClassificationSystemId,
  openedSelect,
  orders,
  prepareClassificationFilterEnter,
  prepareClassificationFilterLeave,
  resetClassificationPageFilters,
  scheduleClassificationCatalogSearch,
  selectClassificationFilter,
  selectConstructionType,
  selectOrder,
  selectSystemType,
  selectedClassificationFilterCount,
  selectedClassificationFilters,
  selectedConstructionType,
  selectedOrderId,
  selectedOrderName,
  selectedSystemType,
  setPage,
  systemTypeImageSource,
  toggleClassificationFilter,
  toggleClassificationSystem,
  toggleSelect,
  visibleClassificationFilterGroups,
  visibleSystemTypes,
}

const systemsPageModel: SystemsPageViewModel = {
  activeSystemFilterCount,
  allVisibleSystemsSelected,
  changeSystemDocumentPage,
  changeSystemDocumentPageSize,
  classModifier,
  comparisonPendingIds,
  currentSystemDocumentRows,
  exportSystemCatalog,
  filterSystemsByClass,
  hasActiveSystemFilters,
  hideBrokenSystemTypeImage,
  isBulkComparisonUpdating,
  isSystemDocumentLoading,
  isSystemFiltering,
  isSystemTypesOpen,
  isSystemsRefreshDone,
  isSystemsRefreshing,
  loadSystemDocuments,
  openSystemHistory,
  openedSelect,
  orders,
  refreshSystemsPage,
  resetSystemFilters,
  scheduleSystemDocumentSearch,
  selectConstructionType,
  selectOrder,
  selectSystemType,
  selectedConstructionType,
  selectedOrderId,
  selectedOrderName,
  selectedSystemCatalogClass,
  selectedSystemCatalogCurator,
  selectedSystemType,
  someVisibleSystemsSelected,
  statusAccentColor,
  systemCatalogClassOptions,
  systemCatalogCuratorOptions,
  systemCatalogError,
  systemCatalogSearch,
  systemCatalogStats,
  systemDocumentPage,
  systemDocumentPageCount,
  systemDocumentPageSize,
  systemDocumentRangeEnd,
  systemDocumentRangeStart,
  systemTypeImageSource,
  systemsConstructionTypes,
  toggleAllSystemComparisons,
  toggleSelect,
  toggleSystemComparison,
  visibleSystemDocumentRows,
  visibleSystemTypes,
}

const changesPageModel: ChangesPageViewModel = {
  activeChangeFilterCount,
  afterFilterAccentModifier,
  afterOptions,
  beforeOptions,
  changeClassificationPage,
  changeClassificationPageSize,
  classModifier,
  classificationChangesEmptyMessage,
  classificationConstructionTypes,
  classificationError,
  classificationLoadingMessage,
  classificationPage,
  classificationPageCount,
  classificationPageSize,
  classificationRangeEnd,
  classificationRangeStart,
  classificationStats,
  currentClassificationRows,
  exportClassificationTable,
  filterChangesByClass,
  hasActiveChangeFilters,
  isChangesRefreshDone,
  isChangesRefreshing,
  isClassificationFiltering,
  isClassificationLoading,
  openedSelect,
  orders,
  refreshChangesPage,
  resetChangesFilters,
  selectAfterChangeFilter,
  selectBeforeChangeFilter,
  selectConstructionType,
  selectOrder,
  selectedAfterFilter,
  selectedBeforeFilter,
  selectedConstructionType,
  selectedOrderId,
  selectedOrderName,
  statusAccentColor,
  toggleSelect,
  visibleClassificationRows,
}

function updateScrollTopVisibility() {
  isScrollTopVisible.value = window.scrollY > 500
}

function scrollToPageTop() {
  window.scrollTo({
    top: 0,
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
  })
}

onMounted(async () => {
  updateClassificationCardColumns()
  updateScrollTopVisibility()
  window.addEventListener('resize', updateClassificationCardColumns)
  window.addEventListener('scroll', updateScrollTopVisibility, { passive: true })
  window.addEventListener('hashchange', syncPageFromLocation)
  await loadOrders()
  await Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable(), loadNavParserSettings(), loadNavParserProgress(), loadNavParserRuns()])
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateClassificationCardColumns)
  window.removeEventListener('scroll', updateScrollTopVisibility)
  window.removeEventListener('hashchange', syncPageFromLocation)
  if (systemDocumentSearchTimer) {
    window.clearTimeout(systemDocumentSearchTimer)
  }
  if (classificationSearchTimer) {
    window.clearTimeout(classificationSearchTimer)
  }
  stopNavParserPolling()
  classificationRequestController?.abort()
  systemCatalogRequestController?.abort()
  classificationCatalogRequestController?.abort()
  systemDocumentRequestController?.abort()
  documentTableRequestController?.abort()
  documentCommentTimers.forEach((timer) => window.clearTimeout(timer))
  classificationEditTimers.forEach((timer) => window.clearTimeout(timer))
  systemCatalogEditTimers.forEach((timer) => window.clearTimeout(timer))
  orderRenameTimers.forEach((timer) => window.clearTimeout(timer))
  orderRenameControllers.forEach((controller) => controller.abort())
  classificationEditControllers.forEach((controller) => controller.abort())
  systemCatalogEditControllers.forEach((controller) => controller.abort())
  documentCommentControllers.forEach((controller) => controller.abort())
})
</script>

<template>
  <div
    class="app"
    :style="{ '--app-min-font-size': `${minimumAppFontSize}px` }"
    @click="openedSelect = null; settingsOrderMenuId = null; isClassificationLegendOpen = false"
  >
    <AppHeader
      :active-page="activePage"
      :items="navItems"
      v-model:is-legend-open="isClassificationLegendOpen"
      @navigate="setPage"
    />

    <main class="page-main">
      <ChangesPage v-if="activePage === 'changes'" :model="changesPageModel" />
      <SystemsPage v-else-if="activePage === 'systems'" :model="systemsPageModel" />
      <ClassificationPage v-else-if="activePage === 'classification'" :model="classificationPageModel" />
      <ComparisonPage v-else-if="activePage === 'comparison'" :model="comparisonPageModel" />
      <SettingsPage v-else-if="activePage === 'settings'" :model="settingsPageModel" />
      <section v-else class="placeholder-page">
        <h1>{{ pageTitle() }}</h1>
      </section>

      <ClassLegendFooter
        v-if="activePage !== 'settings'"
        :include-forbidden="activePage !== 'changes'"
      />
    </main>

    <Transition name="scroll-top">
      <button
        v-if="(activePage === 'changes' || activePage === 'systems') && isScrollTopVisible"
        class="scroll-top-button"
        type="button"
        aria-label="Вернуться в начало страницы"
        title="Наверх"
        @click="scrollToPageTop"
      >
        <i aria-hidden="true" />
      </button>
    </Transition>

    <SystemHistoryModal
      :system="selectedHistorySystem"
      :rows="systemHistoryRows"
      :is-loading="isSystemHistoryLoading"
      :error="systemHistoryError"
      v-model:is-history-open="isHistoryOpen"
      :attachment-file-kind="attachmentFileKind"
      :attachment-file-icon="attachmentFileIcon"
      :attachment-u-r-l="systemDocumentAttachmentUrl"
      @close="closeSystemHistory"
    />
  </div>
</template>
