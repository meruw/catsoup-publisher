<script setup lang="ts">
import { ref, watch } from 'vue';
import { uploads, type SetupPlatform } from '../services/uploads';
import UiButton from './UiButton.vue';
import CredentialState from './CredentialState.vue';

const props = defineProps<{ initialProvider: SetupPlatform; statuses: Record<string,string | undefined> }>();
const emit = defineEmits<{ busy: [value: boolean]; saved: [] }>();
const names = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram' };
const provider = ref(props.initialProvider);

const pending = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const identifier = ref('');
const secret = ref('');
const configID = ref('');
const defaultCallback = 'https://publisher.meruw.dev/auth/meta/';
const callback = ref('');
const callbackDraft = ref('');
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
    if (!id || !client_secret || (provider.value==='instagram' && !configID.value.trim())) return;
    const client = provider.value==='youtube'
      ? { installed: { client_id: id, client_secret } }
      : provider.value==='instagram' ? {app_id:id,app_secret:client_secret,config_id:configID.value.trim()} : { client_key: id, client_secret };
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
    callback.value = callbackDraft.value = '';
    callback.value = callbackDraft.value = await uploads.metaCallback();
  });
}
watch(provider, p => { if (p==='instagram') loadCallback(); }, { immediate: true });
function saveCallback(value: string) {
  void run(async () => {
    callback.value = callbackDraft.value = await uploads.setMetaCallback(value);
    message.value = 'Callback URL saved. Register this exact URL in your Meta app.';
  });
}
function openGoogleSetup(step: string) { void run(() => uploads.googleSetup(step)); }

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
    <div v-if="provider==='instagram'" class="return-url">
     <p v-if="callback">Redirect URI for Meta: <code>{{ callback }}</code></p>
     <details>
      <summary>Callback hosting</summary>
      <form @submit.prevent="saveCallback(callbackDraft)">
       <label class="field-group">HTTPS callback URL<input v-model="callbackDraft" type="url" class="ui-field" autocomplete="off" spellcheck="false" :placeholder="defaultCallback" :disabled="pending" required></label>
       <p>Use the Publisher-hosted page, or host the return page on a domain you control. Register the same URL in Meta.</p>
       <div class="draft-actions"><UiButton type="submit" :disabled="pending || !callbackDraft.trim() || callbackDraft===callback">Save callback</UiButton>
        <UiButton variant="quiet" :disabled="pending || (callback===defaultCallback && callbackDraft===defaultCallback)" @click="saveCallback(defaultCallback)">Use Publisher default</UiButton></div>
      </form>
     </details>
    </div>
    <form class="credential-form" @submit.prevent="saveFields">
     <div class="fields">
      <label class="field-group">{{ provider==='youtube'?'Client ID':provider==='instagram'?'App ID':'Client key' }}<input class="ui-field" v-model="identifier" autocomplete="off" spellcheck="false" :disabled="pending" required></label>
      <label class="field-group">{{ provider==='instagram'?'App secret':'Client secret' }}<input class="ui-field" type="password" v-model="secret" autocomplete="off" spellcheck="false" :disabled="pending" required></label>
      <label v-if="provider==='instagram'" class="field-group">Configuration ID<input class="ui-field" v-model="configID" autocomplete="off" spellcheck="false" :disabled="pending" required></label>
     </div>
     <UiButton variant="primary" type="submit" :disabled="pending || !identifier.trim() || !secret.trim() || (provider==='instagram' && !configID.trim())">Save {{ names[provider] }} credentials</UiButton>
    </form>
    <p v-if="message" class="feedback" role="status">{{ message }}</p>
    <p v-if="error" class="notice error" role="alert">{{ error }}</p>
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
.providers {display:grid;grid-template-columns:repeat(3,minmax(0,1fr));border-bottom:1px solid var(--line);margin-top:var(--space-4)}
.provider {display:grid;align-content:start;gap:var(--space-1);text-align:left;padding:var(--space-3);background:transparent;border:0;border-bottom:2px solid transparent;color:var(--secondary)}
.provider:hover:not(:disabled) {background:var(--sidebar)}
.provider[aria-pressed="true"] {color:var(--text);border-bottom-color:var(--text)}
.setup {padding-block:var(--space-6)}
.setup-heading,.google-help li div {display:flex;align-items:center;justify-content:space-between;gap:var(--space-4)}
.setup-heading {flex-wrap:wrap;margin-bottom:var(--space-2)}
.return-url details {margin-top:var(--space-2)}
.return-url summary {cursor:pointer;width:fit-content}
.return-url form {margin-block:var(--space-3)}
.return-url .draft-actions {margin-top:var(--space-2)}
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
