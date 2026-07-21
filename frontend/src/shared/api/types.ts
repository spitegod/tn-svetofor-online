export type ClassificationChange = {
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

export type ClassificationStats = {
  addedSystems: number
  recommended: number
  allowed: number
  classificationChanges: number
}

export type ClassificationResponse = {
  rows: ClassificationChange[]
  stats: ClassificationStats
  beforeOptions: string[]
  afterOptions: string[]
}

export type SystemCharacteristic = {
  position: number
  name: string
  value: string
}

export type SystemCatalogRow = {
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

export type NavParserSettings = {
  updateIntervalDays: number
  workerCount: number
  requestTimeoutSeconds: number
  retryAttempts: number
  retryDelaySeconds: number
  fallbackSearch: boolean
  lastRunAt: string | null
  lastAttemptAt: string | null
  consecutiveFailures: number
  nextRunAt: string | null
}

export type NavParserLogEntry = {
  time: string
  level: 'info' | 'success' | 'warning' | 'error'
  message: string
}

export type NavParserProgress = {
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

export type NavParserRun = {
  id: number
  source: 'manual' | 'scheduled' | string
  status: 'completed' | 'failed' | 'canceled' | string
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

export type SystemCatalogStats = {
  total: number
  recommended: number
  allowed: number
  forbidden: number
  curators: number
}

export type SystemTypeOption = {
  slug: string
  name: string
  imageUrl: string
  position: number
}

export type SystemCatalogResponse = {
  rows: SystemCatalogRow[]
  stats: SystemCatalogStats
  classOptions: string[]
  curatorOptions: string[]
  systemTypes: SystemTypeOption[]
}

export type SystemDocumentRow = {
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

export type SystemDocumentResponse = {
  rows: SystemDocumentRow[]
  stats: SystemCatalogStats
  classOptions: string[]
  curatorOptions: string[]
}

export type Order = {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

export type ComparisonRow = {
  key: string
  name: string
  values: string[]
}
