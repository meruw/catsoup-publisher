<script setup lang="ts">
import {computed,nextTick,onMounted,ref,watch} from 'vue';
import UiButton from './UiButton.vue';
import {uploads,type Profile} from '../services/uploads';
import {accountKey,modeLabel,transfer} from '../services/publication';
import {type Draft} from '../services/drafts';
const props=defineProps<{draft:Draft;hint:string;flush:()=>Promise<void>}>();
const emit=defineEmits<{busy:[value:boolean];accounts:[]}>();
const account=ref<Profile>({id:'',name:'',api:false,platform:'',language:''});
const receipt=ref({state:'',reference:'',retryAllowed:false});
const loading=ref(true);
const mode=computed(()=>props.draft.deliveryMode || 'draft');
const audience=ref(''),synthetic=ref(false),confirmed=ref(false),paidAccepted=ref(false),reconciled=ref(false),busy=ref(false),error=ref('');
const platform=computed(()=>props.draft.key.platform);
const browserX=computed(()=>platform.value==='x' && !account.value.api);
const localOnly=computed(()=>mode.value==='draft' && !['youtube','tiktok'].includes(platform.value));
const privateVideo=computed(()=>platform.value==='youtube' && ['private','sent','check_publish'].includes(receipt.value.state));
const completed=computed(()=>['published','inbox'].includes(receipt.value.state) || (receipt.value.state==='sent' && platform.value!=='youtube'));
const action=computed(()=>{
 if(privateVideo.value) return 'Publish existing YouTube video';
 if(browserX.value) return 'Copy video path and open X';
 return ({youtube:mode.value==='draft'?'Upload privately to YouTube':'Publish on YouTube',tiktok:'Send to TikTok inbox',instagram:'Publish Reel on Instagram',x:'Publish on X · Paid API'})[platform.value];
});
const note=computed(()=>{
 if(privateVideo.value && mode.value==='draft')return '';
 if(localOnly.value) return 'Saved on this PC';
 if(platform.value==='youtube') return mode.value==='draft'?'Upload a private video. It will not be public.':'Make this video public on the connected YouTube channel.';
 if(platform.value==='tiktok')return 'Finish editing and posting in TikTok.';
 if(browserX.value)return 'Publisher opens the prepared text and copies the video path. Paste it into X’s file picker, then review and post in X.';
 if(platform.value==='x')return 'Post the video and text using your paid X API access. X bills your developer project.';
 return 'Publish a Reel on the connected Instagram profile.';
});
const blocked=computed(()=>!!props.hint || !confirmed.value || (platform.value==='youtube' && !audience.value) || (platform.value==='x' && !browserX.value && !paidAccepted.value));
const receiptLabel=computed(()=>receipt.value.state==='published'?'Published':receipt.value.state==='inbox'?'Sent to inbox':privateVideo.value?(receipt.value.state==='check_publish'?'Check visibility in YouTube Studio':'Private video uploaded'):receipt.value.state==='sent'?(platform.value==='tiktok'?'Sent to inbox':'Published'):'Check the platform');
watch([mode,audience,synthetic,()=>props.draft.revision,()=>props.hint],()=>{confirmed.value=false;paidAccepted.value=false});
async function run(task:()=>Promise<void>){if(busy.value)return;busy.value=true;emit('busy',true);error.value='';try{await task()}catch(e){error.value=String(e)}finally{busy.value=false;emit('busy',false)}}
async function refreshReceipt(){if(account.value.id && !browserX.value)receipt.value=await uploads.receipt(props.draft.key,account.value.id);}
onMounted(async()=>{try{const p=(await uploads.accounts()||[]).find(p=>p.platform===platform.value && p.language===props.draft.key.language);if(p && accountKey(p)===props.draft.key.account){account.value=p;await refreshReceipt()}}catch(e){error.value=String(e)}finally{loading.value=false}});
function send(){void run(async()=>{
 await props.flush();await nextTick();
 try{
 await transfer(props.draft,account.value,receipt.value,confirmed.value,paidAccepted.value,audience.value,synthetic.value,true);
 }finally{confirmed.value=false;paidAccepted.value=false;await refreshReceipt()}
})}
function clearAttempt(){void run(async()=>{await uploads.clear(props.draft.key,account.value.id,props.draft.videoHash,reconciled.value);reconciled.value=false;await refreshReceipt()})}
</script>
<template>
 <section class="send-panel" aria-label="Platform upload">
  <p v-if="loading" role="status">Checking connection...</p>
  <template v-else>
  <div v-if="account.id" class="send-account"><strong :title="account.id">{{account.name}}</strong><span>{{draft.key.language==='es'?'Spanish':'English'}}</span></div>
  <div v-if="receipt.state" class="send-receipt" role="status"><strong>{{receiptLabel}}</strong><span class="mono">{{receipt.reference}}</span><small v-if="!completed && !privateVideo">Check the destination before retrying.</small></div>
  <template v-if="!completed && (!receipt.state || privateVideo)">
   <template v-if="hint">
    <p class="send-blocker" role="status">{{hint}}</p>
   </template>
   <template v-else>
   <p class="delivery-mode-label">{{modeLabel(platform,mode,account)}}</p>
   <p v-if="note" class="send-explanation" :class="{'local-draft-status':localOnly}">{{note}}</p>
   <template v-if="localOnly"></template>
   <template v-else-if="privateVideo && mode==='draft'"></template>
   <template v-else-if="!account.id"><UiButton @click="emit('accounts')">Connect {{platform==='youtube'?'YouTube':platform==='tiktok'?'TikTok':platform==='instagram'?'Instagram':'X'}}</UiButton></template>
   <fieldset v-else class="draft-fields send-confirmation" :disabled="busy">
    <template v-if="platform==='youtube'">
     <div class="field-group"><label for="youtube-audience">Audience</label><select id="youtube-audience" v-model="audience" class="ui-field"><option value="" disabled>Choose an audience</option><option value="general">Not made for kids</option><option value="kids">Made for kids</option></select></div>
     <label><input v-model="synthetic" type="checkbox"> Contains realistic altered or synthetic content</label>
    </template>
    <label><input v-model="confirmed" type="checkbox"> {{browserX?'I checked the active account in X.':`I confirm ${account.name} and this ${mode==='draft'?'private / inbox transfer':'public post'}.`}}</label>
    <label v-if="platform==='x' && !browserX"><input v-model="paidAccepted" type="checkbox"> I accept X API usage charges for this attempt.</label>
    <div class="send-action"><UiButton variant="primary" :busy="busy" :disabled="blocked" @click="send">{{busy?'Sending...':action}}</UiButton></div>
   </fieldset>
   </template>
  </template>
  <template v-else-if="receipt.state && !completed">
   <UiButton :disabled="busy" @click="run(refreshReceipt)">Refresh receipt</UiButton>
   <div v-if="receipt.retryAllowed" class="retry-confirmation"><label><input v-model="reconciled" type="checkbox"> I checked this account and the upload is absent.</label><UiButton :disabled="!reconciled || busy" @click="clearAttempt">Allow another attempt</UiButton></div>
  </template>
  </template>
  <p v-if="error" role="alert" class="notice error">{{error}}</p>
 </section>
