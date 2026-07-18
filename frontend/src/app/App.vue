<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  ArrowDownUp,
  CalendarDays,
  ChevronDown,
  ChevronRight,
  ChevronUp,
  CircleCheck,
  Clock3,
  CloudUpload,
  Database,
  ExternalLink,
  Folder,
  FunnelX,
  Globe2,
  GripVertical,
  Grid2X2,
  House,
  Info,
  Layers3,
  List,
  ListFilter,
  Lock,
  MoreHorizontal,
  EllipsisVertical,
  Plus,
  RefreshCw,
  Repeat2,
  Scale,
  Search,
  TriangleAlert,
  Trash2,
  Unlock,
  UsersRound,
  X,
} from '@lucide/vue'
import genericFileIcon from 'bootstrap-icons/icons/file-earmark.svg'
import docFileIcon from 'bootstrap-icons/icons/filetype-doc.svg'
import docxFileIcon from 'bootstrap-icons/icons/filetype-docx.svg'
import jpgFileIcon from 'bootstrap-icons/icons/filetype-jpg.svg'
import pdfFileIcon from 'bootstrap-icons/icons/filetype-pdf.svg'
import pngFileIcon from 'bootstrap-icons/icons/filetype-png.svg'
import xlsFileIcon from 'bootstrap-icons/icons/filetype-xls.svg'
import xlsxFileIcon from 'bootstrap-icons/icons/filetype-xlsx.svg'
import logo from '@/shared/assets/logo.png'

type ClassificationChange = {
  id: number
  orderId: number
  position: number
  systemName: string
  systemUrl: string
  constructionType: string
  classBefore: string
  classAfter: string
  importedAt: string
}

type ClassificationStats = {
  addedSystems: number
  recommended: number
  allowed: number
  classificationChanges: number
}

type ClassificationResponse = {
  rows: ClassificationChange[]
  stats: ClassificationStats
  beforeOptions: string[]
  afterOptions: string[]
}

type SystemCatalogRow = {
  id: number
  orderId: number
  position: number
  code: string
  systemName: string
  systemUrl: string
  systemClass: string
  curator: string
  importedAt: string
  characteristics: SystemCharacteristic[]
}

type SystemCharacteristic = {
  position: number
  name: string
  value: string
}

type NavParseReport = {
  total: number
  found: number
  fallbackFound: number
  updated: number
  failed: number
  failedSystems: string[]
  notFound: string[]
}

type NavParserSettings = {
  updateIntervalDays: number
  workerCount: number
  requestTimeoutSeconds: number
  retryAttempts: number
  retryDelaySeconds: number
  fallbackSearch: boolean
  lastRunAt: string | null
  nextRunAt: string | null
}

type NavParserLogEntry = {
  time: string
  level: 'info' | 'success' | 'warning' | 'error'
  message: string
}

type NavParserProgress = {
  running: boolean
  source: 'manual' | 'scheduled' | string
  stage: string
  message: string
  percent: number
  processed: number
  total: number
  found: number
  updated: number
  failed: number
  notFound: number
  startedAt: string | null
  finishedAt: string | null
  logs: NavParserLogEntry[]
}

type NavParserRun = {
  id: number
  source: 'manual' | 'scheduled' | string
  status: 'completed' | 'failed' | string
  message: string
  total: number
  found: number
  updated: number
  failed: number
  notFound: number
  startedAt: string
  finishedAt: string
  logs: NavParserLogEntry[]
}

type SystemCatalogStats = {
  total: number
  recommended: number
  allowed: number
  forbidden: number
  curators: number
}

type SystemCatalogResponse = {
  rows: SystemCatalogRow[]
  stats: SystemCatalogStats
  classOptions: string[]
  curatorOptions: string[]
  systemTypes: SystemTypeOption[]
}

type SystemDocumentRow = {
  id: number
  orderId: number
  orderName: string
  systemCatalogId: number
  position: number
  code: string
  systemName: string
  systemUrl: string
  systemClass: string
  curator: string
  comparisonSelected: boolean
  comment: string
  attachmentName: string
  attachmentType: string
  attachmentSize: number
  createdAt: string
  updatedAt: string
  characteristics: SystemCharacteristic[]
}

type SystemDocumentResponse = {
  rows: SystemDocumentRow[]
  stats: SystemCatalogStats
  classOptions: string[]
  curatorOptions: string[]
}

type SystemTypeOption = {
  slug: string
  name: string
  imageUrl: string
  position: number
}

