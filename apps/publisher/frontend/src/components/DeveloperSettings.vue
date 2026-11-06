<script setup lang="ts">
import { ref, watch } from 'vue';
import { uploads, type SetupPlatform } from '../services/uploads';
import UiButton from './UiButton.vue';
import CredentialState from './CredentialState.vue';

const props = defineProps<{ initialProvider: SetupPlatform; statuses: Record<string,string | undefined> }>();
const emit = defineEmits<{ busy: [value: boolean]; saved: [] }>();
const names = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram', x:'X' };
const provider = ref(props.initialProvider);

const pending = ref(false);
const tikTokDirect=ref(false);
function toggleTikTok(event:Event){const control=event.target as HTMLInputElement;const value=control.checked;void run(async()=>{try{await uploads.setTikTokDirectEnabled(value);tikTokDirect.value=value;emit('saved');message.value=value?'Direct Post enabled. Reconnect both TikTok accounts to authorize video.publish.':'Direct Post disabled. Inbox uploads remain available.';}finally{control.checked=tikTokDirect.value;}});}
const fileInput = ref<HTMLInputElement | null>(null);
const identifier = ref('');
const secret = ref('');
const configID = ref('');
const xCallback = ref('http://127.0.0.1:48173/callback/');
const hostedCallback = 'https://publisher.meruw.dev/auth/meta/';
const callback = ref('');
const callbackDraft = ref('');
const hosting = ref<'off'|'hosted'|'custom'>('off');
const accepted = ref(false);
function showHosting(value:string) {
  callback.value=value;
  hosting.value=value===hostedCallback?'hosted':value?'custom':'off';
  callbackDraft.value=hosting.value==='custom'?value:'';
  accepted.value=value===hostedCallback;
}
const message = ref('');
const error = ref('');

async function run(action: () => Promise<void>) {
  if (pending.value) return;
  pending.value = true;
  emit('busy', true);
  message.value = '';
  error.value = '';
  try { await action(); }
  catch (e) { error.value = String(e); }
  finally { pending.value = false; emit('busy', false); }
}
function select(p: SetupPlatform) {
  if (pending.value || provider.value===p) return;
  provider.value=p; identifier.value=secret.value=configID.value=message.value=error.value='';
}
async function persist(raw: string) {
  await uploads.configure(provider.value, raw);
  identifier.value = secret.value = configID.value = '';
  emit('saved');
  message.value = names[provider.value] + ' credentials saved for both languages.';
}
function saveFields() {
  void run(async () => {
    const id = identifier.value.trim(), client_secret = secret.value.trim();
    if (!id || (provider.value!=='x' && !client_secret) || (provider.value==='instagram' && !configID.value.trim())) return;
    const client = provider.value==='x' ? {client_id:id,redirect_uri:xCallback.value.trim()} : provider.value==='youtube'
      ? { installed: { client_id: id, client_secret } }
      : provider.value==='instagram' ? {app_id:id,app_secret:client_secret,config_id:configID.value.trim()} : { client_key: id, client_secret, direct_post:tikTokDirect.value };
    await persist(JSON.stringify(client));
  });
}
function importFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (file) void run(async () => {
    if (file.size > 32768) throw Error('Choose a client JSON file under 32 KiB.');
    await persist(await file.text());
  });
}
function loadCallback() {
  void run(async () => {
    showHosting('');
    showHosting(await uploads.metaCallback());
  });
}
watch(provider, p => { if (p==='instagram') loadCallback(); if(p==='tiktok')void run(async()=>{tikTokDirect.value=await uploads.tikTokDirectEnabled()}); }, { immediate: true });
function saveCallback() {
  void run(async () => {
    if(hosting.value==='hosted' && !accepted.value) throw Error('Accept the hosting notice to continue.');
    if(hosting.value==='custom' && !callbackDraft.value.trim()) throw Error('Enter your HTTPS callback URL.');
    const value = hosting.value==='hosted' ? await uploads.enableHostedMetaCallback()
      : await uploads.setMetaCallback(hosting.value==='custom'?callbackDraft.value.trim():'');
    showHosting(value);
    message.value = value?'Callback saved. Register this exact URL in your Meta app.':'Meta browser sign-in is off.';
  });
}
function openGoogleSetup(step: string) { void run(() => uploads.googleSetup(step)); }
function openXSetup(step: string) { void run(() => uploads.xSetup(step)); }

