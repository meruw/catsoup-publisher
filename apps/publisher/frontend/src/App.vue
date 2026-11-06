<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref } from 'vue';
import VideoPreview from './components/VideoPreview.vue';
import UiButton from './components/UiButton.vue';
import UiSegmented from './components/UiSegmented.vue';
import WindowBar from './components/WindowBar.vue';
import SettingsPage from './components/SettingsPage.vue';
import type { SetupPlatform } from './services/uploads';
import AccountManager from './components/AccountManager.vue';
import PublicationDraft from './components/PublicationDraft.vue';
import UiTooltip from './components/UiTooltip.vue';
import UiIcon from './components/UiIcon.vue';
import { uploads } from './services/uploads';
import type { PostStatus } from './services/drafts';
import { drafts } from './services/drafts';
import { publisher, mediaURL, type Entry, type Episode } from './services/publisher';

const module = ref<'publisher'|'posts'|'accounts'|'settings'>('publisher');
const initialPlatform=ref('');
const PostOverview=defineAsyncComponent(()=>import('./components/PostOverview.vue'));
async function openPost(post:PostStatus){await navigate(async()=>{const accounts=await uploads.accounts()||[];if(!accounts.some(p=>p.platform===post.key.platform&&p.language===post.key.language&&'account:'+p.id===post.key.account))throw Error('This account changed; reconnect it in Accounts');await select(post.key.episode);chooseLanguage(post.key.language);if(variants.value.some(v=>v.id===post.videoId))videoID.value=post.videoId;initialPlatform.value=post.key.platform;module.value='publisher';});}
const DiagnosticsConsole = defineAsyncComponent(() => import('./components/DiagnosticsConsole.vue'));
const consoleOpen = ref(false);
const consoleCollapsed = ref(false);
const diagnosticFilter = ref({service:'',language:''});
function openDiagnostics(service='', language='') {
  diagnosticFilter.value={service,language};
  consoleOpen.value=true;
  consoleCollapsed.value=false;
}
function diagnosticShortcut(event:KeyboardEvent) {
  if(event.ctrlKey && event.shiftKey && event.key.toLowerCase()==='d') {
    event.preventDefault();
    if(consoleOpen.value && consoleCollapsed.value) consoleCollapsed.value=false;
    else if(consoleOpen.value) consoleOpen.value=false;
    else openDiagnostics();
  }
}
const accountBusy = ref(false);
const accountLanguage = ref('es');
const settingsProvider = ref<SetupPlatform|null>(null);
function showAccounts() { void navigate(() => { if(module.value==='publisher') accountLanguage.value=language.value;module.value='accounts'; }); }
function showSettings(provider?:SetupPlatform, lang?:string) {
  void navigate(() => { settingsProvider.value=provider ?? null; if(lang) accountLanguage.value=lang; module.value='settings'; });
}
const entries = ref<Entry[]>([]);
const episode = ref<Episode | null>(null);
const busy = ref(false);
const error = ref('');
const language = ref('es');
const videoID = ref('');
const stage = ref<'prepare'|'publish'>('prepare');
const transferBusy = ref(false);
const editor = ref<InstanceType<typeof PublicationDraft> | null>(null);
const navigating = ref(false);
let stopCloseListener: (() => void) | undefined;
const variants = computed(() => (episode.value?.videos || []).filter(v => v.lang === language.value || v.lang === null));
const video = computed(() => variants.value.find(v => v.id === videoID.value));
const source = computed(() => episode.value && video.value ? mediaURL(episode.value.id, video.value.id) : '');
const languageName = (lang: string) => lang === 'es' ? 'Spanish' : 'English';

