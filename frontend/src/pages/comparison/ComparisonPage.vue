<script setup lang="ts">
import {
  ArrowDownUp,
  ChevronDown,
  CircleCheck,
  GripVertical,
  Layers3,
  ListFilter,
  Plus,
  RefreshCw,
  Repeat2,
  Scale,
  Trash2,
  TriangleAlert,
} from '@lucide/vue'
import xlsxFileIcon from 'bootstrap-icons/icons/filetype-xlsx.svg'

export type ComparisonPageViewModel = {
  [key: string]: any
}

const props = defineProps<{ model: ComparisonPageViewModel }>()

const {
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
} = props.model
</script>

<template>
      <section class="comparison-page">
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
                    <span class="comparison-order__number">{{ Number(index) + 1 }}</span>
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
                    :disabled="comparisonOrderIds.length >= MAX_COMPARISON_ORDERS || availableComparisonOrders().length === 0"
                    aria-label="Добавить распоряжение"
                    :title="comparisonOrderIds.length >= MAX_COMPARISON_ORDERS ? 'Можно сравнить не более 6 распоряжений' : availableComparisonOrders().length ? 'Добавить распоряжение' : 'Все распоряжения уже добавлены'"
                    @click.stop="toggleSelect('comparison-add')"
                  >
                    <Plus :size="20" :stroke-width="2" aria-hidden="true" />
                    <span>Добавить распоряжение</span>
                  </button>
                  <Transition name="select-menu">
                    <div v-if="openedSelect === 'comparison-add' && comparisonOrderIds.length < MAX_COMPARISON_ORDERS && availableComparisonOrders().length" class="custom-select__menu comparison-add-menu">
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

        <div
          class="systems-table comparison-table"
          :class="{ 'is-empty-loading': isComparisonLoading && comparisonRows().length === 0 }"
        >
          <Transition name="table-filter-loading">
            <div
              v-if="isComparisonLoading"
              class="systems-table__filter-loading"
              :class="{ 'systems-table__filter-loading--initial': comparisonRows().length === 0 }"
              role="status"
              aria-live="polite"
            >
              <span>
                <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                Загружаем сравнение…
              </span>
            </div>
          </Transition>
          <table>
            <thead>
              <tr>
                <th>Название системы <ArrowDownUp :size="15" :stroke-width="1.7" aria-hidden="true" /></th>
                <th v-for="(orderId, index) in comparisonOrderIds" :key="orderId">
                  <span class="comparison-table__order">
                    <i>{{ Number(index) + 1 }}</i>
                    <span :title="comparisonOrderName(orderId)">{{ comparisonOrderName(orderId) }}</span>
                  </span>
                </th>
                <th>Действие</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!isComparisonLoading && comparisonRows().length === 0">
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

</template>
