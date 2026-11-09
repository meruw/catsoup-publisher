<script setup lang="ts">
import {computed,onMounted,onUnmounted,ref,watch} from 'vue';
import {transferJobs,activeTransfers,transferUpdates,cancelQueued} from '../services/transfers';
import {drafts,type PostGroup} from '../services/drafts';
import UiButton from './UiButton.vue';
import UiDialog from './UiDialog.vue';
import UiIcon from './UiIcon.vue';
defineProps<{open:boolean}>();
const emit=defineEmits<{close:[];posts:[episode?:string]}>();
const view=ref<'active'|'history'>(activeTransfers.value?'active':'history');
const now=ref(Date.now()),search=ref(''),offset=ref(0),history=ref<PostGroup[]>([]),more=ref(false),busy=ref(false),error=ref('');
let timer:ReturnType<typeof setInterval>|undefined,debounce:ReturnType<typeof setTimeout>|undefined,request=0;
const names:Record<string,string>={youtube:'YouTube',tiktok:'TikTok',instagram:'Instagram',facebook:'Facebook',x:'X'};
const groups=computed(()=>{
 const live=new Set(transferJobs.filter(j=>j.state==='queued'||j.state==='sending').map(j=>j.episode));
 const latest=new Map<string,(typeof transferJobs)[number]>();
 for(const job of [...transferJobs].sort((a,b)=>b.id-a.id))if(live.has(job.episode)&&!latest.has(job.destination))latest.set(job.destination,job);
 return [...live].map(episode=>({episode,title:[...latest.values()].find(j=>j.episode===episode)?.title||episode,jobs:[...latest.values()].filter(j=>j.episode===episode).sort((a,b)=>a.language.localeCompare(b.language)||a.platform.localeCompare(b.platform))}));
});
const activePage=ref(0),visibleGroups=computed(()=>groups.value.slice(activePage.value*8,activePage.value*8+8));
watch(()=>groups.value.length,n=>activePage.value=Math.min(activePage.value,Math.max(0,Math.ceil(n/8)-1)));
const pending=(jobs:typeof transferJobs)=>jobs.filter(j=>j.state==='queued'||j.state==='sending').length;
const elapsed=(start:number,end:number)=>{const s=Math.max(0,Math.floor(((end||now.value)-start)/1000));return Math.floor(s/60)+':'+String(s%60).padStart(2,'0')};
const date=(value:string)=>new Date(value).toLocaleString([], {month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'});
function summary(g:PostGroup){const n=g.counts||{},attention=(n.check||0)+(n.failed||0)+(n.sending||0);return attention?attention+' to check Â· '+g.total+' destinations':g.total+' sent'}
async function load(){
 const run=++request;busy.value=true;error.value='';
 try{const result=await drafts.groups({episodeMatches:[],episode:'',search:search.value.trim(),platform:'',language:'',state:'transfers',offset:offset.value});if(run===request){history.value=result.items||[];more.value=result.more;}}
 catch(e){if(run===request){error.value=String(e);history.value=[];more.value=false;}}
 finally{if(run===request)busy.value=false;}
}
watch(search,()=>{offset.value=0;++request;busy.value=true;clearTimeout(debounce);debounce=setTimeout(load,200)});
watch(view,v=>{if(v==='history')void load()});
watch(transferUpdates,()=>{if(view.value==='history')void load()});
function page(delta:number){offset.value=Math.max(0,offset.value+delta);void load()}
onMounted(()=>{timer=setInterval(()=>now.value=Date.now(),1000);if(view.value==='history')void load()});
onUnmounted(()=>{clearInterval(timer);clearTimeout(debounce);++request});
</script>
<template>
 <UiDialog title="Uploads" close-label="Close uploads" return-focus-id="uploads-launcher" @close="emit('close')">
  <nav class="upload-tabs" aria-label="Upload views"><button :aria-current="view==='active'?'page':undefined" @click="view='active'">In progress<span v-if="activeTransfers"> {{activeTransfers}}</span></button><button :aria-current="view==='history'?'page':undefined" @click="view='history'">History</button></nav>
  <template v-if="view==='active'">
   <p class="upload-note">Keep Publisher open while transfers are running</p>
   <div class="upload-groups" aria-label="Upload activity">
    <details v-for="group in visibleGroups" :key="group.episode" class="upload-group">
     <summary><span class="group-title">{{group.title}}</span><span class="group-count">{{pending(group.jobs)}} active</span><UiIcon name="chevron"/></summary>
     <ul>
      <li v-for="job in group.jobs" :key="job.id">
       <div class="upload-heading"><strong>{{names[job.platform]}} Â· {{job.language.toUpperCase()}}</strong><span v-if="job.started" class="upload-time">{{elapsed(job.started,job.finished)}}</span></div>
       <p class="upload-account">{{job.account}} Â· {{job.mode}}</p>
       <div class="upload-result" :class="job.state" role="status"><span>{{job.state==='queued'?'Waiting for '+names[job.platform]:job.state==='sending'?(job.result||'Sending / awaiting platform confirmation'):job.state==='attention'?'Needs attention':job.result}}</span><UiButton v-if="job.state==='queued'" variant="quiet" size="compact" @click="cancelQueued(job.id)">Cancel</UiButton></div>
       <p v-if="job.state==='sending'&&now-job.started>=60000" class="upload-note">Still waiting for {{names[job.platform]}}. Other platforms can continue. Do not resend this video while this attempt is active.</p>
       <details v-if="job.state==='attention'" class="upload-error"><summary>Details</summary><p>{{job.result}}</p></details>
      </li>
     </ul>
     <UiButton variant="quiet" size="compact" @click="emit('posts',group.episode)">View post details</UiButton>
    </details>
   </div>
   <p v-if="!groups.length" class="upload-empty">No uploads running</p>
   <footer v-if="groups.length>8"><UiButton variant="quiet" :disabled="!activePage" @click="activePage--">Previous</UiButton><span>{{activePage+1}} / {{Math.ceil(groups.length/8)}}</span><UiButton variant="quiet" :disabled="(activePage+1)*8>=groups.length" @click="activePage++">Next</UiButton></footer>
  </template>
  <template v-else>
   <div class="history-search"><input v-model="search" type="search" class="ui-field" aria-label="Search upload history" placeholder="Search videos or accounts" maxlength="200"><UiButton variant="quiet" :disabled="busy" aria-label="Refresh upload history" @click="load"><UiIcon name="refresh"/></UiButton></div>
   <p class="upload-note">Latest saved results by video Â· kept on this PC</p>
   <p v-if="error" role="alert" class="notice error">{{error}}</p>
   <div class="history-list" :aria-busy="busy"><button v-for="group in history" :key="group.episode" class="history-row" @click="emit('posts',group.episode)"><span class="group-title">{{group.title||group.episode}}</span><span>{{summary(group)}}</span><time>{{date(group.updatedAt)}}</time></button></div>
   <p v-if="busy" class="upload-empty" role="status">Loading historyâ€¦</p><p v-else-if="!history.length&&!error" class="upload-empty">{{search?'No matching uploads':'No saved uploads'}}</p>
   <footer v-if="offset||more"><UiButton variant="quiet" :disabled="busy||!offset" @click="page(-20)">Previous</UiButton><span>{{offset+1}}â€“{{offset+history.length}}</span><UiButton variant="quiet" :disabled="busy||!more" @click="page(20)">Next</UiButton></footer>
  </template>
 </UiDialog>
</template>
<style scoped>
.upload-tabs{display:flex;gap:20px;border-bottom:1px solid var(--line);margin-bottom:20px}.upload-tabs button{background:none;border:0;padding:8px 0 12px;color:var(--secondary);border-bottom:2px solid transparent;font-size:var(--text-sm)}.upload-tabs button[aria-current]{border-color:var(--accent);color:var(--text)}.upload-tabs span{margin-left:4px;font-variant-numeric:tabular-nums}.upload-note,.upload-empty{font-size:var(--text-xs);color:var(--secondary);line-height:1.6;margin:12px 0}.upload-empty{padding:20px 0}.upload-group{border-bottom:1px solid var(--line);padding:4px 0}.upload-group>summary{display:flex;align-items:center;gap:12px;cursor:pointer;list-style:none;min-height:56px;padding:12px 0}.upload-group>summary::-webkit-details-marker{display:none}.group-title{font-weight:500;color:var(--text);font-size:var(--text-sm);overflow-wrap:anywhere;min-width:0;flex:1}.group-count{color:var(--secondary);font-size:var(--text-xs);white-space:nowrap}.upload-group[open]>summary :deep(.ui-icon){transform:rotate(180deg)}.upload-group ul{list-style:none;padding:0;margin:0 0 12px}.upload-group li{padding:12px 0;border-top:1px solid var(--line)}.upload-heading,.upload-result,footer{display:flex;align-items:center;justify-content:space-between;gap:12px}.upload-heading strong{font-size:var(--text-sm);font-weight:500}.upload-time,footer span{font-variant-numeric:tabular-nums;color:var(--muted);font-size:var(--text-xs)}.upload-account{font-size:var(--text-xs);overflow-wrap:anywhere;margin:4px 0;color:var(--secondary)}.upload-result{font-size:var(--text-xs);margin-top:8px;min-height:24px}.done{color:var(--success)}.attention{color:var(--warning)}.queued,.cancelled{color:var(--muted)}.upload-error{margin-top:8px;font-size:var(--text-xs);overflow-wrap:anywhere}.upload-error p{margin-top:8px}.history-search{display:flex;align-items:center;gap:8px;padding:4px;margin:-4px}.history-search input{min-width:0;flex:1;font-size:var(--text-sm)}.history-row{display:grid;grid-template-columns:1fr auto;gap:8px 16px;width:100%;border:0;border-bottom:1px solid var(--line);background:transparent;text-align:left;padding:16px 4px}.history-row:hover{background:var(--raised)}.history-row .group-title{grid-column:1/-1}.history-row>span:not(.group-title),.history-row time{font-size:var(--text-xs);color:var(--secondary)}.history-row time{font-variant-numeric:tabular-nums;color:var(--muted)}footer{padding-top:16px;margin-top:8px;border-top:1px solid var(--line)}
</style>
