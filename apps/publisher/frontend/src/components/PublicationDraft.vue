<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import UiButton from './UiButton.vue';
import {uploads,type Profile} from '../services/uploads';
import DirectUpload from './DirectUpload.vue';
import DestinationQueue from './DestinationQueue.vue';
import { drafts, type Draft, type DraftKey } from '../services/drafts';
import type { Episode, Video } from '../services/publisher';

const props = defineProps<{ episode: Episode; language: string; video: Video; locked: boolean; stage:'prepare'|'publish'; reviewed:boolean }>();
const region = ref<HTMLElement | null>(null);
const platforms = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram', x: 'X' };
const emit=defineEmits<{accounts:[];prepare:[]}>();
const profiles=ref<Profile[]>([]);
const suggestedAccount=computed(()=>profiles.value.find(p=>p.platform===platform.value && p.language===props.language)?.name || '');
const platform = ref('youtube');
const account = ref('');
const saved = ref<Draft[]>([]);
const form = ref<Draft | null>(null);
const loading = ref(true);
const saving = ref(false);
const dirty = ref(false);
const error = ref('');
const feedback = ref('');
const reloadPrompt = ref(false);
const accountConfirmed = ref(false);
const preparingUpload = ref(false);
const platformBusy = ref(false);
const uploadError = ref('');
let generation = 0;
let attachVideo = false;
let timer: ReturnType<typeof setTimeout> | undefined;
let inFlight: Promise<void> | undefined;
const languageName = computed(() => props.language === 'es' ? 'Spanish' : 'English');
const name = (key: DraftKey) => `${platforms[key.platform as keyof typeof platforms]} · ${key.account}`;
const stale = computed(() => !!form.value?.revision && (form.value.deliveryFingerprint !== props.episode.fingerprint ||
  form.value.videoHash !== props.episode.videos?.find(v => v.id === form.value?.videoId)?.sha256));
const differentVideo = computed(() => form.value?.videoId !== props.video.id);
const uploadHint = computed(() => {
  if (!props.episode.ready) return 'Available after the delivery is approved for publishing.';
  if (!props.reviewed) return 'Mark the current final video as reviewed in Review before sending.';
  if (!form.value?.revision) return 'Save this draft to prepare an upload.';
  if (stale.value || differentVideo.value || props.video.kind !== 'final' || props.video.lang !== props.language)
    return 'Select and link the current final video for this language.';
  return '';
});
const status = computed(() => saving.value ? 'Saving…' : dirty.value ? 'Unsaved changes' : form.value?.revision ? 'Saved locally' : 'Not saved yet');

