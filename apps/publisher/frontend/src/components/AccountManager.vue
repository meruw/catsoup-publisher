<script setup lang="ts">
import { computed,onMounted,ref } from 'vue';
import { uploads, type Profile } from '../services/uploads';
import UiButton from './UiButton.vue';
import UiSegmented from './UiSegmented.vue';

const emit=defineEmits<{busy:[value:boolean]}>();
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
const message=ref('');
const error=ref('');
const langName=computed(()=>language.value==='es'?'Spanish':'English');
function account(p:Platform){return rows.value.find(a=>a.language===language.value && a.platform===p);}
function reset(){editing.value=null;remove.value=null;pageToken.value='';handle.value='';advanced.value=false;message.value='';error.value='';}
function edit(p:Platform){reset();editing.value=p;handle.value=account(p)?.name||'';}
function switchLanguage(value:string){if(!pending.value){reset();language.value=value;}}
async function run(task:()=>Promise<void>){
 if(pending.value)return;error.value='';pending.value=true;emit('busy',true);
 try{await task();}catch(e){error.value=String(e);}finally{pending.value=false;emit('busy',false);}
}
async function load(){rows.value=await uploads.accounts()||[];}
function link(p:Platform){void run(async()=>{
 try{const profile=await uploads.link(p,language.value,account(p)?.id||'',handle.value,advanced.value?pageToken.value:'');
 await load();reset();message.value=profile.name+' assigned to '+langName.value+'.';}
 finally{pageToken.value='';}
});}
function unlink(p:Platform){void run(async()=>{
 const current=account(p);if(!current)return;
 await uploads.unlink(p,current.language,current.id);await load();reset();message.value='Account removed from Publisher.';
});}
function importFile(event:Event,p:Platform){
 const input=event.target as HTMLInputElement,file=input.files?.[0];input.value='';if(!file)return;
 void run(async()=>{if(file.size>32768)throw Error('Choose a client JSON file under 32 KiB.');
 await uploads.configure(p,await file.text());message.value='Client configuration saved for '+platforms[p]+'.';});
}
onMounted(()=>{void run(load);});
</script>
<template>
 <main class="account-manager" :aria-busy="pending">
  <header class="page-header"><div><h1>Accounts</h1><p>Manage where each language uploads.</p></div></header>
  <div class="section-header"><h2>{{ langName }} accounts</h2>
   <UiSegmented label="Account language" :model-value="language" :disabled="pending" :options="[{value:'es',label:'ES',accessibleLabel:'Spanish accounts'},{value:'en',label:'EN',accessibleLabel:'English accounts'}]" @update:model-value="switchLanguage"/>
  </div>
  <section v-for="(name,p) in platforms" :key="p" class="account-row" :aria-label="name">
   <div class="section-header"><div><h3>{{ name }}</h3>
    <p>{{ account(p)?.name || (p==='x'?'No browser handle assigned':'Not connected') }}</p>
    <small v-if="account(p) && p!=='x'" class="mono">{{ account(p)?.id }}</small>
   </div>
    <div class="draft-actions" v-if="editing !== p && remove !== p">
     <UiButton :disabled="pending" @click="edit(p)">{{ account(p)?'Change':p==='x'?'Set handle':'Connect' }}</UiButton>
     <UiButton v-if="account(p)" variant="quiet" :disabled="pending" @click="reset(); remove=p">Remove</UiButton>
    </div>
   </div>
   <div v-if="remove===p" class="draft-fields"><p>Remove {{ account(p)?.name }} from {{ langName }} uploads?</p>
    <div class="draft-actions"><UiButton :disabled="pending" @click="unlink(p)">Remove account</UiButton><UiButton variant="quiet" :disabled="pending" @click="reset">Cancel</UiButton></div>
   </div>
   <div v-if="editing===p" class="draft-fields">
    <label v-if="p==='instagram'||p==='x'" class="field-group">{{ name }} username<input v-model="handle" class="ui-field" placeholder="@username" :disabled="pending" maxlength="32"></label>
    <p v-if="p==='x'" class="publication-hint">Confirm this handle in the X browser composer when posting.</p>
    <details v-if="p!=='x'"><summary>Developer setup</summary><label class="field-group">Import {{ name }} app credentials (.json)<input type="file" accept=".json" :disabled="pending" @change="importFile($event,p)"></label>
     <label v-if="p==='instagram'"><input type="checkbox" v-model="advanced" :disabled="pending"> Use an existing Facebook Page token</label>
     <input v-if="p==='instagram'&&advanced" type="password" class="ui-field" v-model="pageToken" aria-label="Facebook Page access token" autocomplete="off" :disabled="pending">
    </details>
    <div class="draft-actions"><UiButton variant="primary" :busy="pending" @click="link(p)">{{ p==='x'?'Save handle':p==='instagram'?(advanced?'Connect Page':'Sign in with Facebook'):p==='youtube'?'Sign in with Google':'Sign in with TikTok' }}</UiButton><UiButton variant="quiet" :disabled="pending" @click="reset">Cancel</UiButton></div>
   </div>
  </section>
  <p v-if="pending && editing" role="status">Waiting for {{ platforms[editing] }}…</p>
  <p v-if="message" role="status">{{ message }}</p>
  <p v-if="error" class="notice error" role="alert">{{ error }}</p>
 </main>
</template>
<style scoped>
.account-row {border-bottom:1px solid var(--line); padding:1rem 0;}
.account-row p {margin:.35rem 0;}
.account-row .draft-fields {margin:.8rem 0;}
.account-manager details .field-group {margin:.6rem 0;}
</style>
