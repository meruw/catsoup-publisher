<script setup lang="ts">
import {computed,onMounted,onUnmounted,ref,watch} from 'vue';
import {Events} from '@wailsio/runtime';
import {uploads} from '../services/uploads';
import type {DraftKey} from '../services/drafts';
import type {Receipt} from '../../bindings/catsoup/publisher/internal/drafts/models';
const props=defineProps<{postKey:DraftKey;accountId:string}>();
const receipt=ref<Receipt|null>(null),error=ref('');
const names:Record<string,string>={PROCESSING_UPLOAD:'TikTok is processing the upload',PROCESSING_DOWNLOAD:'TikTok is downloading the video',SEND_TO_USER_INBOX:'TikTok confirmed the inbox notification',PUBLISH_COMPLETE:'Published from TikTok',FAILED:'TikTok processing failed'};
const tracking=computed(()=>receipt.value?.tracking);
const label=computed(()=>names[tracking.value?.status||'']||'Waiting for the first TikTok check');
const stamp=(n:number)=>new Date(n*1000).toLocaleString();
let generation=0,off:(()=>void)|undefined;
async function load(){const run=++generation;try{const r=await uploads.receipt(props.postKey,props.accountId);if(run===generation){receipt.value=r;error.value=''}}catch{if(run===generation)error.value='Could not load the saved TikTok status.'}}
watch(()=>[props.postKey,props.accountId],()=>{receipt.value=null;void load()});
onMounted(()=>{void load();off=Events.On('publisher:receipts-changed',()=>void load())});
onUnmounted(()=>{++generation;off?.()});
</script>
<template>
 <section v-if="receipt?.reference.startsWith('v_inbox_file')||receipt?.inboxHistory?.length" class="tiktok-tracking" aria-label="TikTok tracking">
  <strong>{{label}}</strong>
  <p v-if="tracking?.status==='SEND_TO_USER_INBOX'">TikTok reports the notification as sent. Publisher cannot confirm that it is visible on your phone.</p>
  <p v-else-if="tracking?.reason">{{tracking.reason}}</p>
  <p v-if="tracking?.problem||error" class="tracking-problem">{{tracking?.problem||error}}</p>
  <small v-if="tracking?.checkedAt">Last check · {{stamp(tracking.checkedAt)}}</small>
  <small v-if="!['PUBLISH_COMPLETE','FAILED'].includes(tracking?.status||'')">Checks automatically while Publisher is open and resumes after restart. No video is resent.</small>
  <details v-if="receipt?.inboxHistory?.length"><summary>Previous attempts · {{receipt.inboxHistory.length}}</summary><ul><li v-for="item in receipt.inboxHistory" :key="item.reference"><span>{{names[item.tracking?.status||'']||'Waiting for check'}}</span><small v-if="item.tracking?.checkedAt">{{stamp(item.tracking.checkedAt)}}</small><small v-if="item.tracking?.problem">{{item.tracking.problem}}</small><code>{{item.reference}}</code></li></ul></details>
 </section>
 <p v-else-if="error" role="status">{{error}}</p>
</template>
<style scoped>
.tiktok-tracking{display:grid;gap:10px;padding:18px 0;border-top:1px solid var(--line);margin:16px 0;font-size:var(--text-sm)}.tiktok-tracking strong{font-weight:500}.tiktok-tracking p{margin:0;line-height:1.5}.tiktok-tracking small{color:var(--secondary);font-size:var(--text-xs);line-height:1.5}.tracking-problem{color:var(--warning)}summary{cursor:pointer;padding:8px 0}ul{list-style:none;padding:0;margin:0}li{display:grid;gap:6px;padding:12px 0;border-top:1px solid var(--line)}code{overflow-wrap:anywhere;font-size:var(--text-xs);color:var(--secondary)}
</style>