function edited() {
  generation++;
  dirty.value = true;
  feedback.value = '';
  clearTimeout(timer);
  timer = setTimeout(() => { void flush().catch(() => {}); }, 600);
}
function updateList(draft: Draft) {
  saved.value = [draft, ...saved.value.filter(d => JSON.stringify(d.key) !== JSON.stringify(draft.key))];
}
// Serialize writes; a response updates the revision, never text typed while saving.
function flush(): Promise<void> {
  clearTimeout(timer);
  if (inFlight) return inFlight;
  if (!dirty.value || !form.value) return Promise.resolve();
  inFlight = (async () => {
    await Promise.resolve();
    saving.value = true;
    error.value = '';
    try {
      while (dirty.value && form.value) {
        const snapshot = form.value;
        const sentGeneration = generation;
        const result = await drafts.save({
          key: snapshot.key, title: snapshot.title, description: snapshot.description,
          hashtags: snapshot.hashtags, revision: snapshot.revision,
          videoId: attachVideo ? props.video.id : snapshot.videoId, attachVideo,
        });
        form.value = { ...form.value, revision: result.revision, updatedAt: result.updatedAt,
          videoId: result.videoId, videoHash: result.videoHash, deliveryFingerprint: result.deliveryFingerprint };
        attachVideo = false;
        dirty.value = sentGeneration !== generation;
        updateList(result);
      }
    } catch (e) {
      error.value = `Could not save. Your edits are still here. ${String(e)}`;
      throw e;
    } finally { saving.value = false; inFlight = undefined; }
  })();
  return inFlight;
}
async function open(key: DraftKey, initial = false) {
  if (loading.value || (props.locked && !initial)) return;
  loading.value = true;
  try {
    await flush();
    const result = await drafts.get(key);
    if (!result.revision) {
      result.videoId = props.video.id;
      result.videoHash = props.video.sha256;
      result.deliveryFingerprint = props.episode.fingerprint;
    }
    form.value = result;
    accountConfirmed.value = false; uploadError.value = '';
    platform.value = key.platform; account.value = key.account;
    generation = 0; attachVideo = false; dirty.value = false;
    error.value = ''; feedback.value = ''; reloadPrompt.value = false;
  } catch (e) { error.value ||= `Could not open draft. ${String(e)}`; }
  finally { loading.value = false; }
}
function openDestination() {
  const label=account.value.trim() || suggestedAccount.value;
  if (!label) { error.value = 'Enter an account label first.'; return; }
  void open({ episode: props.episode.id, language: props.language, platform: platform.value, account: label });
}
function selectSaved(event: Event) {
  const select = event.target as HTMLSelectElement;
  const draft = saved.value[Number(select.value)];
  select.value = '';
  if (draft) void open(draft.key);
}
async function saveNow() {
  if (form.value && !form.value.revision && !dirty.value) edited();
  try { await flush(); } catch { /* Error and unsaved state remain visible. */ }
}
async function useSelectedVideo() {
  attachVideo = true; edited(); await saveNow();
}
async function copy(part: string) {
  if (!form.value) return;
  try { await flush(); await drafts.copy(form.value.key, part); feedback.value = part === 'title' ? 'Title copied' : 'Caption copied'; }
  catch (e) { error.value ||= `Could not copy. ${String(e)}`; }
}
async function reveal() {
  try { await drafts.reveal(props.episode.id, props.video.id); feedback.value = 'Video selected in file manager'; }
  catch (e) { error.value = `Could not open the folder. ${String(e)}`; }
}
async function copyUploadPath() {
  if (!form.value || !accountConfirmed.value || uploadHint.value || preparingUpload.value) return;
  preparingUpload.value = true;
  feedback.value = ''; uploadError.value = '';
  try {
    await flush();
    await drafts.copyUploadPath(form.value.key, form.value.revision, accountConfirmed.value);
    feedback.value = 'Video path copied. Paste it into the platform’s file picker.';
  } catch (e) { uploadError.value = String(e); }
  finally { preparingUpload.value = false; accountConfirmed.value = false; }
}
async function reloadSaved() {
  if (!form.value) return;
  loading.value = true; clearTimeout(timer);
  try {
    const result = await drafts.get(form.value.key);
    if (!result.revision) throw new Error('No saved version exists. Your edits were retained.');
    form.value = result; dirty.value = false; attachVideo = false; generation = 0;
    error.value = ''; reloadPrompt.value = false; updateList(result);
  } catch (e) { error.value = String(e); }
  finally { loading.value = false; }
}
onMounted(async () => {
  try {
    profiles.value = await uploads.accounts() || [];
    saved.value = await drafts.list(props.episode.id, props.language) || [];
    loading.value = false;
    if (saved.value[0]) await open(saved.value[0].key, true);
    else { const p=profiles.value.find(p=>p.language===props.language && p.platform==='youtube') || profiles.value.find(p=>p.language===props.language); if(p) await open({episode:props.episode.id,language:props.language,platform:p.platform,account:p.name},true); }
  } catch (e) { error.value = `Could not load saved destinations. ${String(e)}`; }
  finally { loading.value = false; }
});
onUnmounted(() => clearTimeout(timer));
function focusEditor() {
  region.value?.scrollIntoView({ block: 'start' });
  region.value?.querySelector<HTMLElement>(form.value ? '#draft-title' : '#draft-account')?.focus({ preventScroll: true });
}
async function beforeLeaving() {
  if (platformBusy.value) throw new Error('Wait for the platform request to finish.');
  await flush();
}
defineExpose({ flush: beforeLeaving, focusEditor });
</script>

