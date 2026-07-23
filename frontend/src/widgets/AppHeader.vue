<script setup lang="ts">
import { Info } from '@lucide/vue'
import logo from '@/shared/assets/tn-svetofor-logo.svg'
import ClassLegend from '@/shared/ui/ClassLegend.vue'

type NavigationItem = {
  key: string
  label: string
}

defineProps<{
  activePage: string
  items: NavigationItem[]
  isLegendOpen: boolean
}>()

const emit = defineEmits<{
  navigate: [page: string]
  'update:isLegendOpen': [value: boolean]
}>()
</script>

<template>
  <header class="site-header">
    <div class="header-container">
      <div class="main-header__inner">
        <a class="brand-link" href="/" aria-label="ТН-Светофор">
          <img :src="logo" alt="ТН-Светофор" />
        </a>

        <nav class="primary-nav" aria-label="Основная навигация">
          <button
            v-for="item in items"
            :key="item.key"
            class="primary-nav__item"
            :class="{ 'is-active': activePage === item.key }"
            type="button"
            @click.stop="emit('navigate', item.key)"
          >
            {{ item.label }}
          </button>
        </nav>

        <div class="header-class-legend" @click.stop>
          <button
            class="header-class-legend__toggle"
            :class="{ 'is-open': isLegendOpen }"
            type="button"
            aria-label="Показать легенду классов систем"
            title="Легенда классов систем"
            :aria-expanded="isLegendOpen"
            aria-controls="global-class-legend"
            @click="emit('update:isLegendOpen', !isLegendOpen)"
          >
            <Info :size="18" :stroke-width="2" aria-hidden="true" />
          </button>
          <Transition name="header-legend">
            <ClassLegend
              v-if="isLegendOpen"
              id="global-class-legend"
              class="header-class-legend__panel"
            />
          </Transition>
        </div>
      </div>
    </div>
  </header>
</template>
