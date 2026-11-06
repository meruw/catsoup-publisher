<script setup lang="ts">
import {onMounted,onUnmounted,ref,watch} from 'vue';
import {mediaURL} from '../services/publisher';
import {thumbnail} from '../services/thumbnails';
const props=defineProps<{episode:string;video:string;stamp:string}>();
const host=ref<HTMLElement|null>(null),image=ref(''),unavailable=ref(false);
let observer:IntersectionObserver|undefined,controller:AbortController|undefined,visible=false;
async function load(){
 controller?.abort();controller=new AbortController();
 const current=controller;image.value='';unavailable.value=false;
 if(!visible||!props.video)return;
 try{const result=await thumbnail(mediaURL(props.episode,props.video),props.stamp,current.signal);if(!current.signal.aborted)image.value=result;}
 catch{if(!current.signal.aborted)unavailable.value=true;}
}
onMounted(()=>{
 observer=new IntersectionObserver(entries=>{
  visible=entries[0]?.isIntersecting||false;
  if(visible&&!image.value)void load();else if(!visible)controller?.abort();
 },{rootMargin:'80px'});
 if(host.value)observer.observe(host.value);
});
watch(()=>[props.episode,props.video,props.stamp],load);
onUnmounted(()=>{observer?.disconnect();controller?.abort();});
</script>
<template>
 <span ref="host" class="video-thumbnail" aria-hidden="true" :title="unavailable?'Preview unavailable':undefined">
  <img v-if="image" :src="image" alt="" decoding="async" width="48" height="64">
  <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="4" y="3" width="16" height="18" rx="2"/><path d="m10 8 6 4-6 4Z"/></svg>
 </span>
</template>
<style scoped>
.video-thumbnail{width:48px;height:64px;flex:0 0 48px;display:inline-flex;align-items:center;justify-content:center;background:var(--sidebar);color:var(--muted);border-radius:4px;overflow:hidden;vertical-align:middle}
.video-thumbnail img{display:block;width:100%;height:100%;object-fit:contain;background:#211a16}
</style>
