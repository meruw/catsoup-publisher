<script setup lang="ts">
import { watch,ref } from 'vue';
import { uploads,type SetupPlatform } from '../services/uploads';
import DeveloperSettings from './DeveloperSettings.vue';
import PublishingSettings from './PublishingSettings.vue';
import UiButton from './UiButton.vue';

const props=defineProps<{ provider: SetupPlatform|null }>();
const emit=defineEmits<{select:[provider?:SetupPlatform];busy:[value:boolean]}>();
const statuses=ref<Record<string,string|undefined>>({});
const error=ref('');
const busy=ref(false);
const publishing=ref(false);
let sequence=0;
async function refresh(){
 const request=++sequence;
 try {
  const result=await uploads.credentialStatuses();
  if(!result) throw Error('Status unavailable');
  if(request===sequence){statuses.value=result;error.value='';}
 } catch {
  if(request===sequence){statuses.value={};error.value='Could not read credential status.';}
 }
}
function setBusy(value:boolean){busy.value=value;emit('busy',value);}
watch(()=>props.provider, value=>{ if(value){publishing.value=false;void refresh();} }, {immediate:true});
</script>
<template>
 <main class="settings-page" :aria-busy="busy">
  <template v-if="provider">
   <div class="settings-back"><UiButton size="compact" variant="quiet" :disabled="busy" @click="emit('select')">Back to Settings</UiButton></div>
   <DeveloperSettings :initial-provider="provider" :statuses="statuses" @saved="refresh" @busy="setBusy"/>
  </template>
  <template v-else-if="publishing">
   <div class="settings-back"><UiButton size="compact" variant="quiet" :disabled="busy" @click="publishing=false">Back to Settings</UiButton></div>
   <PublishingSettings @busy="setBusy" @setup="emit('select','tiktok')"/>
  </template>
  <template v-else>
   <header class="page-header"><h1>Settings</h1></header>
   <section class="credential-overview" aria-label="Developer credentials">
    <div class="section-header"><div><h2>Developer credentials</h2><p>Manage the app credentials used for account sign-in.</p></div>
     <UiButton @click="emit('select','youtube')">Manage credentials</UiButton>
    </div>
   </section>
   <section class="credential-overview" aria-label="Publishing preferences"><div class="section-header"><div><h2>Publishing defaults</h2><p>Choose the starting mode for each account.</p></div><UiButton @click="publishing=true">Manage publishing</UiButton></div></section>
  </template>
  <p v-if="provider && error" class="notice error" role="alert">{{ error }} <UiButton size="compact" @click="refresh">Retry</UiButton></p>
 </main>
</template>
<style scoped>
.settings-back {padding-top:var(--space-4)}
.credential-overview + .credential-overview {border-top:1px solid var(--line)}
.credential-overview {padding-block:var(--space-6);max-width:48rem}
.credential-overview .section-header {align-items:flex-start;flex-wrap:wrap}
.credential-overview p {margin-top:var(--space-2);color:var(--secondary)}
</style>
