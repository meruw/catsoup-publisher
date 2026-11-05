<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue';
import UiButton from './UiButton.vue';
import { uploads } from '../services/uploads';
import type { Draft } from '../services/drafts';

const props = defineProps<{ draft: Draft; hint: string; flush: () => Promise<void> }>();
const emit = defineEmits<{ busy: [value: boolean] }>();
const token = ref('');
const account = ref({ id: '', name: '' });
const receipt = ref({ state: '', reference: '', retryAllowed: false });
const confirmed = ref(false);
const reconciled = ref(false);
const busy = ref(false);
const error = ref('');
const platform = computed(() => props.draft.key.platform);
const action = computed(() => ({ youtube: 'Upload privately to YouTube', tiktok: 'Send to TikTok inbox', instagram: 'Publish Reel on Instagram' })[platform.value]);
const note = computed(() => ({
 youtube: 'Google will open in your browser. Your video will be private.',
 tiktok: 'Finish editing and posting from your TikTok inbox. Copy the caption from the draft.',
 instagram: 'Publishing makes the Reel visible on the connected Instagram account.',
})[platform.value]);
async function run(task: () => Promise<void>) {
 if (busy.value) return;
 busy.value = true; emit('busy', true); error.value = '';
 try { await task(); } catch(e) { error.value = String(e); }
 finally { busy.value = false; emit('busy', false); }
}
async function refreshReceipt() {
 if (account.value.id) receipt.value = await uploads.receipt(props.draft.key, account.value.id);
}
onMounted(async () => {
 if (platform.value === 'x') return;
 try { account.value = await uploads.connected(props.draft.key); await refreshReceipt(); } catch(e) { error.value = String(e); }
});
function connect() { void run(async () => {
 try { account.value = await uploads.connect(props.draft.key, token.value.trim()); }
 finally { token.value = ''; }
 await refreshReceipt();
}); }
function disconnect() { void run(async () => {
 await uploads.disconnect(props.draft.key); account.value = { id: '', name: '' }; confirmed.value = false; reconciled.value = false;
}); }
function send() { void run(async () => {
 await props.flush(); await nextTick();
 try { receipt.value = await uploads.send(props.draft.key, props.draft.revision, account.value.id, confirmed.value); }
 finally { confirmed.value = false; await refreshReceipt(); }
}); }
function clearAttempt() { void run(async () => {
 await uploads.clear(props.draft.key, account.value.id, props.draft.videoHash, reconciled.value); reconciled.value = false; await refreshReceipt();
}); }
function compose() { void run(async () => {
 await props.flush(); await nextTick(); await uploads.openX(props.draft.key,props.draft.revision);
}); }
</script>
<template>
 <section class="draft-fields" aria-label="Platform upload">
  <template v-if="platform === 'x'">
   <h3>X composer</h3>
   <p class="publication-hint">Open the post text in X, attach the approved MP4 and post from your browser.</p>
   <div class="draft-actions"><UiButton :disabled="!!hint" :busy="busy" @click="compose">Open X composer</UiButton></div>
  </template>
  <template v-else>
   <h3>Platform upload</h3>
   <p class="publication-hint">{{ note }}</p>
   <template v-if="!account.id">
    <label v-if="platform === 'instagram'" class="field-group">Facebook Page access token
     <input v-model="token" type="password" class="ui-field" autocomplete="off" maxlength="8192" spellcheck="false">
    </label>
    <div class="draft-actions"><UiButton :busy="busy" :disabled="platform === 'instagram' && !token.trim()" @click="connect">{{ platform === 'youtube' ? 'Sign in with Google' : platform === 'tiktok' ? 'Sign in with TikTok' : 'Connect account' }}</UiButton></div>
   </template>
   <template v-else>
    <p>Connected to <strong>{{ account.name }}</strong> · {{ account.id }}</p>
    <div class="draft-actions"><UiButton variant="quiet" :disabled="busy" @click="disconnect">Disconnect</UiButton></div>
    <p v-if="receipt.state === 'sent'" role="status">{{ platform === 'instagram' ? 'Reel published' : platform === 'youtube' ? 'Private video uploaded' : 'Sent to TikTok inbox' }} · {{ receipt.reference }}</p>
    <template v-else-if="receipt.state">
     <p role="status">An earlier attempt exists. Check the platform before sending again. {{ receipt.reference }}</p>
     <div class="draft-actions"><UiButton :disabled="busy" @click="run(refreshReceipt)">Refresh receipt</UiButton></div>
     <template v-if="receipt.retryAllowed">
      <label><input v-model="reconciled" type="checkbox"> I checked this account and the upload is absent.</label>
      <div class="draft-actions"><UiButton :disabled="!reconciled || busy" @click="clearAttempt">Allow another attempt</UiButton></div>
     </template>
    </template>
    <p v-else-if="hint" class="publication-hint">{{ hint }}</p>
    <template v-else>
     <label><input v-model="confirmed" type="checkbox"> Send the {{ draft.key.language.toUpperCase() }} video to {{ account.name }}.</label>
     <div class="draft-actions"><UiButton variant="primary" :busy="busy" :disabled="!confirmed" @click="send">{{ busy ? 'Sending…' : action }}</UiButton></div>
    </template>
   </template>
  </template>
  <p v-if="error" role="alert" class="notice error">{{ error }}</p>
 </section>
</template>
