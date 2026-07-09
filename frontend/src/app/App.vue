<script setup lang="ts">
import { ref } from 'vue'
import logo from '@/shared/assets/logo.png'

const navItems = [
  { label: 'Изменения', active: true },
  { label: 'Список систем', active: false },
  { label: 'Классификация', active: false },
  { label: 'Сравнение', active: false },
  { label: 'Настройки', active: false },
]

const constructionTypes = [
  'Все',
  'Промышленное и гражданское строительство',
  'Индивидуальное жилищное строительство',
  'Транспортное и дорожное строительство',
  'Специальные сооружения',
]

const orders = [
  '№ ТД-Р-143 от 24.10.2025',
  '№ ТД-Р-118 от 12.09.2025',
  '№ ТД-Р-097 от 18.08.2025',
]

const classOptions = [
  'Рекомендованная',
  'Разрешенная',
  'Запрещенная',
]

const selectedOrder = ref(orders[0])
const selectedBefore = ref(classOptions[0])
const selectedAfter = ref(classOptions[0])
const openedSelect = ref<string | null>(null)

function toggleSelect(name: string) {
  openedSelect.value = openedSelect.value === name ? null : name
}

function selectValue(name: string, value: string) {
  if (name === 'order') {
    selectedOrder.value = value
  }

  if (name === 'before') {
    selectedBefore.value = value
  }

  if (name === 'after') {
    selectedAfter.value = value
  }

  openedSelect.value = null
}

const systemRows = [
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК АВТО',
    before: 'Новая система',
    after: 'Разрешенная',
    afterStatus: 'allowed',
  },
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК ТРОТУАР',
    before: 'Новая система',
    after: 'Рекомендованная',
    afterStatus: 'recommended',
  },
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК АВТО',
    before: 'Новая система',
    after: 'Разрешенная',
    afterStatus: 'allowed',
  },
  {
    name: 'ТН-СТИЛОБАТ КЛАССИК',
    before: 'Рекомендованная',
    beforeStatus: 'recommended',
    after: 'Разрешенная',
    afterStatus: 'allowed',
  },
]
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
              :key="item.label"
              class="primary-nav__item"
              :class="{ 'is-active': item.active }"
              type="button"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>
      </div>
    </header>

    <main class="page-main">
      <section class="changes-page">
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
              <tr v-for="row in systemRows" :key="`${row.name}-${row.before}-${row.after}`">
                <td>{{ row.name }}</td>
                <td :class="row.beforeStatus && `status-cell status-cell--${row.beforeStatus}`">
                  {{ row.before }}
                </td>
                <td :class="row.afterStatus && `status-cell status-cell--${row.afterStatus}`">
                  {{ row.after }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <button class="more-button" type="button">Показать еще</button>
      </section>
    </main>
  </div>
</template>
