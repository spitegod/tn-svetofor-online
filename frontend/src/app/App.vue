<script setup lang="ts">
import { ref } from 'vue'
import logo from '@/shared/assets/logo.png'
import folderIcon from '@/shared/assets/folder.png'
import openIcon from '@/shared/assets/open.png'
import browseIcon from '@/shared/assets/browse.png'
import trashIcon from '@/shared/assets/trash.png'
import reverseIcon from '@/shared/assets/reverse.png'

const activePage = ref('systems')

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

const systemTypes = [
  { name: 'Все системы', count: 447 },
  { name: 'Плоская крыша', count: 108 },
  { name: 'Крыша стилобата', count: 12 },
  { name: 'Скатная крыша', count: 8 },
  { name: 'Фасад', count: 18 },
  { name: 'Цоколь', count: 5 },
  { name: 'Потолок', count: 1 },
  { name: 'Отмостка', count: 6 },
  { name: 'Полы и перекрытия', count: 43 },
  { name: 'Стены и перегородки', count: 10 },
  { name: 'Звукоизоляция', count: 13 },
  { name: 'Фундамент', count: 62 },
  { name: 'Огнезащита', count: 7 },
  { name: 'Лакокрасочные покрытия', count: 5 },
  { name: 'Подпорные сооружения', count: 2 },
  { name: 'Техническая изоляция', count: 15 },
  { name: 'Емкость, резервуар', count: 9 },
  { name: 'Благоустройство территорий', count: 2 },
  { name: 'Решения для нефтегазового комплекса', count: 7 },
  { name: 'Решения для сельскохозяйственного комплекса', count: 2 },
  { name: 'Решения для природоохранных сооружений', count: 4 },
  { name: 'Решения для гидротехнического строительства', count: 4 },
  { name: 'Решения для горнодобывающей промышленности', count: 5 },
  { name: 'Тоннель', count: 7 },
  { name: 'Дорога', count: 7 },
  { name: 'Мост', count: 5 },
  { name: 'Искусственные водоемы, пруды и пр.', count: 7 },
  { name: 'Полигоны, площадки хранения и пр.', count: 5 },
  { name: 'Автомобильный транспорт', count: 26 },
  { name: 'Железнодорожный транспорт', count: 20 },
  { name: 'Грунтовые плотины и дамбы', count: 2 },
  { name: 'Конструкция летного поля', count: 5 },
  { name: 'Комплексные решения', count: 3 },
]

const orders = [
  '№ ТД-Р-143 от 24.10.2025',
  '№ ТД-Р-118 от 12.09.2025',
  '№ ТД-Р-097 от 18.08.2025',
]

const classOptions = ['Рекомендованная', 'Разрешенная', 'Запрещенная']
const curatorOptions = ['Все кураторы', 'Сендецкий В.', 'Уртенков А.', 'Золотарев М.', 'Кузнецова Н.']

const changesRows = [
  { name: 'ТН-СТИЛОБАТ КЛАССИК АВТО', before: 'Новая система', after: 'Разрешенная' },
  { name: 'ТН-СТИЛОБАТ КЛАССИК ТРОТУАР', before: 'Новая система', after: 'Рекомендованная' },
  { name: 'ТН-СТИЛОБАТ КЛАССИК АВТО', before: 'Новая система', after: 'Разрешенная' },
  { name: 'ТН-СТИЛОБАТ КЛАССИК', before: 'Рекомендованная', after: 'Разрешенная' },
]