type Order = {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

type ComparisonRow = {
  key: string
  name: string
  values: string[]
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api'

const activePage = ref('changes')
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
const selectedOrderId = ref<number | null>(null)
const comparisonOrderIds = ref<number[]>([])
const comparisonCatalogByOrder = ref<Record<number, SystemDocumentRow[]>>({})
const comparisonPendingIds = ref<number[]>([])
const attachmentPendingIds = ref<number[]>([])
const comparisonAllOrders = ref(true)
const isBulkComparisonUpdating = ref(false)
const hiddenComparisonRows = ref<string[]>([])
const comparisonOnlyDifferences = ref(false)
const comparisonSort = ref<'differences-first' | 'name-asc' | 'name-desc'>('differences-first')
const comparisonPageSize = ref('20')
const comparisonPage = ref(1)
const isComparisonLoading = ref(false)
const comparisonError = ref('')
const isOrdersLoading = ref(false)
const ordersError = ref('')
const settingsOrderMenuId = ref<number | null>(null)
const orderRenameTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const classificationEditTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const systemCatalogEditTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const documentCommentTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
let systemDocumentSearchTimer: ReturnType<typeof window.setTimeout> | null = null
let classificationSearchTimer: ReturnType<typeof window.setTimeout> | null = null
let classificationFilterFeedbackTimer: ReturnType<typeof window.setTimeout> | null = null
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
const classificationRows = ref<ClassificationChange[]>([])
const classificationPageSize = ref('20')
const classificationPage = ref(1)
const settingsClassificationPageSize = ref('10')
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
const isClassificationLoading = ref(false)
const isClassificationFiltering = ref(false)
const isChangesRefreshing = ref(false)
const changesLastRefreshedAt = ref('')
const classificationLoadingMessage = ref('Загрузка таблицы...')
const classificationError = ref('')
const classificationConstructionTypes = computed(() => [...constructionTypes, 'Тип не присвоен'].map((name) => ({
  name,
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
const settingsSystemCatalogPageSize = ref('10')
const settingsSystemCatalogPage = ref(1)
const isSettingsSystemCatalogUnlocked = ref(false)
const systemDocumentRows = ref<SystemDocumentRow[]>([])
const systemDocumentPageSize = ref('20')
const systemDocumentPage = ref(1)
const systemsConstructionTypes = computed(() => constructionTypes.map((name) => ({
  name,
  count: systemDocumentRows.value.filter((system) => systemMatchesConstructionType(system, name)).length,
})))
const documentRows = ref<SystemDocumentRow[]>([])
const documentSearch = ref('')
const settingsDocumentsPageSize = ref('10')
const settingsDocumentsPage = ref(1)
const documentError = ref('')
const isDocumentTableLoading = ref(false)
const classificationCatalogRows = ref<SystemCatalogRow[]>([])
const classificationCatalogSearch = ref('')
const classificationCatalogSearchInput = ref('')
const isClassificationSearchPending = ref(false)
const classificationView = ref<'grid' | 'list'>('grid')
const classificationCatalogPageSize = ref('50')
const classificationCatalogPage = ref(1)
const isClassificationCatalogLoading = ref(false)
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
const systemTypeSourceRows = computed(() => activePage.value === 'systems' ? systemDocumentRows.value : classificationCatalogRows.value)
const systemTypes = computed(() => [{ slug: '', name: 'Все системы', imageUrl: '', position: 0 }, ...parsedSystemTypes.value].map((type) => ({
  ...type,
  count: systemTypeSourceRows.value.filter((system) =>
    matchesConstructionType(system) && matchesSystemType(system, type),
  ).length,
})))
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
  if (!query) return classificationFilterGroups.value
  return classificationFilterGroups.value.filter((name) => name.toLocaleLowerCase('ru-RU').includes(query))
})
const selectedClassificationFilterCount = computed(() => Object.keys(selectedClassificationFilters.value).length)
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
  return [...systems].sort((left, right) => left.systemName.localeCompare(right.systemName, 'ru'))
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
const isSystemCatalogLoading = ref(false)
const isSystemDocumentLoading = ref(false)
const systemFilterRequestCount = ref(0)
const isSystemFiltering = computed(() => systemFilterRequestCount.value > 0)
const isSystemsRefreshing = ref(false)
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
const isNavParsing = ref(false)
const navParseMessage = ref('')
const navParseError = ref('')
const navParseNotFound = ref<string[]>([])
const navParserIntervalDays = ref(7)
const navParserProgress = ref<NavParserProgress>({
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
const navParserLogsNewestFirst = computed(() => [...navParserProgress.value.logs].reverse())
const navParserRuns = ref<NavParserRun[]>([])
const isNavParserLogOpen = ref(false)
const openedNavParserRunId = ref<number | null>(null)
const navParseFailedSystems = ref<string[]>([])
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
let navParserPollTimer: ReturnType<typeof window.setInterval> | null = null
let navParserRequestPending = false

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

function selectedOrderUpdatedAt() {
  return orders.value.find((order) => order.id === selectedOrderId.value)?.updatedAt ?? ''
}

function changesPageUpdatedAt() {
  return changesLastRefreshedAt.value || selectedOrderUpdatedAt()
}

function systemsPageUpdatedAt() {
  return systemsLastRefreshedAt.value || selectedOrderUpdatedAt()
}

function addedSystemsHint() {
  const index = orders.value.findIndex((order) => order.id === selectedOrderId.value)
  return index === orders.value.length - 1
    ? 'Относительно пустой базы'
    : 'Относительно предыдущего распоряжения'
}

async function refreshChangesPage() {
  if (isChangesRefreshing.value) {
    return
  }
  isChangesRefreshing.value = true
  try {
    await loadOrders()
    await loadClassificationChanges()
    changesLastRefreshedAt.value = new Date().toISOString()
  } finally {
    isChangesRefreshing.value = false
  }
}

async function refreshSystemsPage() {
  if (isSystemsRefreshing.value) {
    return
  }
  isSystemsRefreshing.value = true
  try {
    await Promise.all([loadSystemCatalog(true), loadSystemDocuments(true)])
    if (!systemCatalogError.value) {
      systemsLastRefreshedAt.value = new Date().toISOString()
    }
  } finally {
    isSystemsRefreshing.value = false
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
    const response = await fetch(`${API_BASE_URL}/orders`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить распоряжения')
    }

    orders.value = await response.json()
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
  const response = await fetch(`${API_BASE_URL}/system-documents?${query.toString()}`)
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

async function createOrder() {
  const name = window.prompt('Название распоряжения')
  if (!name?.trim()) {
    return
  }

  const response = await fetch(`${API_BASE_URL}/orders`, {
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

async function deleteOrder(order: Order) {
  if (!window.confirm(`Удалить ${order.name}? Данные таблиц этого распоряжения тоже будут удалены.`)) {
    return
  }

  const response = await fetch(`${API_BASE_URL}/orders/${order.id}`, { method: 'DELETE' })
  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    ordersError.value = payload?.error ?? 'Не удалось удалить распоряжение'
    return
  }

  orders.value = orders.value.filter((item) => item.id !== order.id)
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

  try {
    const response = await fetch(`${API_BASE_URL}/orders/${order.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: nextName }),
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
    ordersError.value = error instanceof Error ? error.message : 'Не удалось сохранить название распоряжения'
  }
}

async function selectComparisonOrder(index: number, order: Order) {
  comparisonOrderIds.value[index] = order.id
  openedSelect.value = null
  await loadComparisonCatalog(order.id)
}

function comparisonOrderOptions(currentOrderId: number) {
  return orders.value.filter((order) => order.id === currentOrderId || !comparisonOrderIds.value.includes(order.id))
}

function availableComparisonOrders() {
  return orders.value.filter((order) => !comparisonOrderIds.value.includes(order.id))
}

async function addComparisonOrder(order?: Order) {
  const nextOrder = order ?? availableComparisonOrders()[0]
  if (nextOrder) {
    comparisonOrderIds.value.push(nextOrder.id)
    await loadComparisonCatalog(nextOrder.id)
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

  const response = await fetch(`${API_BASE_URL}/comparison/export`, {
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
  isClassificationLoading.value = true
  classificationLoadingMessage.value = 'Загрузка таблицы...'
  classificationError.value = ''

  try {
    const query = new URLSearchParams()
    if (selectedOrderId.value) {
      query.set('orderId', String(selectedOrderId.value))
    }
    const response = await fetch(`${API_BASE_URL}/classification-changes?${query.toString()}`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить таблицу 1')
    }

    applyClassificationPayload(await response.json())
  } catch (error) {
    classificationError.value = error instanceof Error ? error.message : 'Не удалось загрузить таблицу 1'
  } finally {
    isClassificationLoading.value = false
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
    const response = await fetch(`${API_BASE_URL}/classification-changes/import`, {
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
  const response = await fetch(`${API_BASE_URL}/classification-changes/export?${query.toString()}`)
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
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const response = await fetch(`${API_BASE_URL}/classification-changes/${row.id}?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(submitted),
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
    classificationError.value = error instanceof Error ? error.message : 'Не удалось сохранить строку таблицы 1'
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
  isClassificationCatalogLoading.value = true
  classificationCatalogError.value = ''

  try {
    const query = new URLSearchParams()
    if (selectedOrderId.value) {
      query.set('orderId', String(selectedOrderId.value))
    }

    const response = await fetch(`${API_BASE_URL}/system-catalog?${query.toString()}`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить системы из таблицы 2')
    }

    const payload: SystemCatalogResponse = await response.json()
    classificationCatalogRows.value = payload.rows
    parsedSystemTypes.value = payload.systemTypes ?? parsedSystemTypes.value
  } catch (error) {
    classificationCatalogError.value = error instanceof Error ? error.message : 'Не удалось загрузить системы из таблицы 2'
  } finally {
    isClassificationCatalogLoading.value = false
  }
}

async function loadSystemCatalog(silent = false) {
  if (!silent) {
    isSystemCatalogLoading.value = true
  }
  systemCatalogError.value = ''

  try {
    const query = buildSystemCatalogQuery()
    const response = await fetch(`${API_BASE_URL}/system-catalog?${query.toString()}`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить таблицу 2')
    }

    const payload: SystemCatalogResponse = await response.json()
    applySystemCatalogPayload(payload)
  } catch (error) {
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось загрузить таблицу 2'
  } finally {
    if (!silent) {
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
    const response = await fetch(`${API_BASE_URL}/system-catalog/import`, {
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
  const response = await fetch(`${API_BASE_URL}/system-documents/export?${query.toString()}`)
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

async function runNavParser() {
  if (isNavParsing.value) {
    return
  }

  navParserRequestPending = true
  isNavParsing.value = true
  navParseMessage.value = ''
  navParseError.value = ''
  navParseNotFound.value = []
  navParseFailedSystems.value = []
  isNavParserLogOpen.value = true
  startNavParserPolling()
  try {
    const response = await fetch(`${API_BASE_URL}/system-catalog/parse-nav`, { method: 'POST' })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      if (response.status === 409) {
        await loadNavParserProgress()
        return
      }
      throw new Error(payload?.error ?? 'Не удалось выполнить парсинг nav.tn.ru')
    }

    const report: NavParseReport = await response.json()
    navParseMessage.value = `Обновлено ${report.updated} из ${report.total}. Через резервный поиск: ${report.fallbackFound}, не найдено: ${report.notFound.length}, ошибок: ${report.failed}.`
    navParseNotFound.value = report.notFound
    navParseFailedSystems.value = report.failedSystems ?? []
    selectedClassificationFilters.value = {}
    await Promise.all([loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable(), loadNavParserSettings()])
  } catch (error) {
    navParseError.value = error instanceof Error ? error.message : 'Не удалось выполнить парсинг nav.tn.ru'
  } finally {
    navParserRequestPending = false
    await Promise.all([loadNavParserProgress(), loadNavParserRuns()])
    isNavParsing.value = navParserProgress.value.running
    if (!navParserProgress.value.running) {
      stopNavParserPolling()
    }
  }
}

async function loadNavParserProgress() {
  try {
    const response = await fetch(`${API_BASE_URL}/nav-parser/status`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить прогресс парсера')
    }
    const progress: NavParserProgress = await response.json()
    navParserProgress.value = { ...progress, logs: progress.logs ?? [] }
    if (navParserProgress.value.running) {
      isNavParserLogOpen.value = true
    }
    isNavParsing.value = navParserProgress.value.running || navParserRequestPending
    if (navParserProgress.value.running && !navParserPollTimer) {
      startNavParserPolling()
    }
  } catch (error) {
    if (navParserRequestPending) {
      navParseError.value = error instanceof Error ? error.message : 'Не удалось загрузить прогресс парсера'
    }
  }
}

async function loadNavParserRuns() {
  try {
    const response = await fetch(`${API_BASE_URL}/nav-parser/runs?limit=50`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить историю запусков')
    }
    const runs: NavParserRun[] | null = await response.json()
    navParserRuns.value = (runs ?? []).map((run) => ({ ...run, logs: run.logs ?? [] }))
  } catch (error) {
    navSettingsError.value = error instanceof Error ? error.message : 'Не удалось загрузить историю запусков'
  }
}

function startNavParserPolling() {
  if (navParserPollTimer) {
    return
  }
  void loadNavParserProgress()
  navParserPollTimer = window.setInterval(() => {
    void loadNavParserProgress().then(() => {
      if (!navParserProgress.value.running && !navParserRequestPending) {
        stopNavParserPolling()
      }
    })
  }, 750)
}

function stopNavParserPolling() {
  if (!navParserPollTimer) {
    return
  }
  window.clearInterval(navParserPollTimer)
  navParserPollTimer = null
}

function formatNavParserLogTime(value: string) {
  if (!value) {
    return '—'
  }
  return new Intl.DateTimeFormat('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))
}

function formatNavParserRunDate(value: string) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

function formatNavParserRunDuration(run: NavParserRun) {
  const seconds = Math.max(0, Math.round((new Date(run.finishedAt).getTime() - new Date(run.startedAt).getTime()) / 1000))
  if (seconds < 60) {
    return `${seconds} сек.`
  }
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

async function loadNavParserSettings() {
  try {
    const response = await fetch(`${API_BASE_URL}/nav-parser/settings`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить настройки парсера')
    }
    const settings: NavParserSettings = await response.json()
    applyNavParserSettings(settings)
    navSettingsError.value = ''
  } catch (error) {
    navSettingsError.value = error instanceof Error ? error.message : 'Не удалось загрузить настройки парсера'
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
    const response = await fetch(`${API_BASE_URL}/nav-parser/settings`, {
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
    const settings: NavParserSettings = await response.json()
    applyNavParserSettings(settings)
    navSettingsMessage.value = 'Частота обновления сохранена'
  } catch (error) {
    navSettingsError.value = error instanceof Error ? error.message : 'Не удалось сохранить частоту обновления'
  } finally {
    isNavSettingsSaving.value = false
  }
}

function navParserNextRunLabel() {
  return navParserNextRunAt.value
    ? `Следующий запуск: ${formatOrderDateTime(navParserNextRunAt.value)}`
    : 'Следующий запуск — после первого успешного запуска'
}

function applyNavParserSettings(settings: NavParserSettings) {
  navParserIntervalDays.value = settings.updateIntervalDays ?? 1
  navParserWorkerCount.value = settings.workerCount ?? 4
  navParserRequestTimeout.value = settings.requestTimeoutSeconds ?? 35
  navParserRetryAttempts.value = settings.retryAttempts ?? 3
  navParserRetryDelay.value = settings.retryDelaySeconds ?? 2
  navParserFallbackSearch.value = settings.fallbackSearch ?? true
  navParserNextRunAt.value = settings.nextRunAt ?? null
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
  try {
    const query = new URLSearchParams({ orderId: String(row.orderId) })
    const response = await fetch(`${API_BASE_URL}/system-catalog/${row.id}?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(submitted),
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
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось сохранить строку таблицы 2'
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

function setPage(page: string) {
  activePage.value = page
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
    ? `${API_BASE_URL}/nav-system-types/${encodeURIComponent(type.slug)}/image`
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

function removeClassificationFilter(name: string) {
  const next = { ...selectedClassificationFilters.value }
  delete next[name]
  selectedClassificationFilters.value = next
  classificationCatalogPage.value = 1
  openedClassificationSystemId.value = null
}

function collapseClassificationFilters() {
  openedClassificationFilter.value = null
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
  if (!silent) {
    isSystemCatalogLoading.value = true
    isSystemDocumentLoading.value = true
  } else {
    systemFilterRequestCount.value += 1
  }
  systemCatalogError.value = ''
  try {
    const query = buildSystemCatalogQuery()
    const response = await fetch(`${API_BASE_URL}/system-documents?${query.toString()}`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить список систем')
    }
    applySystemDocumentPayload(await response.json())
  } catch (error) {
    systemCatalogError.value = error instanceof Error ? error.message : 'Не удалось загрузить список систем'
  } finally {
    if (!silent) {
      isSystemCatalogLoading.value = false
      isSystemDocumentLoading.value = false
    } else {
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
  isDocumentTableLoading.value = true
  documentError.value = ''
  try {
    const query = new URLSearchParams()
    if (selectedOrderId.value) {
      query.set('orderId', String(selectedOrderId.value))
    }
    const response = await fetch(`${API_BASE_URL}/system-documents?${query.toString()}`)
    if (!response.ok) {
      throw new Error('Не удалось загрузить таблицу 3')
    }
    const payload: SystemDocumentResponse = await response.json()
    documentRows.value = payload.rows
  } catch (error) {
    documentError.value = error instanceof Error ? error.message : 'Не удалось загрузить таблицу 3'
  } finally {
    isDocumentTableLoading.value = false
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
  try {
    const query = new URLSearchParams({ orderId: String(selectedOrderId.value) })
    const response = await fetch(`${API_BASE_URL}/system-documents/${row.id}?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ comment: row.comment }),
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
    documentError.value = error instanceof Error ? error.message : 'Не удалось сохранить комментарий'
  }
}

function systemDocumentAttachmentUrl(row: SystemDocumentRow) {
  const query = new URLSearchParams({ orderId: String(row.orderId) })
  return `${API_BASE_URL}/system-documents/${row.id}/attachment?${query.toString()}`
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
    const response = await fetch(`${API_BASE_URL}/system-documents/${row.id}/attachment?${query.toString()}`, {
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
    const response = await fetch(`${API_BASE_URL}/system-documents/${row.id}/attachment?${query.toString()}`, {
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
    const response = await fetch(`${API_BASE_URL}/system-documents/comparison?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        selected,
        allOrders: comparisonAllOrders.value,
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
    if (comparisonAllOrders.value) {
      await loadComparisonCatalogs()
    } else if (comparisonOrderIds.value.includes(row.orderId)) {
      await loadComparisonCatalog(row.orderId)
    }
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
    const response = await fetch(`${API_BASE_URL}/system-documents/comparison?${query.toString()}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        selected,
        allOrders: comparisonAllOrders.value,
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
    const response = await fetch(`${API_BASE_URL}/system-documents/history?${query.toString()}`)
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
  await loadOrders()
  await Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable(), loadNavParserSettings(), loadNavParserProgress(), loadNavParserRuns()])
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateClassificationCardColumns)
  window.removeEventListener('scroll', updateScrollTopVisibility)
  if (systemDocumentSearchTimer) {
    window.clearTimeout(systemDocumentSearchTimer)
  }
  if (classificationSearchTimer) {
    window.clearTimeout(classificationSearchTimer)
  }
  stopNavParserPolling()
  documentCommentTimers.forEach((timer) => window.clearTimeout(timer))
  classificationEditTimers.forEach((timer) => window.clearTimeout(timer))
  systemCatalogEditTimers.forEach((timer) => window.clearTimeout(timer))
  orderRenameTimers.forEach((timer) => window.clearTimeout(timer))
})
</script>

<template>
  <div class="app" @click="openedSelect = null; settingsOrderMenuId = null">
    <header class="site-header">
      <div class="header-container">
        <div class="main-header__inner">
          <a class="brand-link" href="/" aria-label="Технониколь Светофор онлайн">
            <img :src="logo" alt="Технониколь Светофор онлайн" />
          </a>

          <nav class="primary-nav" aria-label="Основная навигация">
            <button
              v-for="item in navItems"
              :key="item.key"
              class="primary-nav__item"
              :class="{ 'is-active': activePage === item.key }"
              type="button"
              @click.stop="setPage(item.key)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>
      </div>
    </header>

    <main class="page-main">
      <section v-if="activePage === 'changes'" class="changes-page">
        <div class="changes-topbar">
          <div class="select-field">
            <span>Распоряжения</span>
            <div class="custom-select changes-order-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button changes-order-select__button" type="button" @click.stop="toggleSelect('order')">
                <CalendarDays :size="18" :stroke-width="1.8" aria-hidden="true" />
                <span>{{ selectedOrderName() }}</span>
                <ChevronDown class="changes-order-select__chevron" :size="18" :stroke-width="1.8" aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order.id"
                    class="custom-select__option"
                    :class="{ 'is-selected': order.id === selectedOrderId }"
                    type="button"
                    @click="selectOrder(order)"
                  >
                    {{ order.name }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <div class="changes-refresh-panel">
            <span>Последнее обновление: {{ formatOrderDateTime(changesPageUpdatedAt()) }}</span>
            <button type="button" :disabled="isChangesRefreshing" @click="refreshChangesPage">
              <RefreshCw :class="{ 'is-spinning': isChangesRefreshing }" :size="18" :stroke-width="1.8" aria-hidden="true" />
              {{ isChangesRefreshing ? 'Обновление…' : 'Обновить' }}
            </button>
          </div>
        </div>

        <section class="summary-grid" aria-label="Сводка изменений">
          <article class="summary-card summary-card--added">
            <div class="summary-card__icon" aria-hidden="true">
              <Layers3 :size="34" :stroke-width="1.8" />
            </div>
            <div class="summary-card__content">
              <span>Добавлено систем</span>
              <strong>{{ classificationStats.addedSystems }}</strong>
              <small>{{ addedSystemsHint() }}</small>
            </div>
          </article>

          <div class="status-stack changes-status-stack">
            <button class="status-card status-card--recommended" :class="{ 'is-selected': selectedAfterFilter === 'Рекомендованная' }" type="button" @click="filterChangesByClass('Рекомендованная')">
              <CircleCheck :size="25" :stroke-width="1.8" aria-hidden="true" />
              <strong>{{ classificationStats.recommended }}</strong>
              <span>Рекомендованных</span>
              <ChevronRight class="status-card__chevron" :size="20" aria-hidden="true" />
            </button>
            <button class="status-card status-card--allowed" :class="{ 'is-selected': selectedAfterFilter === 'Разрешенная' }" type="button" @click="filterChangesByClass('Разрешенная')">
              <TriangleAlert :size="25" :stroke-width="1.8" aria-hidden="true" />
              <strong>{{ classificationStats.allowed }}</strong>
              <span>Разрешенных</span>
              <ChevronRight class="status-card__chevron" :size="20" aria-hidden="true" />
            </button>
          </div>

          <article class="summary-card summary-card--changed">
            <div class="summary-card__icon" aria-hidden="true">
              <Repeat2 :size="34" :stroke-width="1.8" />
            </div>
            <div class="summary-card__content">
              <span>Изменения в классификации</span>
              <strong>{{ classificationStats.classificationChanges }}</strong>
              <small>{{ classificationStats.classificationChanges ? 'Класс систем изменён' : 'Нет изменений' }}</small>
            </div>
          </article>
        </section>

        <section class="filter-panel" aria-label="Тип строительства">
          <h2>Тип строительства</h2>
          <div class="type-tabs type-tabs--changes">
            <button
              v-for="type in classificationConstructionTypes"
              :key="type.name"
              class="type-tab"
              :class="{ 'type-tab--active': type.name === selectedConstructionType }"
              type="button"
              @click="selectConstructionType(type.name)"
            >
              <span>{{ type.name }}</span>
              <strong>{{ type.count }}</strong>
            </button>
          </div>
        </section>

        <section
          class="table-toolbar"
          :class="afterFilterAccentModifier() && `changes-filter-accent--${afterFilterAccentModifier()}`"
          :style="{ '--before-accent': statusAccentColor(selectedBeforeFilter), '--after-accent': statusAccentColor(selectedAfterFilter) }"
          aria-label="Управление таблицей"
        >
          <div class="select-field">
            <span>Было</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'before', 'is-filtered': selectedBeforeFilter !== 'Все' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('before')">
                <span>{{ selectedBeforeFilter }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'before'" class="custom-select__menu">
                  <button
                    v-for="option in beforeOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedBeforeFilter }"
                    type="button"
                    @click="selectBeforeChangeFilter(option)"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
          <div class="select-field">
            <span>Стало</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'after', 'is-filtered': selectedAfterFilter !== 'Все' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('after')">
                <span>{{ selectedAfterFilter }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'after'" class="custom-select__menu">
                  <button
                    v-for="option in afterOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedAfterFilter }"
                    type="button"
                    @click="selectAfterChangeFilter(option)"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
          <Transition name="changes-reset">
            <button
              v-if="hasActiveChangeFilters"
              class="changes-reset-filters"
              type="button"
              title="Сбросить фильтры"
              aria-label="Сбросить фильтры"
              @click="resetChangesFilters"
            >
              <FunnelX :size="19" :stroke-width="1.8" aria-hidden="true" />
              <span class="systems-reset-filters__count" aria-hidden="true">{{ activeChangeFilterCount }}</span>
            </button>
          </Transition>
          <button class="export-button" type="button" @click="exportClassificationTable">
            Экспорт
            <img class="export-button__xlsx-icon" :src="xlsxFileIcon" alt="" aria-hidden="true" />
          </button>
        </section>

        <p v-if="classificationError" class="table-message table-message--error">{{ classificationError }}</p>

        <div
          class="systems-table changes-table-card"
          :class="afterFilterAccentModifier() && `changes-table-card--${afterFilterAccentModifier()}`"
          :style="{ '--before-accent': statusAccentColor(selectedBeforeFilter), '--after-accent': statusAccentColor(selectedAfterFilter) }"
        >
          <header class="changes-table-card__header">
            <div class="changes-table-card__heading">
              <span class="changes-table-card__icon" aria-hidden="true">
                <ArrowDownUp :size="18" :stroke-width="1.8" />
              </span>
              <div>
                <strong>Системы и изменения классов</strong>
                <span>Сравнение статусов по выбранному распоряжению</span>
              </div>
            </div>
            <span class="changes-table-card__count">{{ currentClassificationRows().length }} систем</span>
          </header>
          <Transition name="table-filter-loading">
            <div v-if="isClassificationLoading || isClassificationFiltering" class="systems-table__filter-loading" role="status" aria-live="polite">
              <span>
                <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                {{ isClassificationLoading ? classificationLoadingMessage : 'Применяем фильтры…' }}
              </span>
            </div>
          </Transition>
          <table>
            <colgroup>
              <col class="changes-table__name-column" />
              <col />
              <col />
            </colgroup>
            <thead>
              <tr>
                <th rowspan="2">Название системы</th>
                <th colspan="2">Класс</th>
              </tr>
              <tr>
                <th>было</th>
                <th>стало</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="currentClassificationRows().length === 0">
                <td class="empty-table-cell" colspan="3">{{ classificationChangesEmptyMessage() }}</td>
              </tr>
              <tr v-for="row in visibleClassificationRows()" :key="row.id" tabindex="-1">
                <td class="changes-system-cell">
                  <a v-if="row.systemUrl" class="changes-system-link" :href="row.systemUrl" target="_blank" rel="noreferrer">
                    <span>{{ row.systemName }}</span>
                    <ExternalLink :size="13" :stroke-width="1.8" aria-hidden="true" />
                  </a>
                  <span v-else class="changes-system-name">{{ row.systemName }}</span>
                  <small v-if="row.constructionType === 'Тип не присвоен'" class="construction-type-status">Тип не присвоен</small>
                </td>
                <td :class="`changes-status-cell changes-status-cell--${classModifier(row.classBefore) || 'neutral'}`">
                  <span :class="`changes-status changes-status--${classModifier(row.classBefore) || 'neutral'}`">
                    <CircleCheck v-if="classModifier(row.classBefore) === 'recommended'" :size="15" aria-hidden="true" />
                    <TriangleAlert v-else-if="classModifier(row.classBefore) === 'allowed'" :size="15" aria-hidden="true" />
                    <X v-else-if="classModifier(row.classBefore) === 'forbidden'" :size="15" aria-hidden="true" />
                    <Plus v-else :size="15" aria-hidden="true" />
                    {{ row.classBefore }}
                  </span>
                </td>
                <td :class="`changes-status-cell changes-status-cell--${classModifier(row.classAfter) || 'neutral'}`">
                  <span :class="`changes-status changes-status--${classModifier(row.classAfter) || 'neutral'}`">
                    <CircleCheck v-if="classModifier(row.classAfter) === 'recommended'" :size="15" aria-hidden="true" />
                    <TriangleAlert v-else-if="classModifier(row.classAfter) === 'allowed'" :size="15" aria-hidden="true" />
                    <X v-else-if="classModifier(row.classAfter) === 'forbidden'" :size="15" aria-hidden="true" />
                    <Info v-else :size="15" aria-hidden="true" />
                    {{ row.classAfter }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer v-if="currentClassificationRows().length > 0" class="table-pagination changes-table-pagination">
          <span class="table-pagination__range">
            Показано {{ classificationRangeStart() }}–{{ classificationRangeEnd() }} из {{ currentClassificationRows().length }}
          </span>
          <div class="table-pagination__controls">
            <label>
              <span>Строк на странице</span>
              <select v-model="classificationPageSize" @change="changeClassificationPageSize">
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="all">Все</option>
              </select>
            </label>
            <div v-if="classificationPageSize !== 'all'" class="table-pagination__pages">
              <button type="button" :disabled="classificationPage === 1" aria-label="Предыдущая страница" @click="changeClassificationPage(classificationPage - 1)">‹</button>
              <strong>{{ classificationPage }} / {{ classificationPageCount() }}</strong>
              <button type="button" :disabled="classificationPage >= classificationPageCount()" aria-label="Следующая страница" @click="changeClassificationPage(classificationPage + 1)">›</button>
            </div>
          </div>
        </footer>
      </section>

      <section v-else-if="activePage === 'systems'" class="systems-page">
        <div class="changes-topbar systems-topbar">
          <div class="select-field">
            <span>Распоряжение</span>
            <div class="custom-select changes-order-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button changes-order-select__button" type="button" @click.stop="toggleSelect('order')">
                <CalendarDays :size="18" :stroke-width="1.8" aria-hidden="true" />
                <span>{{ selectedOrderName() }}</span>
                <ChevronDown class="changes-order-select__chevron" :size="18" :stroke-width="1.8" aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order.id"
                    class="custom-select__option"
                    :class="{ 'is-selected': order.id === selectedOrderId }"
                    type="button"
                    @click="selectOrder(order)"
                  >
                    {{ order.name }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <div class="changes-refresh-panel">
            <span>Последнее обновление: {{ formatOrderDateTime(systemsPageUpdatedAt()) }}</span>
            <button type="button" :disabled="isSystemsRefreshing" @click="refreshSystemsPage">
              <RefreshCw :class="{ 'is-spinning': isSystemsRefreshing }" :size="18" :stroke-width="1.8" aria-hidden="true" />
              {{ isSystemsRefreshing ? 'Обновление…' : 'Обновить' }}
            </button>
          </div>
        </div>

        <section class="systems-summary" aria-label="Сводка систем">
          <article class="systems-metric-card">
            <span class="systems-metric-card__accent" aria-hidden="true" />
            <span class="systems-metric-card__icon" aria-hidden="true">
              <Layers3 :size="38" :stroke-width="1.7" />
            </span>
            <span class="systems-metric-card__value">
              <strong>{{ systemCatalogStats.total }}</strong>
              <small>систем</small>
            </span>
            <Layers3 class="systems-metric-card__watermark" :size="118" :stroke-width="1.2" aria-hidden="true" />
          </article>

          <div class="status-stack systems-status-stack">
            <button class="status-card status-card--recommended" :class="{ 'is-selected': selectedSystemCatalogClass === 'Рекомендованная' }" type="button" @click="filterSystemsByClass('Рекомендованная')">
              <strong>{{ systemCatalogStats.recommended }}</strong>
              <span>Рекомендованных</span>
              <ChevronRight :size="19" aria-hidden="true" />
            </button>
            <button class="status-card status-card--allowed" :class="{ 'is-selected': selectedSystemCatalogClass === 'Разрешенная' }" type="button" @click="filterSystemsByClass('Разрешенная')">
              <strong>{{ systemCatalogStats.allowed }}</strong>
              <span>Разрешенных</span>
              <ChevronRight :size="19" aria-hidden="true" />
            </button>
            <button class="status-card status-card--forbidden" :class="{ 'is-selected': selectedSystemCatalogClass === 'Запрещенная' }" type="button" @click="filterSystemsByClass('Запрещенная')">
              <strong>{{ systemCatalogStats.forbidden }}</strong>
              <span>Запрещенных</span>
              <ChevronRight :size="19" aria-hidden="true" />
            </button>
          </div>

          <article class="systems-metric-card">
            <span class="systems-metric-card__accent" aria-hidden="true" />
            <span class="systems-metric-card__icon" aria-hidden="true">
              <UsersRound :size="38" :stroke-width="1.7" />
            </span>
            <span class="systems-metric-card__value">
              <strong>{{ systemCatalogStats.curators }}</strong>
              <small>кураторов</small>
            </span>
            <UsersRound class="systems-metric-card__watermark" :size="118" :stroke-width="1.2" aria-hidden="true" />
          </article>
        </section>

        <div class="systems-tools">
          <section class="filter-panel" aria-label="Тип строительства">
            <h2>Тип строительства</h2>
            <div class="type-tabs type-tabs--changes type-tabs--systems">
              <button
                v-for="type in systemsConstructionTypes"
                :key="type.name"
                class="type-tab"
                :class="{ 'type-tab--active': type.name === selectedConstructionType }"
                type="button"
                @click="selectConstructionType(type.name)"
              >
                <span>{{ type.name }}</span>
                <strong>{{ type.count }}</strong>
              </button>
            </div>
          </section>
        </div>

        <section class="system-type-panel" :class="{ 'is-open': isSystemTypesOpen }" aria-label="Тип системы">
          <button
            class="system-type-toggle"
            type="button"
            :aria-expanded="isSystemTypesOpen"
            @click="isSystemTypesOpen = !isSystemTypesOpen"
          >
            <span class="system-type-toggle__title">Тип системы</span>
            <span class="system-type-toggle__selected">Выбрано: {{ selectedSystemType.name }}</span>
            <i aria-hidden="true" />
          </button>

          <Transition name="system-type-body">
            <div v-if="isSystemTypesOpen" class="system-type-body">
              <div class="system-type-grid">
                <button
                  v-for="type in systemTypes"
                  :key="type.name"
                  class="system-type-card"
                  :class="{ 'is-active': type.name === selectedSystemType.name }"
                  type="button"
                  @click="selectSystemType(type)"
                >
                  <span class="system-type-card__image" aria-hidden="true">
                    <Layers3 :size="25" :stroke-width="1.6" />
                    <img v-if="type.imageUrl" :src="systemTypeImageSource(type)" alt="" loading="lazy" decoding="async" @error="hideBrokenSystemTypeImage" />
                  </span>
                  <span class="system-type-card__content">
                    <strong>{{ type.name }}</strong>
                    <span>{{ type.count }} систем</span>
                  </span>
                </button>
              </div>
            </div>
          </Transition>
        </section>

        <section class="table-toolbar systems-table-toolbar" aria-label="Управление таблицей">
          <label class="search-field systems-name-search" :class="{ 'is-filtered': systemCatalogSearch.trim() }">
            <span>Поиск системы</span>
            <span class="systems-search-control">
              <Search :size="18" :stroke-width="1.8" aria-hidden="true" />
              <input v-model="systemCatalogSearch" type="search" placeholder="Введите название системы" @input="scheduleSystemDocumentSearch" />
            </span>
          </label>

          <div class="select-field systems-class-filter" :style="{ '--class-accent': statusAccentColor(selectedSystemCatalogClass) }">
            <span>Класс</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'class', 'is-filtered': selectedSystemCatalogClass !== 'Все' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('class')">
                <span>{{ selectedSystemCatalogClass }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'class'" class="custom-select__menu">
                  <button
                    v-for="option in systemCatalogClassOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedSystemCatalogClass }"
                    type="button"
                    @click="selectedSystemCatalogClass = option; openedSelect = null; loadSystemDocuments(true)"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <div class="select-field">
            <span>Куратор</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'curator', 'is-filtered': selectedSystemCatalogCurator !== 'Все кураторы' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('curator')">
                <span>{{ selectedSystemCatalogCurator }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'curator'" class="custom-select__menu">
                  <button
                    v-for="option in systemCatalogCuratorOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedSystemCatalogCurator }"
                    type="button"
                    @click="selectedSystemCatalogCurator = option; openedSelect = null; loadSystemDocuments(true)"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <span class="systems-reset-slot">
            <Transition name="changes-reset">
              <button
                v-if="hasActiveSystemFilters"
                class="changes-reset-filters systems-reset-filters"
                type="button"
                title="Сбросить фильтры"
                aria-label="Сбросить фильтры"
                @click="resetSystemFilters"
              >
                <FunnelX :size="19" :stroke-width="1.8" aria-hidden="true" />
                <span class="systems-reset-filters__count" aria-hidden="true">{{ activeSystemFilterCount }}</span>
              </button>
            </Transition>
          </span>

          <button class="export-button" type="button" @click="exportSystemCatalog">
            Экспорт
            <img class="export-button__xlsx-icon" :src="xlsxFileIcon" alt="" aria-hidden="true" />
          </button>
        </section>

        <p v-if="systemCatalogError" class="table-message table-message--error">{{ systemCatalogError }}</p>

        <section class="comparison-table-controls" aria-label="Управление выбором для сравнения">
          <div class="comparison-table-controls__heading-wrap">
            <span class="comparison-table-controls__icon" aria-hidden="true">
              <Scale :size="24" :stroke-width="1.8" />
            </span>
            <div class="comparison-table-controls__heading">
              <strong>Добавление в сравнение</strong>
              <span>Отметьте нужные системы в таблице</span>
            </div>
          </div>
          <div class="comparison-table-controls__actions">
            <label class="toolbar-checkbox" :class="{ 'is-partial': someVisibleSystemsSelected }">
              <input
                type="checkbox"
                :checked="allVisibleSystemsSelected"
                :disabled="isBulkComparisonUpdating || currentSystemDocumentRows().length === 0"
                @change="toggleAllSystemComparisons"
              />
              <span aria-hidden="true" />
              <strong>Выбрать все</strong>
            </label>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'comparison-scope' }">
              <button class="custom-select__button" type="button" aria-label="Область применения выбора" :disabled="isBulkComparisonUpdating" @click.stop="toggleSelect('comparison-scope')">
                <span>{{ comparisonAllOrders ? 'Все распоряжения' : 'Текущее распоряжение' }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'comparison-scope'" class="custom-select__menu">
                  <button class="custom-select__option" :class="{ 'is-selected': !comparisonAllOrders }" type="button" @click="comparisonAllOrders = false; openedSelect = null">
                    Текущее распоряжение
                  </button>
                  <button class="custom-select__option" :class="{ 'is-selected': comparisonAllOrders }" type="button" @click="comparisonAllOrders = true; openedSelect = null">
                    Все распоряжения
                  </button>
                </div>
              </Transition>
            </div>
          </div>

        </section>

        <div class="systems-table systems-table--catalog">
          <Transition name="table-filter-loading">
            <div v-if="isSystemDocumentLoading || isSystemFiltering" class="systems-table__filter-loading" role="status" aria-live="polite">
              <span>
                <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                {{ isSystemDocumentLoading ? 'Загружаем список систем…' : 'Применяем фильтры…' }}
              </span>
            </div>
          </Transition>
          <table>
            <thead>
              <tr>
                <th>Шифр</th>
                <th>Название системы</th>
                <th>Класс</th>
                <th>Куратор</th>
                <th>Сравнение</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="currentSystemDocumentRows().length === 0">
                <td class="empty-table-cell" colspan="5">В таблице 3 этого распоряжения пока нет систем</td>
              </tr>
              <tr v-for="row in visibleSystemDocumentRows()" :key="row.id" tabindex="-1">
                <td class="systems-code-cell"><span>{{ row.code }}</span></td>
                <td class="systems-name-cell">
                  <a v-if="row.systemUrl" :href="row.systemUrl" target="_blank" rel="noreferrer">
                    {{ row.systemName }}
                  </a>
                  <span v-else>{{ row.systemName }}</span>
                </td>
                <td :class="`status-cell systems-catalog-status-cell systems-catalog-status-cell--${classModifier(row.systemClass)}`">
                  <span :class="`systems-class-badge systems-class-badge--${classModifier(row.systemClass)}`">
                    <i aria-hidden="true" />
                    {{ row.systemClass }}
                  </span>
                  <button class="status-cell__icon systems-history-icon" type="button" :aria-label="`История ${row.systemName}`" @click.stop="openSystemHistory(row)">
                    <Folder :size="19" :stroke-width="2" aria-hidden="true" />
                  </button>
                </td>
                <td class="systems-curator-cell">{{ row.curator }}</td>
                <td>
                  <label class="compare-checkbox" :class="{ 'is-pending': comparisonPendingIds.includes(row.id) }">
                    <input
                      type="checkbox"
                      :checked="row.comparisonSelected"
                      :disabled="comparisonPendingIds.includes(row.id)"
                      :aria-label="`Добавить ${row.systemName} в сравнение`"
                      @change="toggleSystemComparison(row, $event)"
                    />
                    <span class="compare-mark" :class="{ 'is-checked': row.comparisonSelected }" aria-hidden="true">
                      {{ row.comparisonSelected ? '✓' : '' }}
                    </span>
                  </label>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer v-if="currentSystemDocumentRows().length > 0" class="table-pagination">
          <span class="table-pagination__range">
            Показано {{ systemDocumentRangeStart() }}–{{ systemDocumentRangeEnd() }} из {{ currentSystemDocumentRows().length }}
          </span>
          <div class="table-pagination__controls">
            <label>
              <span>Строк на странице</span>
              <select v-model="systemDocumentPageSize" @change="changeSystemDocumentPageSize">
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="all">Все</option>
              </select>
            </label>
            <div v-if="systemDocumentPageSize !== 'all'" class="table-pagination__pages">
              <button type="button" :disabled="systemDocumentPage === 1" aria-label="Предыдущая страница" @click="changeSystemDocumentPage(systemDocumentPage - 1)">‹</button>
              <strong>{{ systemDocumentPage }} / {{ systemDocumentPageCount() }}</strong>
              <button type="button" :disabled="systemDocumentPage >= systemDocumentPageCount()" aria-label="Следующая страница" @click="changeSystemDocumentPage(systemDocumentPage + 1)">›</button>
            </div>
          </div>
        </footer>

      </section>

      <section v-else-if="activePage === 'classification'" class="classification-page">
        <div class="classification-topline">
          <div class="select-field">
            <span>Распоряжение</span>
            <div class="custom-select changes-order-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button changes-order-select__button" type="button" @click.stop="toggleSelect('order')">
                <CalendarDays :size="18" :stroke-width="1.8" aria-hidden="true" />
                <span>{{ selectedOrderName() }}</span>
                <ChevronDown class="changes-order-select__chevron" :size="18" :stroke-width="1.8" aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order.id"
                    class="custom-select__option"
                    :class="{ 'is-selected': order.id === selectedOrderId }"
                    type="button"
                    @click="selectOrder(order)"
                  >
                    {{ order.name }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <label class="search-field classification-search" :class="{ 'is-pending': isClassificationSearchPending }">
            <span>Поиск</span>
            <span class="systems-search-control">
              <Search :size="18" :stroke-width="1.8" aria-hidden="true" />
              <input
                v-model="classificationCatalogSearchInput"
                type="search"
                placeholder="Поиск по названию или коду ЕКН"
                :aria-busy="isClassificationSearchPending"
                @input="scheduleClassificationCatalogSearch"
              />
            </span>
          </label>

          <div class="classification-found">
            <span class="classification-found__icon" aria-hidden="true">
              <Layers3 :size="24" :stroke-width="1.8" />
            </span>
            <span>
              <small>Найдено систем</small>
              <strong>{{ classificationSystems.length }}</strong>
              <em>{{ isClassificationSearchPending ? 'Обновляем результаты…' : 'по выбранным критериям' }}</em>
            </span>
          </div>
        </div>

        <section class="filter-panel classification-construction" aria-label="Тип строительства">
          <h2>Тип строительства</h2>
          <div class="type-tabs type-tabs--changes type-tabs--systems">
            <button
              v-for="type in classificationCatalogConstructionTypes"
              :key="type.name"
              class="type-tab"
              :class="{ 'type-tab--active': type.name === selectedConstructionType }"
              type="button"
              @click="selectConstructionType(type.name)"
            >
              <span>{{ type.name }}</span>
              <strong>{{ type.count }}</strong>
            </button>
          </div>
        </section>

        <section class="system-type-panel classification-system-types" :class="{ 'is-open': isSystemTypesOpen }" aria-label="Тип системы">
          <button
            class="system-type-toggle"
            type="button"
            :aria-expanded="isSystemTypesOpen"
            @click="isSystemTypesOpen = !isSystemTypesOpen"
          >
            <span class="system-type-toggle__title">Тип системы</span>
            <span class="system-type-toggle__selected">Выбрано: {{ selectedSystemType.name }}</span>
            <i aria-hidden="true" />
          </button>

          <Transition name="system-type-body">
            <div v-if="isSystemTypesOpen" class="system-type-body">
              <div class="system-type-grid">
                <button
                  v-for="type in systemTypes"
                  :key="type.name"
                  class="system-type-card"
                  :class="{ 'is-active': type.name === selectedSystemType.name }"
                  type="button"
                  @click="selectSystemType(type)"
                >
                  <span class="system-type-card__image" aria-hidden="true">
                    <Layers3 :size="25" :stroke-width="1.6" />
                    <img v-if="type.imageUrl" :src="systemTypeImageSource(type)" alt="" loading="lazy" decoding="async" @error="hideBrokenSystemTypeImage" />
                  </span>
                  <span class="system-type-card__content">
                    <strong>{{ type.name }}</strong>
                    <span>{{ type.count }} систем</span>
                  </span>
                </button>
              </div>
            </div>
          </Transition>
        </section>

        <div class="classification-layout">
          <div class="classification-main">
            <section class="classification-results-toolbar" tabindex="-1" aria-label="Настройки отображения">
              <div class="classification-results-toolbar__count">
                <span class="classification-results-toolbar__icon" aria-hidden="true">
                  <Layers3 :size="18" :stroke-width="1.8" />
                </span>
                <span>
                  <small>Найдено систем</small>
                  <strong>{{ classificationSystems.length }}</strong>
                </span>
              </div>
              <span class="classification-results-toolbar__view-label">Вид</span>
              <div class="classification-view-toggle" aria-label="Вид списка">
                <button :class="{ 'is-active': classificationView === 'grid' }" type="button" aria-label="Карточки" @click="classificationView = 'grid'">
                  <Grid2X2 :size="18" :stroke-width="1.8" aria-hidden="true" />
                </button>
                <button :class="{ 'is-active': classificationView === 'list' }" type="button" aria-label="Список" @click="classificationView = 'list'">
                  <List :size="19" :stroke-width="1.8" aria-hidden="true" />
                </button>
              </div>
            </section>

            <section class="classification-cards" :class="{ 'is-list-view': classificationView === 'list' }" aria-label="Системы классификации" aria-live="polite">
            <p v-if="classificationCatalogError" class="table-message table-message--error classification-cards__message">
              {{ classificationCatalogError }}
            </p>
            <p v-else-if="isClassificationCatalogLoading" class="table-message classification-cards__message">
              Загрузка систем из таблицы 2...
            </p>
            <p v-else-if="classificationSystems.length === 0" class="table-message classification-cards__message">
              {{ classificationEmptyMessage }}
            </p>
            <div v-for="(row, rowIndex) in classificationSystemRows" :key="rowIndex" class="classification-card-row">
              <article
                v-for="system in row"
                :key="system.id"
                v-memo="[system.id === openedClassificationSystemId]"
                class="classification-card"
                :class="`classification-card--${classModifier(system.systemClass)}`"
              >
                <header class="classification-card__header">
                  <a :href="system.systemUrl || 'https://nav.tn.ru/systems/'" target="_blank" rel="noreferrer">{{ system.systemName }}</a>
                  <a class="classification-card__source" :href="system.systemUrl || 'https://nav.tn.ru/systems/'" target="_blank" rel="noreferrer" aria-label="Открыть на nav.tn.ru">
                    <Globe2 :size="15" :stroke-width="1.9" aria-hidden="true" />
                  </a>
                </header>

                <span class="classification-card__code">{{ system.code }}</span>
                <span class="classification-card__curator">Куратор: {{ system.curator || 'не указан' }}</span>
                <span :class="`classification-card__status classification-card__status--${classModifier(system.systemClass)}`">
                  <i aria-hidden="true" />
                  {{ system.systemClass }}
                </span>

                <button
                  class="classification-card__more"
                  :class="{ 'is-open': openedClassificationSystemId === system.id }"
                  type="button"
                  :aria-expanded="openedClassificationSystemId === system.id"
                  aria-label="Показать характеристики системы"
                  @click="toggleClassificationSystem(system.id)"
                >
                  <MoreHorizontal :size="19" :stroke-width="2" aria-hidden="true" />
                </button>
              </article>

              <div
                v-if="openedClassificationSystem && row.some((system) => system.id === openedClassificationSystemId)"
                class="classification-details-shell"
              >
                  <section class="classification-details">
                    <div class="classification-details__header">
                      <div class="classification-details__title">
                        <span aria-hidden="true"><ListFilter :size="19" :stroke-width="1.8" /></span>
                        <div>
                          <small>Характеристики системы</small>
                          <strong>{{ openedClassificationSystem.systemName }}</strong>
                        </div>
                      </div>
                      <div class="classification-details__actions">
                        <a v-if="openedClassificationSystem.systemUrl" :href="openedClassificationSystem.systemUrl" target="_blank" rel="noreferrer">
                          Открыть на nav.tn.ru
                          <ExternalLink :size="14" :stroke-width="1.9" aria-hidden="true" />
                        </a>
                        <button type="button" aria-label="Закрыть характеристики" @click="toggleClassificationSystem(openedClassificationSystem.id)">
                          <X :size="17" :stroke-width="2" aria-hidden="true" />
                        </button>
                      </div>
                    </div>
                    <table v-if="openedClassificationSystem.characteristics?.length">
                      <thead>
                        <tr>
                          <th>Наименование показателя</th>
                          <th>Значение</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="characteristic in openedClassificationSystem.characteristics" :key="`${openedClassificationSystem.id}-${characteristic.position}`">
                          <th>{{ characteristic.name }}</th>
                          <td>{{ characteristic.value }}</td>
                        </tr>
                      </tbody>
                    </table>
                    <div v-else class="classification-details__empty">
                      <span class="classification-details__empty-icon" aria-hidden="true">i</span>
                      <div class="classification-details__empty-copy">
                        <strong>Характеристики пока не загружены</strong>
                        <p>Запустите парсер, чтобы добавить показатели и значения системы.</p>
                      </div>
                      <button type="button" class="classification-details__empty-action" @click="setPage('settings')">
                        Перейти в настройки
                      </button>
                    </div>
                  </section>
              </div>
            </div>
            </section>

            <footer v-if="classificationSystems.length" class="table-pagination classification-cards-pagination">
              <span class="table-pagination__range">
                Показано {{ classificationCatalogRangeStart() }}–{{ classificationCatalogRangeEnd() }} из {{ classificationSystems.length }}
              </span>
              <div class="table-pagination__controls">
                <label>
                  <span>Карточек на странице</span>
                  <select v-model="classificationCatalogPageSize" @change="changeClassificationCatalogPageSize">
                    <option value="50">50</option>
                    <option value="100">100</option>
                    <option value="200">200</option>
                    <option value="all">Все</option>
                  </select>
                </label>
                <div v-if="classificationCatalogPageSize !== 'all'" class="table-pagination__pages">
                  <button type="button" :disabled="classificationCatalogPage === 1" aria-label="Предыдущая страница" @click="changeClassificationCatalogPage(classificationCatalogPage - 1)">‹</button>
                  <strong>{{ classificationCatalogPage }} / {{ classificationCatalogPageCount() }}</strong>
                  <button type="button" :disabled="classificationCatalogPage >= classificationCatalogPageCount()" aria-label="Следующая страница" @click="changeClassificationCatalogPage(classificationCatalogPage + 1)">›</button>
                </div>
              </div>
            </footer>
          </div>

          <aside class="classification-sidebar" aria-label="Фильтры классификации">
            <header class="classification-sidebar__header">
              <div class="classification-sidebar__heading">
                <span class="classification-sidebar__heading-icon" aria-hidden="true">
                  <ListFilter :size="18" :stroke-width="1.9" />
                </span>
                <div>
                  <strong>Фильтры</strong>
                  <span>{{ classificationSystems.length }} из {{ classificationBaseSystems.length }} систем</span>
                </div>
              </div>
              <button
                type="button"
                :disabled="selectedClassificationFilterCount === 0"
                aria-label="Сбросить боковые фильтры"
                @click="clearClassificationFilters"
              >
                <RefreshCw :size="14" :stroke-width="1.8" aria-hidden="true" />
                Сбросить
              </button>
            </header>
            <div class="classification-sidebar__tools">
              <label class="classification-sidebar__search">
                <Search :size="15" :stroke-width="1.8" aria-hidden="true" />
                <input v-model="classificationFilterSearch" type="search" placeholder="Найти характеристику" />
                <button v-if="classificationFilterSearch" type="button" aria-label="Очистить поиск фильтров" @click="classificationFilterSearch = ''">
                  <X :size="14" :stroke-width="2" aria-hidden="true" />
                </button>
              </label>
              <button
                class="classification-sidebar__collapse"
                type="button"
                :disabled="!openedClassificationFilter"
                @click="collapseClassificationFilters"
              >
                <ChevronUp :size="14" :stroke-width="1.9" aria-hidden="true" />
                Свернуть все
              </button>
            </div>
            <div v-if="selectedClassificationFilterCount" class="classification-sidebar__chips" aria-label="Выбранные фильтры">
              <button
                v-for="(value, name) in selectedClassificationFilters"
                :key="name"
                type="button"
                :title="`Убрать фильтр «${name}»`"
                @click="removeClassificationFilter(name)"
              >
                <span>{{ name }}: <strong>{{ value }}</strong></span>
                <X :size="13" :stroke-width="2" aria-hidden="true" />
              </button>
            </div>
            <p v-if="classificationFilterGroups.length === 0" class="table-message classification-sidebar__empty">
              Для выбранного типа нет доступных характеристик.
            </p>
            <p v-else-if="visibleClassificationFilterGroups.length === 0" class="table-message classification-sidebar__empty">
              Характеристики не найдены.
            </p>
            <div
              v-for="filter in visibleClassificationFilterGroups"
              :key="filter"
              class="classification-sidebar__group"
              :class="{ 'is-open': openedClassificationFilter === filter, 'is-selected': selectedClassificationFilters[filter] }"
            >
              <button
                class="classification-sidebar__item"
                type="button"
                @click="toggleClassificationFilter(filter)"
              >
                <span class="classification-sidebar__label">
                  <strong>{{ filter }}</strong>
                  <small v-if="selectedClassificationFilters[filter]">{{ selectedClassificationFilters[filter] }}</small>
                </span>
                <i aria-hidden="true" />
              </button>
              <Transition name="classification-filter-options">
                <div v-if="openedClassificationFilter === filter" class="classification-sidebar__options">
                  <button type="button" :class="{ 'is-selected': !selectedClassificationFilters[filter] }" @click="selectClassificationFilter(filter, '')">
                    <span>Все</span>
                    <small>{{ classificationFilterAvailableCount(filter) }}</small>
                  </button>
                  <button
                    v-for="option in classificationFilterOptions(filter)"
                    :key="option"
                    type="button"
                    :class="{ 'is-selected': selectedClassificationFilters[filter] === option }"
                    @click="selectClassificationFilter(filter, option)"
                  >
                    <span>{{ option }}</span>
                    <small>{{ classificationFilterOptionCount(filter, option) }}</small>
                  </button>
                </div>
              </Transition>
            </div>
          </aside>
        </div>
      </section>

      <section v-else-if="activePage === 'comparison'" class="comparison-page">
        <section class="comparison-controls">
          <div class="comparison-controls__intro">
            <span class="comparison-controls__intro-icon" aria-hidden="true">
              <Scale :size="24" :stroke-width="1.8" />
            </span>
            <div>
              <h1>Сравнение систем</h1>
              <p>Выберите распоряжения, изменения классов систем по которым нужно сравнить.</p>
            </div>
          </div>

          <div class="comparison-controls__orders">
            <span class="comparison-controls__label">Сравниваемые распоряжения</span>
            <div class="comparison-controls__row">
              <div
                v-for="(orderId, index) in comparisonOrderIds"
                :key="orderId"
                class="custom-select comparison-order"
                :class="{
                  'is-open': openedSelect === `comparison-${index}`,
                  'is-dragging': draggedComparisonOrderId === orderId,
                  'is-drop-target': comparisonDropIndex === index && draggedComparisonOrderId !== orderId,
                }"
                @dragenter.prevent="enterComparisonOrderDrop(index)"
                @dragover.prevent
                @drop.prevent="dropComparisonOrder(index)"
              >
                <div class="comparison-order__control">
                  <button class="custom-select__button" type="button" @click.stop="toggleSelect(`comparison-${index}`)">
                    <span class="comparison-order__number">{{ index + 1 }}</span>
                    <span class="comparison-order__name">{{ comparisonOrderName(orderId) }}</span>
                    <ChevronDown class="comparison-order__chevron" :size="17" :stroke-width="1.8" aria-hidden="true" />
                  </button>
                  <span
                    class="comparison-order__drag"
                    draggable="true"
                    :aria-grabbed="draggedComparisonOrderId === orderId"
                    title="Перетащите, чтобы изменить порядок"
                    @dragstart="startComparisonOrderDrag($event, orderId)"
                    @dragend="endComparisonOrderDrag"
                  >
                    <GripVertical :size="18" :stroke-width="2" aria-hidden="true" />
                  </span>
                </div>
                <Transition name="select-menu">
                  <div v-if="openedSelect === `comparison-${index}`" class="custom-select__menu comparison-order-menu">
                    <button
                      v-for="option in comparisonOrderOptions(orderId)"
                      :key="option.id"
                      class="custom-select__option"
                      :class="{ 'is-selected': option.id === orderId }"
                      type="button"
                      @click="selectComparisonOrder(index, option)"
                    >
                      {{ option.name }}
                    </button>
                    <span class="comparison-order-menu__separator" aria-hidden="true" />
                    <button class="comparison-order-menu__delete" type="button" @click="removeComparisonOrder(orderId)">
                      <Trash2 :size="17" :stroke-width="1.8" aria-hidden="true" />
                      <span>Удалить распоряжение</span>
                    </button>
                  </div>
                </Transition>
              </div>

              <div class="comparison-add-wrap">
                <div class="custom-select comparison-add-select" :class="{ 'is-open': openedSelect === 'comparison-add' }">
                  <button
                    class="comparison-add-button"
                    type="button"
                    :disabled="availableComparisonOrders().length === 0"
                    aria-label="Добавить распоряжение"
                    :title="availableComparisonOrders().length ? 'Добавить распоряжение' : 'Все распоряжения уже добавлены'"
                    @click.stop="toggleSelect('comparison-add')"
                  >
                    <Plus :size="20" :stroke-width="2" aria-hidden="true" />
                    <span>Добавить распоряжение</span>
                  </button>
                  <Transition name="select-menu">
                    <div v-if="openedSelect === 'comparison-add' && availableComparisonOrders().length" class="custom-select__menu comparison-add-menu">
                      <button v-for="order in availableComparisonOrders()" :key="order.id" class="custom-select__option" type="button" @click="addComparisonOrder(order)">
                        {{ order.name }}
                      </button>
                    </div>
                  </Transition>
                </div>
              </div>
            </div>
          </div>
        </section>

        <p v-if="comparisonError" class="table-message table-message--error">{{ comparisonError }}</p>
        <p v-else-if="isComparisonLoading" class="table-message">Загрузка сравнения...</p>

        <section class="comparison-toolbar" aria-label="Управление сравнением">
          <div class="comparison-toolbar__summary">
            <span class="comparison-toolbar__icon" aria-hidden="true"><Layers3 :size="22" :stroke-width="1.8" /></span>
            <span>
              <small>Систем в таблице</small>
              <strong>{{ comparisonRows().length }} <em>шт.</em></strong>
            </span>
          </div>
          <div class="comparison-toolbar__actions">
            <div class="custom-select comparison-difference-filter" :class="{ 'is-open': openedSelect === 'comparison-filter', 'is-active': comparisonOnlyDifferences }">
              <button class="comparison-difference-toggle" type="button" aria-label="Фильтр сравнения" @click.stop="toggleSelect('comparison-filter')">
                <Repeat2 v-if="comparisonOnlyDifferences" :size="17" :stroke-width="1.8" aria-hidden="true" />
                <Layers3 v-else :size="17" :stroke-width="1.8" aria-hidden="true" />
                <span>{{ comparisonOnlyDifferences ? 'Только различия' : 'Все системы' }}</span>
                <ChevronDown class="comparison-difference-chevron" :size="16" :stroke-width="1.8" aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'comparison-filter'" class="custom-select__menu comparison-difference-menu">
                  <button type="button" :class="{ 'is-selected': !comparisonOnlyDifferences }" @click="selectComparisonDifferenceFilter(false)">
                    <Layers3 :size="16" :stroke-width="1.8" aria-hidden="true" />
                    <span>Все системы</span>
                  </button>
                  <button type="button" :class="{ 'is-selected': comparisonOnlyDifferences }" @click="selectComparisonDifferenceFilter(true)">
                    <Repeat2 :size="16" :stroke-width="1.8" aria-hidden="true" />
                    <span>Только различия</span>
                  </button>
                </div>
              </Transition>
            </div>
            <div class="custom-select comparison-sort" :class="{ 'is-open': openedSelect === 'comparison-sort' }">
              <button class="comparison-sort__button" type="button" aria-label="Сортировка сравнения" @click.stop="toggleSelect('comparison-sort')">
                <ListFilter v-if="comparisonSort === 'differences-first'" :size="17" :stroke-width="1.8" aria-hidden="true" />
                <ArrowDownUp v-else :size="17" :stroke-width="1.8" aria-hidden="true" />
                <span>{{ comparisonSortLabel() }}</span>
                <ChevronDown class="comparison-sort__chevron" :size="16" :stroke-width="1.8" aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'comparison-sort'" class="custom-select__menu comparison-sort__menu">
                  <button type="button" :class="{ 'is-selected': comparisonSort === 'differences-first' }" @click="selectComparisonSort('differences-first')">
                    <ListFilter :size="16" :stroke-width="1.8" aria-hidden="true" />
                    <span>Различия сначала</span>
                  </button>
                  <button type="button" :class="{ 'is-selected': comparisonSort === 'name-asc' }" @click="selectComparisonSort('name-asc')">
                    <ArrowDownUp :size="16" :stroke-width="1.8" aria-hidden="true" />
                    <span>Без префикса (А–Я)</span>
                  </button>
                  <button type="button" :class="{ 'is-selected': comparisonSort === 'name-desc' }" @click="selectComparisonSort('name-desc')">
                    <ArrowDownUp :size="16" :stroke-width="1.8" aria-hidden="true" />
                    <span>Без префикса (Я–А)</span>
                  </button>
                </div>
              </Transition>
            </div>
            <button class="export-button comparison-export-button" type="button" :disabled="comparisonRows().length === 0" @click="exportComparisonTable">
              <span>Экспорт</span>
              <img class="export-button__xlsx-icon" :src="xlsxFileIcon" alt="" aria-hidden="true" />
            </button>
          </div>
        </section>

        <div class="systems-table comparison-table">
          <table>
            <thead>
              <tr>
                <th>Название системы <ArrowDownUp :size="15" :stroke-width="1.7" aria-hidden="true" /></th>
                <th v-for="(orderId, index) in comparisonOrderIds" :key="orderId">
                  <span class="comparison-table__order"><i>{{ index + 1 }}</i>{{ comparisonOrderName(orderId) }}</span>
                </th>
                <th>Действие</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="comparisonRows().length === 0">
                <td class="empty-table-cell" :colspan="comparisonOrderIds.length + 2">
                  В выбранных распоряжениях пока нет данных таблицы 2
                </td>
              </tr>
              <tr v-for="row in visibleComparisonRows()" :key="row.key" :class="{ 'has-difference': comparisonRowHasDifference(row) }">
                <td class="comparison-system-cell">
                  <span>{{ row.name }}</span>
                  <small v-if="comparisonRowHasDifference(row)">
                    <Repeat2 :size="12" :stroke-width="1.9" aria-hidden="true" />
                    Есть различия
                  </small>
                </td>
                <td
                  v-for="(_, index) in comparisonOrderIds"
                  :key="`${row.name}-${index}`"
                  :class="`comparison-value-cell comparison-value-cell--${classModifier(comparisonValue(row, index)) || 'empty'}`"
                >
                  <span :class="`comparison-status comparison-status--${classModifier(comparisonValue(row, index)) || 'empty'}`">
                    <CircleCheck v-if="classModifier(comparisonValue(row, index)) === 'recommended'" :size="14" :stroke-width="2.2" aria-hidden="true" />
                    <TriangleAlert v-else-if="classModifier(comparisonValue(row, index)) === 'allowed'" :size="15" :stroke-width="2" aria-hidden="true" />
                    <X v-else-if="classModifier(comparisonValue(row, index)) === 'forbidden'" :size="14" :stroke-width="2.2" aria-hidden="true" />
                    {{ comparisonValue(row, index) }}
                  </span>
                </td>
                <td>
                  <button class="comparison-delete-button" type="button" aria-label="Удалить из сравнения" @click="hideComparisonRow(row)">
                    <Trash2 :size="17" :stroke-width="1.8" aria-hidden="true" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer v-if="comparisonRows().length > 0" class="table-pagination comparison-table-pagination">
          <span class="table-pagination__range">
            Показано {{ comparisonRangeStart() }}–{{ comparisonRangeEnd() }} из {{ comparisonRows().length }}
          </span>
          <div class="table-pagination__controls">
            <label>
              <span>Строк на странице</span>
              <select v-model="comparisonPageSize" @change="changeComparisonPageSize">
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="all">Все</option>
              </select>
            </label>
            <div v-if="comparisonPageSize !== 'all'" class="table-pagination__pages">
              <button type="button" :disabled="comparisonPage === 1" aria-label="Предыдущая страница" @click="changeComparisonPage(comparisonPage - 1)">‹</button>
              <strong>{{ comparisonPage }} / {{ comparisonPageCount() }}</strong>
              <button type="button" :disabled="comparisonPage >= comparisonPageCount()" aria-label="Следующая страница" @click="changeComparisonPage(comparisonPage + 1)">›</button>
            </div>
          </div>
        </footer>
      </section>

      <section v-else-if="activePage === 'settings'" class="settings-page">
        <section class="settings-section parser-settings" aria-labelledby="parser-settings-title">
          <div class="parser-settings__main">
            <div class="parser-settings__identity">
              <span class="parser-settings__icon" aria-hidden="true">
                <RefreshCw :size="42" :stroke-width="1.8" />
              </span>
              <div class="parser-settings__content">
                <h1 id="parser-settings-title">Парсинг навигатора</h1>
                <div class="parser-settings__schedule">
                  <span>Автозапуск каждые <strong>{{ navParserIntervalDays }}</strong> дн.</span>
                  <small>{{ navParserNextRunLabel() }}</small>
                </div>
              </div>
            </div>
            <div class="parser-settings__controls">
              <button class="import-button parser-settings__button" type="button" :disabled="isNavParsing" @click="runNavParser">
                {{ isNavParsing ? 'Парсинг выполняется…' : 'Запустить парсер' }}
              </button>
            </div>
          </div>
          <section v-if="navParserProgress.running || navParserProgress.logs.length" class="parser-progress" aria-live="polite">
            <header class="parser-progress__header">
              <div>
                <span class="parser-progress__state" :class="{ 'is-running': navParserProgress.running }">
                  <RefreshCw v-if="navParserProgress.running" :size="15" :stroke-width="2" aria-hidden="true" />
                  <CircleCheck v-else-if="navParserProgress.percent === 100" :size="15" :stroke-width="2" aria-hidden="true" />
                  <Info v-else :size="15" :stroke-width="2" aria-hidden="true" />
                  {{ navParserProgress.stage }}
                </span>
                <strong>{{ navParserProgress.message }}</strong>
              </div>
              <b>{{ navParserProgress.percent }}%</b>
            </header>
            <div
              class="parser-progress__track"
              role="progressbar"
              :aria-valuenow="navParserProgress.percent"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="`Прогресс парсера: ${navParserProgress.percent}%`"
            >
              <span :style="{ width: `${navParserProgress.percent}%` }" />
            </div>
            <div class="parser-progress__stats">
              <span><small>Обработано</small><strong>{{ navParserProgress.processed }} / {{ navParserProgress.total }}</strong></span>
              <span><small>Найдено</small><strong>{{ navParserProgress.found }}</strong></span>
              <span><small>Обновлено</small><strong>{{ navParserProgress.updated }}</strong></span>
              <span><small>Не найдено</small><strong>{{ navParserProgress.notFound }}</strong></span>
              <span><small>Ошибки</small><strong>{{ navParserProgress.failed }}</strong></span>
            </div>
            <section class="parser-log">
              <button
                class="parser-log__toggle"
                type="button"
                :aria-expanded="isNavParserLogOpen"
                @click="isNavParserLogOpen = !isNavParserLogOpen"
              >
                <span>Журнал выполнения</span>
                <span class="parser-log__toggle-meta">
                  <small>{{ navParserProgress.logs.length }} записей</small>
                  <ChevronDown :size="16" :stroke-width="2" :class="{ 'is-open': isNavParserLogOpen }" aria-hidden="true" />
                </span>
              </button>
              <Transition name="parser-log-body">
                <div v-if="isNavParserLogOpen" class="parser-log__body">
                  <div class="parser-log__entries">
                    <p v-for="(entry, index) in navParserLogsNewestFirst" :key="`${entry.time}-${index}`" :class="`parser-log__entry parser-log__entry--${entry.level}`">
                      <time>{{ formatNavParserLogTime(entry.time) }}</time>
                      <i aria-hidden="true" />
                      <span>{{ entry.message }}</span>
                    </p>
                  </div>
                </div>
              </Transition>
            </section>
          </section>
          <section class="parser-history" aria-labelledby="parser-history-title">
            <header class="parser-history__header">
              <div>
                <h2 id="parser-history-title">Журнал всех запусков</h2>
              </div>
              <span>{{ navParserRuns.length }} запусков</span>
            </header>
            <p v-if="!navParserRuns.length" class="parser-history__empty">Завершённых запусков пока нет.</p>
            <div v-else class="parser-history__list">
              <article v-for="run in navParserRuns" :key="run.id" class="parser-history__run" :class="`is-${run.status}`">
                <button
                  class="parser-history__run-toggle"
                  type="button"
                  :aria-expanded="openedNavParserRunId === run.id"
                  @click="openedNavParserRunId = openedNavParserRunId === run.id ? null : run.id"
                >
                  <span class="parser-history__status" aria-hidden="true">
                    <CircleCheck v-if="run.status === 'completed'" :size="17" :stroke-width="2" />
                    <TriangleAlert v-else :size="17" :stroke-width="2" />
                  </span>
                  <span class="parser-history__identity">
                    <strong>{{ formatNavParserRunDate(run.startedAt) }}</strong>
                    <small>{{ navParserSourceLabel(run.source) }} · {{ formatNavParserRunDuration(run) }}</small>
                  </span>
                  <span class="parser-history__summary">
                    <span>Обновлено <b>{{ run.updated }}</b></span>
                    <span>Не найдено <b>{{ run.notFound }}</b></span>
                    <span>Ошибки <b>{{ run.failed }}</b></span>
                  </span>
                  <ChevronDown :size="18" :stroke-width="2" :class="{ 'is-open': openedNavParserRunId === run.id }" aria-hidden="true" />
                </button>
                <Transition name="parser-log-body">
                  <div v-if="openedNavParserRunId === run.id" class="parser-log__body parser-history__log-body">
                    <div class="parser-log__entries">
                      <p v-for="(entry, index) in navParserRunLogsNewestFirst(run)" :key="`${run.id}-${entry.time}-${index}`" :class="`parser-log__entry parser-log__entry--${entry.level}`">
                        <time>{{ formatNavParserLogTime(entry.time) }}</time>
                        <i aria-hidden="true" />
                        <span>{{ entry.message }}</span>
                      </p>
                    </div>
                  </div>
                </Transition>
              </article>
            </div>
          </section>
          <section class="parser-options-section">
            <button
              class="parser-options__toggle"
              type="button"
              :aria-expanded="isNavParserSettingsOpen"
              @click="isNavParserSettingsOpen = !isNavParserSettingsOpen"
            >
              <span>Настройки</span>
              <ChevronDown :class="{ 'is-open': isNavParserSettingsOpen }" :size="17" :stroke-width="1.9" aria-hidden="true" />
            </button>
            <Transition name="parser-options">
              <div v-if="isNavParserSettingsOpen" class="parser-options__shell">
                <div class="parser-options">
                  <div class="parser-options__heading">
                    <div>
                      <strong>Расширенные настройки</strong>
                    </div>
                    <button class="parser-options__save" type="button" :disabled="isNavSettingsSaving" @click="saveNavParserSettings">
                      {{ isNavSettingsSaving ? 'Сохранение…' : 'Сохранить настройки' }}
                    </button>
                  </div>
                  <div class="parser-options__grid">
                    <label>
                      <span>Период обновления</span>
                      <small>Через сколько дней парсер сам запустится снова.</small>
                      <span class="parser-options__input"><input v-model.number="navParserIntervalDays" type="number" min="1" max="365" /><em>дней</em></span>
                    </label>
                    <label>
                      <span>Параллельные запросы</span>
                      <small>Сколько страниц nav.tn.ru загружать одновременно. Рекомендуемое значение — 4.</small>
                      <span class="parser-options__input"><input v-model.number="navParserWorkerCount" type="number" min="1" max="10" /><em>шт.</em></span>
                    </label>
                    <label>
                      <span>Тайм-аут запроса</span>
                      <small>Сколько ждать ответа сайта, прежде чем считать запрос неудачным. Рекомендуется 35 секунд.</small>
                      <span class="parser-options__input"><input v-model.number="navParserRequestTimeout" type="number" min="5" max="120" /><em>сек.</em></span>
                    </label>
                    <label>
                      <span>Количество попыток</span>
                      <small>Сколько раз повторить запрос, если сайт временно не ответил. Безопасное значение — 3.</small>
                      <span class="parser-options__input"><input v-model.number="navParserRetryAttempts" type="number" min="1" max="5" /><em>раз</em></span>
                    </label>
                    <label>
                      <span>Задержка между попытками</span>
                      <small>Пауза перед повторным запросом.</small>
                      <span class="parser-options__input"><input v-model.number="navParserRetryDelay" type="number" min="1" max="30" /><em>сек.</em></span>
                    </label>
                    <label class="parser-options__switch">
                      <span>
                        <strong>Резервный поиск</strong>
                        <small>Если система не найдена в общем каталоге, искать её отдельно через поиск nav.tn.ru.</small>
                      </span>
                      <input v-model="navParserFallbackSearch" type="checkbox" />
                    </label>
                  </div>
                </div>
              </div>
            </Transition>
          </section>
          <p v-if="navSettingsError" class="table-message table-message--error">{{ navSettingsError }}</p>
          <p v-else-if="navSettingsMessage" class="table-message table-message--success">{{ navSettingsMessage }}</p>
          <p v-if="navParseError" class="table-message table-message--error">{{ navParseError }}</p>
          <p v-else-if="navParseMessage" class="table-message table-message--success">{{ navParseMessage }}</p>
          <details v-if="navParseNotFound.length" class="parser-settings__not-found">
            <summary>Не найденные на nav.tn.ru системы ({{ navParseNotFound.length }})</summary>
            <ul>
              <li v-for="systemName in navParseNotFound" :key="systemName">{{ systemName }}</li>
            </ul>
          </details>
          <details v-if="navParseFailedSystems.length" class="parser-settings__not-found parser-settings__failed">
            <summary>Ошибки загрузки систем ({{ navParseFailedSystems.length }})</summary>
            <ul>
              <li v-for="systemName in navParseFailedSystems" :key="systemName">{{ systemName }}</li>
            </ul>
          </details>
        </section>

        <section class="settings-section orders-settings" aria-labelledby="orders-db-title">
          <div class="settings-section__header settings-orders-header">
            <div class="settings-orders-heading">
              <span class="settings-orders-heading__icon" aria-hidden="true">
                <Database :size="22" :stroke-width="1.8" />
              </span>
              <div>
                <span class="settings-section__eyebrow">Распоряжения</span>
                <span class="settings-orders-heading__title">
                  <h2 id="orders-db-title">Управление базами данных</h2>
                  <em>Баз: {{ orders.length }}</em>
                </span>
              </div>
            </div>
            <button class="settings-create-order" type="button" @click="createOrder">
              <Plus :size="18" :stroke-width="1.8" aria-hidden="true" />
              Создать новую БД
            </button>
          </div>

          <div class="systems-table settings-orders-table settings-table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Распоряжение</th>
                  <th>Дата создания</th>
                  <th>Последняя актуализация</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="order in orders" :key="order.id">
                  <td>
                    <span class="settings-order-name">
                      <i aria-hidden="true"><Database :size="17" :stroke-width="1.8" /></i>
                      <input
                        v-model="order.name"
                        class="order-name-input"
                        type="text"
                        aria-label="Название распоряжения"
                        @input="scheduleOrderRename(order)"
                        @blur="saveOrderName(order)"
                        @keyup.enter="($event.target as HTMLInputElement).blur()"
                      />
                    </span>
                  </td>
                  <td class="settings-order-date"><span>Создана</span><strong>{{ formatOrderDateTime(order.createdAt) }}</strong></td>
                  <td class="settings-order-date settings-order-date--updated"><span>Обновлена</span><strong>{{ formatOrderDateTime(order.updatedAt) }}</strong></td>
                  <td class="settings-order-actions">
                    <button class="settings-order-menu-button" type="button" aria-label="Действия с распоряжением" @click.stop="settingsOrderMenuId = settingsOrderMenuId === order.id ? null : order.id">
                      <EllipsisVertical :size="19" :stroke-width="1.9" aria-hidden="true" />
                    </button>
                    <Transition name="select-menu">
                      <div v-if="settingsOrderMenuId === order.id" class="settings-order-menu">
                        <button type="button" @click="settingsOrderMenuId = null; deleteOrder(order)">
                          <Trash2 :size="16" :stroke-width="1.8" aria-hidden="true" />
                          Удалить БД
                        </button>
                      </div>
                    </Transition>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <p v-if="ordersError" class="table-message table-message--error">{{ ordersError }}</p>
          <p v-else-if="isOrdersLoading" class="table-message">Загрузка распоряжений...</p>
        </section>

        <section class="settings-section settings-editor" aria-labelledby="edit-db-title">
          <div class="settings-section__header">
            <h2 id="edit-db-title">Редактирование БД</h2>
          </div>

          <div class="settings-edit-select">
            <span>Выбрать БД распоряжения для редактирования</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'settings-order' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('settings-order')">
                <span>{{ selectedOrderName() }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'settings-order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order.id"
                    class="custom-select__option"
                    :class="{ 'is-selected': order.id === selectedOrderId }"
                    type="button"
                    @click="selectOrder(order)"
                  >
                    {{ order.name }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <section class="settings-table-block settings-classification-block" aria-label="Таблица 1">
            <div class="settings-table-toolbar">
              <span>Таблица 1</span>
              <label class="settings-search">
                <input
                  v-model="tableSearch"
                  type="search"
                  placeholder="Поиск по названию или ЕКН"
                  @input="settingsClassificationPage = 1"
                />
              </label>
              <button class="import-button" type="button" :disabled="isClassificationLoading" @click="openTableImport">
                <CloudUpload :size="17" :stroke-width="1.8" aria-hidden="true" />
                {{ isClassificationLoading ? 'Импорт...' : 'Импортировать таблицу' }}
              </button>
              <input
                ref="importFileInput"
                class="visually-hidden-input"
                type="file"
                accept=".xlsx"
                @change="importTableFile"
              />
            </div>

            <p v-if="classificationError" class="table-message table-message--error">{{ classificationError }}</p>

            <div class="systems-table settings-data-table settings-paginated-table settings-classification-table">
              <table>
                <thead>
                  <tr>
                    <th rowspan="2">Название системы</th>
                    <th colspan="2">
                      <span class="settings-class-header">
                        <span>Класс</span>
                        <button
                          type="button"
                          :class="{ 'is-unlocked': isSettingsClassificationUnlocked }"
                          :aria-label="isSettingsClassificationUnlocked ? 'Заблокировать редактирование' : 'Разрешить редактирование'"
                          :title="isSettingsClassificationUnlocked ? 'Редактирование разрешено' : 'Редактирование заблокировано'"
                          @click="isSettingsClassificationUnlocked = !isSettingsClassificationUnlocked"
                        >
                          <Unlock v-if="isSettingsClassificationUnlocked" :size="16" :stroke-width="2" aria-hidden="true" />
                          <Lock v-else :size="16" :stroke-width="2" aria-hidden="true" />
                        </button>
                      </span>
                    </th>
                  </tr>
                  <tr>
                    <th><span class="settings-class-heading settings-class-heading--before">Было</span></th>
                    <th><span class="settings-class-heading settings-class-heading--after">Стало</span></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="currentSettingsClassificationRows().length === 0">
                    <td class="empty-table-cell" colspan="3">В этом распоряжении пока нет данных таблицы 1</td>
                  </tr>
                  <tr v-for="row in visibleSettingsClassificationRows()" :key="`settings-${row.id}`">
                    <td>
                      <input
                        v-model="row.systemName"
                        class="settings-cell-input"
                        type="text"
                        :disabled="!isSettingsClassificationUnlocked"
                        aria-label="Название системы"
                        @input="scheduleClassificationRowSave(row)"
                        @blur="saveClassificationRow(row)"
                      />
                    </td>
                    <td :class="classModifier(row.classBefore) && `status-cell status-cell--${classModifier(row.classBefore)}`">
                      <select v-model="row.classBefore" :disabled="!isSettingsClassificationUnlocked" :class="`settings-cell-select settings-cell-select--${classModifier(row.classBefore) || 'new'}`" aria-label="Класс было" @change="saveClassificationRow(row)">
                        <option value="Новая система">Новая система</option>
                        <option v-for="option in classOptions" :key="`before-${option}`" :value="option">{{ option }}</option>
                      </select>
                    </td>
                    <td :class="classModifier(row.classAfter) && `status-cell status-cell--${classModifier(row.classAfter)}`">
                      <select v-model="row.classAfter" :disabled="!isSettingsClassificationUnlocked" :class="`settings-cell-select settings-cell-select--${classModifier(row.classAfter) || 'new'}`" aria-label="Класс стало" @change="saveClassificationRow(row)">
                        <option v-for="option in classOptions" :key="`after-${option}`" :value="option">{{ option }}</option>
                      </select>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <footer v-if="currentSettingsClassificationRows().length > 0" class="table-pagination settings-table-pagination">
              <span class="table-pagination__range">
                Показано {{ settingsClassificationRangeStart() }}–{{ settingsClassificationRangeEnd() }} из {{ currentSettingsClassificationRows().length }}
              </span>
              <div class="table-pagination__controls">
                <label>
                  <span>Записей на странице</span>
                  <select v-model="settingsClassificationPageSize" @change="changeSettingsClassificationPageSize">
                    <option value="10">10</option>
                    <option value="20">20</option>
                    <option value="50">50</option>
                    <option value="100">100</option>
                    <option value="all">Все</option>
                  </select>
                </label>
                <div v-if="settingsClassificationPageSize !== 'all'" class="table-pagination__pages">
                  <button type="button" :disabled="settingsClassificationPage === 1" aria-label="Предыдущая страница" @click="changeSettingsClassificationPage(settingsClassificationPage - 1)">‹</button>
                  <strong>{{ settingsClassificationPage }} / {{ settingsClassificationPageCount() }}</strong>
                  <button type="button" :disabled="settingsClassificationPage >= settingsClassificationPageCount()" aria-label="Следующая страница" @click="changeSettingsClassificationPage(settingsClassificationPage + 1)">›</button>
                </div>
              </div>
            </footer>
          </section>

          <section class="settings-table-block settings-system-catalog-block" aria-label="Таблица 2">
            <div class="settings-table-toolbar">
              <span>Таблица 2</span>
              <label class="settings-search">
                <input
                  v-model="systemCatalogSearch"
                  type="search"
                  placeholder="Поиск по названию или ЕКН"
                  @input="settingsSystemCatalogPage = 1"
                />
              </label>
              <button class="import-button" type="button" :disabled="isSystemCatalogLoading" @click="openSystemCatalogImport">
                <CloudUpload :size="17" :stroke-width="1.8" aria-hidden="true" />
                {{ isSystemCatalogLoading ? 'Импорт...' : 'Импортировать таблицу' }}
              </button>
              <input
                ref="systemCatalogFileInput"
                class="visually-hidden-input"
                type="file"
                accept=".xlsx"
                @change="importSystemCatalogFile"
              />
            </div>

            <p v-if="systemCatalogError" class="table-message table-message--error">{{ systemCatalogError }}</p>

            <div class="systems-table settings-data-table settings-paginated-table settings-classification-table settings-system-catalog-table">
              <table>
                <thead>
                  <tr>
                    <th>Шифр</th>
                    <th>Название системы</th>
                    <th>Класс</th>
                    <th>
                      <span class="settings-class-header settings-curator-header">
                        <span>Куратор</span>
                        <button
                          type="button"
                          :class="{ 'is-unlocked': isSettingsSystemCatalogUnlocked }"
                          :aria-label="isSettingsSystemCatalogUnlocked ? 'Заблокировать редактирование' : 'Разрешить редактирование'"
                          :title="isSettingsSystemCatalogUnlocked ? 'Редактирование разрешено' : 'Редактирование заблокировано'"
                          @click="isSettingsSystemCatalogUnlocked = !isSettingsSystemCatalogUnlocked"
                        >
                          <Unlock v-if="isSettingsSystemCatalogUnlocked" :size="16" :stroke-width="2" aria-hidden="true" />
                          <Lock v-else :size="16" :stroke-width="2" aria-hidden="true" />
                        </button>
                      </span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="currentSettingsSystemCatalogRows().length === 0">
                    <td class="empty-table-cell" colspan="4">В этом распоряжении пока нет данных таблицы 2</td>
                  </tr>
                  <tr v-for="row in visibleSettingsSystemCatalogRows()" :key="`settings-system-${row.id}`">
                    <td>
                      <input
                        v-model="row.code"
                        class="settings-cell-input"
                        type="text"
                        :disabled="!isSettingsSystemCatalogUnlocked"
                        aria-label="Шифр системы"
                        @input="scheduleSystemCatalogRowSave(row)"
                        @blur="saveSystemCatalogRow(row)"
                      />
                    </td>
                    <td>
                      <input
                        v-model="row.systemName"
                        class="settings-cell-input"
                        type="text"
                        :disabled="!isSettingsSystemCatalogUnlocked"
                        aria-label="Название системы"
                        @input="scheduleSystemCatalogRowSave(row)"
                        @blur="saveSystemCatalogRow(row)"
                      />
                    </td>
                    <td :class="`status-cell status-cell--${classModifier(row.systemClass)}`">
                      <select v-model="row.systemClass" :disabled="!isSettingsSystemCatalogUnlocked" :class="`settings-cell-select settings-cell-select--${classModifier(row.systemClass) || 'new'}`" aria-label="Класс системы" @change="saveSystemCatalogRow(row)">
                        <option v-for="option in classOptions" :key="`catalog-${option}`" :value="option">{{ option }}</option>
                      </select>
                    </td>
                    <td>
                      <input
                        v-model="row.curator"
                        class="settings-cell-input"
                        type="text"
                        :disabled="!isSettingsSystemCatalogUnlocked"
                        aria-label="Куратор"
                        @input="scheduleSystemCatalogRowSave(row)"
                        @blur="saveSystemCatalogRow(row)"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <footer v-if="currentSettingsSystemCatalogRows().length > 0" class="table-pagination settings-table-pagination">
              <span class="table-pagination__range">
                Показано {{ settingsSystemCatalogRangeStart() }}–{{ settingsSystemCatalogRangeEnd() }} из {{ currentSettingsSystemCatalogRows().length }}
              </span>
              <div class="table-pagination__controls">
                <label>
                  <span>Записей на странице</span>
                  <select v-model="settingsSystemCatalogPageSize" @change="changeSettingsSystemCatalogPageSize">
                    <option value="10">10</option>
                    <option value="20">20</option>
                    <option value="50">50</option>
                    <option value="100">100</option>
                    <option value="all">Все</option>
                  </select>
                </label>
                <div v-if="settingsSystemCatalogPageSize !== 'all'" class="table-pagination__pages">
                  <button type="button" :disabled="settingsSystemCatalogPage === 1" aria-label="Предыдущая страница" @click="changeSettingsSystemCatalogPage(settingsSystemCatalogPage - 1)">‹</button>
                  <strong>{{ settingsSystemCatalogPage }} / {{ settingsSystemCatalogPageCount() }}</strong>
                  <button type="button" :disabled="settingsSystemCatalogPage >= settingsSystemCatalogPageCount()" aria-label="Следующая страница" @click="changeSettingsSystemCatalogPage(settingsSystemCatalogPage + 1)">›</button>
                </div>
              </div>
            </footer>
          </section>

          <section class="settings-table-block settings-table-block--documents settings-documents-block" aria-label="Таблица 3">
            <div class="settings-table-toolbar">
              <span class="settings-table-title">
                <strong>Таблица 3</strong>
                <small>Комментарии и документы по системам</small>
              </span>
              <label class="settings-search">
                <input v-model="documentSearch" type="search" placeholder="Поиск по названию или ЕКН" @input="settingsDocumentsPage = 1" />
              </label>
            </div>

            <p class="settings-table-note">
              <Info :size="17" :stroke-width="1.8" aria-hidden="true" />
              Комментарии сохраняются автоматически. К каждой системе можно прикрепить один файл PDF, DOC или DOCX размером до 25 МБ.
            </p>

            <p v-if="documentError" class="table-message table-message--error">{{ documentError }}</p>
            <p v-else-if="isDocumentTableLoading" class="table-message">Загрузка таблицы 3...</p>

            <div class="systems-table settings-docs-table settings-paginated-table">
              <table>
                <thead>
                  <tr>
                    <th>Название системы</th>
                    <th>Комментарий</th>
                    <th>Документ</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="filteredDocumentRows.length === 0">
                    <td class="empty-table-cell" colspan="4">В этом распоряжении пока нет данных таблицы 3</td>
                  </tr>
                  <tr v-for="row in visibleDocumentRows" :key="`document-${row.id}`">
                    <td>
                      <strong>{{ row.systemName }}</strong>
                      <small class="settings-docs-table__code">{{ row.code }}</small>
                    </td>
                    <td>
                      <textarea
                        :id="`document-comment-${row.id}`"
                        v-model="row.comment"
                        class="document-comment-input"
                        rows="2"
                        placeholder="Добавьте комментарий…"
                        :aria-label="`Комментарий к ${row.systemName}`"
                        @input="scheduleDocumentCommentSave(row)"
                        @blur="saveDocumentComment(row)"
                      />
                    </td>
                    <td>
                      <a
                        v-if="row.attachmentName"
                        class="settings-document-link"
                        :href="systemDocumentAttachmentUrl(row)"
                        target="_blank"
                        rel="noreferrer"
                        :title="`Открыть ${row.attachmentName}`"
                      >
                        <img :src="attachmentFileIcon(row.attachmentName)" alt="" aria-hidden="true" />
                        <span>
                          <strong>{{ row.attachmentName }}</strong>
                          <small>{{ formatFileSize(row.attachmentSize) }}</small>
                        </span>
                      </a>
                      <span v-else class="document-empty-cell">
                        <img :src="genericFileIcon" alt="" aria-hidden="true" />
                        <span>Файл не прикреплён</span>
                      </span>
                    </td>
                    <td>
                      <div class="document-actions">
                        <input
                          :id="`document-attachment-${row.id}`"
                          class="visually-hidden-input"
                          type="file"
                          accept=".pdf,.doc,.docx,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                          @change="uploadSystemDocumentAttachment(row, $event)"
                        />
                        <button
                          class="document-upload-button"
                          type="button"
                          :disabled="attachmentPendingIds.includes(row.id)"
                          :aria-label="row.attachmentName ? 'Изменить документ' : 'Загрузить документ'"
                          @click="openAttachmentPicker(row)"
                        >
                          <RefreshCw v-if="row.attachmentName" :size="16" :stroke-width="1.8" aria-hidden="true" />
                          <CloudUpload v-else :size="16" :stroke-width="1.8" aria-hidden="true" />
                          {{ row.attachmentName ? 'Заменить' : 'Прикрепить' }}
                        </button>
                        <button
                          class="icon-action-button icon-action-button--danger"
                          type="button"
                          :disabled="!row.attachmentName || attachmentPendingIds.includes(row.id)"
                          :aria-label="row.attachmentName ? `Удалить документ ${row.attachmentName}` : 'Документ не загружен'"
                          @click="deleteSystemDocumentAttachment(row)"
                        >
                          <Trash2 :size="17" :stroke-width="1.8" aria-hidden="true" />
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <footer class="table-pagination settings-table-pagination">
              <span>Показано {{ settingsDocumentsRangeStart() }}–{{ settingsDocumentsRangeEnd() }} из {{ filteredDocumentRows.length }}</span>
              <div class="table-pagination__controls">
                <label>
                  Записей на странице
                  <select v-model="settingsDocumentsPageSize" @change="changeSettingsDocumentsPageSize">
                    <option value="10">10</option>
                    <option value="20">20</option>
                    <option value="50">50</option>
                    <option value="100">100</option>
                    <option value="all">Все</option>
                  </select>
                </label>
                <div v-if="settingsDocumentsPageSize !== 'all'" class="table-pagination__pages">
                  <button type="button" :disabled="settingsDocumentsPage === 1" aria-label="Предыдущая страница" @click="changeSettingsDocumentsPage(settingsDocumentsPage - 1)">‹</button>
                  <strong>{{ settingsDocumentsPage }} / {{ settingsDocumentsPageCount() }}</strong>
                  <button type="button" :disabled="settingsDocumentsPage >= settingsDocumentsPageCount()" aria-label="Следующая страница" @click="changeSettingsDocumentsPage(settingsDocumentsPage + 1)">›</button>
                </div>
              </div>
            </footer>
          </section>
        </section>
      </section>

      <section v-else class="placeholder-page">
        <h1>{{ pageTitle() }}</h1>
      </section>
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

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="selectedHistorySystem" class="modal-overlay" @click="closeSystemHistory">
          <section class="system-history-card" aria-label="История изменений системы" @click.stop>
            <button class="modal-close-button" type="button" aria-label="Закрыть" @click="closeSystemHistory">
              <X :size="27" :stroke-width="2" aria-hidden="true" />
            </button>

            <header class="system-history-header">
              <span class="system-history-header__icon" aria-hidden="true">
                <House :size="31" :stroke-width="1.8" />
              </span>
              <div class="system-history-header__content">
                <div class="system-history-header__title">
                  <h2>{{ selectedHistorySystem.systemName }}</h2>
                  <a class="system-history-source" :href="selectedHistorySystem.systemUrl || 'https://nav.tn.ru/systems/'" target="_blank" rel="noreferrer" aria-label="Открыть на nav.tn.ru">
                    <Globe2 :size="19" :stroke-width="2" aria-hidden="true" />
                  </a>
                </div>
                <p>История изменений распоряжений и связанных документов</p>
              </div>
            </header>

            <div v-if="isSystemHistoryLoading" class="system-history-loading">
              <RefreshCw :size="20" :stroke-width="1.8" aria-hidden="true" />
              Загрузка истории…
            </div>
            <p v-else-if="systemHistoryError" class="table-message table-message--error">{{ systemHistoryError }}</p>

            <template v-else-if="systemHistoryRows.length">
              <section class="history-current-section">
                <div class="history-columns" aria-hidden="true">
                  <span>Распоряжение</span>
                  <span>Комментарий</span>
                  <span>Документ</span>
                </div>

                <article v-for="row in systemHistoryRows.slice(0, 1)" :key="`history-current-${row.id}`" class="history-entry history-entry--current">
                  <div class="history-entry__order">
                    <span>Текущая версия</span>
                    <strong>{{ row.orderName }}</strong>
                  </div>
                  <p :class="{ 'history-comment--empty': !row.comment }">{{ row.comment || 'Комментарий не добавлен' }}</p>
                  <a v-if="row.attachmentName" class="history-document" :href="systemDocumentAttachmentUrl(row)" target="_blank" rel="noreferrer">
                    <span class="history-document__type-icon" :class="`history-document__type-icon--${attachmentFileKind(row.attachmentName)}`" aria-hidden="true">
                      <img :src="attachmentFileIcon(row.attachmentName)" alt="" />
                    </span>
                    <span>{{ row.attachmentName }}</span>
                    <ExternalLink :size="17" :stroke-width="1.8" aria-hidden="true" />
                  </a>
                  <span v-else class="history-document history-document--empty">Документ не прикреплён</span>
                </article>
              </section>

              <button v-if="systemHistoryRows.length > 1" class="history-toggle" type="button" @click="isHistoryOpen = !isHistoryOpen">
                <span class="history-toggle__line" aria-hidden="true" />
                <ChevronUp :class="{ 'is-collapsed': !isHistoryOpen }" :size="20" :stroke-width="2" aria-hidden="true" />
                {{ isHistoryOpen ? 'Скрыть историю изменений' : 'Показать историю изменений' }}
                <span class="history-toggle__line" aria-hidden="true" />
              </button>

              <Transition name="history-more">
                <section v-if="isHistoryOpen && systemHistoryRows.length > 1" class="history-archive">
                  <header>
                    <span aria-hidden="true"><Clock3 :size="18" :stroke-width="1.8" /></span>
                    <strong>История изменений</strong>
                  </header>
                  <div class="history-timeline">
                    <article v-for="row in systemHistoryRows.slice(1)" :key="`history-${row.id}`" class="history-entry history-entry--past">
                      <div class="history-entry__order">
                        <strong>{{ row.orderName }}</strong>
                      </div>
                      <p :class="{ 'history-comment--empty': !row.comment }">{{ row.comment || 'Комментарий не добавлен' }}</p>
                      <a v-if="row.attachmentName" class="history-document" :href="systemDocumentAttachmentUrl(row)" target="_blank" rel="noreferrer">
                        <span class="history-document__type-icon" :class="`history-document__type-icon--${attachmentFileKind(row.attachmentName)}`" aria-hidden="true">
                          <img :src="attachmentFileIcon(row.attachmentName)" alt="" />
                        </span>
                        <span>{{ row.attachmentName }}</span>
                        <ExternalLink :size="17" :stroke-width="1.8" aria-hidden="true" />
                      </a>
                      <span v-else class="history-document history-document--empty">Документ не прикреплён</span>
                    </article>
                  </div>
                </section>
              </Transition>

              <footer class="system-history-footer">
                <Info :size="17" :stroke-width="1.8" aria-hidden="true" />
                Самые новые изменения отображаются первыми.
              </footer>
            </template>

            <p v-else class="system-history-empty">История изменений пока отсутствует</p>
          </section>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
