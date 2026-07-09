<script setup lang="ts">
import { ref } from 'vue'
import logo from '@/shared/assets/logo.png'
import folderIcon from '@/shared/assets/folder.png'
import openIcon from '@/shared/assets/open.png'
import browseIcon from '@/shared/assets/browse.png'

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
