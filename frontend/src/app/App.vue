<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import logo from '@/shared/assets/logo.png'
import folderIcon from '@/shared/assets/folder.png'
import browseIcon from '@/shared/assets/browse.png'
import trashIcon from '@/shared/assets/trash.png'

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
  updated: number
  failed: number
  notFound: string[]
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
const comparisonAllOrders = ref(true)
const isBulkComparisonUpdating = ref(false)
const hiddenComparisonRows = ref<string[]>([])
const isComparisonLoading = ref(false)
const comparisonError = ref('')
const isOrdersLoading = ref(false)
const ordersError = ref('')
const orderRenameTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const classificationEditTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const systemCatalogEditTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
const documentCommentTimers = new Map<number, ReturnType<typeof window.setTimeout>>()
let systemDocumentSearchTimer: ReturnType<typeof window.setTimeout> | null = null
const selectedSystemTypeSlug = ref('')
const isSystemTypesOpen = ref(false)
const selectedHistorySystem = ref<SystemDocumentRow | null>(null)
const systemHistoryRows = ref<SystemDocumentRow[]>([])
const isSystemHistoryLoading = ref(false)
const systemHistoryError = ref('')
const isHistoryOpen = ref(false)
const openedSelect = ref<string | null>(null)
const openedComparisonMenu = ref<number | null>(null)
const draggedComparisonOrderId = ref<number | null>(null)
const comparisonDropIndex = ref<number | null>(null)
const importFileInput = ref<HTMLInputElement | null>(null)
const systemCatalogFileInput = ref<HTMLInputElement | null>(null)
const classificationRows = ref<ClassificationChange[]>([])
const classificationVisibleLimit = ref(20)
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
const classificationLoadingMessage = ref('Загрузка таблицы...')
const classificationError = ref('')
const classificationConstructionTypes = computed(() => [...constructionTypes, 'Тип не присвоен'].map((name) => ({
  name,
  count: name === 'Все'
    ? classificationRows.value.length
    : classificationRows.value.filter((row) => row.constructionType === name).length,
})))
const systemCatalogRows = ref<SystemCatalogRow[]>([])
const systemDocumentRows = ref<SystemDocumentRow[]>([])
const documentRows = ref<SystemDocumentRow[]>([])
const documentSearch = ref('')
const documentError = ref('')
const isDocumentTableLoading = ref(false)
const classificationCatalogRows = ref<SystemCatalogRow[]>([])
const classificationCatalogSearch = ref('')
const isClassificationCatalogLoading = ref(false)
const classificationCatalogError = ref('')
const parsedSystemTypes = ref<SystemTypeOption[]>([])
const openedClassificationSystemId = ref<number | null>(null)
const classificationCardColumns = ref(3)
const openedClassificationFilter = ref<string | null>(null)
const selectedClassificationFilters = ref<Record<string, string>>({})
const systemTypeSourceRows = computed(() => activePage.value === 'systems' ? systemDocumentRows.value : classificationCatalogRows.value)
const systemTypes = computed(() => [{ slug: '', name: 'Все системы', position: 0 }, ...parsedSystemTypes.value].map((type) => ({
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
const selectedClassificationFilterCount = computed(() => Object.keys(selectedClassificationFilters.value).length)
const classificationSystems = computed(() => {
  const query = classificationCatalogSearch.value.trim().toLocaleLowerCase('ru-RU')
  const selectedFilters = Object.entries(selectedClassificationFilters.value)
  return classificationBaseSystems.value.filter((system) => {
    const matchesSearch = !query ||
      system.systemName.toLocaleLowerCase('ru-RU').includes(query) ||
      system.code.toLocaleLowerCase('ru-RU').includes(query)
    const matchesFilters = selectedFilters.every(([name, value]) =>
      system.characteristics?.some((characteristic) => characteristic.name === name && characteristic.value === value),
    )
    return matchesSearch && matchesFilters
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
const classificationSystemRows = computed(() => {
  const rows: SystemCatalogRow[][] = []
  for (let index = 0; index < classificationSystems.value.length; index += classificationCardColumns.value) {
    rows.push(classificationSystems.value.slice(index, index + classificationCardColumns.value))
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
const systemCatalogError = ref('')
const isNavParsing = ref(false)
const navParseMessage = ref('')
const navParseError = ref('')
const navParseNotFound = ref<string[]>([])

function selectedOrderName() {
  return orders.value.find((order) => order.id === selectedOrderId.value)?.name ?? 'Распоряжение не выбрано'
}

function comparisonOrderName(orderId: number) {
  return orders.value.find((order) => order.id === orderId)?.name ?? 'Распоряжение удалено'
}

function formatOrderDate(value: string) {
  if (!value) {
    return ''
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: '2-digit',
  }).format(new Date(value))
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
  openedSelect.value = null
  classificationCatalogSearch.value = ''
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
  openedComparisonMenu.value = null
}

function removeComparisonOrder(orderId: number) {
  comparisonOrderIds.value = comparisonOrderIds.value.filter((id) => id !== orderId)
  openedComparisonMenu.value = null
}

function startComparisonOrderDrag(event: DragEvent, orderId: number) {
  draggedComparisonOrderId.value = orderId
  comparisonDropIndex.value = comparisonOrderIds.value.indexOf(orderId)
  openedSelect.value = null
  openedComparisonMenu.value = null

  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', String(orderId))
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

  return [...completeKeys, ...partialKeys]
    .filter((key) => !hiddenComparisonRows.value.includes(key))
    .map((key) => ({
      key,
      name: namesByKey.get(key) ?? key,
      values: comparisonOrderIds.value.map((orderId) => valuesByOrder.get(orderId)?.get(key) ?? 'н/д'),
    }))
}

function hideComparisonRow(row: ComparisonRow) {
  if (!hiddenComparisonRows.value.includes(row.key)) {
    hiddenComparisonRows.value = [...hiddenComparisonRows.value, row.key]
  }
}

function comparisonValue(row: ComparisonRow, index: number) {
  return row.values[index] ?? 'н/д'
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
  classificationVisibleLimit.value = 20
  classificationStats.value = payload.stats
  beforeOptions.value = payload.beforeOptions.length > 1 ? payload.beforeOptions : beforeOptions.value
  afterOptions.value = payload.afterOptions.length > 1 ? payload.afterOptions : afterOptions.value
}

async function loadClassificationChanges() {
  isClassificationLoading.value = true
  classificationLoadingMessage.value = 'Загрузка таблицы...'
  classificationError.value = ''

  try {
    const query = buildClassificationQuery()
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
  if (selectedConstructionType.value === 'Все') {
    return classificationRows.value
  }

  return classificationRows.value.filter((row) => row.constructionType === selectedConstructionType.value)
}

function visibleClassificationRows() {
  return currentClassificationRows().slice(0, classificationVisibleLimit.value)
}

function showMoreClassificationRows() {
  classificationVisibleLimit.value += 20
}

function nextClassificationRowsCount() {
  return Math.min(20, currentClassificationRows().length - classificationVisibleLimit.value)
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

async function loadSystemCatalog() {
  isSystemCatalogLoading.value = true
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
    isSystemCatalogLoading.value = false
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
  if (!selectedOrderId.value || isNavParsing.value) {
    return
  }

  isNavParsing.value = true
  navParseMessage.value = ''
  navParseError.value = ''
  navParseNotFound.value = []
  try {
    const query = new URLSearchParams({ orderId: String(selectedOrderId.value) })
    const response = await fetch(`${API_BASE_URL}/system-catalog/parse-nav?${query.toString()}`, { method: 'POST' })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось выполнить парсинг nav.tn.ru')
    }

    const report: NavParseReport = await response.json()
    navParseMessage.value = `Обновлено ${report.updated} из ${report.total}. Найдено: ${report.found}, не найдено: ${report.notFound.length}, ошибок: ${report.failed}.`
    navParseNotFound.value = report.notFound
    selectedClassificationFilters.value = {}
    await Promise.all([loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable()])
  } catch (error) {
    navParseError.value = error instanceof Error ? error.message : 'Не удалось выполнить парсинг nav.tn.ru'
  } finally {
    isNavParsing.value = false
  }
}

function currentSystemCatalogRows() {
  return systemCatalogRows.value
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

function setPage(page: string) {
  activePage.value = page
  openedSelect.value = null
  openedComparisonMenu.value = null
  if (page === 'changes') {
    void loadClassificationChanges()
  } else if (page === 'systems') {
    void loadSystemDocuments()
  } else if (page === 'classification') {
    void loadClassificationCatalog()
  } else if (page === 'settings') {
    void Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadDocumentTable()])
  }
}

function toggleSelect(name: string) {
  openedComparisonMenu.value = null
  openedSelect.value = openedSelect.value === name ? null : name
}

function toggleComparisonMenu(orderId: number) {
  openedSelect.value = null
  openedComparisonMenu.value = openedComparisonMenu.value === orderId ? null : orderId
}

function matchesConstructionType(system: { characteristics?: SystemCharacteristic[] }) {
  return selectedConstructionType.value === 'Все' ||
    system.characteristics?.some((characteristic) =>
      characteristic.name === 'Сегмент строительства' && characteristic.value.includes(selectedConstructionType.value),
    )
}

function matchesSystemType(system: { characteristics?: SystemCharacteristic[] }, type: SystemTypeOption) {
  return type.slug === '' || system.characteristics?.some((characteristic) =>
    characteristic.name === 'Тип системы' && characteristic.value === type.name,
  )
}

function selectSystemType(type: SystemTypeOption) {
  selectedSystemTypeSlug.value = type.slug
  openedClassificationSystemId.value = null
  clearClassificationFilters()
}

function selectConstructionType(type: string) {
  selectedConstructionType.value = type
  selectedSystemTypeSlug.value = ''
  openedClassificationSystemId.value = null
  clearClassificationFilters()
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
  openedClassificationFilter.value = null
  openedClassificationSystemId.value = null
}

function clearClassificationFilters() {
  selectedClassificationFilters.value = {}
  openedClassificationFilter.value = null
  openedClassificationSystemId.value = null
}

function classificationRowPositions() {
  return [...document.querySelectorAll<HTMLElement>('.classification-card-row')].map((element) => ({
    element,
    top: element.getBoundingClientRect().top,
  }))
}

function applySystemDocumentPayload(payload: SystemDocumentResponse) {
  systemDocumentRows.value = payload.rows
  systemCatalogStats.value = payload.stats
  systemCatalogClassOptions.value = payload.classOptions.length > 1 ? payload.classOptions : ['Все', ...classOptions]
  systemCatalogCuratorOptions.value = payload.curatorOptions.length > 1
    ? ['Все кураторы', ...payload.curatorOptions.filter((option) => option !== 'Все')]
    : ['Все кураторы']
}

async function loadSystemDocuments() {
  isSystemCatalogLoading.value = true
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
    isSystemCatalogLoading.value = false
  }
}

function scheduleSystemDocumentSearch() {
  if (systemDocumentSearchTimer) {
    window.clearTimeout(systemDocumentSearchTimer)
  }
  systemDocumentSearchTimer = window.setTimeout(() => {
    systemDocumentSearchTimer = null
    loadSystemDocuments()
  }, 250)
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

async function deleteSystemDocument(row: SystemDocumentRow) {
  if (!selectedOrderId.value || !window.confirm(`Удалить «${row.systemName}» из таблицы 3 текущего распоряжения?`)) {
    return
  }
  const commentTimer = documentCommentTimers.get(row.id)
  if (commentTimer) {
    window.clearTimeout(commentTimer)
    documentCommentTimers.delete(row.id)
  }
  documentError.value = ''
  try {
    const query = new URLSearchParams({ orderId: String(selectedOrderId.value) })
    const response = await fetch(`${API_BASE_URL}/system-documents/${row.id}?${query.toString()}`, { method: 'DELETE' })
    if (!response.ok) {
      const payload = await response.json().catch(() => null)
      throw new Error(payload?.error ?? 'Не удалось удалить запись таблицы 3')
    }
    await Promise.all([loadDocumentTable(), loadSystemDocuments()])
    if (selectedHistorySystem.value?.id === row.id) {
      closeSystemHistory()
    }
  } catch (error) {
    documentError.value = error instanceof Error ? error.message : 'Не удалось удалить запись таблицы 3'
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
  isHistoryOpen.value = false
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

function pageTitle() {
  return navItems.find((item) => item.key === activePage.value)?.label ?? ''
}

onMounted(async () => {
  updateClassificationCardColumns()
  window.addEventListener('resize', updateClassificationCardColumns)
  await loadOrders()
  await Promise.all([loadClassificationChanges(), loadSystemCatalog(), loadClassificationCatalog(), loadSystemDocuments(), loadDocumentTable()])
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateClassificationCardColumns)
  if (systemDocumentSearchTimer) {
    window.clearTimeout(systemDocumentSearchTimer)
  }
  documentCommentTimers.forEach((timer) => window.clearTimeout(timer))
  classificationEditTimers.forEach((timer) => window.clearTimeout(timer))
  systemCatalogEditTimers.forEach((timer) => window.clearTimeout(timer))
  orderRenameTimers.forEach((timer) => window.clearTimeout(timer))
})
</script>

<template>
  <div class="app" @click="openedSelect = null; openedComparisonMenu = null">
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
        <div class="order-line">
          <div class="select-field">
            <span>Распоряжение</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('order')">
                <span>{{ selectedOrderName() }}</span>
                <i aria-hidden="true" />
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
        </div>

        <section class="summary-grid" aria-label="Сводка изменений">
          <article class="summary-card">
            <span>Добавлено</span>
            <strong>{{ classificationStats.addedSystems }} систем</strong>
          </article>

          <div class="status-stack">
            <article class="status-card status-card--recommended">
              <strong>{{ classificationStats.recommended }}</strong>
              <span>Рекомендованных</span>
            </article>
            <article class="status-card status-card--allowed">
              <strong>{{ classificationStats.allowed }}</strong>
              <span>Разрешенных</span>
            </article>
          </div>

          <article class="summary-card">
            <span>Изм. классификация</span>
            <strong>{{ classificationStats.classificationChanges }} систем</strong>
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

        <section class="table-toolbar" aria-label="Управление таблицей">
          <div class="select-field">
            <span>Было</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'before' }">
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
                    @click="selectedBeforeFilter = option; openedSelect = null; loadClassificationChanges()"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
          <div class="select-field">
            <span>Стало</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'after' }">
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
                    @click="selectedAfterFilter = option; openedSelect = null; loadClassificationChanges()"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
          <button class="export-button" type="button" @click="exportClassificationTable">Экспортировать таблицу</button>
        </section>

        <p v-if="classificationError" class="table-message table-message--error">{{ classificationError }}</p>
        <p v-else-if="isClassificationLoading" class="table-message">{{ classificationLoadingMessage }}</p>

        <div class="systems-table">
          <table>
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
              <tr v-for="row in visibleClassificationRows()" :key="row.id">
                <td>
                  <a v-if="row.systemUrl" :href="row.systemUrl" target="_blank" rel="noreferrer">
                    {{ row.systemName }}
                  </a>
                  <span v-else>{{ row.systemName }}</span>
                  <small v-if="row.constructionType === 'Тип не присвоен'" class="construction-type-status">Тип не присвоен</small>
                </td>
                <td :class="classModifier(row.classBefore) && `status-cell status-cell--${classModifier(row.classBefore)}`">
                  {{ row.classBefore }}
                </td>
                <td :class="classModifier(row.classAfter) && `status-cell status-cell--${classModifier(row.classAfter)}`">
                  {{ row.classAfter }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <button
          v-if="currentClassificationRows().length > classificationVisibleLimit"
          class="more-button"
          type="button"
          @click="showMoreClassificationRows"
        >
          <span>Показать ещё {{ nextClassificationRowsCount() }}</span>
          <i aria-hidden="true" />
        </button>
      </section>

      <section v-else-if="activePage === 'systems'" class="systems-page">
        <div class="order-line">
          <div class="select-field">
            <span>Распоряжение</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('order')">
                <span>{{ selectedOrderName() }}</span>
                <i aria-hidden="true" />
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
        </div>

        <section class="systems-summary" aria-label="Сводка систем">
          <article class="summary-card">
            <strong>{{ systemCatalogStats.total }}</strong>
            <span>систем</span>
          </article>

          <div class="status-stack">
            <article class="status-card status-card--recommended">
              <strong>{{ systemCatalogStats.recommended }}</strong>
              <span>Рекомендованных</span>
            </article>
            <article class="status-card status-card--allowed">
              <strong>{{ systemCatalogStats.allowed }}</strong>
              <span>Разрешенных</span>
            </article>
            <article class="status-card status-card--forbidden">
              <strong>{{ systemCatalogStats.forbidden }}</strong>
              <span>Запрещенных</span>
            </article>
          </div>

          <article class="summary-card">
            <strong>{{ systemCatalogStats.curators }}</strong>
            <span>кураторов</span>
          </article>
        </section>

        <div class="systems-tools">
          <section class="filter-panel" aria-label="Тип строительства">
            <h2>Тип строительства</h2>
            <div class="type-tabs">
              <button
                v-for="type in constructionTypes"
                :key="type"
                class="type-tab"
                :class="{ 'type-tab--active': type === selectedConstructionType }"
                type="button"
                @click="selectConstructionType(type)"
              >
                {{ type }}
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
              <label class="search-field">
                <span>Поиск</span>
                <input v-model="systemCatalogSearch" type="search" placeholder="Поиск по названию или ЕКН" @input="scheduleSystemDocumentSearch" />
              </label>

              <div class="system-type-grid">
                <button
                  v-for="type in systemTypes"
                  :key="type.name"
                  class="system-type-card"
                  :class="{ 'is-active': type.name === selectedSystemType.name }"
                  type="button"
                  @click="selectSystemType(type)"
                >
                  <strong>{{ type.name }}</strong>
                  <span>{{ type.count }} систем</span>
                </button>
              </div>
            </div>
          </Transition>
        </section>

        <section class="table-toolbar systems-table-toolbar" aria-label="Управление таблицей">
          <div class="select-field">
            <span>Класс</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'class' }">
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
                    @click="selectedSystemCatalogClass = option; openedSelect = null; loadSystemDocuments()"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <div class="select-field">
            <span>Куратор</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'curator' }">
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
                    @click="selectedSystemCatalogCurator = option; openedSelect = null; loadSystemDocuments()"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <div class="comparison-bulk-controls" aria-label="Массовый выбор для сравнения">
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
            <label class="toolbar-checkbox">
              <input v-model="comparisonAllOrders" type="checkbox" :disabled="isBulkComparisonUpdating" />
              <span aria-hidden="true" />
              <strong>Все распоряжения</strong>
            </label>
          </div>

          <button class="export-button" type="button" @click="exportSystemCatalog">Экспортировать таблицу</button>
        </section>

        <p v-if="systemCatalogError" class="table-message table-message--error">{{ systemCatalogError }}</p>
        <p v-else-if="isSystemCatalogLoading" class="table-message">Загрузка таблицы...</p>

        <div class="systems-table systems-table--scroll">
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
              <tr v-for="row in currentSystemDocumentRows()" :key="row.id">
                <td>{{ row.code }}</td>
                <td>
                  <a v-if="row.systemUrl" :href="row.systemUrl" target="_blank" rel="noreferrer">
                    {{ row.systemName }}
                  </a>
                  <span v-else>{{ row.systemName }}</span>
                </td>
                <td :class="`status-cell status-cell--${classModifier(row.systemClass)}`">
                  <span>{{ row.systemClass }}</span>
                  <button class="status-cell__icon" type="button" @click.stop="openSystemHistory(row)">
                    <img :src="folderIcon" alt="" aria-hidden="true" />
                  </button>
                </td>
                <td>{{ row.curator }}</td>
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

      </section>

      <section v-else-if="activePage === 'classification'" class="classification-page">
        <div class="classification-topline">
          <div class="select-field">
            <span>Распоряжение</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('order')">
                <span>{{ selectedOrderName() }}</span>
                <i aria-hidden="true" />
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

          <label class="search-field classification-search">
            <span>Поиск</span>
            <input v-model="classificationCatalogSearch" type="search" placeholder="Поиск по названию или ЕКН" />
          </label>
        </div>

        <section class="filter-panel classification-construction" aria-label="Тип строительства">
          <h2>Тип строительства</h2>
          <div class="type-tabs">
            <button
              v-for="type in constructionTypes"
              :key="type"
              class="type-tab"
              :class="{ 'type-tab--active': type === selectedConstructionType }"
              type="button"
              @click="selectConstructionType(type)"
            >
              {{ type }}
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
                  <strong>{{ type.name }}</strong>
                  <span>{{ type.count }} систем</span>
                </button>
              </div>
            </div>
          </Transition>
        </section>

        <div class="classification-layout">
          <section class="classification-cards" aria-label="Системы классификации" aria-live="polite">
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
                    <img :src="browseIcon" alt="" aria-hidden="true" />
                  </a>
                </header>

                <dl class="classification-card__meta">
                  <div>
                    <dt>Шифр</dt>
                    <dd>{{ system.code }}</dd>
                  </div>
                  <div>
                    <dt>Куратор</dt>
                    <dd>{{ system.curator || 'Куратор не указан' }}</dd>
                  </div>
                </dl>

                <button
                  class="classification-card__more"
                  :class="{ 'is-open': openedClassificationSystemId === system.id }"
                  type="button"
                  :aria-expanded="openedClassificationSystemId === system.id"
                  aria-label="Показать характеристики системы"
                  @click="toggleClassificationSystem(system.id)"
                >
                  <i aria-hidden="true" />
                </button>
              </article>

              <div
                v-if="openedClassificationSystem && row.some((system) => system.id === openedClassificationSystemId)"
                class="classification-details-shell"
              >
                  <section class="classification-details">
                    <div class="classification-details__header">
                      <strong>{{ openedClassificationSystem.systemName }}</strong>
                      <div class="classification-details__actions">
                        <a v-if="openedClassificationSystem.systemUrl" :href="openedClassificationSystem.systemUrl" target="_blank" rel="noreferrer">Открыть на nav.tn.ru</a>
                        <button type="button" aria-label="Закрыть характеристики" @click="toggleClassificationSystem(openedClassificationSystem.id)">×</button>
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

          <aside class="classification-sidebar" aria-label="Фильтры классификации">
            <header class="classification-sidebar__header">
              <div>
                <strong>Фильтры</strong>
                <span>{{ classificationSystems.length }} из {{ classificationBaseSystems.length }} систем</span>
              </div>
              <button v-if="selectedClassificationFilterCount" type="button" @click="clearClassificationFilters">
                Сбросить {{ selectedClassificationFilterCount }}
              </button>
            </header>
            <p v-if="classificationFilterGroups.length === 0" class="table-message classification-sidebar__empty">
              Для выбранного типа нет доступных характеристик.
            </p>
            <div
              v-for="filter in classificationFilterGroups"
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
            </div>
          </aside>
        </div>
      </section>

      <section v-else-if="activePage === 'comparison'" class="comparison-page">
        <div class="comparison-controls">
          <h1>Выбрать распоряжения, изменения классов систем по которым нужно сравнить</h1>

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
              draggable="true"
              :aria-grabbed="draggedComparisonOrderId === orderId"
              title="Перетащите, чтобы изменить порядок"
              @dragstart="startComparisonOrderDrag($event, orderId)"
              @dragenter.prevent="enterComparisonOrderDrop(index)"
              @dragover.prevent
              @drop.prevent="dropComparisonOrder(index)"
              @dragend="endComparisonOrderDrag"
            >
              <div class="comparison-order__control">
                <button class="custom-select__button" type="button" @click.stop="toggleSelect(`comparison-${index}`)">
                  <span>{{ comparisonOrderName(orderId) }}</span>
                </button>
                <button
                  class="comparison-order__more"
                  type="button"
                  aria-label="Открыть действия"
                  @click.stop="toggleComparisonMenu(orderId)"
                >
                  <span aria-hidden="true">•••</span>
                </button>
              </div>
              <Transition name="select-menu">
                <div v-if="openedComparisonMenu === orderId" class="comparison-action-menu">
                  <button type="button" @click="removeComparisonOrder(orderId)">
                    <img :src="trashIcon" alt="" aria-hidden="true" />
                    <span>Удалить</span>
                  </button>
                </div>
              </Transition>
              <Transition name="select-menu">
                <div v-if="openedSelect === `comparison-${index}`" class="custom-select__menu">
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
                </div>
              </Transition>
            </div>

            <div
              class="custom-select comparison-add-select"
              :class="{ 'is-open': openedSelect === 'comparison-add' }"
            >
              <button
                class="comparison-add-button"
                type="button"
                :disabled="availableComparisonOrders().length === 0"
                aria-label="Добавить распоряжение"
                :title="availableComparisonOrders().length ? 'Добавить распоряжение' : 'Все распоряжения уже добавлены'"
                @click.stop="toggleSelect('comparison-add')"
              >
                +
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'comparison-add' && availableComparisonOrders().length" class="custom-select__menu comparison-add-menu">
                  <button
                    v-for="order in availableComparisonOrders()"
                    :key="order.id"
                    class="custom-select__option"
                    type="button"
                    @click="addComparisonOrder(order)"
                  >
                    {{ order.name }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
        </div>

        <p v-if="comparisonError" class="table-message table-message--error">{{ comparisonError }}</p>
        <p v-else-if="isComparisonLoading" class="table-message">Загрузка сравнения...</p>

        <div class="systems-table comparison-table">
          <table>
            <thead>
              <tr>
                <th>Название системы</th>
                <th v-for="orderId in comparisonOrderIds" :key="orderId">
                  {{ comparisonOrderName(orderId).replace('№ ', '№\u00A0').replace(' от ', ' от\n') }}
                </th>
                <th>Удалить из сравнения</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="comparisonRows().length === 0">
                <td class="empty-table-cell" :colspan="comparisonOrderIds.length + 2">
                  В выбранных распоряжениях пока нет данных таблицы 2
                </td>
              </tr>
              <tr v-for="row in comparisonRows()" :key="row.key">
                <td>{{ row.name }}</td>
                <td
                  v-for="(_, index) in comparisonOrderIds"
                  :key="`${row.name}-${index}`"
                  :class="classModifier(comparisonValue(row, index)) && `status-cell status-cell--${classModifier(comparisonValue(row, index))}`"
                >
                  {{ comparisonValue(row, index) }}
                </td>
                <td>
                  <button class="comparison-delete-button" type="button" aria-label="Удалить из сравнения" @click="hideComparisonRow(row)">
                    <img :src="trashIcon" alt="" aria-hidden="true" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="activePage === 'settings'" class="settings-page">
        <section class="settings-section parser-settings" aria-labelledby="parser-settings-title">
          <h1 id="parser-settings-title">Парсинг навигатора</h1>
          <p>Загрузить с nav.tn.ru ссылки и характеристики систем для БД «{{ selectedOrderName() }}».</p>
          <button class="import-button parser-settings__button" type="button" :disabled="isNavParsing || !selectedOrderId" @click="runNavParser">
            {{ isNavParsing ? 'Парсинг выполняется…' : 'Запустить парсер' }}
          </button>
          <p v-if="navParseError" class="table-message table-message--error">{{ navParseError }}</p>
          <p v-else-if="navParseMessage" class="table-message table-message--success">{{ navParseMessage }}</p>
          <details v-if="navParseNotFound.length" class="parser-settings__not-found">
            <summary>Не найденные на nav.tn.ru системы ({{ navParseNotFound.length }})</summary>
            <ul>
              <li v-for="systemName in navParseNotFound" :key="systemName">{{ systemName }}</li>
            </ul>
          </details>
        </section>

        <section class="settings-section" aria-labelledby="orders-db-title">
          <div class="settings-section__header">
            <h2 id="orders-db-title">Управление БД Распоряжений</h2>
          </div>

          <div class="systems-table settings-orders-table settings-table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Распоряжение</th>
                  <th>Дата создания</th>
                  <th>Последняя актуализация</th>
                  <th>Удалить БД</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="order in orders" :key="order.id">
                  <td>
                    <input
                      v-model="order.name"
                      class="order-name-input"
                      type="text"
                      aria-label="Название распоряжения"
                      @input="scheduleOrderRename(order)"
                      @blur="saveOrderName(order)"
                      @keyup.enter="($event.target as HTMLInputElement).blur()"
                    />
                  </td>
                  <td>{{ formatOrderDate(order.createdAt) }}</td>
                  <td>{{ formatOrderDate(order.updatedAt) }}</td>
                  <td>
                    <button class="icon-action-button" type="button" aria-label="Удалить БД" @click="deleteOrder(order)">
                      <img :src="trashIcon" alt="" aria-hidden="true" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="create-order-line">
            <span>Создать новую БД распоряжений</span>
            <button class="small-red-button" type="button" @click="createOrder">+</button>
          </div>
          <p v-if="ordersError" class="table-message table-message--error">{{ ordersError }}</p>
          <p v-else-if="isOrdersLoading" class="table-message">Загрузка распоряжений...</p>
        </section>

        <section class="settings-section" aria-labelledby="edit-db-title">
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

          <section class="settings-table-block" aria-label="Таблица 1">
            <div class="settings-table-toolbar">
              <span>Таблица 1</span>
              <label class="settings-search">
                <input
                  v-model="tableSearch"
                  type="search"
                  placeholder="Поиск по названию или ЕКН"
                  @keyup.enter="loadClassificationChanges"
                />
              </label>
              <button class="import-button" type="button" :disabled="isClassificationLoading" @click="openTableImport">
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

            <div class="systems-table settings-data-table settings-table-scroll">
              <table>
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
                    <td class="empty-table-cell" colspan="3">В этом распоряжении пока нет данных таблицы 1</td>
                  </tr>
                  <tr v-for="row in currentClassificationRows()" :key="`settings-${row.id}`">
                    <td>
                      <input
                        v-model="row.systemName"
                        class="settings-cell-input"
                        type="text"
                        aria-label="Название системы"
                        @input="scheduleClassificationRowSave(row)"
                        @blur="saveClassificationRow(row)"
                      />
                    </td>
                    <td :class="classModifier(row.classBefore) && `status-cell status-cell--${classModifier(row.classBefore)}`">
                      <select v-model="row.classBefore" class="settings-cell-select" aria-label="Класс было" @change="saveClassificationRow(row)">
                        <option value="Новая система">Новая система</option>
                        <option v-for="option in classOptions" :key="`before-${option}`" :value="option">{{ option }}</option>
                      </select>
                    </td>
                    <td :class="classModifier(row.classAfter) && `status-cell status-cell--${classModifier(row.classAfter)}`">
                      <select v-model="row.classAfter" class="settings-cell-select" aria-label="Класс стало" @change="saveClassificationRow(row)">
                        <option v-for="option in classOptions" :key="`after-${option}`" :value="option">{{ option }}</option>
                      </select>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="settings-table-block" aria-label="Таблица 2">
            <div class="settings-table-toolbar">
              <span>Таблица 2</span>
              <label class="settings-search">
                <input
                  v-model="systemCatalogSearch"
                  type="search"
                  placeholder="Поиск по названию или ЕКН"
                  @keyup.enter="loadSystemCatalog"
                />
              </label>
              <button class="import-button" type="button" :disabled="isSystemCatalogLoading" @click="openSystemCatalogImport">
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

            <div class="systems-table settings-data-table settings-table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Шифр</th>
                    <th>Название системы</th>
                    <th>Класс</th>
                    <th>Куратор</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="currentSystemCatalogRows().length === 0">
                    <td class="empty-table-cell" colspan="4">В этом распоряжении пока нет данных таблицы 2</td>
                  </tr>
                  <tr v-for="row in currentSystemCatalogRows()" :key="`settings-system-${row.id}`">
                    <td>
                      <input
                        v-model="row.code"
                        class="settings-cell-input"
                        type="text"
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
                        aria-label="Название системы"
                        @input="scheduleSystemCatalogRowSave(row)"
                        @blur="saveSystemCatalogRow(row)"
                      />
                    </td>
                    <td :class="`status-cell status-cell--${classModifier(row.systemClass)}`">
                      <select v-model="row.systemClass" class="settings-cell-select" aria-label="Класс системы" @change="saveSystemCatalogRow(row)">
                        <option v-for="option in classOptions" :key="`catalog-${option}`" :value="option">{{ option }}</option>
                      </select>
                    </td>
                    <td>
                      <input
                        v-model="row.curator"
                        class="settings-cell-input"
                        type="text"
                        aria-label="Куратор"
                        @input="scheduleSystemCatalogRowSave(row)"
                        @blur="saveSystemCatalogRow(row)"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="settings-table-block" aria-label="Таблица 3">
            <div class="settings-table-toolbar">
              <span>Таблица 3</span>
              <label class="settings-search">
                <input v-model="documentSearch" type="search" placeholder="Поиск по названию или ЕКН" />
              </label>
            </div>

            <p v-if="documentError" class="table-message table-message--error">{{ documentError }}</p>
            <p v-else-if="isDocumentTableLoading" class="table-message">Загрузка таблицы 3...</p>

            <div class="systems-table settings-docs-table settings-table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Название системы</th>
                    <th>Комментарий</th>
                    <th>Документ</th>
                    <th>Ред. документа</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="filteredDocumentRows.length === 0">
                    <td class="empty-table-cell" colspan="4">В этом распоряжении пока нет данных таблицы 3</td>
                  </tr>
                  <tr v-for="row in filteredDocumentRows" :key="`document-${row.id}`">
                    <td>
                      <strong>{{ row.systemName }}</strong>
                      <small class="settings-docs-table__code">{{ row.code }}</small>
                    </td>
                    <td>
                      <textarea
                        v-model="row.comment"
                        class="document-comment-input"
                        rows="2"
                        placeholder="Добавьте комментарий…"
                        :aria-label="`Комментарий к ${row.systemName}`"
                        @input="scheduleDocumentCommentSave(row)"
                        @blur="saveDocumentComment(row)"
                      />
                    </td>
                    <td class="document-empty-cell">—</td>
                    <td>
                      <div class="document-actions">
                        <button class="icon-action-button icon-action-button--danger" type="button" :aria-label="`Удалить ${row.systemName} из таблицы 3`" @click="deleteSystemDocument(row)">
                          <img :src="trashIcon" alt="" aria-hidden="true" />
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </section>
      </section>

      <section v-else class="placeholder-page">
        <h1>{{ pageTitle() }}</h1>
      </section>
    </main>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="selectedHistorySystem" class="modal-overlay" @click="closeSystemHistory">
          <section class="system-history-card" aria-label="История изменений системы" @click.stop>
            <button class="modal-close-button" type="button" aria-label="Закрыть" @click="closeSystemHistory">
              ×
            </button>

            <header class="system-history-header">
              <h2>{{ selectedHistorySystem.systemName }}</h2>
              <a class="system-history-source" :href="selectedHistorySystem.systemUrl || 'https://nav.tn.ru/systems/'" target="_blank" rel="noreferrer" aria-label="Открыть на nav.tn.ru">
                <img :src="browseIcon" alt="" aria-hidden="true" />
              </a>
            </header>

            <p v-if="isSystemHistoryLoading" class="table-message">Загрузка истории...</p>
            <p v-else-if="systemHistoryError" class="table-message table-message--error">{{ systemHistoryError }}</p>

            <div v-else-if="systemHistoryRows.length" class="history-table">
              <table>
                <thead>
                  <tr>
                    <th>Распоряжение</th>
                    <th>Комментарий</th>
                    <th>Документ</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in systemHistoryRows.slice(0, 1)" :key="`history-current-${row.id}`">
                    <td>{{ row.orderName }}</td>
                    <td :class="{ 'history-comment--empty': !row.comment }">{{ row.comment || 'Комментарий не добавлен' }}</td>
                    <td class="document-empty-cell">—</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <button v-if="systemHistoryRows.length > 1" class="history-toggle" type="button" @click="isHistoryOpen = !isHistoryOpen">
              {{ isHistoryOpen ? 'скрыть историю изменений' : 'развернуть историю изменений' }}
              <i aria-hidden="true" />
            </button>

            <Transition name="history-more">
              <div v-if="isHistoryOpen" class="history-table history-table--muted">
                <table>
                  <tbody>
                    <tr v-for="row in systemHistoryRows.slice(1)" :key="`history-${row.id}`">
                      <td>{{ row.orderName }}</td>
                      <td :class="{ 'history-comment--empty': !row.comment }">{{ row.comment || 'Комментарий не добавлен' }}</td>
                      <td class="document-empty-cell">—</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </Transition>
          </section>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
