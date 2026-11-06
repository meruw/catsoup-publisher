<script setup lang="ts">
import { computed,onMounted,ref } from 'vue';
import { uploads, type Profile, type SetupPlatform } from '../services/uploads';
import UiButton from './UiButton.vue';
import UiSegmented from './UiSegmented.vue';

const emit=defineEmits<{busy:[value:boolean];settings:[platform:SetupPlatform,language:string];language:[value:string];diagnostics:[service:string,language:string]}>();
const platforms = {youtube:'YouTube',tiktok:'TikTok',instagram:'Instagram',x:'X'};
type Platform = keyof typeof platforms;
const rows=ref<Profile[]>([]);
const props=defineProps<{initialLanguage:string}>();
const language=ref(props.initialLanguage);
const pending=ref(false);
const editing=ref<Platform|null>(null);
const remove=ref<Platform|null>(null);
const advanced=ref(false);
const handle=ref('');
const pageToken=ref('');
const xMode=ref('browser');
const paidAccepted=ref(false);
const message=ref('');
const error=ref('');
const metaCallbackReady=ref(false);
const configured=ref<Partial<Record<SetupPlatform,boolean>>>({});
const langName=computed(()=>language.value==='es'?'Spanish':'English');
function account(p:Platform){return rows.value.find(a=>a.language===language.value && a.platform===p);}
function reset(){editing.value=null;remove.value=null;pageToken.value='';handle.value='';advanced.value=false;paidAccepted.value=false;message.value='';error.value='';}
function edit(p:Platform){
 reset();
 if(p!=='x' && (configured.value[p]===false || (p==='instagram' && !metaCallbackReady.value))) {
   emit('settings',p,language.value);
   return;
 }
 editing.value=p;
 xMode.value=account(p)?.api?'api':'browser';
 handle.value=account(p)?.name||'';
}
function switchLanguage(value:string){if(!pending.value){reset();language.value=value;emit('language',value);}}
async function run(task:()=>Promise<void>){
 if(pending.value)return;error.value='';pending.value=true;emit('busy',true);
 try{await task();}catch(e){error.value=String(e);}finally{pending.value=false;emit('busy',false);}
}
async function load(){rows.value=await uploads.accounts()||[];}
function link(p:Platform){void run(async()=>{
 try{const profile=p==='x' && xMode.value==='api'
  ? await uploads.linkX(language.value,account(p)?.id||'',handle.value,paidAccepted.value)
  : await uploads.link(p,language.value,account(p)?.id||'',handle.value,advanced.value?pageToken.value:'');
 await load();reset();message.value=profile.name+' assigned to '+langName.value+'.';}
 finally{pageToken.value='';}
});}
function unlink(p:Platform){void run(async()=>{
 const current=account(p);if(!current)return;
 await uploads.unlink(p,current.language,current.id);await load();reset();message.value='Account removed from Publisher.';
});}
onMounted(()=>{void run(async()=>{
 await load();
 await Promise.all((['youtube','tiktok','instagram','x'] as const).map(async p=>{
   configured.value[p]=await uploads.clientReady(p);
   if(p==='instagram') metaCallbackReady.value=Boolean(await uploads.metaCallback());
 }));
});});
</script>
<template>
 <main class="account-manager" :aria-busy="pending">
  <header class="page-header"><div><h1>Accounts</h1><p>Manage where each language uploads.</p></div></header>
  <div class="section-header accounts-language"><h2>{{ langName }} accounts</h2>
   <UiSegmented label="Account language" :model-value="language" :disabled="pending" :options="[{value:'es',label:'ES',accessibleLabel:'Spanish accounts'},{value:'en',label:'EN',accessibleLabel:'English accounts'}]" @update:model-value="switchLanguage"/>
  </div>
  <section v-for="(name,p) in platforms" :key="p" class="account-row" :aria-label="name">
   <div class="account-summary">
    <h3>{{ name }}</h3>
    <div class="destination" :class="{assigned:account(p)}">
     <span class="account-name"><span v-if="account(p)" class="status-dot" aria-hidden="true"></span>{{ account(p)?.name || (p==='x'?'No handle assigned':'Not connected') }}</span>
     <small v-if="p==='tiktok' && account(p)">{{account(p)?.directPost?'Inbox + Direct Post':'Inbox only'}}</small>
     <small v-if="p==='x' && account(p)" :title="account(p)?.api?'X bills API usage to your developer project. Publisher adds no posting fee.':'A saved handle is a reminder; check the active account in your browser.'">{{ account(p)?.api?'API · Paid':'Browser · Free' }}</small>
     <details v-if="editing!==p && remove!==p && (account(p) || (p==='instagram' && (!configured.instagram || !metaCallbackReady)))" :key="language">
      <summary>{{ account(p)?'Account details':'Other sign-in options' }}</summary>
      <small v-if="account(p) && p!=='x'" class="mono">ID &middot; {{ account(p)?.id }}</small>
      <div class="draft-actions">
       <UiButton v-if="p==='instagram' && (!configured.instagram || !metaCallbackReady)" size="compact" variant="quiet" :disabled="pending" @click="reset(); editing=p; advanced=true; handle=account(p)?.name||''">Use Page token</UiButton>
       <UiButton v-if="account(p)" size="compact" variant="quiet" :disabled="pending" @click="reset(); remove=p">Remove</UiButton>
      </div>
     </details>
    </div>
    <UiButton v-if="editing!==p && remove!==p" class="row-action" :disabled="pending || (p!=='x' && configured[p]===undefined)" @click="edit(p)">{{ p!=='x' && configured[p]===false?'Set up credentials':p==='instagram' && !metaCallbackReady?'Set up callback':account(p)?'Change':p==='x'?'Set handle':'Connect' }}</UiButton>
   </div>
   <div v-if="remove===p" class="draft-fields"><p>Remove {{ account(p)?.name }} from {{ langName }} uploads?</p>
    <div class="draft-actions"><UiButton :disabled="pending" @click="unlink(p)">Remove account</UiButton><UiButton variant="quiet" :disabled="pending" @click="reset">Cancel</UiButton></div>
   </div>
   <div v-if="editing===p" class="draft-fields">
    <UiSegmented v-if="p==='x'" label="X publishing mode" v-model="xMode" :disabled="pending" :options="[{value:'browser',label:'Browser · Free'},{value:'api',label:'API · Paid'}]" @update:model-value="paidAccepted=false"/>
    <label v-if="p==='instagram'||p==='x'" class="field-group">{{ name }} username<input v-model="handle" class="ui-field" placeholder="@username" :disabled="pending" maxlength="32"></label>
    <p v-if="p==='x' && xMode==='browser'" class="publication-hint">Use X's account switcher in one browser. Confirm this handle before posting; Publisher cannot select the browser account.</p>
    <template v-if="p==='x' && xMode==='api'">
     <p class="publication-hint">Sign in with the username above. Videos can be published directly from Publisher.</p>
     <label v-if="configured.x"><input v-model="paidAccepted" type="checkbox" :disabled="pending"> I understand X may charge my developer project for API usage.</label>
     <UiButton v-else :disabled="pending" @click="emit('settings','x',language)">Set up X credentials</UiButton>
    </template>
    <details v-if="p==='instagram'" :open="advanced"><summary>Other sign-in options</summary>
     <label><input type="checkbox" v-model="advanced" :disabled="pending"> Use an existing Facebook Page token</label>
     <input v-if="advanced" type="password" class="ui-field" v-model="pageToken" aria-label="Facebook Page access token" autocomplete="off" :disabled="pending">
    </details>
    <div class="draft-actions"><UiButton v-if="(p==='x' && (xMode==='browser'||configured.x))||(p!=='x' && configured[p] && (p!=='instagram'||metaCallbackReady))||(p==='instagram'&&advanced)" variant="primary" :busy="pending" :disabled="(advanced && !pageToken.trim()) || ((p==='instagram'||p==='x') && !handle.trim()) || (p==='x' && xMode==='api' && !paidAccepted)" @click="link(p)">{{ p==='x'?(xMode==='api'?'Sign in with X':'Save handle'):p==='instagram'?(advanced?'Connect Page':'Sign in with Facebook'):p==='youtube'?'Sign in with Google':'Sign in with TikTok' }}</UiButton><UiButton variant="quiet" :disabled="pending" @click="reset">Cancel</UiButton></div>
    <p v-if="pending" role="status">Connecting to {{ name }}…</p>
    <div v-else-if="error"><p class="notice error" role="alert">{{ error }}</p><UiButton size="compact" variant="quiet" @click="emit('diagnostics',p,language)">View diagnostics</UiButton></div>
   </div>
  </section>
  <p v-if="message" role="status">{{ message }}</p>
  <p v-if="error && !editing" class="notice error" role="alert">{{ error }}</p>
 </main>
