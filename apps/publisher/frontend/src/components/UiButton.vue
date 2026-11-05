<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'quiet' | 'danger';
  size?: 'default' | 'compact';
  type?: 'button' | 'submit' | 'reset';
  disabled?: boolean;
  busy?: boolean;
}>(), { variant: 'secondary', size: 'default', type: 'button', disabled: false, busy: false });
</script>

<template>
  <button :type="type" class="ui-button" :class="[`ui-button--${variant}`, { 'ui-button--compact': size === 'compact' }]"
    :disabled="disabled || busy" :aria-busy="busy || undefined"><slot /></button>
</template>
