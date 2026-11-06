<script setup lang="ts">
import {computed,onMounted,ref} from 'vue';
import UiButton from './UiButton.vue';
import {drafts,type Draft} from '../services/drafts';
import {uploads,type Profile} from '../services/uploads';
import type {Episode,Video} from '../services/publisher';
import {accountKey,draftBlocker,localDraft,modeLabel,receiptBlocker,transfer,type TransferReceipt} from '../services/publication';
const props=defineProps<{episode:Episode;language:string;video:Video;reviewed:boolean;profiles:Profile[];saved:Draft[]}>();
const emit=defineEmits<{busy:[value:boolean]}>();
type Row={profile:Profile;draft?:Draft;selected:boolean;reason:string;result:string;receipt:TransferReceipt};
const rows=ref<Row[]>([]),loading=ref(true),busy=ref(false),finished=ref(false),error=ref('');
const confirmed=ref(false),paid=ref(false),audience=ref(''),synthetic=ref(false);
const selected=computed(()=>rows.value.filter(r=>r.selected));
const needsAudience=computed(()=>selected.value.some(r=>r.profile.platform==='youtube'));
const needsPaid=computed(()=>selected.value.some(r=>r.profile.platform==='x'&&r.profile.api&&r.draft?.deliveryMode==='publish'));
const ready=computed(()=>!loading.value&&!busy.value&&!finished.value&&selected.value.length>0&&confirmed.value&&(!needsAudience.value||!!audience.value)&&(!needsPaid.value||paid.value));
onMounted(async()=>{
 try{
  for(const profile of props.profiles.filter(p=>p.language===props.language)){
   const draft=props.saved.find(d=>d.key.platform===profile.platform&&d.key.account===accountKey(profile));
   let reason=draftBlocker(draft,profile,props.episode,props.video,props.reviewed);
   const receipt=draft&&!localDraft(draft)&&!(profile.platform==='x'&&!profile.api)?await uploads.receipt(draft.key,profile.id):{state:'',reference:'',retryAllowed:false};
   if(!reason&&draft)reason=receiptBlocker(draft,receipt);
   rows.value.push({profile,draft,selected:!reason,reason,result:'',receipt});
  }
 }catch(e){error.value=String(e);}finally{loading.value=false;}
});
async function sendAll(){
 if(!ready.value||error.value)return;
 busy.value=true;emit('busy',true);
 try{
  // Check every selected snapshot before making the first external request.
  const current=await uploads.accounts() || [];
  for(const row of selected.value){
   const d=row.draft!;const latest=await drafts.get(d.key);
   const profile=current.find(p=>p.platform===row.profile.platform&&p.language===props.language);
   if(profile?.id!==row.profile.id||profile.api!==row.profile.api||latest.revision!==d.revision)throw new Error('A draft or account changed. Close this dialog and review the destinations again.');
   const blocked=draftBlocker(latest,profile,props.episode,props.video,props.reviewed);if(blocked)throw new Error(blocked);
   if(!localDraft(d)&&!(profile.platform==='x'&&!profile.api)){
    row.receipt=await uploads.receipt(d.key,profile.id);
    const reason=receiptBlocker(d,row.receipt);if(reason)throw new Error(`${profile.name}: ${reason}`);
   }
  }
  finished.value=true;
  for(const row of selected.value){
   row.result='Sending…';
   try{row.result=await transfer(row.draft!,row.profile,row.receipt,true,paid.value,audience.value,synthetic.value);}
   catch(e){row.result='Needs checking';throw e;}
  }
 }catch(e){error.value=String(e);for(const row of selected.value){if(!row.result)row.result='Not sent';}}
 finally{busy.value=false;emit('busy',false);}
}
</script>
<template>
 <p class="batch-language">{{language==='es'?'Spanish':'English'}} accounts</p>
 <p v-if="loading" role="status">Checking destinations…</p>
 <fieldset :disabled="loading||busy||finished" class="draft-fields batch-rows">
  <label v-for="row in rows" :key="row.profile.platform" class="batch-row">
   <input v-model="row.selected" type="checkbox" :disabled="!!row.reason" @change="confirmed=false;paid=false">
   <span><strong>{{row.profile.platform==='x'?'X':row.profile.platform==='youtube'?'YouTube':row.profile.platform==='tiktok'?'TikTok':'Instagram'}} / {{row.profile.name}}</strong><small>{{row.reason||modeLabel(row.profile.platform,row.draft?.deliveryMode||'draft',row.profile)}}</small><small v-if="row.result" role="status">{{row.result}}</small></span>
  </label>
 </fieldset>
 <fieldset v-if="!finished" :disabled="loading||busy" class="draft-fields batch-confirmation">
  <div v-if="needsAudience" class="field-group"><label for="batch-audience">YouTube audience</label><select id="batch-audience" v-model="audience" class="ui-field" @change="confirmed=false"><option value="" disabled>Choose an audience</option><option value="general">Not made for kids</option><option value="kids">Made for kids</option></select></div>
  <label v-if="needsAudience"><input v-model="synthetic" type="checkbox" @change="confirmed=false"> Contains realistic altered or synthetic content</label>
  <label v-if="needsPaid"><input v-model="paid" type="checkbox"> I accept X API usage charges for this attempt.</label>
  <label><input v-model="confirmed" type="checkbox"> I confirm the selected accounts and delivery modes.</label>
  <UiButton variant="primary" :busy="busy" :disabled="!ready||!!error" @click="sendAll">Send to {{selected.length}} destinations</UiButton>
 </fieldset>
 <p v-if="error" class="notice error" role="alert">{{error}}</p>
</template>
<style scoped>
.batch-language{color:var(--secondary);margin-bottom:20px}.batch-rows{display:grid;gap:20px;margin-bottom:28px}.batch-row{display:flex;align-items:flex-start;gap:12px}.batch-row>span{display:grid;gap:4px}.batch-row strong{font-weight:500}.batch-row small{color:var(--secondary);font-size:var(--text-sm)}.batch-confirmation{display:grid;gap:20px}.batch-confirmation label{display:flex;gap:8px;align-items:flex-start}.batch-confirmation button{justify-self:end}.batch-confirmation .field-group{margin:0}
</style>