</template>
<style scoped>
.send-account{display:flex;justify-content:space-between;align-items:baseline;gap:16px;margin-bottom:24px;font-size:var(--text-md);overflow-wrap:anywhere}.send-account strong{font-weight:500}.send-account span{color:var(--muted);font-size:var(--text-sm)}
.delivery-mode-label{font-weight:500;margin-bottom:12px}
.send-mode{display:flex;flex-wrap:wrap;gap:24px;border:0;padding:0;margin:0}.send-mode legend{font-size:var(--text-sm);margin-bottom:12px}.send-mode label{display:flex;align-items:center;gap:8px;color:var(--text)}
.send-explanation{font-size:var(--text-sm);color:var(--secondary);margin:16px 0 24px;line-height:1.5}.local-draft-status{color:var(--success);margin-bottom:0}
.send-blocker{font-size:var(--text-md);color:var(--warning);line-height:1.6;margin:0 0 20px}
.send-confirmation{display:grid;gap:20px}.send-confirmation>label{display:flex;gap:8px;align-items:flex-start}.send-confirmation .field-group{margin:0}.send-action{display:flex;justify-content:flex-end;margin-top:4px}
.send-receipt{display:grid;gap:8px;margin-bottom:24px;overflow-wrap:anywhere}.send-receipt strong{font-weight:500;color:var(--success)}.send-receipt span,.send-receipt small{font-size:var(--text-xs);color:var(--secondary)}
.retry-confirmation{display:grid;gap:16px;margin-top:20px}
</style>