</script>
<template>
 <div :aria-busy="pending">
  <header class="page-header"><div><h1>Developer credentials</h1><p>Configure each provider once for both languages.</p></div></header>
  <div class="credential-workspace">
   <div class="providers" aria-label="Provider settings" role="group">
    <button v-for="(name,p) in names" :key="p" type="button" class="provider" :aria-pressed="provider===p" :disabled="pending" @click="select(p)">
     <strong>{{ name }}</strong><CredentialState :state="statuses[p]"/>
    </button>
   </div>
   <section class="setup" :key="provider" :aria-label="names[provider]+' settings'">
    <div class="setup-heading">
     <h2>{{ names[provider] }} credentials</h2>
     <UiButton :disabled="pending" @click="fileInput?.click()">Import JSON…</UiButton>
     <input ref="fileInput" type="file" hidden accept=".json,application/json" :aria-label="'Import '+names[provider]+' client JSON'" :disabled="pending" @change="importFile">
    </div>
    <CredentialState :state="statuses[provider]" detail/>
    <div v-if="provider==='x'" class="return-url">
     <p><strong>Optional paid API</strong> · Browser posting stays free.</p>
     <p>X bills API requests to your developer project. Saving these settings does not connect an account or publish anything.</p>
    </div>
    <div v-if="provider==='instagram'" class="return-url">
     <p v-if="callback">Redirect URI for Meta: <code>{{ callback }}</code></p>
     <p v-if="callback===hostedCallback">CATsoup hosting is on. GitHub Pages receives the temporary sign-in code; your tokens stay on this PC.</p>
     <p v-if="!callback && !pending">Browser sign-in is off until you choose a callback.</p>
     <details>
      <summary>Callback hosting</summary>
      <form @submit.prevent="saveCallback">
       <fieldset :disabled="pending" class="hosting-options">
        <legend>Where should Meta return after sign-in?</legend>
        <label><input v-model="hosting" type="radio" value="off"> Off</label>
        <label><input v-model="hosting" type="radio" value="custom"> My HTTPS domain</label>
        <label><input v-model="hosting" type="radio" value="hosted"> CATsoup-hosted callback</label>
       </fieldset>
       <label v-if="hosting==='custom'" class="field-group">HTTPS callback URL<input v-model="callbackDraft" type="url" class="ui-field" autocomplete="off" spellcheck="false" placeholder="https://studio.example.org/auth/meta/" :disabled="pending" required></label>
       <p v-if="hosting==='custom'">Host the return page on a domain you control and register its exact URL in Meta.</p>
       <div v-if="hosting==='hosted'" class="hosting-notice">
        <p>During Facebook sign-in, GitHub Pages serves the return page at <code>{{ hostedCallback }}</code> and receives the temporary authorization code and state in the URL. Publisher exchanges the code on your PC. Tokens and your app secret stay on your PC. New sign-ins require this site to be available.</p>
        <label><input v-model="accepted" type="checkbox" :disabled="pending"> I agree to use the CATsoup-hosted callback.</label>
       </div>
       <UiButton type="submit" :disabled="pending || (hosting==='hosted' && !accepted) || (hosting==='custom' && !callbackDraft.trim())">Save callback</UiButton>
      </form>
     </details>
    </div>
    <section v-if="provider==='tiktok'" class="return-url">
     <label><input type="checkbox" :checked="tikTokDirect" :disabled="pending || !statuses.tiktok || statuses.tiktok==='missing'" @change="toggleTikTok"> Enable Direct Post</label>
     <p>Enable Direct Post and video.publish in your TikTok developer app first, then reconnect each account. Public visibility requires TikTok's audit; unaudited clients are restricted to private accounts and Only me posts.</p>
    </section>
    <form class="credential-form" @submit.prevent="saveFields">
     <div class="fields">
      <label class="field-group">{{ provider==='youtube'||provider==='x'?'Client ID':provider==='instagram'?'App ID':'Client key' }}<input class="ui-field" v-model="identifier" autocomplete="off" spellcheck="false" :disabled="pending" required></label>
      <label v-if="provider!=='x'" class="field-group">{{ provider==='instagram'?'App secret':'Client secret' }}<input class="ui-field" type="password" v-model="secret" autocomplete="off" spellcheck="false" :disabled="pending" required></label>
      <label v-if="provider==='x'" class="field-group">Callback URL<input class="ui-field" v-model="xCallback" autocomplete="off" spellcheck="false" :disabled="pending" required><small>Register this exact URL in X. Change the port in both places if it is occupied.</small></label>
      <label v-if="provider==='instagram'" class="field-group">Configuration ID<input class="ui-field" v-model="configID" autocomplete="off" spellcheck="false" :disabled="pending" required></label>
     </div>
     <UiButton variant="primary" type="submit" :disabled="pending || !identifier.trim() || (provider!=='x' && !secret.trim()) || (provider==='instagram' && !configID.trim())">Save {{ names[provider] }} credentials</UiButton>
    </form>
    <p v-if="message" class="feedback" role="status">{{ message }}</p>
    <p v-if="error" class="notice error" role="alert">{{ error }}</p>
    <details v-if="provider==='x'" class="google-help"><summary>X Developer setup</summary>
     <ol><li>Create an app with OAuth 2.0 enabled and choose <strong>Native App</strong> (public client). Use its Client ID.</li><li>Register the callback shown above. Publisher requests tweet.read, users.read, tweet.write, media.write and offline.access.</li><li>Review X's usage prices and billing controls, then connect each language in Accounts → X → API · Paid.</li></ol>
     <div class="draft-actions"><UiButton size="compact" :disabled="pending" @click="openXSetup('console')">Developer console</UiButton><UiButton size="compact" variant="quiet" :disabled="pending" @click="openXSetup('pricing')">X pricing</UiButton><UiButton size="compact" variant="quiet" :disabled="pending" @click="openXSetup('guide')">Setup guide</UiButton></div>
    </details>
    <details v-if="provider==='youtube'" class="google-help"><summary>Google Cloud setup</summary>
     <ol>
      <li><div><span>Enable the YouTube Data API v3 in your project.</span><UiButton size="compact" :disabled="pending" @click="openGoogleSetup('api')">Open API library</UiButton></div></li>
      <li><div><span>Configure consent and add your test accounts.</span><UiButton size="compact" :disabled="pending" @click="openGoogleSetup('audience')">Open audience settings</UiButton></div></li>
      <li><div><span>Create a <strong>Desktop app</strong> client and download its JSON.</span><UiButton size="compact" :disabled="pending" @click="openGoogleSetup('client')">Open Google clients</UiButton></div></li>
     </ol>
    </details>
   </section>
  </div>
 </div>
