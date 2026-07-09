<script setup lang="ts">
import logo from '@/shared/assets/logo.png'

const navItems = [
  { label: 'Изменения', active: false },
  { label: 'Список систем', active: true },
  { label: 'Классификация', active: false },
  { label: 'Сравнение', active: false },
  { label: 'Настройки', active: false },
]
</script>

<template>
  <div class="app">
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

    <main class="page-main" />
  </div>
</template>
