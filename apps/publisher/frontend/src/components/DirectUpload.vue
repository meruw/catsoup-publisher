<script setup lang="ts">
import {computed,nextTick,onMounted,ref,watch} from 'vue';
import UiButton from './UiButton.vue';
import {uploads} from '../services/uploads';
import {drafts,type Draft} from '../services/drafts';
const props=defineProps<{draft:Draft;hint:string;flush:()=>Promise<void>}>();
const emit=defineEmits<{busy:[value:boolean];accounts:[]}>();
const account=ref({id:'',name:'',api:false});
const receipt=ref({state:'',reference:'',retryAllowed:false});
const mode=ref<'draft'|'publish'>('draft');
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
 if(localOnly.value) return 'This draft stays on your PC. This connector does not create drafts on the platform.';
 if(platform.value==='youtube') return mode.value==='draft'?'Upload a private video. It will not be public.':'Make this video public on the connected YouTube channel.';
 if(platform.value==='tiktok')return 'Send the video to the TikTok inbox, then finish editing and posting in TikTok. Copy the caption from Prepare.';
 if(browserX.value)return 'Publisher opens the prepared text and copies the video path. Paste it into X’s file picker, then review and post in X.';
 if(platform.value==='x')return 'Post the video and text using your paid X API access. X bills your developer project.';
 return 'Publish a Reel on the connected Instagram profile.';
});
const blocked=computed(()=>!!props.hint || !confirmed.value || (platform.value==='youtube' && !audience.value) || (platform.value==='x' && !browserX.value && !paidAccepted.value));
const receiptLabel=computed(()=>receipt.value.state==='published'?'Published':receipt.value.state==='inbox'?'Sent to inbox':privateVideo.value?(receipt.value.state==='check_publish'?'Check visibility in YouTube Studio':'Private video uploaded'):receipt.value.state==='sent'?(platform.value==='tiktok'?'Sent to inbox':'Published'):'Check the platform');
watch([mode,audience,synthetic,()=>props.draft.revision,()=>props.hint],()=>{confirmed.value=false;paidAccepted.value=false});
async function run(task:()=>Promise<void>){if(busy.value)return;busy.value=true;emit('busy',true);error.value='';try{await task()}catch(e){error.value=String(e)}finally{busy.value=false;emit('busy',false)}}
async function refreshReceipt(){if(account.value.id && !browserX.value)receipt.value=await uploads.receipt(props.draft.key,account.value.id);}
onMounted(async()=>{try{const p=(await uploads.accounts()||[]).find(p=>p.platform===platform.value && p.language===props.draft.key.language);if(p){account.value=p;await refreshReceipt()}}catch(e){error.value=String(e)}});
function send(){void run(async()=>{
 await props.flush();await nextTick();
 try{
 if(browserX.value){await drafts.copyUploadPath(props.draft.key,props.draft.revision,confirmed.value);await uploads.openX(props.draft.key,props.draft.revision);return;}
 receipt.value=privateVideo.value
 ? await uploads.publishYouTube(props.draft.key,props.draft.revision,account.value.id,confirmed.value,audience.value,synthetic.value)
 : await uploads.sendMode(props.draft.key,props.draft.revision,account.value.id,mode.value,confirmed.value,paidAccepted.value,audience.value,synthetic.value);
 }finally{confirmed.value=false;paidAccepted.value=false;await refreshReceipt()}
})}
function clearAttempt(){void run(async()=>{await uploads.clear(props.draft.key,account.value.id,props.draft.videoHash,reconciled.value);reconciled.value=false;await refreshReceipt()})}
</script>
<template>
 <section class="send-panel" aria-label="Platform upload">
  <div v-if="receipt.state" class="send-receipt" role="status"><strong>{{receiptLabel}}</strong><span class="mono">{{receipt.reference}}</span><small>{{completed?'This file will not be sent again.':privateVideo?'Publishing uses this same video ID.':'An earlier attempt exists. Check the destination before retrying.'}}</small></div>
  <template v-if="!completed && (!receipt.state || privateVideo)">
   <fieldset class="send-mode" :disabled="busy"><legend>Delivery mode</legend>
    <label><input v-model="mode" type="radio" value="draft" name="delivery-mode"> Keep as draft</label>
    <label :class="{'mode-unavailable':platform==='tiktok'}"><input v-model="mode" type="radio" value="publish" name="delivery-mode" :disabled="platform==='tiktok'"> Publish now</label>
   </fieldset>
   <p v-if="platform==='tiktok'" class="publication-hint">Direct Post is not enabled in this connector. Public posting requires TikTok approval and its publishing flow.</p>
   <p class="send-explanation">{{note}}</p>
   <p v-if="localOnly" class="local-draft-status" role="status">{{draft.revision?'Saved locally · Return to this destination whenever you are ready':'Save this draft in Prepare to keep it on this PC'}}</p>
   <template v-else-if="privateVideo && mode==='draft'"><p class="publication-hint">Your private upload is already recorded. Choose Publish now to make it public without uploading again.</p></template>
   <template v-else-if="!account.id"><p>No account assigned for {{draft.key.language.toUpperCase()}}</p><UiButton @click="emit('accounts')">Manage accounts</UiButton></template>
   <template v-else>
    <dl class="send-destination"><div><dt>{{draft.key.language==='es'?'Spanish':'English'}} destination</dt><dd><strong>{{account.name}}</strong><small class="mono">{{account.id}}</small></dd></div></dl>
    <p v-if="hint" class="notice warning">{{hint}}</p>
    <fieldset v-else class="draft-fields send-confirmation" :disabled="busy">
     <template v-if="platform==='youtube'">
      <div class="field-group"><label for="youtube-audience">Audience</label><select id="youtube-audience" v-model="audience" class="ui-field"><option value="" disabled>Choose an audience</option><option value="general">Not made for kids</option><option value="kids">Made for kids</option></select></div>
      <label><input v-model="synthetic" type="checkbox"> Contains realistic altered or synthetic content</label>
      <p v-if="privateVideo" class="publication-hint">Publishing an existing video needs YouTube management permission. Reconnect in Accounts if Google rejects it.</p>
     </template>
     <label><input v-model="confirmed" type="checkbox"> {{browserX?'I checked the active account in X.':`I confirm ${account.name} and this ${mode==='draft'?'private / inbox transfer':'public post'}.`}}</label>
     <label v-if="platform==='x' && !browserX"><input v-model="paidAccepted" type="checkbox"> I accept X API usage charges for this attempt.</label>
     <div class="draft-actions"><UiButton variant="primary" :busy="busy" :disabled="blocked" :title="platform==='x' && !browserX?'X bills API usage to your developer project. Publisher adds no posting fee.':undefined" @click="send">{{busy?'Sending…':action}}</UiButton></div>
    </fieldset>
   </template>
  </template>
  <template v-else-if="receipt.state && !completed">
   <div class="draft-actions"><UiButton :disabled="busy" @click="run(refreshReceipt)">Refresh receipt</UiButton></div>
   <template v-if="receipt.retryAllowed"><label><input v-model="reconciled" type="checkbox"> I checked this account and the upload is absent.</label><UiButton :disabled="!reconciled || busy" @click="clearAttempt">Allow another attempt</UiButton></template>
  </template>
  <p v-if="error" role="alert" class="notice error">{{error}}</p>
 </section>
</template>
<style scoped>
.send-mode{display:flex;flex-wrap:wrap;gap:24px;border:0;padding:0;margin:0}.send-mode legend{font-size:var(--text-sm);margin-bottom:12px}.send-mode label{display:flex;align-items:center;gap:8px;color:var(--text)}.mode-unavailable{opacity:.6}
.send-explanation{font-size:var(--text-sm);color:var(--secondary);margin:16px 0;max-width:72ch}.send-destination{margin-bottom:20px}.send-destination small{display:block;color:var(--muted);margin-top:4px;font-size:var(--text-xs)}
.send-confirmation{display:grid;gap:16px}.send-confirmation>label{display:flex;gap:8px;align-items:flex-start}.send-confirmation .field-group{margin:0;max-width:360px}
.send-receipt{display:grid;gap:8px;border-bottom:1px solid var(--line);padding-bottom:20px;margin-bottom:24px;overflow-wrap:anywhere}.send-receipt strong{font-weight:500;color:var(--success)}.send-receipt span,.send-receipt small{font-size:var(--text-xs);color:var(--secondary)}
.local-draft-status{font-size:var(--text-sm);padding:16px 0;color:var(--success)}
</style>