</template>
<style scoped>
.credential-workspace {max-width:48rem;container-type:inline-size}
.providers {display:grid;grid-template-columns:repeat(4,minmax(0,1fr));border-bottom:1px solid var(--line);margin-top:var(--space-4)}
.provider {display:grid;align-content:start;gap:var(--space-1);text-align:left;padding:var(--space-3);background:transparent;border:0;border-bottom:2px solid transparent;color:var(--secondary)}
.provider:hover:not(:disabled) {background:var(--sidebar)}
.provider[aria-pressed="true"] {color:var(--text);border-bottom-color:var(--text)}
.setup {padding-block:var(--space-6)}
.setup-heading,.google-help li div {display:flex;align-items:center;justify-content:space-between;gap:var(--space-4)}
.setup-heading {flex-wrap:wrap;margin-bottom:var(--space-2)}
.return-url details {margin-top:var(--space-2)}
.return-url summary {cursor:pointer;width:fit-content}
.return-url form {margin-block:var(--space-3)}
.hosting-options {display:grid;gap:var(--space-2);padding:0;border:0;margin-bottom:var(--space-3)}
.hosting-options legend {margin-bottom:var(--space-2)}
.hosting-notice {max-width:60ch;margin-block:var(--space-3)}
.hosting-notice label {display:block;margin-top:var(--space-2)}
.return-url form>.ui-button {margin-top:var(--space-3)}
.credential-form {margin-top:var(--space-5)}
.fields {display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4)}
.return-url {margin-top:var(--space-3);overflow-wrap:anywhere}
.feedback {color:var(--success);margin-top:var(--space-3)}
.google-help {margin-top:var(--space-6);padding-top:var(--space-4);border-top:1px solid var(--line)}
.google-help summary {cursor:pointer;width:fit-content;font-weight:var(--weight-medium)}
.google-help ol {margin:var(--space-2) 0 0;padding-inline-start:var(--space-5)}
.google-help li {padding:var(--space-2) 0}
.google-help li button {flex:none}
@container (max-width:32rem) {
 .fields {grid-template-columns:1fr;gap:0}
 .google-help li div {align-items:flex-start;flex-direction:column;gap:var(--space-2)}
 .provider {padding-inline:var(--space-1)}
}
</style>