</template>
<style scoped>
.accounts-language {padding-block:var(--space-5) var(--space-2)}
.account-manager {container-type:inline-size}
.account-row {padding-block:var(--space-4);border-bottom:1px solid var(--line)}
.account-summary {display:grid;grid-template-columns:minmax(6rem,1fr) minmax(0,2fr) 10rem;align-items:baseline;gap:var(--space-3)}
.account-summary h3 {font-size:var(--text-body);font-weight:var(--weight-strong)}
.destination {min-width:0;color:var(--muted)}
.account-name {display:inline-flex;align-items:center;gap:var(--space-2);overflow-wrap:anywhere}
.assigned .account-name {color:var(--text);font-weight:var(--weight-medium)}
.assigned .status-dot {background:var(--success)}
.destination details {margin-top:var(--space-1);font-size:var(--text-sm)}
.destination summary {cursor:pointer;color:var(--secondary);width:fit-content}
.destination small {display:block;margin-block:var(--space-2);overflow-wrap:anywhere}
.account-row .draft-fields {display:grid;gap:var(--space-3)}
.row-action {justify-self:end}
.account-row .draft-fields {margin-top:var(--space-4)}
@container (max-width: 30rem) {
 .account-summary {grid-template-columns:minmax(0,1fr) auto;row-gap:var(--space-1)}
 .destination {grid-column:1}
 .row-action {grid-column:2;grid-row:1 / span 2;align-self:center}
}
</style>
