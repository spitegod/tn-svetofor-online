<script setup lang="ts">
import {
  CloudUpload,
  Database,
  EllipsisVertical,
  Info,
  Layers3,
  Lock,
  Plus,
  RefreshCw,
  Trash2,
  Unlock,
} from '@lucide/vue'
import genericFileIcon from 'bootstrap-icons/icons/file-earmark.svg'
import NavParserPanel, { type NavParserPanelViewModel } from '@/features/nav-parser/NavParserPanel.vue'

export type SettingsPageViewModel = {
  [key: string]: any
}

const props = defineProps<{ model: SettingsPageViewModel }>()

const {
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
} = props.model

const navParserPanelModel: NavParserPanelViewModel = {
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
}
</script>

<template>
      <section class="settings-page">
        <NavParserPanel :model="navParserPanelModel" />

        <div class="settings-orders-group">
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
            <div class="settings-orders-actions">
              <button class="settings-create-order settings-create-order--secondary" type="button" :disabled="isOrderWorkbookImporting" @click="createOrder">
                <Plus :size="18" :stroke-width="1.8" aria-hidden="true" />
                Создать вручную
              </button>
              <button class="settings-create-order" type="button" :disabled="isOrderWorkbookImporting" @click="openOrderWorkbookImport">
                <CloudUpload :size="18" :stroke-width="1.8" aria-hidden="true" />
                {{ isOrderWorkbookImporting ? 'Импорт...' : 'Импорт XLSX' }}
              </button>
              <input
                ref="orderWorkbookInput"
                class="visually-hidden-input"
                type="file"
                accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                @change="importOrderWorkbook"
              />
            </div>
          </div>

          <div
            class="systems-table settings-orders-table settings-table-scroll"
            :class="{ 'is-empty-loading': isOrdersLoading && orders.length === 0 }"
          >
            <Transition name="table-filter-loading">
              <div
                v-if="isOrdersLoading"
                class="systems-table__filter-loading"
                :class="{ 'systems-table__filter-loading--initial': orders.length === 0 }"
                role="status"
                aria-live="polite"
              >
                <span>
                  <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                  Загружаем распоряжения…
                </span>
              </div>
            </Transition>
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
                <tr v-if="!isOrdersLoading && orders.length === 0">
                  <td class="empty-table-cell" colspan="4">Распоряжений пока нет</td>
                </tr>
                <tr v-for="order in visibleOrders" :key="order.id">
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
        </section>
        <footer v-if="!isOrdersLoading && orders.length > 0" class="table-pagination settings-orders-pagination">
          <span class="table-pagination__range">
            Показано {{ ordersRangeStart() }}–{{ ordersRangeEnd() }} из {{ orders.length }}
          </span>
          <div class="table-pagination__controls">
            <label>
              <span>Записей на странице</span>
              <select v-model="ordersPageSize" @change="changeOrdersPageSize">
                <option value="5">5</option>
                <option value="10">10</option>
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="all">Все</option>
              </select>
            </label>
            <div v-if="ordersPageSize !== 'all'" class="table-pagination__pages">
              <button type="button" :disabled="ordersPage === 1" aria-label="Предыдущая страница" @click="changeOrdersPage(ordersPage - 1)">‹</button>
              <strong>{{ ordersPage }} / {{ ordersPageCount() }}</strong>
              <button type="button" :disabled="ordersPage >= ordersPageCount()" aria-label="Следующая страница" @click="changeOrdersPage(ordersPage + 1)">›</button>
            </div>
          </div>
        </footer>
        </div>

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

            <div
              class="systems-table settings-data-table settings-paginated-table settings-classification-table"
              :class="{ 'is-empty-loading': isClassificationLoading && currentSettingsClassificationRows().length === 0 }"
            >
              <Transition name="table-filter-loading">
                <div
                  v-if="isClassificationLoading"
                  class="systems-table__filter-loading"
                  :class="{ 'systems-table__filter-loading--initial': currentSettingsClassificationRows().length === 0 }"
                  role="status"
                  aria-live="polite"
                >
                  <span>
                    <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                    {{ classificationLoadingMessage }}
                  </span>
                </div>
              </Transition>
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
                  <tr v-if="!isClassificationLoading && currentSettingsClassificationRows().length === 0">
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
            <footer v-if="!isClassificationLoading && currentSettingsClassificationRows().length > 0" class="table-pagination settings-table-pagination">
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

            <div
              class="systems-table settings-data-table settings-paginated-table settings-classification-table settings-system-catalog-table"
              :class="{ 'is-empty-loading': isSystemCatalogLoading && currentSettingsSystemCatalogRows().length === 0 }"
            >
              <Transition name="table-filter-loading">
                <div
                  v-if="isSystemCatalogLoading"
                  class="systems-table__filter-loading"
                  :class="{ 'systems-table__filter-loading--initial': currentSettingsSystemCatalogRows().length === 0 }"
                  role="status"
                  aria-live="polite"
                >
                  <span>
                    <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                    Загружаем таблицу 2…
                  </span>
                </div>
              </Transition>
              <table>
                <thead>
                  <tr>
                    <th>Шифр</th>
                    <th>Система</th>
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
                  <tr v-if="!isSystemCatalogLoading && currentSettingsSystemCatalogRows().length === 0">
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
            <footer v-if="!isSystemCatalogLoading && currentSettingsSystemCatalogRows().length > 0" class="table-pagination settings-table-pagination">
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
                <input v-model="documentSearch" type="search" placeholder="Поиск по названию или шифру" @input="settingsDocumentsPage = 1" />
              </label>
            </div>

            <p class="settings-table-note">
              <Info :size="17" :stroke-width="1.8" aria-hidden="true" />
              Комментарии сохраняются автоматически. К каждой системе можно прикрепить один файл PDF, DOC или DOCX размером до 25 МБ.
            </p>

            <p v-if="documentError" class="table-message table-message--error">{{ documentError }}</p>

            <div
              class="systems-table settings-docs-table settings-paginated-table"
              :class="{ 'is-empty-loading': isDocumentTableLoading && filteredDocumentRows.length === 0 }"
            >
              <Transition name="table-filter-loading">
                <div
                  v-if="isDocumentTableLoading"
                  class="systems-table__filter-loading"
                  :class="{ 'systems-table__filter-loading--initial': filteredDocumentRows.length === 0 }"
                  role="status"
                  aria-live="polite"
                >
                  <span>
                    <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                    Загружаем таблицу 3…
                  </span>
                </div>
              </Transition>
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
                  <tr v-if="!isDocumentTableLoading && filteredDocumentRows.length === 0">
                    <td class="empty-table-cell" colspan="4">В этом распоряжении пока нет данных таблицы 3</td>
                  </tr>
                  <tr v-for="row in visibleDocumentRows" :key="`document-${row.id}`">
                    <td class="settings-docs-system-cell">
                      <div class="settings-docs-system">
                        <span class="settings-docs-system__icon" aria-hidden="true">
                          <Layers3 :size="19" :stroke-width="1.8" />
                        </span>
                        <span class="settings-docs-system__identity">
                          <strong :title="row.systemName">{{ row.systemName }}</strong>
                          <small class="settings-docs-table__code" :class="{ 'is-empty': !row.code.trim() }">
                            {{ row.code.trim() || 'Шифр не присвоен' }}
                          </small>
                        </span>
                      </div>
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
            <footer v-if="!isDocumentTableLoading && filteredDocumentRows.length > 0" class="table-pagination settings-table-pagination">
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

        <section class="settings-section appearance-settings" aria-labelledby="appearance-settings-title">
          <div class="appearance-settings__intro">
            <span class="settings-section__eyebrow">Интерфейс</span>
            <h1 id="appearance-settings-title">Внешний вид</h1>
            <p>Минимальный размер текста во всём интерфейсе</p>
          </div>
          <div
            class="appearance-settings__segment"
            :class="`is-${fontSizePreset}`"
            role="radiogroup"
            aria-label="Минимальный размер текста"
          >
            <span class="appearance-settings__indicator" aria-hidden="true" />
            <button
              v-for="preset in fontSizePresets"
              :key="preset.key"
              :class="{ 'is-selected': fontSizePreset === preset.key }"
              type="button"
              role="radio"
              :aria-checked="fontSizePreset === preset.key"
              @click="fontSizePreset = preset.key"
            >
              <strong>{{ preset.label }}</strong>
            </button>
          </div>
        </section>
      </section>

</template>
