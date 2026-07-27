<script setup lang="ts">
import {
  CalendarDays,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  FunnelX,
  Globe2,
  Grid2X2,
  Layers3,
  List,
  ListFilter,
  MoreHorizontal,
  RefreshCw,
  Search,
} from '@lucide/vue'

export type ClassificationPageViewModel = {
  [key: string]: any
}

const props = defineProps<{ model: ClassificationPageViewModel }>()

const {
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
  isClassificationPrimaryFiltersOpen,
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
} = props.model
</script>

<template>
      <section class="classification-page">
        <section class="changes-filters classification-page-filters" :class="{ 'is-collapsed': !isClassificationPrimaryFiltersOpen }" aria-label="Основные параметры классификации">
          <header class="changes-filters__header">
            <button
              class="changes-filters__heading"
              type="button"
              :aria-expanded="isClassificationPrimaryFiltersOpen"
              @click="isClassificationPrimaryFiltersOpen = !isClassificationPrimaryFiltersOpen"
            >
              <span aria-hidden="true"><ListFilter :size="19" :stroke-width="1.9" /></span>
              <div>
                <h2>Основные параметры</h2>
                <small v-if="!isClassificationPrimaryFiltersOpen">{{ selectedConstructionType }}</small>
              </div>
              <ChevronDown class="changes-filters__collapse-chevron" :class="{ 'is-open': isClassificationPrimaryFiltersOpen }" :size="18" aria-hidden="true" />
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
            </div>
          </header>

          <Transition name="primary-filters">
            <div v-if="isClassificationPrimaryFiltersOpen" class="primary-filters-body">
              <div class="primary-filters-body__inner">
          <div class="changes-filters__group">
            <h3>Тип строительства</h3>
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
          </div>

          <div class="changes-filters__group classification-page-filters__types">
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

          <div class="changes-filters__group classification-page-filters__search-row">
            <label class="search-field classification-search" :class="{ 'is-pending': isClassificationSearchPending, 'is-filtered': classificationCatalogSearchInput.trim() }">
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
            <span class="systems-reset-slot">
              <Transition name="changes-reset">
                <button
                  v-if="hasActiveClassificationPageFilters"
                  class="changes-reset-filters systems-reset-filters"
                  type="button"
                  title="Сбросить основные параметры"
                  aria-label="Сбросить основные параметры"
                  @click="resetClassificationPageFilters"
                >
                  <FunnelX :size="19" :stroke-width="1.8" aria-hidden="true" />
                  <span class="systems-reset-filters__count" aria-hidden="true">{{ activeClassificationPageFilterCount }}</span>
                </button>
              </Transition>
            </span>
          </div>
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

            <section
              class="classification-cards"
              :class="{
                'is-list-view': classificationView === 'list',
                'is-empty-loading': isClassificationCatalogLoading && classificationSystems.length === 0,
              }"
              aria-label="Системы классификации"
              aria-live="polite"
            >
            <Transition name="table-filter-loading">
              <div
                v-if="isClassificationCatalogLoading"
                class="systems-table__filter-loading"
                :class="{ 'systems-table__filter-loading--initial': classificationSystems.length === 0 }"
                role="status"
              >
                <span>
                  <RefreshCw :size="18" :stroke-width="1.8" aria-hidden="true" />
                  Загружаем системы классификации…
                </span>
              </div>
            </Transition>
            <p v-if="classificationCatalogError" class="table-message table-message--error classification-cards__message">
              {{ classificationCatalogError }}
            </p>
            <p v-else-if="!isClassificationCatalogLoading && classificationSystems.length === 0" class="table-message classification-cards__message">
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
                v-if="openedClassificationSystem && row.some((system: any) => system.id === openedClassificationSystemId)"
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

          <aside class="classification-sidebar" aria-label="Характеристики системы">
            <header class="classification-sidebar__header">
              <div class="classification-sidebar__heading">
                <span class="classification-sidebar__heading-icon" aria-hidden="true">
                  <ListFilter :size="18" :stroke-width="1.9" />
                </span>
                <div>
                  <strong>Характеристики системы</strong>
                  <span>{{ classificationSystems.length }} из {{ classificationBaseSystems.length }} систем</span>
                </div>
              </div>
              <button
                type="button"
                :disabled="selectedClassificationFilterCount === 0"
                aria-label="Сбросить характеристики системы"
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
            <div class="classification-sidebar__body">
              <p v-if="!isClassificationCatalogLoading && classificationFilterGroups.length === 0" class="table-message classification-sidebar__empty">
                Для выбранного типа нет доступных характеристик.
              </p>
              <p v-else-if="!isClassificationCatalogLoading && visibleClassificationFilterGroups.length === 0" class="table-message classification-sidebar__empty">
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
                <Transition
                  name="classification-filter-options"
                  @before-enter="prepareClassificationFilterEnter"
                  @enter="animateClassificationFilterEnter"
                  @after-enter="finishClassificationFilterMotion"
                  @before-leave="prepareClassificationFilterLeave"
                  @leave="animateClassificationFilterLeave"
                  @after-leave="finishClassificationFilterMotion"
                >
                  <div v-if="openedClassificationFilter === filter" class="classification-sidebar__options-shell">
                    <div class="classification-sidebar__options">
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
                </Transition>
              </div>
            </div>
          </aside>
        </div>
      </section>

</template>
