<script setup lang="ts">
import {
  CalendarDays,
  ChevronDown,
  ChevronRight,
  CircleCheck,
  Folder,
  FunnelX,
  Info,
  Layers3,
  ListFilter,
  RefreshCw,
  Scale,
  Search,
  UsersRound,
} from '@lucide/vue'
import xlsxFileIcon from 'bootstrap-icons/icons/filetype-xlsx.svg'

export type SystemsPageViewModel = {
  [key: string]: any
}

const props = defineProps<{ model: SystemsPageViewModel }>()

const {
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
  isSystemsFiltersOpen,
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
} = props.model
</script>

<template>
      <section class="systems-page">
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

        <section class="changes-filters systems-filters" :class="{ 'is-collapsed': !isSystemsFiltersOpen }" aria-label="Фильтры списка систем">
          <header class="changes-filters__header">
            <button
              class="changes-filters__heading"
              type="button"
              :aria-expanded="isSystemsFiltersOpen"
              @click="isSystemsFiltersOpen = !isSystemsFiltersOpen"
            >
              <span aria-hidden="true"><ListFilter :size="19" :stroke-width="1.9" /></span>
              <div>
                <h2>Фильтры</h2>
                <small v-if="!isSystemsFiltersOpen">{{ selectedConstructionType }}</small>
              </div>
              <ChevronDown class="changes-filters__collapse-chevron" :class="{ 'is-open': isSystemsFiltersOpen }" :size="18" aria-hidden="true" />
            </button>
            <div class="changes-filters__header-actions">
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
                <button type="button" :class="{ 'is-success': isSystemsRefreshDone }" :disabled="isSystemsRefreshing" @click="refreshSystemsPage">
                  <CircleCheck v-if="isSystemsRefreshDone" :size="18" :stroke-width="1.9" aria-hidden="true" />
                  <RefreshCw v-else :class="{ 'is-spinning': isSystemsRefreshing }" :size="18" :stroke-width="1.8" aria-hidden="true" />
                  {{ isSystemsRefreshing ? 'Обновление…' : isSystemsRefreshDone ? 'Обновлено' : 'Обновить' }}
                </button>
              </div>
            </div>
          </header>

          <Transition name="primary-filters">
            <div v-if="isSystemsFiltersOpen" class="primary-filters-body">
              <div class="primary-filters-body__inner">
          <div class="changes-filters__group systems-filters__construction">
            <h3>Тип строительства</h3>
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
          </div>

          <div class="changes-filters__group systems-filters__types">
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
                      v-for="type in visibleSystemTypes"
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
          </div>

          <div class="changes-filters__group systems-filters__parameters">
            <section class="table-toolbar systems-table-toolbar" aria-label="Параметры системы">
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
          </div>
              </div>
            </div>
          </Transition>
        </section>

        <p v-if="systemCatalogError" class="table-message table-message--error">{{ systemCatalogError }}</p>

        <section class="comparison-table-controls" aria-label="Управление выбором для сравнения">
          <div class="comparison-table-controls__heading-wrap">
            <span class="comparison-table-controls__icon" aria-hidden="true">
              <Scale :size="24" :stroke-width="1.8" />
            </span>
            <div class="comparison-table-controls__heading">
              <strong>Добавление в сравнение</strong>
              <span
                class="comparison-table-controls__help"
                tabindex="0"
                aria-label="Отметьте нужные системы в таблице для сравнения изменений класса системы в зависимости от версии распоряжения"
              >
                <Info :size="15" :stroke-width="2" aria-hidden="true" />
                <span role="tooltip">Отметьте нужные системы в таблице для сравнения изменений класса системы в зависимости от версии распоряжения.</span>
              </span>
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
          </div>

        </section>

        <div
          class="systems-table systems-table--catalog"
          :class="{ 'is-empty-loading': (isSystemDocumentLoading || isSystemFiltering) && currentSystemDocumentRows().length === 0 }"
        >
          <Transition name="table-filter-loading">
            <div
              v-if="isSystemDocumentLoading || isSystemFiltering"
              class="systems-table__filter-loading"
              :class="{ 'systems-table__filter-loading--initial': currentSystemDocumentRows().length === 0 }"
              role="status"
              aria-live="polite"
            >
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
              <tr v-if="!isSystemDocumentLoading && !isSystemFiltering && currentSystemDocumentRows().length === 0">
                <td class="empty-table-cell" colspan="5">В таблице 3 этого распоряжения пока нет систем</td>
              </tr>
              <tr v-for="row in visibleSystemDocumentRows()" :key="row.id" tabindex="-1">
                <td class="systems-code-cell">
                  <span :class="{ 'systems-code-cell__empty': !row.code.trim() }">{{ row.code.trim() || 'Не присвоен' }}</span>
                </td>
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

</template>
