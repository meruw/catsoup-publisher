<script setup lang="ts">
import refresh from '../assets/icons/refresh-cw.svg?raw';
import minus from '../assets/icons/minus.svg?raw';
import square from '../assets/icons/square.svg?raw';
import copy from '../assets/icons/copy.svg?raw';
import close from '../assets/icons/x.svg?raw';
import chevron from '../assets/icons/chevron-down.svg?raw';
const icons={refresh,minus,square,copy,close,chevron};
defineProps<{name:keyof typeof icons}>();
// Only these reviewed, vendored SVG literals are rendered; no user/provider HTML.
</script>
<template><span class="ui-icon" aria-hidden="true" v-html="icons[name]"></span></template>
<style scoped>.ui-icon{display:inline-flex;width:16px;height:16px;flex-shrink:0;pointer-events:none}.ui-icon :deep(svg){width:100%;height:100%;fill:none;stroke:currentColor;stroke-width:1.75}</style>
