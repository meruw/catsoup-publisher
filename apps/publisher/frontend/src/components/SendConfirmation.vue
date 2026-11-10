<script setup lang="ts">
import {computed,ref,watch} from 'vue';
import UiButton from './UiButton.vue';
import {mediaURL} from '../services/publisher';
const props=defineProps<{language:string;episode:string;videoId:string;title:string;destinations:{platform:string;name:string;mode:string}[];busy?:boolean;disabled?:boolean;prepare?:boolean}>();
const emit=defineEmits<{confirm:[];back:[]}>();
const checked=ref(false),submitted=ref(false);
const languageName=computed(()=>props.language==='es'?'Spanish':props.language==='en'?'English':'Unknown language');
const validLanguage=computed(()=>['es','en'].includes(props.language));
watch(()=>JSON.stringify([props.language,props.episode,props.videoId,props.title,props.destinations]),()=>{checked.value=false;submitted.value=false});
watch(()=>props.busy,(value,previous)=>{if(previous&&!value)submitted.value=false});
function confirm(){if(!checked.value||!validLanguage.value||props.busy||props.disabled||submitted.value)return;submitted.value=true;emit('confirm');}
const names:Record<string,string>={youtube:'YouTube',instagram:'Instagram',facebook:'Facebook',tiktok:'TikTok',x:'X'};
</script>
<template>
 <section class="send-final" aria-label="Final language confirmation">
  <div class="send-final-language"><h3>{{languageName}} video</h3><span>{{language.toUpperCase()}}</span></div>
  <p class="send-final-title">{{title||episode}}</p>
  <video :key="episode+videoId" :src="mediaURL(episode,videoId)" controls preload="metadata" :aria-label="languageName+' video preview'"/>
  <p class="send-final-file">{{videoId}} <span>· {{episode}}</span></p>
  <ul aria-label="Confirmed destinations"><li v-for="(destination,index) in destinations" :key="index"><strong>{{names[destination.platform]||destination.platform}} / {{destination.name}}</strong><span>{{destination.mode}}</span></li></ul>
  <label><input v-model="checked" type="checkbox" :disabled="busy||submitted"> I confirm this is the {{languageName}} version for these accounts.</label>
  <div class="send-final-actions"><UiButton :disabled="busy||submitted" @click="emit('back')">Back</UiButton><UiButton variant="primary" :busy="busy" :disabled="!checked||!validLanguage||disabled||submitted" @click="confirm">{{prepare?'Prepare':'Send'}} {{languageName}} video</UiButton></div>
 </section>
</template>
<style scoped>
.send-final{display:grid;gap:16px}.send-final-language{display:flex;align-items:center;justify-content:space-between;gap:16px}.send-final h3{font-size:var(--text-section);font-weight:500;margin:0}.send-final-language>span{font-weight:500;color:var(--accent-ink)}.send-final-title{margin:0;overflow-wrap:anywhere}.send-final video{display:block;width:100%;max-height:200px;background:#171717;border-radius:8px}.send-final-file{font-size:var(--text-xs);color:var(--secondary);margin:0;overflow-wrap:anywhere}.send-final-file span{color:var(--muted)}.send-final ul{list-style:none;padding:0;margin:0;border-block:1px solid var(--line)}.send-final li{display:grid;gap:4px;padding:12px 0}.send-final li+li{border-top:1px solid var(--line)}.send-final li strong{font-weight:500}.send-final li span{font-size:var(--text-sm);color:var(--secondary)}.send-final label{display:flex;align-items:flex-start;gap:8px;line-height:1.5}.send-final-actions{display:flex;justify-content:space-between;gap:12px;margin-top:8px}
</style>
