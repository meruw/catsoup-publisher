<script setup lang="ts">
import {ref} from 'vue';
defineProps<{text:string}>();
const dismissed=ref(false);
</script>
<template>
  <span class="help-tip" :class="{dismissed}" @mouseenter="dismissed=false" @focusin="dismissed=false" @keydown.esc.stop="dismissed=true">
    <button type="button" class="help-trigger" :aria-label="text" @click="dismissed=false">?</button>
    <span class="help-content" role="tooltip">{{text}}</span>
  </span>
</template>
<style scoped>
.help-tip{position:relative;display:inline-flex;vertical-align:middle;margin-left:4px}
.help-trigger{display:grid;place-items:center;width:24px;height:24px;border:0;border-radius:4px;background:transparent;color:var(--muted);font-size:12px}
.help-trigger:hover{background:var(--raised);color:var(--text)}
.help-content{display:none;position:absolute;z-index:5;top:calc(100% + 8px);right:0;width:240px;max-width:60vw;padding:12px;background:var(--text);color:var(--page);border-radius:8px;font-size:12px;font-weight:400;line-height:1.5;text-align:left}
.help-tip:not(.dismissed):is(:hover,:focus-within) .help-content{display:block}
</style>