function chooseLanguage(lang: string) {
  stage.value='prepare';
  language.value = lang;
  videoID.value = variants.value.find(v => v.kind === 'final' && v.narration)?.id || variants.value[0]?.id || '';
}
async function select(id: string) {
  initialPlatform.value='';
  busy.value = true;
  error.value = '';
  episode.value = null;
  try {
    episode.value = await publisher.get(id);
    chooseLanguage(episode.value.languages?.includes(language.value) ? language.value : episode.value.languages?.[0] || 'es');
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
async function refresh() {
  if (busy.value) return;
  busy.value = true;
  error.value = '';
  try {
    entries.value = await publisher.list() || [];
    const target = entries.value.find(e => e.id === episode.value?.id) || entries.value.find(e => e.valid);
    if (target) await select(target.id);
    else episode.value = null;
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
async function navigate(action: () => void | Promise<void>) {
  if (navigating.value || busy.value || accountBusy.value) return;
  navigating.value = true;
  try { await editor.value?.flush(); error.value = ''; await action(); }
  catch (e) { error.value = `Could not leave this view. ${String(e)}`; }
  finally { navigating.value = false; }
}
function switchVideo(event: Event) {
  const select = event.target as HTMLSelectElement;
  const next = select.value; select.value = videoID.value;
  void navigate(() => { videoID.value = next; });
}

onMounted(async () => {
  window.addEventListener('keydown',diagnosticShortcut);
  stopCloseListener = drafts.onCloseRequested(() => {
    if (module.value==='accounts') void drafts.confirmClose();
    else void navigate(async () => { await drafts.confirmClose(); });
  });
  try { await drafts.enableCloseGuard(); }
  catch { error.value = 'Close protection is unavailable. Save your draft before closing.'; }
  await refresh();
});
onUnmounted(() => {stopCloseListener?.();window.removeEventListener('keydown',diagnosticShortcut);});
</script>

<template>
  <WindowBar />
  <DiagnosticsConsole v-if="consoleOpen" v-model:collapsed="consoleCollapsed" :service="diagnosticFilter.service" :language="diagnosticFilter.language" @close="consoleOpen=false"/>
  <div class="workspace">
    <aside class="sidebar" aria-label="Modules and episodes">
      <div class="module-label">Modules</div>
      <nav class="module-nav" aria-label="Modules">
        <button class="module-item" :aria-current="module==='publisher'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="navigate(async () => { module='publisher'; await refresh(); })">Publisher</button>
        <button class="module-item" :aria-current="module==='posts'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="navigate(()=>{module='posts'})">Posts</button>
        <button class="module-item" :aria-current="module==='accounts'?'page':undefined" :disabled="navigating || accountBusy" @click="showAccounts">Accounts</button>
        <button class="module-item" :aria-current="module==='settings'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="showSettings()">Settings</button>
      </nav>
      <div class="library-label"><span>Episodes</span><span class="mono">{{ entries.length }}</span></div>
      <nav aria-label="Publisher episodes">
        <button v-for="entry in entries" :key="entry.id" class="episode-button"
          :aria-current="episode?.id === entry.id ? 'true' : undefined"
          :disabled="busy || navigating || accountBusy" @click="navigate(async () => { module='publisher'; await select(entry.id); })">
          <span>{{ entry.title }}</span>
          <small v-if="!entry.valid">Needs attention</small>
        </button>
      </nav>
      <p v-if="!entries.length && !busy" class="muted">No deliveries yet</p>
      <div class="sidebar-tools"><button class="console-launcher" :aria-expanded="consoleOpen && !consoleCollapsed" aria-controls="diagnostics-console" aria-keyshortcuts="Control+Shift+D" title="Console (Ctrl+Shift+D)" @click="consoleOpen ? consoleCollapsed=false : openDiagnostics()"><span aria-hidden="true" class="console-symbol">&gt;_</span><span>Console</span><kbd>Ctrl ⇧ D</kbd></button></div>
    </aside>
    <AccountManager v-if="module==='accounts'" :initial-language="accountLanguage" @settings="showSettings" @language="accountLanguage=$event" @busy="accountBusy=$event" @diagnostics="openDiagnostics"/>
    <SettingsPage v-else-if="module==='settings'" :provider="settingsProvider" @select="showSettings" @busy="accountBusy=$event"/>
    <PostOverview v-else-if="module==='posts'" :entries="entries" :open-error="error" @open="openPost"/>
    <main v-else class="publisher-main" :aria-busy="busy">
      <header class="page-header">
        <h1>{{ episode?.title || 'Publisher' }}</h1>
        <div class="page-actions">

          <UiSegmented v-if="episode" :model-value="language" label="Workspace language" :disabled="navigating" :options="(episode.languages || []).map(lang => ({value:lang,label:lang.toUpperCase(),accessibleLabel:languageName(lang)}))" @update:model-value="lang=>navigate(()=>chooseLanguage(lang))"/>
          <UiTooltip text="Refresh delivery files" :disabled="navigating"><UiButton variant="quiet" :busy="busy" :disabled="navigating" aria-label="Refresh" @click="navigate(refresh)"><UiIcon name="refresh"/></UiButton></UiTooltip>
          <UiTooltip v-if="episode" class="batch-header-action" :disabled="navigating || transferBusy || !editor || !!editor.batchBlocker" :text="editor?.batchBlocker || 'Confirm delivery to all connected accounts in this language, using each post’s saved mode.'"><UiButton id="publish-all" variant="primary" :disabled="navigating || transferBusy || !editor || !!editor.batchBlocker" @click="editor?.publishAll()">Publish all…</UiButton></UiTooltip>
        </div>
      </header>
      <div v-if="error" class="notice error" role="alert">{{ error }}</div>
      <div v-if="busy" class="notice" role="status">Checking delivery files…</div>
      <section v-if="episode && !busy" class="episode-workspace" aria-label="Selected delivery">
        <div class="editor-layout">
          <section class="preview-panel" aria-label="Video preview">
            <VideoPreview v-if="source" :key="source" :source="source" :language="languageName(language)" :active="stage==='prepare'" />
            <div class="variant-control">
              <label for="variant">Version</label>
              <select id="variant" :value="videoID" :disabled="navigating" class="ui-field" @change="switchVideo">
                <option v-for="v in variants" :key="v.id" :value="v.id">{{ v.kind === 'final' ? 'Final' : 'Clean' }} · {{ v.narration ? 'with narration' : 'no narration' }}</option>
              </select>
            </div>
            <UiTooltip v-if="!episode.ready" class="production-lock" :disabled="true" :text="(episode.reason ? episode.reason+' ' : '')+'This delivery is not authorized for upload. Approve it in the production project and refresh Publisher. Test deliveries remain preview-only.'"><span>Production approval missing</span></UiTooltip>
            <details class="delivery-check"><summary>Video details</summary>
            <div v-if="video" class="video-specs mono" aria-label="Video specifications">
              <span>{{ video.measured.resolution }}</span><span>{{ video.measured.fps }} fps</span><span>{{ video.measured.duration_s }} s</span>
            </div>
            <dl class="video-facts">
              <div><dt>Files</dt><dd>Verified</dd></div>
              <div><dt>Publishing</dt><dd>{{ episode.ready ? 'Approved' : 'Not approved' }}</dd></div>
            </dl>
            <p v-if="episode.reason" class="notice">{{ episode.reason }}</p>
            <p v-for="warning in episode.warnings" :key="warning" class="notice warning">{{ warning }}</p>
            </details>
          </section>
          <PublicationDraft v-if="video" :key="episode.id + ':' + language" ref="editor"
            :initial-platform="initialPlatform" :episode="episode" :language="language" :video="video" :locked="navigating" :stage="stage" @busy="transferBusy=$event" @accounts="showAccounts" @prepare="navigate(()=>{stage='prepare'})" @publish="navigate(()=>{stage='publish'})" />
        </div>
      </section>
      <section v-if="!episode && !busy && !error" class="empty">
        <h2>No delivery to preview</h2><p>Add a production delivery, then refresh</p>
      </section>
    </main>
  </div>
</template>
