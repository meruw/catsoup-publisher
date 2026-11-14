<script setup lang="ts">
import {onMounted,ref,computed} from 'vue';
import {drafts,type PublishingDefault} from '../services/drafts';
import {uploads,type Profile} from '../services/uploads';
import {modeOptions} from '../services/publication';
import UiSegmented from './UiSegmented.vue';
import UiButton from './UiButton.vue';
import UiTooltip from './UiTooltip.vue';
const emit=defineEmits<{busy:[value:boolean];setup:[]}>();
const profiles=ref<Profile[]>([]),defaults=ref<PublishingDefault[]>([]),language=ref('es'),busy=ref(false),error=ref(''),feedback=ref('');
const hosting=ref({enabled:false,available:false});
const platforms={youtube:'YouTube',tiktok:'TikTok',instagram:'Instagram',facebook:'Facebook',x:'X'};
const rows=computed(()=>Object.entries(platforms).map(([id,label])=>({id,label,profile:profiles.value.find(p=>p.language===language.value&&p.platform===id)})));
const mode=(p?:Profile)=>defaults.value.find(d=>d.platform===p?.platform&&d.language===p?.language&&d.accountId===p?.id)?.mode||'draft';
async function load(){try{profiles.value=await uploads.accounts()||[];defaults.value=await drafts.publishingDefaults()||[];hosting.value=await uploads.instagramHosting();}catch(e){error.value=String(e)}}
async function setHosting(event:Event){const control=event.target as HTMLInputElement;const enabled=control.checked;if(busy.value)return;busy.value=true;emit('busy',true);error.value='';feedback.value='';try{await uploads.setInstagramHosting(enabled,enabled);hosting.value=await uploads.instagramHosting();feedback.value='Instagram upload method saved for both languages.';}catch(e){control.checked=hosting.value.enabled;error.value=String(e)}finally{busy.value=false;emit('busy',false)}}
async function save(p:Profile,event:Event){const control=event.target as HTMLSelectElement;const selected=control.value;const previous=mode(p);if(busy.value)return;busy.value=true;emit('busy',true);error.value='';feedback.value='';try{const value={platform:p.platform,language:p.language,accountId:p.id,mode:selected};await drafts.setPublishingDefault(value);defaults.value=defaults.value.filter(d=>!(d.platform===p.platform&&d.language===p.language&&d.accountId===p.id)).concat(value);feedback.value='Default saved for '+p.name;}catch(e){control.value=previous;error.value=String(e)}finally{busy.value=false;emit('busy',false)}}
onMounted(load);
</script>
<template>
 <section class="publishing-settings" aria-label="Publishing defaults">
  <header><h2>Publishing defaults</h2><UiSegmented v-model="language" label="Default language" :disabled="busy" :options="[{value:'es',label:'ES',accessibleLabel:'Spanish defaults'},{value:'en',label:'EN',accessibleLabel:'English defaults'}]"/></header>
  <p class="muted">For new drafts. Saved posts keep their current mode and approval.</p>
  <div v-for="row in rows" :key="row.id" class="default-row">
   <span><strong>{{row.label}}</strong><small>{{row.profile?.name || 'Connect an account in Accounts'}}</small></span>
   <UiTooltip :disabled="busy||!row.profile||row.id==='tiktok'" :text="row.id==='tiktok'?'TikTok defaults to inbox. Choose Direct Post per post after authorizing video.publish; privacy must be selected for each send.':!row.profile?'Connect this account first.':'Used for new drafts for this account. Existing posts are unchanged.'">
    <select class="ui-field" :aria-label="row.label+' default mode'" :value="mode(row.profile)" :disabled="busy||!row.profile||row.id==='tiktok'" @change="row.profile && save(row.profile,$event)"><option v-for="option in modeOptions(row.id,row.id==='tiktok'?undefined:row.profile)" :key="option.value" :value="option.value">{{option.label}}</option></select>
   </UiTooltip>
  </div>
  <section class="hosting-option" aria-label="Instagram upload method">
   <label><input type="checkbox" :checked="hosting.enabled" :disabled="busy||(!hosting.available&&!hosting.enabled)" @change="setHosting"> Use temporary HTTPS for Instagram</label>
   <p class="muted">Share only the approved video through Cloudflare while Instagram downloads it. Anyone with the temporary link can access that video. The link closes when the transfer finishes or stops. Applies to ES and EN.</p>
   <p v-if="!hosting.available" class="muted">Install cloudflared in the workspace .tools folder to enable this option.</p>
   <p v-else class="muted">Free experimental transport; availability depends on Cloudflare and Meta.</p>
  </section>
  <p v-if="feedback" role="status" class="ui-status ui-status--success">{{feedback}}</p><p v-if="error" role="alert" class="notice error">{{error}}</p>
  <div class="publishing-setup"><UiButton variant="quiet" @click="emit('setup')">Set up TikTok Direct Post</UiButton></div>
 </section>
</template>
<style scoped>
.hosting-option{padding:24px 0;border-bottom:1px solid var(--line)}.hosting-option label{display:flex;align-items:center;gap:10px;font-weight:500}.hosting-option p{font-size:var(--text-sm);line-height:1.5;max-width:40rem;margin:10px 0 0}.hosting-option input{flex-shrink:0}
.publishing-setup{margin-top:20px}.publishing-settings{max-width:48rem;padding:24px 0;border-top:1px solid var(--line)}header{display:flex;justify-content:space-between;align-items:center;gap:24px;margin-bottom:12px}.publishing-settings>p{font-size:var(--text-sm);margin-bottom:20px}.default-row{display:flex;align-items:center;justify-content:space-between;gap:24px;padding:16px 0;border-bottom:1px solid var(--line)}.default-row>span{display:grid;gap:4px;min-width:0}.default-row strong{font-weight:500}.default-row small{color:var(--muted);overflow-wrap:anywhere}.default-row .control-tip{width:180px;flex-shrink:0}.default-row select{width:100%}
</style>
