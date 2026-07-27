<script setup lang="ts">
import {
  ArrowDownUp,
  CalendarDays,
  ChevronDown,
  ChevronRight,
  CircleCheck,
  ExternalLink,
  FunnelX,
  Info,
  Layers3,
  ListFilter,
  Plus,
  RefreshCw,
  Repeat2,
  TriangleAlert,
} from '@lucide/vue'
import xlsxFileIcon from 'bootstrap-icons/icons/filetype-xlsx.svg'

export type ChangesPageViewModel = {
  [key: string]: any
}

const props = defineProps<{ model: ChangesPageViewModel }>()

const {
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
  isChangesFiltersOpen,
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
} = props.model
</script>

<template>
      <section class="changes-page">
        <section class="summary-grid" aria-label="Сводка изменений">
          <article class="summary-card summary-card--changed">
            <div class="summary-card__icon" aria-hidden="true">
              <Repeat2 :size="34" :stroke-width="1.8" />
            </div>
            <div class="summary-card__content">
              <span>Изменения в классификации</span>
              <strong>{{ classificationStats.classificationChanges }}</strong>
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

          <article class="summary-card summary-card--added">
            <div class="summary-card__icon" aria-hidden="true">
              <Layers3 :size="34" :stroke-width="1.8" />
            </div>
            <div class="summary-card__content">
              <span>Добавлено систем</span>
              <strong>{{ classificationStats.addedSystems }}</strong>
            </div>
          </article>
        </section>

        <section
          class="changes-filters"
          :class="[
            afterFilterAccentModifier() && `changes-filter-accent--${afterFilterAccentModifier()}`,
            { 'is-collapsed': !isChangesFiltersOpen },
          ]"
          :style="{ '--before-accent': statusAccentColor(selectedBeforeFilter), '--after-accent': statusAccentColor(selectedAfterFilter) }"
          aria-label="Фильтры изменений"
        >
          <header class="changes-filters__header">
            <button
              class="changes-filters__heading"
              type="button"
              :aria-expanded="isChangesFiltersOpen"
              @click="isChangesFiltersOpen = !isChangesFiltersOpen"
            >
              <span aria-hidden="true"><ListFilter :size="19" :stroke-width="1.9" /></span>
              <div>
                <h2>Фильтры</h2>
                <small v-if="!isChangesFiltersOpen">{{ selectedConstructionType }}</small>
              </div>
              <ChevronDown class="changes-filters__collapse-chevron" :class="{ 'is-open': isChangesFiltersOpen }" :size="18" aria-hidden="true" />
            </button>
            <div class="changes-filters__header-actions">
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
                <button type="button" :class="{ 'is-success': isChangesRefreshDone }" :disabled="isChangesRefreshing" @click="refreshChangesPage">
                  <CircleCheck v-if="isChangesRefreshDone" :size="18" :stroke-width="1.9" aria-hidden="true" />
                  <RefreshCw v-else :class="{ 'is-spinning': isChangesRefreshing }" :size="18" :stroke-width="1.8" aria-hidden="true" />
                  {{ isChangesRefreshing ? 'Обновление…' : isChangesRefreshDone ? 'Обновлено' : 'Обновить' }}
                </button>
              </div>
            </div>
          </header>

          <Transition name="primary-filters">
            <div v-if="isChangesFiltersOpen" class="primary-filters-body">
              <div class="primary-filters-body__inner">
                <div class="changes-filters__group changes-filters__construction">
                  <h3>Тип строительства</h3>
                  <div class="type-tabs type-tabs--changes">
                    <button
                      v-for="type in classificationConstructionTypes"
                      :key="type.name"
                      class="type-tab"
                      :class="{ 'type-tab--active': type.name === selectedConstructionType }"
                      type="button"
                      @click="selectConstructionType(type.name)"
                    >
                      <span>{{ type.label }}</span>
                      <strong>{{ type.count }}</strong>
                    </button>
                  </div>
                </div>

                <div class="changes-filters__group changes-filters__classes">
                  <h3>Класс системы</h3>
                  <div class="table-toolbar changes-filters__toolbar">
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
                  </div>
                </div>
              </div>
            </div>
          </Transition>
        </section>

        <p v-if="classificationError" class="table-message table-message--error">{{ classificationError }}</p>

        <div
          class="systems-table changes-table-card"
          :class="[
            afterFilterAccentModifier() && `changes-table-card--${afterFilterAccentModifier()}`,
            { 'is-empty-loading': (isClassificationLoading || isClassificationFiltering) && currentClassificationRows().length === 0 },
          ]"
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
            <div
              v-if="isClassificationLoading || isClassificationFiltering"
              class="systems-table__filter-loading"
              :class="{ 'systems-table__filter-loading--initial': currentClassificationRows().length === 0 }"
              role="status"
              aria-live="polite"
            >
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
              <tr v-if="!isClassificationLoading && !isClassificationFiltering && currentClassificationRows().length === 0">
                <td class="empty-table-cell" colspan="3">{{ classificationChangesEmptyMessage() }}</td>
              </tr>
              <tr v-for="row in visibleClassificationRows()" :key="row.id" tabindex="-1">
                <td class="changes-system-cell">
                  <div class="changes-system-title">
                    <a v-if="row.systemUrl" class="changes-system-link" :href="row.systemUrl" target="_blank" rel="noreferrer">
                      <span>{{ row.systemName }}</span>
                      <ExternalLink :size="13" :stroke-width="1.8" aria-hidden="true" />
                    </a>
                    <span v-else class="changes-system-name">{{ row.systemName }}</span>
                    <span
                      v-if="row.constructionType === 'Тип не присвоен'"
                      class="construction-type-info"
                      tabindex="0"
                      aria-label="Для данной системы параметр Тип строительства не был найден на nav.tn.ru"
                    >
                      <Info :size="13" :stroke-width="2.1" aria-hidden="true" />
                      <span class="construction-type-info__tooltip" role="tooltip">Для данной системы параметр «Тип строительства» не был найден на nav.tn.ru</span>
                    </span>
                  </div>
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

</template>
