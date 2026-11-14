<script setup lang="ts">
import TikTokTracking from './TikTokTracking.vue';
import {Events} from '@wailsio/runtime';
import SendConfirmation from './SendConfirmation.vue';
import TikTokWebPrepare from './TikTokWebPrepare.vue';
import UiSegmented from './UiSegmented.vue';
import {computed,nextTick,onMounted,onUnmounted,ref,watch} from 'vue';
import UiTooltip from './UiTooltip.vue';
import UiButton from './UiButton.vue';
import {mediaURL} from '../services/publisher';
import {uploads,type Profile,type TikTokCreator,type TikTokOptions} from '../services/uploads';
import {accountKey,modeLabel,transfer} from '../services/publication';
import {enqueueTransfers,transferPending,receiptSummary} from '../services/transfers';
import {type Draft} from '../services/drafts';
const props=defineProps<{draft:Draft;hint:string;flush:()=>Promise<void>}>();
const emit=defineEmits<{busy:[value:boolean];accounts:[];queued:[]}>();
const account=ref<Profile>({id:'',name:'',api:false,directPost:false,platform:'',language:''});
const receipt=ref<{state:string;reference:string;retryAllowed:boolean;inboxHistory?:{reference:string;confirmedAt:number;archivedAt:number}[]|null}>({state:'',reference:'',retryAllowed:false});
const resendOpen=ref(false),resendConfirmed=ref(false),resendReady=ref(false);
const loading=ref(true),finalReview=ref(false);
watch(()=>JSON.stringify([props.draft,account.value]),()=>{finalReview.value=false;confirmed.value=false;resendConfirmed.value=false;resendOpen.value=false});
const transport=ref('browser');
const creator=ref<TikTokCreator|null>(null),creatorError=ref('');
const options=ref<TikTokOptions>({privacy:'',comments:false,duet:false,stitch:false,commercial:false,ownBrand:false,branded:false,synthetic:false,consent:false});
const directTikTok=computed(()=>platform.value==='tiktok'&&mode.value==='publish');
const privacyNames:Record<string,string>={PUBLIC_TO_EVERYONE:'Everyone',MUTUAL_FOLLOW_FRIENDS:'Friends',FOLLOWER_OF_CREATOR:'Followers',SELF_ONLY:'Only me'};
const tikTokBlocked=computed(()=>directTikTok.value&&(!creator.value||!options.value.privacy||!options.value.consent||(options.value.commercial&&!options.value.ownBrand&&!options.value.branded)||(options.value.branded&&options.value.privacy==='SELF_ONLY')));
watch(options,()=>{confirmed.value=false},{deep:true});
watch(()=>options.value.branded,()=>{options.value.consent=false});
watch(()=>options.value.commercial,value=>{if(!value){options.value.ownBrand=false;options.value.branded=false;}});
async function loadCreator(){try{creator.value=await uploads.tikTokCreator(props.draft.key,account.value.id);creatorError.value='';}catch(e){creator.value=null;creatorError.value=String(e)}}

