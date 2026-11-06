<script setup lang="ts">
import {onMounted,ref,computed} from 'vue';
import {drafts,type PublishingDefault} from '../services/drafts';
import {uploads,type Profile} from '../services/uploads';
import {modeOptions} from '../services/publication';
import UiSegmented from './UiSegmented.vue';
import UiTooltip from './UiTooltip.vue';
const emit=defineEmits<{busy:[value:boolean]}>();
const profiles=ref<Profile[]>([]),defaults=ref<PublishingDefault[]>([]),language=ref('es'),busy=ref(false),error=ref(''),feedback=ref('');
const platforms={youtube:'YouTube',tiktok:'TikTok',instagram:'Instagram',x:'X'};
const rows=computed(()=>Object.entries(platforms).map(([id,label])=>({id,label,profile:profiles.value.find(p=>p.language===language.value&&p.platform===id)})));
const mode=(p?:Profile)=>defaults.value.find(d=>d.platform===p?.platform&&d.language===p?.language&&d.accountId===p?.id)?.mode||'draft';
async function load(){try{profiles.value=await uploads.accounts()||[];defaults.value=await drafts.publishingDefaults()||[];}catch(e){error.value=String(e)}}
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
  <p v-if="feedback" role="status" class="ui-status ui-status--success">{{feedback}}</p><p v-if="error" role="alert" class="notice error">{{error}}</p>
 </section>
</template>
<style scoped>
.publishing-settings{max-width:48rem;padding:24px 0;border-top:1px solid var(--line)}header{display:flex;justify-content:space-between;align-items:center;gap:24px;margin-bottom:12px}.publishing-settings>p{font-size:var(--text-sm);margin-bottom:20px}.default-row{display:flex;align-items:center;justify-content:space-between;gap:24px;padding:16px 0;border-bottom:1px solid var(--line)}.default-row>span{display:grid;gap:4px;min-width:0}.default-row strong{font-weight:500}.default-row small{color:var(--muted);overflow-wrap:anywhere}.default-row .control-tip{width:180px;flex-shrink:0}.default-row select{width:100%}
</style>
