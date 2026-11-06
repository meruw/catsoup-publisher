<script setup lang="ts">
import {onMounted,onBeforeUnmount,ref} from 'vue';
import UiButton from './UiButton.vue';
const props=defineProps<{title:string;busy?:boolean;returnFocusId?:string}>();
const emit=defineEmits<{close:[]}>();
const dialog=ref<HTMLDialogElement|null>(null);
let restoreFocus=false;
function close(){if(!props.busy){restoreFocus=true;emit('close');}}
onMounted(()=>dialog.value?.showModal());
onBeforeUnmount(()=>{
  dialog.value?.close();
  // Navigation unlocks its controls in the same Vue update as the unmount.
  if(restoreFocus)requestAnimationFrame(()=>{if(props.returnFocusId)document.getElementById(props.returnFocusId)?.focus();});
});
</script>
<template>
  <dialog ref="dialog" class="ui-dialog" aria-labelledby="publish-dialog-title" @cancel.prevent="close">
    <header><h2 id="publish-dialog-title">{{title}}</h2><UiButton variant="quiet" size="compact" :disabled="busy" aria-label="Close publishing options" title="Close" @click="close"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18"/></svg></UiButton></header>
    <slot/>
  </dialog>
</template>
<style scoped>
.ui-dialog{width:480px;max-width:calc(100vw - 48px);max-height:calc(100vh - 96px);margin:auto;padding:28px;border:1px solid var(--line);border-radius:12px;background:var(--page);color:var(--text);overflow:auto;box-shadow:0 16px 64px #0002}
.ui-dialog::backdrop{background:#18141066}
.ui-dialog>header{display:flex;align-items:center;justify-content:space-between;gap:20px;margin-bottom:24px}
.ui-dialog h2{font-size:var(--text-section);font-weight:500}
@media(max-height:640px){.ui-dialog{max-height:calc(100vh - 48px);padding:24px}}
</style>