const mode=computed(()=>props.draft.deliveryMode || 'draft');
const audience=ref(''),synthetic=ref(false),confirmed=ref(false),paidAccepted=ref(false),reconciled=ref(false),busy=ref(false),error=ref('');
const platform=computed(()=>props.draft.key.platform);
const browserX=computed(()=>platform.value==='x' && !account.value.api);
const localOnly=computed(()=>mode.value==='draft' && !['youtube','tiktok'].includes(platform.value));
const privateVideo=computed(()=>platform.value==='youtube' && ['private','sent','check_publish'].includes(receipt.value.state));
const completed=computed(()=>['published','inbox','restricted'].includes(receipt.value.state) || (receipt.value.state==='sent' && platform.value!=='youtube') || (receipt.value.state==='private'&&platform.value==='tiktok'));
const failedInstagram=computed(()=>platform.value==='instagram' && ['failed','check'].includes(receipt.value.state));
const action=computed(()=>{
 if(failedInstagram.value)return 'Check and finish on Instagram';
 if(privateVideo.value) return 'Publish existing YouTube video';
 if(browserX.value) return 'Copy video path and open X';
 return ({youtube:mode.value==='draft'?'Upload privately to YouTube':'Publish on YouTube',tiktok:mode.value==='publish'?'Post on TikTok':'Send to TikTok inbox',instagram:'Publish Reel on Instagram',facebook:'Publish Reel on Facebook',x:'Publish on X · Paid API'})[platform.value];
});
const note=computed(()=>{
 if(privateVideo.value && mode.value==='draft')return '';
 if(localOnly.value) return 'Saved on this PC';
 if(platform.value==='youtube') return mode.value==='draft'?'Upload a private video. It will not be public.':'Make this video public on the connected YouTube channel.';
 if(platform.value==='tiktok')return directTikTok.value?'Choose who can see this post. TikTok may take time to process and moderate it.':'Finish editing and posting in TikTok.';
 if(browserX.value)return 'Publisher opens the prepared text and copies the video path. Paste it into X’s file picker, then review and post in X.';
 if(platform.value==='x')return 'Post the video and text using your paid X API access. X bills your developer project.';
 return platform.value==='facebook'?'Publish a Reel on the connected Facebook Page.':'Publish a Reel on the connected Instagram profile. Audio is prepared at 128 kb/s; your original file and video stream stay unchanged.';
});
const blocked=computed(()=>transferPending(props.draft)||tikTokBlocked.value||!!props.hint || !confirmed.value || (platform.value==='youtube' && !audience.value) || (platform.value==='x' && !browserX.value && !paidAccepted.value));
const receiptLabel=computed(()=>receipt.value.state==='failed'?'Failed':receipt.value.state==='restricted'?'Posted with limited visibility':receipt.value.state==='private'&&platform.value==='tiktok'?'Posted · Only me':receipt.value.state==='published'?'Published':receipt.value.state==='inbox'?'TikTok confirmed inbox':privateVideo.value?(receipt.value.state==='check_publish'?'Check visibility in YouTube Studio':'Private video uploaded'):receipt.value.state==='sent'?(platform.value==='tiktok'?'Sent to inbox':'Published'):'Check the platform');
watch([mode,audience,synthetic,()=>props.draft.revision,()=>props.hint],()=>{confirmed.value=false;paidAccepted.value=false});
async function run(task:()=>Promise<void>){if(busy.value)return;busy.value=true;emit('busy',true);error.value='';try{await task()}catch(e){finalReview.value=false;error.value=String(e)}finally{busy.value=false;emit('busy',false)}}
function openPolicy(policy:string){void run(async()=>{await uploads.tikTokPolicy(policy)})}
async function refreshReceipt(){if(account.value.id && !browserX.value){receipt.value=await uploads.receipt(props.draft.key,account.value.id);if(platform.value==='tiktok'&&receipt.value.state==='inbox')transport.value='api';}}
let offTracking:(()=>void)|undefined;
onMounted(()=>{offTracking=Events.On('publisher:receipts-changed',()=>{void refreshReceipt().catch(()=>{})})});
onUnmounted(()=>offTracking?.());
onMounted(async()=>{try{const p=(await uploads.accounts()||[]).find(p=>p.platform===platform.value && p.language===props.draft.key.language);if(p && accountKey(p)===props.draft.key.account){account.value=p;await refreshReceipt();if(directTikTok.value)await loadCreator()}}catch(e){error.value=String(e)}finally{loading.value=false}});
function send(){if(!finalReview.value||blocked.value)return;void run(async()=>{
 await props.flush();await nextTick();
 if(!finalReview.value||blocked.value)throw new Error('The language, video or destination changed. Confirm again.');
 const d:Draft=JSON.parse(JSON.stringify(props.draft));
 const p:Profile=JSON.parse(JSON.stringify(account.value));
 const r=JSON.parse(JSON.stringify(receipt.value));
 const opts:TikTokOptions=JSON.parse(JSON.stringify(options.value));
 const ok=confirmed.value,paid=paidAccepted.value,group=audience.value,ai=synthetic.value;
 const direct=directTikTok.value;
 if(browserX.value){await transfer(d,p,r,ok,paid,group,ai,true);confirmed.value=false;return;}
 enqueueTransfers([{draft:d,profile:p,mode:modeLabel(d.key.platform,d.deliveryMode,p),send:async()=>direct
  ? receiptSummary((await uploads.sendTikTok(d.key,d.revision,p.id,ok,opts)).state)
  : transfer(d,p,r,ok,paid,group,ai,true)}]);
 confirmed.value=false;paidAccepted.value=false;
 busy.value=false;emit('busy',false);emit('queued');
})}
function allowInboxResend(){void run(async()=>{
 if(transferPending(props.draft))throw new Error('Wait for the active transfer to finish.');
 await uploads.allowInboxResend(props.draft.key,account.value.id,props.draft.videoHash,receipt.value.reference,resendConfirmed.value);
 resendOpen.value=false;resendConfirmed.value=false;confirmed.value=false;resendReady.value=true;await refreshReceipt();
})}
function clearAttempt(){void run(async()=>{await uploads.clear(props.draft.key,account.value.id,props.draft.videoHash,reconciled.value);reconciled.value=false;await refreshReceipt()})}
</script>
<template>
 <section class="send-panel" aria-label="Platform upload">
  <p v-if="loading" role="status">Checking connection...</p>
  <template v-else>
  <SendConfirmation v-if="finalReview" :language="draft.key.language" :episode="draft.key.episode" :video-id="draft.videoId" :title="draft.title" :destinations="[{platform,name:account.name,mode:directTikTok ? 'Direct Post · '+privacyNames[options.privacy] : modeLabel(platform,mode,account)}]" :busy="busy" :disabled="blocked" @back="finalReview=false" @confirm="send"/>
  <template v-else>
  <div v-if="account.id" class="send-account"><strong :title="account.id">{{account.name}}</strong><span>{{draft.key.language==='es'?'Spanish':'English'}}</span></div>
  <UiSegmented v-if="platform==='tiktok'" v-model="transport" label="TikTok upload method" :options="[{value:'browser',label:'Browser'},{value:'api',label:'Inbox / API'}]"/>
  <TikTokWebPrepare v-if="platform==='tiktok' && transport==='browser'" :draft="draft" :account="account" :receipt="receipt.state" :hint="hint" :flush="flush" @queued="emit('queued')"/>
  <template v-else>
  <TikTokTracking :key="receipt.reference" v-if="platform==='tiktok' && account.id" :post-key="draft.key" :account-id="account.id"/>
  <div v-if="receipt.state" class="send-receipt" role="status"><strong>{{receiptLabel}}</strong><span class="mono">{{receipt.reference}}</span><small v-if="failedInstagram">Check the saved result before publishing. A replacement is created only after Meta confirms the container failed.</small><small v-else-if="!completed && !privateVideo">Check the destination before retrying.</small></div>
  <template v-if="platform==='tiktok'">
   <div v-if="receipt.state==='inbox'" class="inbox-resend">
    <UiButton v-if="!resendOpen" :disabled="busy||transferPending(draft)" @click="resendOpen=true">Resend to inbox…</UiButton>
    <div v-else class="retry-confirmation">
     <p>The earlier transfer may still arrive. Its receipt will be kept. This enables one new send; it does not send yet.</p>
     <label><input v-model="resendConfirmed" type="checkbox"> I want to resend to {{account.name}} ({{draft.key.language==='es'?'Spanish':'English'}}), even if both copies arrive.</label>
     <div class="resend-actions"><UiButton :disabled="busy||!resendConfirmed||transferPending(draft)" @click="allowInboxResend">Enable another send</UiButton><UiButton variant="quiet" :disabled="busy" @click="resendOpen=false;resendConfirmed=false">Cancel</UiButton></div>
    </div>
   </div>
   <p v-if="resendReady" role="status">Ready for another send. Confirm the account and language below.</p>
  </template>
  <template v-if="!completed && (!receipt.state || privateVideo || failedInstagram)">
   <template v-if="hint">
    <p class="send-blocker" role="status">{{hint}}</p>
   </template>
   <template v-else>
   <p class="delivery-mode-label">{{modeLabel(platform,mode,account)}}</p>
   <p v-if="transferPending(draft)" role="status">This video is in Uploads</p>
   <p v-if="note" class="send-explanation" :class="{'local-draft-status':localOnly}">{{note}}</p>
   <template v-if="localOnly"></template>
   <template v-else-if="privateVideo && mode==='draft'"></template>
   <template v-else-if="!account.id"><UiButton @click="emit('accounts')">Connect {{platform==='youtube'?'YouTube':platform==='tiktok'?'TikTok':platform==='instagram'?'Instagram':platform==='facebook'?'Facebook':'X'}}</UiButton></template>
   <fieldset v-else class="draft-fields send-confirmation" :disabled="busy">
    <template v-if="directTikTok">
     <p v-if="creatorError" role="alert" class="notice error">{{creatorError}}</p>
     <template v-if="creator">
      <p>{{creator.creator_nickname}} · @{{creator.creator_username}}</p>
      <video class="tiktok-preview" :src="mediaURL(draft.key.episode,draft.videoId)" controls preload="metadata" aria-label="TikTok post preview"/>
      <details><summary>Caption</summary><p class="tiktok-caption">{{[draft.title,draft.description,draft.hashtags].filter(Boolean).join('\n\n')}}</p><small>Close this dialog to edit the caption and approve the revised post.</small></details>
      <div class="field-group"><label for="tiktok-privacy">Who can watch</label><select id="tiktok-privacy" v-model="options.privacy" class="ui-field"><option value="" disabled>Choose visibility</option><option v-for="p in creator.privacy_level_options||[]" :key="p" :value="p" :disabled="options.branded&&p==='SELF_ONLY'">{{privacyNames[p]}}</option></select></div>
      <div class="tiktok-interactions"><label><input v-model="options.comments" type="checkbox" :disabled="creator.comment_disabled"> Allow comments</label><label><input v-model="options.duet" type="checkbox" :disabled="creator.duet_disabled"> Allow Duet</label><label><input v-model="options.stitch" type="checkbox" :disabled="creator.stitch_disabled"> Allow Stitch</label></div>
      <label><input v-model="options.commercial" type="checkbox"> This post promotes a brand, product or service</label>
      <div v-if="options.commercial" class="tiktok-disclosures"><label><input v-model="options.ownBrand" type="checkbox"> Your brand</label><label title="Branded content cannot be private"><input v-model="options.branded" type="checkbox" :disabled="options.privacy==='SELF_ONLY'"> Branded content · Paid partnership</label><p v-if="options.branded">Your video will be labeled as Paid partnership</p><p v-else-if="options.ownBrand">Your video will be labeled as Promotional content</p></div>
      <label><input v-model="options.synthetic" type="checkbox"> AI-generated content</label>
      <label><input v-model="options.consent" type="checkbox"><span>By posting, I agree to TikTok's <button type="button" class="policy-link" @click.prevent="openPolicy('music')">Music Usage Confirmation</button><span v-if="options.branded"> and <button type="button" class="policy-link" @click.prevent="openPolicy('branded')">Branded Content Policy</button></span>.</span></label>
     </template>
    </template>
    <template v-if="platform==='youtube'">
     <div class="field-group"><label for="youtube-audience">Audience</label><select id="youtube-audience" v-model="audience" class="ui-field"><option value="" disabled>Choose an audience</option><option value="general">Not made for kids</option><option value="kids">Made for kids</option></select></div>
     <label><input v-model="synthetic" type="checkbox"> Contains realistic altered or synthetic content</label>
    </template>
    <label><input v-model="confirmed" type="checkbox"> {{directTikTok?`I confirm ${account.name} and ${privacyNames[options.privacy]||'the selected visibility'}.`:browserX?'I checked the active account in X.':`I confirm ${account.name} and this ${mode==='draft'?'private / inbox transfer':'public post'}.`}}</label>
    <label v-if="platform==='x' && !browserX"><input v-model="paidAccepted" type="checkbox"> I accept X API usage charges for this attempt.</label>
    <div class="send-action"><UiTooltip :disabled="!!blocked" :text="directTikTok && tikTokBlocked ? 'Choose visibility, complete any commercial disclosure and accept the posting terms.' : ''"><UiButton variant="primary" :busy="busy" :disabled="blocked" @click="finalReview=true">{{busy?'Sending...':action}}</UiButton></UiTooltip></div>
   </fieldset>
   </template>
  </template>
  <template v-else-if="receipt.state && !completed">
   <UiButton :disabled="busy" @click="run(refreshReceipt)">Refresh receipt</UiButton>
   <div v-if="receipt.retryAllowed" class="retry-confirmation"><label><input v-model="reconciled" type="checkbox"> I checked this account and the upload is absent.</label><UiButton :disabled="!reconciled || busy" @click="clearAttempt">Allow another attempt</UiButton></div>
  </template>
  </template>
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
.send-receipt{display:grid;gap:8px;margin-bottom:24px;overflow-wrap:anywhere}.send-receipt strong{font-weight:500}.send-receipt span,.send-receipt small{font-size:var(--text-xs);color:var(--secondary)}
.retry-confirmation{display:grid;gap:16px;margin-top:20px}
.inbox-resend{margin:16px 0}.resend-actions{display:flex;gap:12px;flex-wrap:wrap}.retry-confirmation p{margin:0;font-size:var(--text-sm);color:var(--secondary);line-height:1.5}.retry-confirmation label{display:flex;gap:8px;align-items:flex-start}.inbox-history{margin:20px 0;font-size:var(--text-sm)}.inbox-history ul{padding:0;list-style:none}.inbox-history li{display:grid;gap:6px;padding:12px 0;overflow-wrap:anywhere}.inbox-history .mono{font-size:var(--text-xs)}
.policy-link{border:0;background:none;padding:0;font:inherit;color:var(--accent-ink);text-decoration:underline;cursor:pointer}
.tiktok-preview{display:block;width:100%;max-height:200px;background:#171717;border-radius:8px}.tiktok-caption{white-space:pre-wrap;overflow-wrap:anywhere;margin:12px 0}.tiktok-interactions,.tiktok-disclosures{display:grid;gap:12px}.tiktok-interactions label,.tiktok-disclosures label{display:flex;gap:8px;align-items:flex-start}
</style>
