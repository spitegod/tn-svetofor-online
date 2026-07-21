<script setup lang="ts">
import { Info } from '@lucide/vue'
import ClassLegend from '@/shared/ui/ClassLegend.vue'

defineProps<{
  includeForbidden?: boolean
}>()
</script>

<template>
  <footer
    class="page-class-footer"
    :class="{ 'page-class-footer--without-forbidden': !includeForbidden }"
    aria-label="Легенда классов систем"
  >
    <div class="page-class-footer__heading">
      <Info :size="17" :stroke-width="1.9" aria-hidden="true" />
      <strong>Классификация систем</strong>
    </div>
    <ClassLegend class="page-class-footer__legend" :include-forbidden="includeForbidden" />
  </footer>
</template>