const systemNames = [
  'ТН-СТИЛОБАТ КЛАССИК АВТО',
  'ТН-СТИЛОБАТ КЛАССИК ТРОТУАР',
  'ТН-КРОВЛЯ Гарант',
  'ТН-КРОВЛЯ Смарт PIR',
  'ТН-КРОВЛЯ Классик',
  'ТН-ФАСАД Комби',
  'ТН-ФАСАД Профи',
  'ТН-ЦОКОЛЬ Эксперт',
  'ТН-ПОЛ Акустик',
  'ТН-ПОЛ Проф',
  'ТН-СТЕНА Стандарт',
  'ТН-СТЕНА Проф',
  'ТН-ФУНДАМЕНТ Термо',
  'ТН-ФУНДАМЕНТ Термо Проф',
  'ТН-ОГНЕЗАЩИТА Конструктив',
  'ТН-ТЕХИЗОЛЯЦИЯ Трубопровод',
  'ТН-РЕЗЕРВУАР Стандарт',
  'ТН-БЛАГОУСТРОЙСТВО Пешеход',
  'ТН-АВТОДОРОГА Стандарт',
  'ТН-АВТОДОРОГА Проф',
  'ТН-ЖД Платформа',
  'ТН-МОСТ Гидро',
  'ТН-ТОННЕЛЬ Проф',
  'ТН-ДОРОГА Дренаж',
  'ТН-ЛЕТНОЕ ПОЛЕ',
  'ТН-ОТМОСТКА Термо',
  'ТН-СКАТНАЯ КРЫША Мансарда',
  'ТН-КРОВЛЯ Экспресс',
  'ТН-ЗВУКОИЗОЛЯЦИЯ Акустик',
  'ТН-ПЕРЕГОРОДКА Лайт',
  'ТН-ПЛОСКАЯ КРЫША Балласт',
  'ТН-КРОВЛЯ Инверс',
  'ТН-ФАСАД Лайт',
  'ТН-ЦОКОЛЬ Комфорт',
  'ТН-ПОЛ Балкон',
  'ТН-ФУНДАМЕНТ Универсал',
  'ТН-ПОДПОРНАЯ СТЕНА',
  'ТН-ПРУД Гео',
  'ТН-ПОЛИГОН Защита',
  'ТН-ПОЛИГОН Защита Проф',
]

const systemRows = systemNames.map((name, index) => ({
  code: `ПК-${String(10000001 + index)}`,
  name,
  systemClass: classOptions[index % classOptions.length],
  curator: curatorOptions[(index % (curatorOptions.length - 1)) + 1],
  compared: index % 3 === 0,
}))

const classificationFilterGroups = [
  'Рекомендации ТЕХНОНИКОЛЬ',
  'Тип несущего основания системы',
  'Тип кровли по расположению слоев',
  'Тип крыши по степени эксплуатации',
  'Допустимая интенсивность эксплуатационной нагрузки согласно СП17.13330.2017',
  'Наличие теплоизоляционного слоя',
  'Тип теплоизоляции',
  'Метод крепления теплоизоляционного слоя',
  'Тип гидроизоляции',
]

const classificationSystems = systemRows.slice(0, 15).map((row, index) => ({
  ...row,
  systemClass:
    index < 9
      ? 'Рекомендованная'
      : index < 12
        ? 'Разрешенная'
        : 'Запрещенная',
  base: index % 2 === 0 ? 'ПВХ-мембрана' : 'Плоских И.',
}))

const comparisonOrders = [
  '№ ТД-Р-143 от 24.10.2025',
  '№ ТД-Р-126 от 13.09.2022',
  '№ ТД-Р-85 от 05.05.2020',
]

const comparisonRows = [
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК АВТО',
    values: ['Разрешенная', 'Разрешенная', 'Рекомендованная'],
  },
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК ТРОТУАР',
    values: ['Рекомендованная', 'Рекомендованная', 'Рекомендованная'],
  },
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК АВТО',
    values: ['Рекомендованная', 'Разрешенная', 'Разрешенная'],
  },
  {
    name: 'ТН-КРОВЛЯ Гарант',
    values: ['Запрещенная', 'Разрешенная', 'н/д'],
  },
]

const databaseOrders = [
  { name: '№ ТД-Р-143 от 24.10.2025', createdAt: '24.10.25', updatedAt: '09.07.26' },
  { name: '№ ТД-Р-126 от 13.09.2022', createdAt: '13.09.22', updatedAt: '09.07.26' },
  { name: '№ ТД-Р-85 от 05.05.2020', createdAt: '05.05.20', updatedAt: '09.07.26' },
]

const documentRows = Array.from({ length: 4 }, (_, index) => ({
  id: index + 1,
  name: 'ТН-СТИЛОБАТ КЛАССИК АВТО',
  comment:
    'Тут комментарий может отображаться в несколько строк, если много текста нужно оставить',
  document: 'название документа',
}))

