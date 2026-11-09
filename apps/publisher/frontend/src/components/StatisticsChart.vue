<script setup lang="ts">
import {computed,onMounted,onUnmounted,ref,watch} from 'vue';
type Sample={updatedAt:string;items:{label:string;value:string}[]};
const props=defineProps<{samples:Sample[];metric:string}>();
const days=ref(28),selected=ref(-1),width=ref(720),element=ref<HTMLElement|null>(null);
let observer:ResizeObserver|undefined,frame=0;
onMounted(()=>{observer=new ResizeObserver(entries=>{const next=Math.floor(entries[0]?.contentRect.width||720);cancelAnimationFrame(frame);frame=requestAnimationFrame(()=>{if(width.value!==next)width.value=next})});if(element.value)observer.observe(element.value)});
onUnmounted(()=>{observer?.disconnect();cancelAnimationFrame(frame)});
const observations=computed(()=>props.samples.filter(s=>Date.parse(s.updatedAt)>=Date.now()-days.value*86400000));
const points=computed(()=>observations.value.map(s=>({time:Date.parse(s.updatedAt),value:s.items.find(m=>m.label===props.metric)?.value})).filter(p=>Number.isFinite(p.time)));
const available=computed(()=>points.value.filter(p=>p.value!==undefined));
const ceiling=computed(()=>Math.max(1,...available.value.map(p=>Number(p.value))));
const start=computed(()=>points.value[0]?.time||0),end=computed(()=>points.value.at(-1)?.time||0);
const x=(t:number)=>48+(t-start.value)/Math.max(1,end.value-start.value)*Math.max(1,width.value-96);
const y=(v:string)=>164-Number(v)/ceiling.value*144;
const path=computed(()=>{let connected=false;return points.value.map(p=>{if(p.value===undefined){connected=false;return ''}const segment=`${connected?'L':'M'}${x(p.time)},${y(p.value)}`;connected=true;return segment}).join(' ')});
const current=computed(()=>available.value[Math.min(Math.max(0,selected.value),available.value.length-1)]);
watch([()=>props.metric,()=>props.samples,days],()=>{selected.value=Math.max(0,available.value.length-1)},{immediate:true});
function hover(e:PointerEvent){
 const rect=(e.currentTarget as SVGElement).getBoundingClientRect();
 const t=start.value+((e.clientX-rect.left)/rect.width*width.value-48)/Math.max(1,width.value-96)*(end.value-start.value);
 selected.value=available.value.reduce((best,p,i,a)=>Math.abs(p.time-t)<Math.abs(a[best]!.time-t)?i:best,0);
}
function keyboard(e:KeyboardEvent){
 if(!['ArrowLeft','ArrowRight','Home','End'].includes(e.key))return;
 e.preventDefault();selected.value=e.key==='Home'?0:e.key==='End'?available.value.length-1:Math.min(available.value.length-1,Math.max(0,selected.value+(e.key==='ArrowRight'?1:-1)));
}
const compact=(n:number)=>new Intl.NumberFormat(undefined,{notation:'compact',maximumFractionDigits:1}).format(n);
const count=(v:string)=>{try{return new Intl.NumberFormat().format(BigInt(v))}catch{return v}};
const date=(t:number)=>new Date(t).toLocaleString([],{month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'});
</script>
<template>
 <div ref="element" class="statistics-chart">
  <div class="chart-toolbar"><span>{{metric}} over time</span><select v-model="days" class="ui-field" aria-label="Chart period"><option :value="7">Last 7 days</option><option :value="28">Last 28 days</option><option :value="90">Last 90 days</option><option :value="365">Last year</option></select></div>
  <template v-if="available.length>1">
   <svg :viewBox="`0 0 ${width} 200`" role="img" :aria-label="metric+' observed totals. Use left and right arrows to inspect measurements.'" tabindex="0" @pointermove="hover" @keydown="keyboard">
    <g v-for="fraction in [0,0.5,1]" :key="fraction"><line x1="48" :x2="width-48" :y1="164-fraction*144" :y2="164-fraction*144" class="grid"/><text x="38" :y="168-fraction*144" text-anchor="end">{{compact(ceiling*fraction)}}</text></g>
    <path :d="path" class="trend"/>
    <g v-if="current&&current.value!==undefined"><line :x1="x(current.time)" :x2="x(current.time)" y1="20" y2="164" class="cursor"/><circle :cx="x(current.time)" :cy="y(current.value)" r="4"/></g>
    <text x="48" y="192">{{date(start)}}</text><text :x="width-48" y="192" text-anchor="end">{{date(end)}}</text>
   </svg>
   <p class="chart-reading" aria-live="polite"><template v-if="current&&current.value!==undefined"><time>{{date(current.time)}}</time><strong>{{count(current.value)}} {{metric.toLowerCase()}}</strong></template></p>
  </template>
  <p v-else class="chart-empty">{{available.length?'First measurement saved. Refresh later to see how it changes.':'No measurements in this period.'}}</p>
  <p class="chart-caption">Local history · recorded on refresh</p>
 </div>
</template>
<style scoped>
.statistics-chart{margin-top:20px}.chart-toolbar{display:flex;align-items:center;justify-content:space-between;gap:16px;font-size:var(--text-xs);color:var(--secondary)}.chart-toolbar select{width:auto;max-width:170px;font-size:var(--text-xs);min-height:32px;padding-top:6px;padding-bottom:6px}.statistics-chart svg{display:block;width:100%;height:200px;margin-top:16px;overflow:visible;touch-action:pan-y}.statistics-chart text{fill:var(--secondary);font-size:11px;font-variant-numeric:tabular-nums}.grid{stroke:var(--line);stroke-width:1}.trend{fill:none;stroke:var(--accent);stroke-width:2;stroke-linejoin:round;vector-effect:non-scaling-stroke}.cursor{stroke:var(--secondary);stroke-dasharray:3 4;opacity:.45}.statistics-chart circle{fill:var(--accent);stroke:var(--surface);stroke-width:2}.chart-reading{display:flex;flex-wrap:wrap;gap:8px 20px;justify-content:space-between;font-size:var(--text-xs);margin:8px 0;color:var(--secondary)}.chart-reading strong{font-weight:500;color:var(--text)}.chart-caption{font-size:var(--text-xs);color:var(--muted);margin:12px 0 0}.chart-empty{font-size:var(--text-sm);color:var(--secondary);padding:28px 0;margin:0;line-height:1.6}
</style>
