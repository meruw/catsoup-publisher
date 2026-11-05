<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue';
import UiButton from './UiButton.vue';
import { uploads } from '../services/uploads';
import type { Draft } from '../services/drafts';

const props = defineProps<{ draft: Draft; hint: string; flush: () => Promise<void> }>();
const emit = defineEmits<{ busy: [value: boolean]; accounts: [] }>();
const account = ref({ id: '', name: '' });
const receipt = ref({ state: '', reference: '', retryAllowed: false });
const confirmed = ref(false);
const reconciled = ref(false);
const busy = ref(false);
const error = ref('');
const platform = computed(() => props.draft.key.platform);
const action = computed(() => ({ youtube: 'Upload privately to YouTube', tiktok: 'Send to TikTok inbox', instagram: 'Publish Reel on Instagram' })[platform.value]);
const note = computed(() => ({
 youtube: 'Your video will be uploaded privately.',
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
 try { const profiles=await uploads.accounts() || []; const current=profiles.find(p=>p.platform===platform.value && p.language===props.draft.key.language); if(current){account.value=current;if(platform.value!=='x') await refreshReceipt();} } catch(e) { error.value = String(e); }
});
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
   <p class="publication-hint">Check {{ account.name || 'the active account' }} in the browser, attach the approved MP4 and post.</p>
   <div class="draft-actions"><UiButton :disabled="!!hint" :busy="busy" @click="compose">Open X composer</UiButton></div>
  </template>
  <template v-else>
   <h3>Platform upload</h3>
   <p class="publication-hint">{{ note }}</p>
   <template v-if="!account.id">
    <p>No account assigned for {{ draft.key.language.toUpperCase() }}.</p>
    <div class="draft-actions"><UiButton @click="emit('accounts')">Manage accounts</UiButton></div>
   </template>
   <template v-else>
    <p>{{ draft.key.language.toUpperCase() }} destination: <strong>{{ account.name }}</strong> · {{ account.id }}</p>
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