<template>
  <section ref="region" class="publication" aria-label="Publication">
    <div class="section-header"><h2>{{stage==='prepare'?'Prepare your post':'Publish to a destination'}}</h2><span class="publication-language">{{ languageName }}</span></div>
    <fieldset :disabled="locked || loading || preparingUpload || platformBusy" class="draft-fields">
      <DestinationQueue v-if="stage==='publish' && saved.length" :key="saved.map(d=>d.revision).join(':')+':'+platformBusy" :drafts="saved" :selected="form?JSON.stringify(form.key):''" :locked="locked || loading || platformBusy" @select="d=>open(d.key)"/>
      <div v-if="stage==='prepare' && saved.length" class="field-group saved-destinations">
        <label for="saved-destination">Saved destinations</label>
        <select id="saved-destination" class="ui-field" @change="selectSaved"><option value="">Choose a draft…</option><option v-for="(draft, index) in saved" :key="JSON.stringify(draft.key)" :value="index">{{ name(draft.key) }}</option></select>
      </div>
      <form v-show="stage==='prepare'" class="destination-row" @submit.prevent="openDestination">
        <div class="field-group"><label for="draft-platform">Platform</label><select id="draft-platform" v-model="platform" @change="account=''" class="ui-field"><option v-for="(label, value) in platforms" :value="value" :key="value">{{ label }}</option></select></div>
        <div class="field-group"><label for="draft-account">Account label</label><input id="draft-account" v-model="account" class="ui-field" maxlength="100" :placeholder="suggestedAccount || (language==='es'?'@catsoup_es':'@catsoup_en')" autocomplete="off"></div>
        <UiButton type="submit" :busy="loading">Open draft</UiButton>
      </form>
      <template v-if="form">
        <div class="draft-heading"><h3>{{ name(form.key) }} · {{ languageName }}</h3><span role="status" :class="{ warning: dirty }">{{ status }}</span></div>
        <div class="draft-video">
          <span>Linked video: <span class="mono">{{ form.videoId }}</span></span>
          <UiButton v-if="differentVideo || stale" size="compact" :disabled="saving" @click="useSelectedVideo">Use selected video</UiButton>
        </div>
        <p v-if="stale" class="notice warning">The delivery changed since this draft was linked. Review the current video before using it. Your text is preserved.</p>
        <p v-else-if="differentVideo" class="notice warning">The preview shows a different version. Choose “Use selected video” to link it to this draft.</p>
        <template v-if="stage==='prepare'">
        <div class="field-group"><label for="draft-title">Title</label><input id="draft-title" v-model="form.title" class="ui-field" maxlength="500" @input="edited"></div>
        <div class="field-group"><label for="draft-description">Description</label><textarea id="draft-description" v-model="form.description" class="ui-field" rows="4" maxlength="10000" @input="edited"></textarea></div>
        <div class="field-group"><label for="draft-hashtags">Hashtags</label><textarea id="draft-hashtags" v-model="form.hashtags" class="ui-field" rows="2" maxlength="2000" placeholder="#catsoup" @input="edited"></textarea></div>
        <div class="draft-actions">
          <UiButton variant="primary" :busy="saving" :disabled="!dirty && !!form.revision" @click="saveNow">Save draft</UiButton>
          <UiButton :disabled="saving || !form.title.trim()" @click="copy('title')">Copy title</UiButton>
          <UiButton :disabled="saving || !(form.description.trim() || form.hashtags.trim())" @click="copy('caption')">Copy caption</UiButton>
          <UiButton variant="quiet" @click="reveal">Open video folder</UiButton>
        </div>
        <p class="publication-hint">Changes save after you pause typing. Caption copies description and hashtags. Check platform limits before uploading.</p>
        </template>
        <template v-else>
        <section class="post-proof" aria-label="Post preview"><div class="section-header"><h3>Saved text</h3><UiButton variant="quiet" size="compact" @click="emit('prepare')">Edit in Prepare</UiButton></div><strong>{{form.title || 'No title'}}</strong><p>{{form.description}}</p><p class="post-tags">{{form.hashtags}}</p></section>
        <DirectUpload :key="JSON.stringify([form.key, form.videoHash])" :draft="form" :hint="uploadHint" :flush="flush" @busy="platformBusy = $event" @accounts="emit('accounts')" />
        <details v-if="form.key.platform!=='x'" class="manual-handoff">
        <summary>Finish in the browser</summary>
        <div class="draft-actions"><UiButton @click="copy('title')">Copy title</UiButton><UiButton @click="copy('caption')">Copy caption</UiButton></div>
        <p v-if="uploadHint" class="publication-hint">{{ uploadHint }}</p>
        <template v-else>
          <label><input v-model="accountConfirmed" type="checkbox"> I checked that {{ name(form.key) }} is open in my browser.</label>
          <div class="draft-actions"><UiButton :disabled="!accountConfirmed || saving" :busy="preparingUpload" @click="copyUploadPath">Copy upload file path</UiButton></div>
        </template>
        <p v-if="uploadError" class="notice error" role="alert">{{ uploadError }}</p>
        </details>
        </template>
      </template>
      <p v-else-if="stage==='publish'" class="notice">Prepare a local draft for a destination first. <UiButton variant="quiet" @click="emit('prepare')">Go to Prepare</UiButton></p>
    </fieldset>
    <p v-if="feedback" class="ui-status ui-status--success" role="status">{{ feedback }}</p>
    <div v-if="error" class="notice error" role="alert">
      <p>{{ error }}</p>
      <UiButton v-if="form?.revision && !reloadPrompt" variant="quiet" size="compact" :disabled="saving || locked" @click="reloadPrompt = true">Reload saved draft</UiButton>
      <div v-if="reloadPrompt" class="reload-prompt"><p>Replace your unsaved changes with the saved version?</p><UiButton size="compact" :disabled="loading || saving || locked" @click="reloadSaved">Reload and replace edits</UiButton><UiButton size="compact" variant="quiet" @click="reloadPrompt = false">Keep editing</UiButton></div>
    </div>
  </section>
</template>