const selectedOrder = ref(orders[0])
const selectedBefore = ref(classOptions[0])
const selectedAfter = ref(classOptions[1])
const selectedClass = ref(classOptions[0])
const selectedCurator = ref(curatorOptions[0])
const selectedSystemType = ref(systemTypes[0])
const isSystemTypesOpen = ref(false)
const selectedHistorySystem = ref<(typeof systemRows)[number] | null>(null)
const isHistoryOpen = ref(false)
const openedSelect = ref<string | null>(null)

function setPage(page: string) {
  activePage.value = page
  openedSelect.value = null
}

function toggleSelect(name: string) {
  openedSelect.value = openedSelect.value === name ? null : name
}

function selectValue(name: string, value: string) {
  const setters: Record<string, (nextValue: string) => void> = {
    order: (nextValue) => (selectedOrder.value = nextValue),
    before: (nextValue) => (selectedBefore.value = nextValue),
    after: (nextValue) => (selectedAfter.value = nextValue),
    class: (nextValue) => (selectedClass.value = nextValue),
    curator: (nextValue) => (selectedCurator.value = nextValue),
  }

  setters[name]?.(value)
  openedSelect.value = null
}

function selectSystemType(type: (typeof systemTypes)[number]) {
  selectedSystemType.value = type
}

function openSystemHistory(system: (typeof systemRows)[number]) {
  selectedHistorySystem.value = system
  isHistoryOpen.value = false
}

