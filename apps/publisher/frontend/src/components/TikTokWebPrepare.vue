<script setup lang="ts">
import {computed,nextTick,ref} from 'vue';
import type {Draft} from '../services/drafts';
import type {Profile} from '../services/uploads';
import {prepareTikTokWeb} from '../services/tiktokWeb';
import {enqueueTransfers,transferPending} from '../services/transfers';
import TikTokBrowserSession from './TikTokBrowserSession.vue';
import UiButton from './UiButton.vue';
const props=defineProps<{draft:Draft;account:Profile;receipt:string;hint:string;flush:()=>Promise<void>}>();
const emit=defineEmits<{queued:[]}>();
const username=ref(''),error=ref(''),busy=ref(false);
const blocker=computed(()=>props.receipt?'This video already has a TikTok receipt. Check its result in Posts; another upload is blocked.':props.hint.startsWith('Enable Direct Post')?'':props.hint);
async function prepare(){
 busy.value=true;error.value='';
 try{
  await props.flush();await nextTick();
  const d:Draft=JSON.parse(JSON.stringify(props.draft)),p:Profile=JSON.parse(JSON.stringify(props.account));
  enqueueTransfers([{draft:d,profile:p,mode:'Prepare in browser',send:()=>prepareTikTokWeb(d,p.id)}]);emit('queued');
 }catch(e){error.value=String(e)}finally{busy.value=false}
}
</script>
<template>
 <div class="web-preparation">
  <p>Attach the video and fill its caption in your {{draft.key.language==='es'?'Spanish':'English'}} browser session. You review visibility and click Publish in TikTok.</p>
  <TikTokBrowserSession v-if="account.id" :key="account.id" :language="draft.key.language" :account="account.id" @verified="username=$event"/>
  <p v-if="blocker" class="notice" role="status">{{blocker}}</p>
  <p v-if="receipt.startsWith('web_')" role="status">{{receipt==='web_ready'?'Ready in TikTok. Publishing has not been confirmed.':'Check the open TikTok browser. Publisher will not attach this video again.'}}</p>
  <UiButton variant="primary" :busy="busy" :disabled="!!blocker || !username || transferPending(draft)" @click="prepare">Prepare in TikTok</UiButton>
  <small>Keep Publisher open until you finish in TikTok</small>
  <p v-if="error" role="alert" class="notice error">{{error}}</p>
 </div>
</template>
<style scoped>
.web-preparation{display:grid;gap:var(--space-4);padding-top:var(--space-4)}
.web-preparation>p{margin:0;line-height:1.5;color:var(--secondary)}
.web-preparation>button{justify-self:start}.web-preparation>small{color:var(--muted)}
</style>
