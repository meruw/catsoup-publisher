<script setup lang="ts">
import {onUnmounted,ref,watch} from 'vue';
import {readStatistics,statisticsKey,statisticsNote} from '../services/statistics';
import type {PostStatus} from '../services/drafts';
import UiButton from './UiButton.vue';
import StatisticsChart from './StatisticsChart.vue';
const props=defineProps<{post:PostStatus;heading?:string}>();
const metrics=ref<{label:string;value:string}[]>([]),busy=ref(false),error=ref(''),note=ref(''),updated=ref('');
const history=ref<{updatedAt:string;items:{label:string;value:string}[]}[]>([]),historyNote=ref(''),metric=ref('');
let generation=0;
const reset=()=>{++generation;metrics.value=[];history.value=[];historyNote.value='';metric.value='';busy.value=false;error.value='';note.value='';updated.value='';};
watch(()=>statisticsKey(props.post),()=>{reset();note.value=statisticsNote(props.post);if(!note.value)void refresh(false);},{immediate:true});
onUnmounted(()=>{++generation});
async function refresh(force=true){
 if(statisticsNote(props.post))return false;
 const request=++generation;busy.value=true;error.value='';
 try {const result=await readStatistics(props.post,force);if(request!==generation)return false;
 metrics.value=result.items||[];updated.value=result.updatedAt;note.value=result.note;
 history.value=(result.history||[]).map(s=>({...s,items:s.items||[]}));historyNote.value=result.historyNote||'';
 if(!metrics.value.some(m=>m.label===metric.value))metric.value=metrics.value[0]?.label||'';
 return true;
 } catch(e){if(request===generation)error.value=String(e).replace(/^RuntimeError:\s*/,'');return false;}
 finally{if(request===generation)busy.value=false;}
}
defineExpose({refresh,eligible:()=>!statisticsNote(props.post)});
const count=(value:string)=>{try{return new Intl.NumberFormat().format(BigInt(value))}catch{return value}};
</script>
<template>
 <section class="post-statistics" aria-label="Post statistics" :aria-busy="busy">
  <header><h3>{{heading || 'Statistics'}}</h3><UiButton variant="quiet" size="compact" :busy="busy" v-if="!statisticsNote(post)" :disabled="busy" @click="refresh(true)">{{busy?'Loadingâ€¦':updated?'Refresh statistics':'Retry statistics'}}</UiButton></header>
  <dl v-if="metrics.length" class="metric-strip"><div v-for="item in metrics" :key="item.label" :class="{active:metric===item.label}"><dt><button :aria-pressed="metric===item.label" @click="metric=item.label">{{item.label}}</button></dt><dd>{{count(item.value)}}</dd></div></dl>
  <StatisticsChart v-if="history.length&&metric" :samples="history" :metric="metric"/>
  <p v-if="historyNote" class="statistics-error" role="status">{{historyNote}}</p>
  <p v-if="updated" class="statistics-note">{{error?'Last successful update':'Updated'}} {{new Date(updated).toLocaleString()}}</p>
  <p v-else-if="!note&&!error" class="statistics-note">{{busy?'Reading available countsâ€¦':'No counts loaded'}}</p>
  <p v-if="note" class="statistics-note">{{note}}</p>
  <p v-if="error" role="alert" class="statistics-error">{{error}} Counts may require renewed account permissions.</p>
 </section>
</template>
<style scoped>
.post-statistics{border-top:1px solid var(--line);padding:20px 0;margin:8px 0 20px}.post-statistics header{display:flex;align-items:center;justify-content:space-between;gap:16px}.post-statistics h3{font-size:var(--text-sm);font-weight:500;margin:0}.metric-strip{display:flex;flex-wrap:wrap;gap:20px 32px;margin:20px 0 12px}.metric-strip>div{display:block;border:0;padding:0}.metric-strip dt{font-size:var(--text-xs);color:var(--secondary)}.metric-strip dd{font-size:22px;font-variant-numeric:tabular-nums;margin:6px 0 0}.statistics-note,.statistics-error{font-size:var(--text-xs);line-height:1.6;margin:12px 0 0;color:var(--secondary)}.statistics-error{color:var(--danger)}
.metric-strip>div{position:relative;padding-bottom:12px;border-bottom:2px solid transparent}.metric-strip>div.active{border-bottom-color:var(--accent)}.metric-strip dt button{font:inherit;color:inherit;background:none;border:0;padding:0;cursor:pointer}.metric-strip dt button::after{content:"";position:absolute;inset:0}.metric-strip dt button:focus-visible{outline-offset:5px}
</style>