function closeSystemHistory() {
  selectedHistorySystem.value = null
  isHistoryOpen.value = false
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
</script>

<template>
  <div class="app" @click="openedSelect = null">
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
                <span>{{ selectedOrder }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order"
                    class="custom-select__option"
                    :class="{ 'is-selected': order === selectedOrder }"
                    type="button"
                    @click="selectValue('order', order)"
                  >
                    {{ order }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
        </div>

        <section class="summary-grid" aria-label="Сводка изменений">
          <article class="summary-card">
            <span>Добавлено</span>
            <strong>30 систем</strong>
          </article>

          <div class="status-stack">
            <article class="status-card status-card--recommended">
              <strong>20</strong>
              <span>Рекомендованных</span>
            </article>
            <article class="status-card status-card--allowed">
              <strong>10</strong>
              <span>Разрешенных</span>
            </article>
          </div>

          <article class="summary-card">
            <span>Изм. классификация</span>
            <strong>6 систем</strong>
          </article>
        </section>

        <section class="filter-panel" aria-label="Тип строительства">
          <h2>Тип строительства</h2>
          <div class="type-tabs">
            <button
              v-for="type in constructionTypes"
              :key="type"
              class="type-tab"
              :class="{ 'type-tab--active': type === 'Промышленное и гражданское строительство' }"
              type="button"
            >
              {{ type }}
            </button>
          </div>
        </section>

        <section class="table-toolbar" aria-label="Управление таблицей">
          <div class="select-field">
            <span>Было</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'before' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('before')">
                <span>{{ selectedBefore }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'before'" class="custom-select__menu">
                  <button
                    v-for="option in classOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedBefore }"
                    type="button"
                    @click="selectValue('before', option)"
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
                <span>{{ selectedAfter }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'after'" class="custom-select__menu">
                  <button
                    v-for="option in classOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedAfter }"
                    type="button"
                    @click="selectValue('after', option)"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
          <button class="export-button" type="button">Экспортировать таблицу</button>
        </section>

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
              <tr v-for="row in changesRows" :key="`${row.name}-${row.before}-${row.after}`">
                <td>{{ row.name }}</td>
                <td :class="classModifier(row.before) && `status-cell status-cell--${classModifier(row.before)}`">
                  {{ row.before }}
                </td>
                <td :class="classModifier(row.after) && `status-cell status-cell--${classModifier(row.after)}`">
                  {{ row.after }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <button class="more-button" type="button">Показать еще</button>
      </section>

      <section v-else-if="activePage === 'systems'" class="systems-page">
        <div class="order-line">
          <div class="select-field">
            <span>Распоряжение</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'order' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('order')">
                <span>{{ selectedOrder }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order"
                    class="custom-select__option"
                    :class="{ 'is-selected': order === selectedOrder }"
                    type="button"
                    @click="selectValue('order', order)"
                  >
                    {{ order }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>
        </div>

        <section class="systems-summary" aria-label="Сводка систем">
          <article class="summary-card">
            <strong>300</strong>
            <span>систем</span>
          </article>

          <div class="status-stack">
            <article class="status-card status-card--recommended">
              <strong>200</strong>
              <span>Рекомендованных</span>
            </article>
            <article class="status-card status-card--allowed">
              <strong>80</strong>
              <span>Разрешенных</span>
            </article>
            <article class="status-card status-card--forbidden">
              <strong>20</strong>
              <span>Запрещенных</span>
            </article>
          </div>

          <article class="summary-card">
            <strong>15</strong>
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
                :class="{ 'type-tab--active': type === 'Промышленное и гражданское строительство' }"
                type="button"
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
                <input type="search" placeholder="Поиск по названию или ЕКН" />
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
                <span>{{ selectedClass }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'class'" class="custom-select__menu">
                  <button
                    v-for="option in classOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedClass }"
                    type="button"
                    @click="selectValue('class', option)"
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
                <span>{{ selectedCurator }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'curator'" class="custom-select__menu">
                  <button
                    v-for="option in curatorOptions"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === selectedCurator }"
                    type="button"
                    @click="selectValue('curator', option)"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <button class="export-button" type="button">Экспортировать таблицу</button>
        </section>

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
              <tr v-for="row in systemRows" :key="row.code">
                <td>{{ row.code }}</td>
                <td>{{ row.name }}</td>
                <td :class="`status-cell status-cell--${classModifier(row.systemClass)}`">
                  <span>{{ row.systemClass }}</span>
                  <button class="status-cell__icon" type="button" @click.stop="openSystemHistory(row)">
                    <img :src="folderIcon" alt="" aria-hidden="true" />
                  </button>
                </td>
                <td>{{ row.curator }}</td>
                <td>
                  <span class="compare-mark" :class="{ 'is-checked': row.compared }">
                    {{ row.compared ? '✓' : '' }}
                  </span>
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
                <span>{{ selectedOrder }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order"
                    class="custom-select__option"
                    :class="{ 'is-selected': order === selectedOrder }"
                    type="button"
                    @click="selectValue('order', order)"
                  >
                    {{ order }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <label class="search-field classification-search">
            <span>Поиск</span>
            <input type="search" placeholder="Поиск по названию или ЕКН" />
          </label>
        </div>

        <section class="filter-panel classification-construction" aria-label="Тип строительства">
          <h2>Тип строительства</h2>
          <div class="type-tabs">
            <button
              v-for="type in constructionTypes"
              :key="type"
              class="type-tab"
              :class="{ 'type-tab--active': type === 'Промышленное и гражданское строительство' }"
              type="button"
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
                  v-for="type in systemTypes.slice(0, 16)"
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
          <section class="classification-cards" aria-label="Системы классификации">
            <article
              v-for="system in classificationSystems"
              :key="system.code"
              class="classification-card"
              :class="`classification-card--${classModifier(system.systemClass)}`"
            >
              <header class="classification-card__header">
                <a href="https://nav.tn.ru/systems/" target="_blank" rel="noreferrer">{{ system.name }}</a>
                <a class="classification-card__source" href="https://nav.tn.ru/systems/" target="_blank" rel="noreferrer" aria-label="Открыть на nav.tn.ru">
                  <img :src="browseIcon" alt="" aria-hidden="true" />
                </a>
              </header>

              <dl class="classification-card__meta">
                <div>
                  <dt>Шифр</dt>
                  <dd>{{ system.code }}</dd>
                </div>
                <div>
                  <dt>Класс</dt>
                  <dd>{{ system.base }}</dd>
                </div>
              </dl>

              <button class="classification-card__more" type="button" aria-label="Открыть действия">
                <i aria-hidden="true" />
              </button>
            </article>
          </section>

          <aside class="classification-sidebar" aria-label="Фильтры классификации">
            <button
              v-for="filter in classificationFilterGroups"
              :key="filter"
              class="classification-sidebar__item"
              type="button"
            >
              <span>{{ filter }}</span>
              <i aria-hidden="true" />
            </button>
          </aside>
        </div>
      </section>

      <section v-else-if="activePage === 'comparison'" class="comparison-page">
        <div class="comparison-controls">
          <h1>Выбрать распоряжения, изменения классов систем по которым нужно сравнить</h1>

          <div class="comparison-controls__row">
            <div
              v-for="(order, index) in comparisonOrders"
              :key="order"
              class="custom-select comparison-order"
              :class="{ 'is-open': openedSelect === `comparison-${index}` }"
            >
              <button class="custom-select__button" type="button" @click.stop="toggleSelect(`comparison-${index}`)">
                <span>{{ order }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === `comparison-${index}`" class="custom-select__menu">
                  <button
                    v-for="option in orders"
                    :key="option"
                    class="custom-select__option"
                    :class="{ 'is-selected': option === order }"
                    type="button"
                    @click="openedSelect = null"
                  >
                    {{ option }}
                  </button>
                </div>
              </Transition>
            </div>

            <button class="comparison-add-button" type="button" aria-label="Добавить распоряжение">+</button>
          </div>
        </div>

        <div class="systems-table comparison-table">
          <table>
            <thead>
              <tr>
                <th>Название системы</th>
                <th v-for="order in comparisonOrders" :key="order">
                  {{ order.replace('№ ', '№\u00A0').replace(' от ', ' от\n') }}
                </th>
                <th>Удалить из сравнения</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in comparisonRows" :key="`${row.name}-${row.values.join('-')}`">
                <td>{{ row.name }}</td>
                <td
                  v-for="value in row.values"
                  :key="value"
                  :class="classModifier(value) && `status-cell status-cell--${classModifier(value)}`"
                >
                  {{ value }}
                </td>
                <td>
                  <button class="comparison-delete-button" type="button" aria-label="Удалить из сравнения">
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
          <label class="settings-inline-field">
            <span>Частота обновления БД фильтров и ссылок на системы</span>
            <input type="text" value="XX дней" />
          </label>
        </section>

        <section class="settings-section" aria-labelledby="orders-db-title">
          <div class="settings-section__header">
            <h2 id="orders-db-title">Управление БД Распоряжений</h2>
          </div>

          <div class="systems-table settings-orders-table">
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
                <tr v-for="order in databaseOrders" :key="order.name">
                  <td>{{ order.name }}</td>
                  <td>{{ order.createdAt }}</td>
                  <td>{{ order.updatedAt }}</td>
                  <td>
                    <button class="icon-action-button" type="button" aria-label="Удалить БД">
                      <img :src="trashIcon" alt="" aria-hidden="true" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="create-order-line">
            <span>Создать новую БД распоряжений</span>
            <button class="small-red-button" type="button">+</button>
          </div>
        </section>

        <section class="settings-section" aria-labelledby="edit-db-title">
          <div class="settings-section__header">
            <h2 id="edit-db-title">Редактирование БД</h2>
          </div>

          <div class="settings-edit-select">
            <span>Выбрать БД распоряжения для редактирования</span>
            <div class="custom-select" :class="{ 'is-open': openedSelect === 'settings-order' }">
              <button class="custom-select__button" type="button" @click.stop="toggleSelect('settings-order')">
                <span>{{ selectedOrder }}</span>
                <i aria-hidden="true" />
              </button>
              <Transition name="select-menu">
                <div v-if="openedSelect === 'settings-order'" class="custom-select__menu">
                  <button
                    v-for="order in orders"
                    :key="order"
                    class="custom-select__option"
                    :class="{ 'is-selected': order === selectedOrder }"
                    type="button"
                    @click="selectValue('order', order)"
                  >
                    {{ order }}
                  </button>
                </div>
              </Transition>
            </div>
          </div>

          <section class="settings-table-block" aria-label="Таблица 1">
            <div class="settings-table-toolbar">
              <span>Таблица 1</span>
              <label class="settings-search">
                <input type="search" placeholder="Поиск по названию или ЕКН" />
              </label>
              <button class="import-button" type="button">Импортировать таблицу</button>
            </div>

            <div class="systems-table settings-data-table">
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
                  <tr v-for="row in changesRows" :key="`settings-${row.name}-${row.before}-${row.after}`">
                    <td>{{ row.name }}</td>
                    <td :class="classModifier(row.before) && `status-cell status-cell--${classModifier(row.before)}`">
                      {{ row.before }}
                    </td>
                    <td :class="classModifier(row.after) && `status-cell status-cell--${classModifier(row.after)}`">
                      {{ row.after }}
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
                <input type="search" placeholder="Поиск по названию или ЕКН" />
              </label>
              <button class="import-button" type="button">Импортировать таблицу</button>
            </div>

            <div class="systems-table settings-data-table">
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
                  <tr v-for="row in systemRows.slice(0, 4)" :key="`settings-system-${row.code}`">
                    <td>{{ row.code }}</td>
                    <td>{{ row.name }}</td>
                    <td :class="`status-cell status-cell--${classModifier(row.systemClass)}`">
                      {{ row.systemClass }}
                    </td>
                    <td>{{ row.curator }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="settings-table-block" aria-label="Таблица 3">
            <div class="settings-table-toolbar">
              <span>Таблица 3</span>
              <label class="settings-search">
                <input type="search" placeholder="Поиск по названию или ЕКН" />
              </label>
            </div>

            <div class="systems-table settings-docs-table">
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
                  <tr v-for="row in documentRows" :key="`document-${row.id}`">
                    <td>{{ row.name }}</td>
                    <td>{{ row.comment }}</td>
                    <td>
                      <a class="settings-document-link" href="#">
                        <img :src="openIcon" alt="" aria-hidden="true" />
                        {{ row.document }}
                      </a>
                    </td>
                    <td>
                      <div class="document-actions">
                        <button class="icon-action-button" type="button" aria-label="Открыть папку">
                          <img :src="folderIcon" alt="" aria-hidden="true" />
                        </button>
                        <button class="icon-action-button" type="button" aria-label="Обновить документ">
                          <img :src="reverseIcon" alt="" aria-hidden="true" />
                        </button>
                        <button class="icon-action-button icon-action-button--danger" type="button" aria-label="Удалить документ">
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
              <h2>{{ selectedHistorySystem.name }}</h2>
              <a class="system-history-source" href="https://nav.tn.ru/systems/" target="_blank" rel="noreferrer" aria-label="Открыть на nav.tn.ru">
                <img :src="browseIcon" alt="" aria-hidden="true" />
              </a>
            </header>

            <div class="history-table">
              <table>
                <thead>
                  <tr>
                    <th>Распоряжение</th>
                    <th>Комментарий</th>
                    <th>Документ</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td>№ ТД-Р-143 от 24.10.2025</td>
                    <td>Тут комментарий может отображаться в несколько строк, если много текста нужно оставить</td>
                    <td>
                      <a class="document-link" href="#">
                        <img :src="openIcon" alt="" aria-hidden="true" />
                        открыть
                      </a>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <button class="history-toggle" type="button" @click="isHistoryOpen = !isHistoryOpen">
              {{ isHistoryOpen ? 'скрыть историю изменений' : 'развернуть историю изменений' }}
              <i aria-hidden="true" />
            </button>

            <Transition name="history-more">
              <div v-if="isHistoryOpen" class="history-table history-table--muted">
                <table>
                  <tbody>
                    <tr>
                      <td>№ ТД-Р-126 от 13.09.2022</td>
                      <td>Тут комментарий может отображаться в несколько строк, если много текста нужно оставить</td>
                      <td>
                        <a class="document-link" href="#">
                          <img :src="openIcon" alt="" aria-hidden="true" />
                          открыть
                        </a>
                      </td>
                    </tr>
                    <tr>
                      <td>№ ТД-Р-85 от 05.05.2020</td>
                      <td>Тут комментарий может отображаться в несколько строк, если много текста нужно оставить</td>
                      <td>
                        <a class="document-link" href="#">
                          <img :src="openIcon" alt="" aria-hidden="true" />
                          открыть
                        </a>
                      </td>
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
