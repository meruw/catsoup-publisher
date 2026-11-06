<script setup lang="ts">
import {computed,onMounted,onUnmounted,ref,watch} from 'vue';
import {drafts,type PostStatus,type PostGroup} from '../services/drafts';
import type {Entry} from '../services/publisher';
import UiButton from './UiButton.vue';
import UiIcon from './UiIcon.vue';
import UiDialog from './UiDialog.vue';
import UiTooltip from './UiTooltip.vue';
const props=defineProps<{entries:Entry[];openError:string}>();
const emit=defineEmits<{open:[post:PostStatus]}>();
const platforms:Record<string,string>={youtube:'YouTube',tiktok:'TikTok',instagram:'Instagram',x:'X'};
const states:Record<string,string>={draft:'Draft',approved:'Approved',published:'Published',restricted:'Limited visibility',private:'Private video',inbox:'TikTok inbox',sending:'Sending',check:'Check result',failed:'Failed'};
const hints:Record<string,string>={draft:'Saved on this PC; no completed transfer is recorded for this file.',approved:'Approved locally. Nothing has been sent yet.',published:'The provider confirmed publishing. Later changes on the platform are not monitored.',restricted:'Posted with friends or followers visibility, not public visibility.',private:'Uploaded with private visibility.',inbox:'Delivered to TikTok’s inbox. Finish posting in TikTok.',sending:'A transfer is in progress. Do not resend it.',processing:'The provider is processing this video.',check:'The result is uncertain. Check the destination before allowing another attempt.',failed:'The provider confirmed that processing failed.',browser:'Opened in the browser. Publisher cannot confirm whether you posted there.'};

