<script setup lang="ts">
import UiButton from './UiButton.vue';
defineProps<{ modelValue: string; label: string; disabled?: boolean; options: { value: string; label: string; accessibleLabel?: string }[] }>();
defineEmits<{ 'update:modelValue': [value: string] }>();
</script>

<template>
  <div class="ui-segmented" role="group" :aria-label="label">
    <UiButton v-for="option in options" :key="option.value" variant="quiet" size="compact"
      :aria-pressed="modelValue === option.value" :aria-label="option.accessibleLabel || option.label" :disabled="disabled"
      @click="$emit('update:modelValue', option.value)">{{ option.label }}</UiButton>
  </div>
</template>
