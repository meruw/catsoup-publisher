<script setup lang="ts">
import {onMounted,ref} from 'vue';
import UiButton from './UiButton.vue';
import {tikTokWeb} from '../services/tiktokWeb';
const props=defineProps<{language:string;account:string}>();
const emit=defineEmits<{verified:[username:string]}>();
const username=ref(''),saved=ref(''),busy=ref(false),error=ref(''),opened=ref(false);
async function run(action:()=>Promise<void>){busy.value=true;error.value='';try{await action()}catch(e){error.value=String(e)}finally{busy.value=false}}
onMounted(()=>void run(async()=>{saved.value=await tikTokWeb.username(props.language,props.account);username.value=saved.value;emit('verified',saved.value)}));
function open(){void run(async()=>{await tikTokWeb.open(props.language,props.account);opened.value=true})}
function verify(){void run(async()=>{saved.value=await tikTokWeb.verify(props.language,props.account,username.value);emit('verified',saved.value);opened.value=false})}
</script>
<template>
 <details class="browser-session" :open="!saved">
  <summary>{{saved?`Browser session · @${saved}`:'Set up browser session'}}</summary>
  <div class="browser-setup">
   <p>Chrome keeps a separate session for {{language==='es'?'Spanish':'English'}}. Sign in once, then verify its username.</p>
   <label class="field-group">TikTok username<input v-model="username" class="ui-field" placeholder="@username" autocomplete="off" maxlength="25" :disabled="busy"></label>
   <div class="draft-actions"><UiButton :disabled="busy" @click="open">Open TikTok browser</UiButton><UiButton :disabled="busy || !username.trim()" variant="quiet" @click="verify">Verify account</UiButton></div>
   <p v-if="opened" role="status">Sign in in the opened window, then choose Verify account</p>
   <p v-if="error" class="notice error" role="alert">{{error}}</p>
  </div>
 </details>
</template>
<style scoped>
.browser-session{margin-block:var(--space-3);font-size:var(--text-sm)}
summary{cursor:pointer;color:var(--secondary)}
.browser-setup{display:grid;gap:var(--space-3);padding-top:var(--space-3);max-width:32rem}
.browser-setup p{margin:0;color:var(--secondary);line-height:1.5}.browser-setup .field-group{margin:0}
</style>