const search=ref(''),platform=ref(''),language=ref(''),state=ref('');
const offset=ref(0),groups=ref<PostGroup[]>([]),more=ref(false),busy=ref(false),error=ref('');
const group=ref<PostGroup|null>(null),rows=ref<PostStatus[]>([]),detailLanguage=ref(''),detailOffset=ref(0),detailMore=ref(false),selected=ref<PostStatus|null>(null);
const filterOpen=ref(false),pendingFilters=ref({platform:'',language:'',state:''});
const activeFilters=computed(()=>[platforms[platform.value],language.value==='es'?'Spanish':language.value==='en'?'English':'',states[state.value]].filter(Boolean));
const title=(g:PostGroup)=>props.entries.find(e=>e.id===g.episode)?.title||g.title||g.episode;
const date=(s:string)=>new Date(s).toLocaleString([], {month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'});
const languages=(g:PostGroup)=>g.languages.split(',').sort().reverse().map(s=>s.toUpperCase()).join(' / ');
const summary=(g:PostGroup)=>{
 const n=g.counts||{};
 if((n.failed||0)+(n.check||0))return `${(n.failed||0)+(n.check||0)} need attention`;
 if(n.sending)return `${n.sending} sending`;
 if(n.published)return `${n.published} of ${g.total} published`;
 if(n.inbox||n.private||n.restricted)return `${(n.inbox||0)+(n.private||0)+(n.restricted||0)} sent · not public`;
 if(n.approved)return `${n.approved} of ${g.total} approved`;
 return g.total===1?'1 draft':`${g.total} drafts`;
};
const breakdown=(g:PostGroup)=>Object.entries(g.counts||{}).map(([s,n])=>`${n} ${states[s]||s}`).join(' · ');
const canOpen=computed(()=>selected.value?.connected&&props.entries.some(e=>e.id===selected.value?.key.episode&&e.valid));
let generation=0,timer:ReturnType<typeof setTimeout>|undefined;
const blankFilter=()=>({episodeMatches:[] as string[],episode:'',search:'',platform:'',language:'',state:'',offset:0});
async function load(){
 const request=++generation;busy.value=true;error.value='';
 try{
  if(group.value){
   const result=await drafts.statuses({...blankFilter(),episode:group.value.episode,language:detailLanguage.value,offset:detailOffset.value});
   if(request===generation){rows.value=result.items||[];detailMore.value=result.more;}
  }else{
   const result=await drafts.groups({...blankFilter(),search:search.value.trim(),episodeMatches:search.value.trim()?props.entries.filter(e=>e.title.toLowerCase().includes(search.value.trim().toLowerCase())).map(e=>e.id):[],platform:platform.value,language:language.value,state:state.value,offset:offset.value});
   if(request===generation){groups.value=result.items||[];more.value=result.more;}
  }
 }catch(e){if(request===generation){error.value=String(e);rows.value=[];groups.value=[];more.value=false;detailMore.value=false;}}
 finally{if(request===generation)busy.value=false;}
}
watch([search,platform,language,state],()=>{offset.value=0;clearTimeout(timer);++generation;busy.value=true;timer=setTimeout(load,200)});
watch(detailLanguage,()=>{detailOffset.value=0;void load()});
function openGroup(g:PostGroup){clearTimeout(timer);group.value=g;rows.value=[];detailLanguage.value='';detailOffset.value=0;void load();}
function back(){group.value=null;selected.value=null;void load();}
function page(delta:number){if(group.value)detailOffset.value=Math.max(0,detailOffset.value+delta);else offset.value=Math.max(0,offset.value+delta);void load();}
function filters(){pendingFilters.value={platform:platform.value,language:language.value,state:state.value};filterOpen.value=true;}
function applyFilters(){platform.value=pendingFilters.value.platform;language.value=pendingFilters.value.language;state.value=pendingFilters.value.state;filterOpen.value=false;}
function resetFilters(){platform.value='';language.value='';state.value='';}
onMounted(load);onUnmounted(()=>{clearTimeout(timer);++generation});
</script>
<template>
 <main class="publisher-main posts-main">
  <header class="page-header"><h1 v-if="!group">Posts</h1><UiButton v-else variant="quiet" @click="back">Back to Posts</UiButton><UiTooltip text="Reload saved posts and transfer results"><UiButton variant="quiet" :disabled="busy" aria-label="Refresh posts" @click="load"><UiIcon name="refresh"/></UiButton></UiTooltip></header>
  <template v-if="!group">
   <div class="posts-toolbar"><input v-model="search" type="search" class="ui-field posts-search" placeholder="Search videos or accounts" aria-label="Search posts"><UiButton variant="quiet" @click="filters">Filters<span v-if="activeFilters.length"> · {{activeFilters.length}}</span></UiButton></div>
   <div v-if="activeFilters.length" class="active-filters"><span>{{activeFilters.join(' · ')}}</span><UiButton size="compact" variant="quiet" @click="resetFilters">Clear filters</UiButton></div>
   <div class="posts-list" :aria-busy="busy">
    <div v-if="groups.length" class="posts-columns" aria-hidden="true"><span>Video</span><span>Status</span><span>Updated</span></div>
    <button v-for="item in groups" :key="item.episode" class="video-row" @click="openGroup(item)">
     <span class="video-name"><strong>{{title(item)}}</strong><small>{{languages(item)}} · {{item.platforms}} {{item.platforms===1?'platform':'platforms'}}</small></span>
     <span class="video-state" :class="{'needs-attention':item.counts?.failed||item.counts?.check}" :title="breakdown(item)">{{summary(item)}}</span>
     <time :datetime="item.updatedAt">{{date(item.updatedAt)}}</time>
    </button>
   </div>
   <p v-if="!groups.length&&!error" class="posts-empty" role="status">{{busy?'Loading videos…':search||activeFilters.length?'No matching videos':'Your saved videos will appear here'}}</p>
   <footer class="posts-footer"><span>Saved on this PC</span><div v-if="offset||more"><UiButton variant="quiet" :disabled="busy||!offset" @click="page(-20)">Previous</UiButton><span>{{offset+1}}–{{offset+groups.length}}</span><UiButton variant="quiet" :disabled="busy||!more" @click="page(20)">Next</UiButton></div></footer>
  </template>
  <template v-else>
   <header class="video-detail-heading"><h2>{{title(group)}}</h2><p>{{breakdown(group)}}</p></header>
   <nav class="language-tabs" aria-label="Video languages"><button v-for="(label,value) in {'':'All languages',es:'Spanish',en:'English'}" :key="value" :aria-current="detailLanguage===value?'page':undefined" @click="detailLanguage=value">{{label}}</button></nav>
   <div class="destination-list" :aria-busy="busy"><button v-for="row in rows" :key="JSON.stringify(row.key)" class="destination-row" @click="selected=row"><span><strong>{{platforms[row.key.platform]}}</strong><small>{{row.accountName.replace(/^account:/,'')}} · {{row.key.language.toUpperCase()}}{{row.connected?'':' · Disconnected'}}</small></span><span :class="'state-'+row.state">{{states[row.state]||'Check result'}}</span></button></div>
   <p v-if="!rows.length&&!error" class="posts-empty" role="status">{{busy?'Loading destinations…':'No saved posts in this language'}}</p>
   <footer v-if="detailOffset||detailMore" class="posts-footer"><UiButton variant="quiet" :disabled="busy||!detailOffset" @click="page(-50)">Previous</UiButton><UiButton variant="quiet" :disabled="busy||!detailMore" @click="page(50)">Next</UiButton></footer>
  </template>
  <p v-if="error||openError" class="notice error" role="alert">{{error||openError}}</p>
  <UiDialog v-if="filterOpen" title="Filter videos" close-label="Close filters" @close="filterOpen=false"><form class="filter-form" @submit.prevent="applyFilters">
   <label>Platform<select v-model="pendingFilters.platform" class="ui-field" aria-label="Filter platform"><option value="">All platforms</option><option v-for="(label,id) in platforms" :value="id" :key="id">{{label}}</option></select></label>
   <label>Language<select v-model="pendingFilters.language" class="ui-field" aria-label="Filter language"><option value="">Both languages</option><option value="es">Spanish</option><option value="en">English</option></select></label>
   <label>Status<select v-model="pendingFilters.state" class="ui-field" aria-label="Filter status"><option value="">All statuses</option><option v-for="(label,id) in states" :value="id" :key="id">{{label}}</option></select></label>
   <div class="filter-actions"><UiButton variant="quiet" @click="pendingFilters={platform:'',language:'',state:''}">Reset filters</UiButton><UiButton variant="primary" type="submit">Apply filters</UiButton></div>
  </form></UiDialog>
  <UiDialog v-if="selected" close-label="Close post details" :title="platforms[selected.key.platform]+' / '+selected.accountName.replace(/^account:/,'')" @close="selected=null"><p class="post-detail-state" :class="'state-'+selected.state">{{states[selected.state]||'Check result'}}</p><p class="muted">{{hints[selected.state]}}</p><dl class="post-details"><div><dt>Language</dt><dd>{{selected.key.language==='es'?'Spanish':'English'}}</dd></div><div><dt>Post title</dt><dd>{{selected.title||'Untitled'}}</dd></div><div><dt>Video version</dt><dd>{{selected.videoId}}</dd></div><div><dt>Last update</dt><dd>{{date(selected.updatedAt)}}</dd></div><div v-if="selected.reference"><dt>Reference</dt><dd class="mono">{{selected.reference}}</dd></div></dl><UiButton :disabled="!canOpen" @click="emit('open',selected)">Open post</UiButton><p v-if="!canOpen" class="muted">Reconnect this account and make the delivery available to edit this post.</p></UiDialog>
 </main>
</template>
<style scoped>
.posts-main{min-width:0}.posts-toolbar{display:flex;align-items:center;gap:16px;padding:24px 4px 20px;margin:0 -4px}.posts-search{flex:1;max-width:440px;min-width:0}.posts-search:focus-visible{outline-offset:2px}.active-filters{display:flex;align-items:center;gap:12px;margin-bottom:16px;color:var(--secondary);font-size:var(--text-sm)}
.posts-columns,.video-row{display:grid;grid-template-columns:minmax(0,1fr) 180px 140px;gap:24px;text-align:left;align-items:center}.posts-columns{font-size:var(--text-xs);color:var(--muted);padding:12px 16px}.video-row,.destination-row{width:100%;border:0;border-top:1px solid var(--line);background:transparent;color:var(--text);font:inherit;font-size:var(--text-sm);padding:20px 16px;cursor:pointer}.video-row:hover,.destination-row:hover{background:var(--sidebar)}.video-row:focus-visible,.destination-row:focus-visible{outline-offset:-3px}.video-row strong,.destination-row strong{display:block;font-weight:500}.video-row small,.destination-row small{display:block;color:var(--muted);font-size:var(--text-xs);margin-top:8px}.video-name{min-width:0;overflow-wrap:anywhere}.video-row time{font-size:var(--text-xs);color:var(--muted)}.video-state{font-size:var(--text-sm)}.needs-attention,.state-check{color:var(--warning)}.state-failed{color:var(--danger)}.state-published,.state-approved{color:var(--success)}
.posts-footer{display:flex;justify-content:space-between;align-items:center;margin:24px 0;color:var(--muted);font-size:var(--text-xs)}.posts-footer>div{display:flex;align-items:center;gap:12px}.posts-empty{padding:48px 0;color:var(--muted)}.filter-form{display:grid;gap:24px}.filter-form>label{display:grid;gap:8px;font-size:var(--text-sm)}.filter-actions{display:flex;justify-content:space-between;gap:16px;margin-top:8px}
.video-detail-heading{padding:28px 0 24px}.video-detail-heading p{color:var(--secondary);font-size:var(--text-sm);margin-top:12px}.language-tabs{display:flex;gap:24px;margin-bottom:8px}.language-tabs button{padding:12px 0;background:none;border:0;border-bottom:2px solid transparent;color:var(--secondary);font:inherit;font-size:var(--text-sm);cursor:pointer}.language-tabs button[aria-current]{border-color:var(--text);color:var(--text)}.destination-row{display:flex;justify-content:space-between;align-items:center;gap:24px;text-align:left}.destination-row>span:first-child{min-width:0;overflow-wrap:anywhere}.destination-row>span:last-child{flex-shrink:0}.post-detail-state{font-weight:500;margin-bottom:12px}.post-details{margin:24px 0}.post-details>div{display:flex;justify-content:space-between;gap:24px;padding:12px 0;border-top:1px solid var(--line)}.post-details dd{margin:0;overflow-wrap:anywhere;text-align:right}
@media(max-width:1000px){.posts-columns,.video-row{grid-template-columns:minmax(0,1fr) 160px;gap:16px}.posts-columns>span:last-child,.video-row>time{display:none}}@media(max-width:600px){.posts-columns{display:none}.video-row{grid-template-columns:1fr}.posts-toolbar{gap:8px}.language-tabs{gap:16px}}
</style>
