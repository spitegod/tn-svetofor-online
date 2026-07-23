<script setup lang="ts">
import AppearanceSettings, {
  type FontSizePreset,
  type FontSizePresetOption,
} from '@/features/appearance/AppearanceSettings.vue'

defineProps<{
  fontSizePreset: FontSizePreset
  fontSizePresets: FontSizePresetOption[]
}>()

const emit = defineEmits<{
  'update:fontSizePreset': [value: FontSizePreset]
}>()
</script>

<template>
  <section class="settings-page settings-page--appearance">
    <AppearanceSettings
      :model-value="fontSizePreset"
      :presets="fontSizePresets"
      @update:model-value="emit('update:fontSizePreset', $event)"
    />
  </section>
</template>
