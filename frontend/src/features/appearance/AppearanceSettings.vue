<script setup lang="ts">
export type FontSizePreset = 'small' | 'standard' | 'large'

export type FontSizePresetOption = {
  key: FontSizePreset
  label: string
  size: number
}

defineProps<{
  modelValue: FontSizePreset
  presets: FontSizePresetOption[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: FontSizePreset]
}>()
</script>

<template>
  <section class="settings-section appearance-settings" aria-labelledby="appearance-settings-title">
    <div class="appearance-settings__intro">
      <span class="settings-section__eyebrow">Интерфейс</span>
      <h1 id="appearance-settings-title">Внешний вид</h1>
      <p>Минимальный размер текста во всём интерфейсе</p>
    </div>
    <div
      class="appearance-settings__segment"
      :class="`is-${modelValue}`"
      role="radiogroup"
      aria-label="Минимальный размер текста"
    >
      <span class="appearance-settings__indicator" aria-hidden="true" />
      <button
        v-for="preset in presets"
        :key="preset.key"
        :class="{ 'is-selected': modelValue === preset.key }"
        type="button"
        role="radio"
        :aria-checked="modelValue === preset.key"
        @click="emit('update:modelValue', preset.key)"
      >
        <strong>{{ preset.label }}</strong>
      </button>
    </div>
  </section>
</template>
