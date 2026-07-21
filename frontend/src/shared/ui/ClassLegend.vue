<script setup lang="ts">
withDefaults(defineProps<{ includeForbidden?: boolean }>(), {
  includeForbidden: true,
})
</script>

<template>
  <div class="classification-class-legend">
    <article class="classification-class-legend__item classification-class-legend__item--recommended">
      <span class="classification-class-legend__marker" aria-hidden="true" />
      <p><strong>Рекомендованные</strong> — системы ТН, обладающие исключительными характеристиками, направленными на повышение безопасности, энергоэффективности и долговечности зданий и сооружений.</p>
    </article>
    <article class="classification-class-legend__item classification-class-legend__item--allowed">
      <span class="classification-class-legend__marker" aria-hidden="true" />
      <p><strong>Разрешённые</strong> — системы ТН, обладающие высокими характеристиками, направленными на повышение безопасности, энергоэффективности и долговечности зданий и сооружений, но имеющие более низкий приоритет по сравнению с рекомендованными системами.</p>
    </article>
    <article v-if="includeForbidden" class="classification-class-legend__item classification-class-legend__item--forbidden">
      <span class="classification-class-legend__marker" aria-hidden="true" />
      <p><strong>Запрещённые</strong> — системы ТН, противоречащие нормам безопасности и внутренним нормам ТН и/или не имеющие необходимой разрешительной документации для применения в стране присутствия. К этой категории относятся решения, не соответствующие законодательству и/или нормативной документации по механической, пожарной, санитарно-эпидемиологической, экологической и другим видам безопасности.</p>
    </article>
  </div>
</template>
