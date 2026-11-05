<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import UiButton from './UiButton.vue';
import { drafts, type Draft, type DraftKey } from '../services/drafts';
import type { Episode, Video } from '../services/publisher';

const props = defineProps<{ episode: Episode; language: string; video: Video; locked: boolean }>();
const region = ref<HTMLElement | null>(null);
const platforms = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram', x: 'X' };
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
let generation = 0;
let attachVideo = false;
let timer: ReturnType<typeof setTimeout> | undefined;
let inFlight: Promise<void> | undefined;
const languageName = computed(() => props.language === 'es' ? 'Spanish' : 'English');
const name = (key: DraftKey) => `${platforms[key.platform as keyof typeof platforms]} · ${key.account}`;
const stale = computed(() => !!form.value?.revision && (form.value.deliveryFingerprint !== props.episode.fingerprint ||
  form.value.videoHash !== props.episode.videos?.find(v => v.id === form.value?.videoId)?.sha256));
const differentVideo = computed(() => form.value?.videoId !== props.video.id);
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
    platform.value = key.platform; account.value = key.account;
    generation = 0; attachVideo = false; dirty.value = false;
    error.value = ''; feedback.value = ''; reloadPrompt.value = false;
  } catch (e) { error.value ||= `Could not open draft. ${String(e)}`; }
  finally { loading.value = false; }
}
function openDestination() {
  if (!account.value.trim()) { error.value = 'Enter an account label first.'; return; }
  void open({ episode: props.episode.id, language: props.language, platform: platform.value, account: account.value.trim() });
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
    saved.value = await drafts.list(props.episode.id, props.language) || [];
    loading.value = false;
    if (saved.value[0]) await open(saved.value[0].key, true);
  } catch (e) { error.value = `Could not load saved destinations. ${String(e)}`; }
  finally { loading.value = false; }
});
onUnmounted(() => clearTimeout(timer));
function focusEditor() {
  region.value?.scrollIntoView({ block: 'start' });
  region.value?.querySelector<HTMLElement>(form.value ? '#draft-title' : '#draft-account')?.focus({ preventScroll: true });
}
defineExpose({ flush, focusEditor });
</script>

<template>
  <section ref="region" class="publication" aria-label="Publication">
    <div class="section-header"><h2>Publication</h2><span class="publication-language">{{ languageName }}</span></div>
    <p class="section-description">Prepare a local draft. Account labels do not connect or publish to an account.</p>
    <fieldset :disabled="locked || loading" class="draft-fields">
      <div v-if="saved.length" class="field-group saved-destinations">
        <label for="saved-destination">Saved destinations</label>
        <select id="saved-destination" class="ui-field" @change="selectSaved"><option value="">Choose a draft…</option><option v-for="(draft, index) in saved" :key="JSON.stringify(draft.key)" :value="index">{{ name(draft.key) }}</option></select>
      </div>
      <form class="destination-row" @submit.prevent="openDestination">
        <div class="field-group"><label for="draft-platform">Platform</label><select id="draft-platform" v-model="platform" class="ui-field"><option v-for="(label, value) in platforms" :value="value" :key="value">{{ label }}</option></select></div>
        <div class="field-group"><label for="draft-account">Account label</label><input id="draft-account" v-model="account" class="ui-field" maxlength="100" placeholder="@catsoup_es" autocomplete="off"></div>
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
    </fieldset>
    <p v-if="feedback" class="ui-status ui-status--success" role="status">{{ feedback }}</p>
    <div v-if="error" class="notice error" role="alert">
      <p>{{ error }}</p>
      <UiButton v-if="form?.revision && !reloadPrompt" variant="quiet" size="compact" :disabled="saving || locked" @click="reloadPrompt = true">Reload saved draft</UiButton>
      <div v-if="reloadPrompt" class="reload-prompt"><p>Replace your unsaved changes with the saved version?</p><UiButton size="compact" :disabled="loading || saving || locked" @click="reloadSaved">Reload and replace edits</UiButton><UiButton size="compact" variant="quiet" @click="reloadPrompt = false">Keep editing</UiButton></div>
    </div>
  </section>
</template>
