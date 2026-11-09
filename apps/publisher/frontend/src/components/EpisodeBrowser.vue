<script setup lang="ts">
import {computed,ref,watch} from 'vue';
import type {Entry} from '../services/publisher';
import UiButton from './UiButton.vue';
const props=defineProps<{entries:Entry[];selected?:string;disabled:boolean}>();
const emit=defineEmits<{select:[id:string]}>();
const search=ref(''),page=ref(0),size=6;
const matches=computed(()=>{const q=search.value.trim().toLocaleLowerCase();return props.entries.filter(e=>!q||`${e.title} ${e.id}`.toLocaleLowerCase().includes(q));});
const pages=computed(()=>Math.max(1,Math.ceil(matches.value.length/size)));
const visible=computed(()=>matches.value.slice(page.value*size,(page.value+1)*size));
watch(search,()=>page.value=0);
watch(pages,n=>page.value=Math.min(page.value,n-1));
watch(()=>props.selected,id=>{const index=matches.value.findIndex(e=>e.id===id);if(index>=0)page.value=Math.floor(index/size);});
</script>
<template>
 <section class="episode-browser" aria-label="Episode library">
  <input v-model="search" type="search" class="ui-field" aria-label="Search episodes" placeholder="Search episodes" maxlength="200">
  <nav aria-label="Publisher episodes">
   <button v-for="entry in visible" :key="entry.id" class="episode-button" :title="entry.title+(entry.valid?'':' · Needs attention')" :aria-current="selected===entry.id?'true':undefined" :disabled="disabled" @click="emit('select',entry.id)">
    <span>{{entry.title}}</span><small v-if="!entry.valid">Needs attention</small>
   </button>
  </nav>
  <p v-if="!matches.length" class="episode-empty">{{entries.length?'No matching episodes':'No deliveries yet'}}</p>
  <footer v-if="pages>1" class="episode-pages"><UiButton variant="quiet" size="compact" aria-label="Previous episodes" :disabled="page===0" @click="page--">Previous</UiButton><span>{{page+1}} / {{pages}}</span><UiButton variant="quiet" size="compact" aria-label="Next episodes" :disabled="page+1>=pages" @click="page++">Next</UiButton></footer>
 </section>
</template>
<style scoped>
.episode-browser{display:grid;gap:8px;min-height:0;margin-top:12px}.episode-browser input{width:100%;min-width:0;font-size:var(--text-sm)}nav{min-height:0;overflow-y:auto;padding:4px;margin:-4px;max-height:248px;align-content:start}.episode-button{padding:8px 12px;min-height:36px}.episode-button span{font-size:var(--text-sm);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.episode-empty{font-size:var(--text-xs);color:var(--muted);padding:8px}.episode-pages{padding:4px 0;display:flex;align-items:center;justify-content:space-between;gap:4px}.episode-pages span{font-size:var(--text-xs);color:var(--muted);font-variant-numeric:tabular-nums}.episode-pages :deep(button){padding:4px;min-width:0;font-size:var(--text-xs)}
</style>
