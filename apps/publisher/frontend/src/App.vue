<script setup lang="ts">
import {tikTokWeb} from './services/tiktokWeb';
import UiDialog from './components/UiDialog.vue';
import {activeTransfers,transferJobs,transferUpdates} from './services/transfers';
import TransferActivity from './components/TransferActivity.vue';
import EpisodeBrowser from './components/EpisodeBrowser.vue';
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, watch } from 'vue';
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
import {pendingDeliveries} from './services/completion';
import type { Completion, PostStatus } from './services/drafts';
import { drafts } from './services/drafts';
import { publisher, mediaURL, type Entry, type Episode } from './services/publisher';

const module = ref<'publisher'|'posts'|'accounts'|'settings'>('publisher');
const browserCloseWarning=ref(false);
async function finishClose(){browserCloseWarning.value=false;if(module.value==='accounts')await drafts.confirmClose();else await navigate(async()=>{await drafts.confirmClose()});}
const initialPlatform=ref('');
const initialPublish=ref(false),postsEpisode=ref('');
const PostOverview=defineAsyncComponent(()=>import('./components/PostOverview.vue'));
async function openPost(post:PostStatus,reviewUpload=false){await navigate(async()=>{const accounts=await uploads.accounts()||[];if(!accounts.some(p=>p.platform===post.key.platform&&p.language===post.key.language&&'account:'+p.id===post.key.account))throw Error('This account changed; reconnect it in Accounts');await select(post.key.episode);chooseLanguage(post.key.language);if(variants.value.some(v=>v.id===post.videoId))videoID.value=post.videoId;initialPlatform.value=post.key.platform;initialPublish.value=reviewUpload;module.value='publisher';});}
function showPosts(id=''){void navigate(()=>{postsEpisode.value=id;uploadsOpen.value=false;module.value='posts'});}
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
const uploadsOpen=ref(false);
const accountBusy = ref(false);
const accountLanguage = ref('es');
const settingsProvider = ref<SetupPlatform|null>(null);
function showAccounts() { void navigate(() => { if(module.value==='publisher') accountLanguage.value=language.value;module.value='accounts'; }); }
function showSettings(provider?:SetupPlatform, lang?:string) {
  void navigate(() => { settingsProvider.value=provider ?? null; if(lang) accountLanguage.value=lang; module.value='settings'; });
}
const entries = ref<Entry[]>([]);
const completed=ref<Completion[]>([]);
const pendingEntries=computed(()=>pendingDeliveries(entries.value,completed.value));
let completionRequest=0;
async function refreshCompletion(){
 const request=++completionRequest;
 try{const result=await drafts.completedLanguages();if(request===completionRequest)completed.value=result||[];}
 catch{if(request===completionRequest)completed.value=[];}
}
watch(transferUpdates,()=>{void refreshCompletion()});
watch(accountBusy,busy=>{if(!busy)void refreshCompletion()});
const episode = ref<Episode | null>(null);
const busy = ref(false);
const importing = ref(false);
const importMessage = ref('');
const importError = ref(false);
async function importDelivery() {
  importing.value=true; busy.value=true; importMessage.value=''; importError.value=false;
  try {
    const result=await publisher.importDelivery();
    if(!result) { importMessage.value=''; return; }
    entries.value=await publisher.list() || [];
    await refreshCompletion();
    module.value='publisher';
    await select(result.id);
    importMessage.value=result.existing ? 'Already imported' : 'Delivery imported';
  } catch(e) { importError.value=true; importMessage.value=String(e); }
  finally { importing.value=false; busy.value=false; }
}
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
  initialPublish.value=false;
  stage.value='prepare';
  language.value = lang;
  videoID.value = variants.value.find(v => v.kind === 'final' && v.narration)?.id || variants.value[0]?.id || '';
}
async function select(id: string) {
  initialPlatform.value='';
  initialPublish.value=false;
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
    await refreshCompletion();
    const target = pendingEntries.value.find(e => e.id === episode.value?.id) || pendingEntries.value.find(e => e.valid);
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
  stopCloseListener = drafts.onCloseRequested(async () => {
    if(activeTransfers.value){uploadsOpen.value=true;return;}
    try{if(await tikTokWeb.hasPreparedBrowser()){browserCloseWarning.value=true;return;}await finishClose();}
    catch{error.value='Could not check browser work. Finish in TikTok before closing Publisher.';}
  });
  try { await drafts.enableCloseGuard(); }
  catch { error.value = 'Close protection is unavailable. Save your draft before closing.'; }
  await refresh();
});
onUnmounted(() => {stopCloseListener?.();window.removeEventListener('keydown',diagnosticShortcut);});
</script>

<template>
  <WindowBar />
  <UiDialog v-if="browserCloseWarning" title="TikTok is still open" @close="browserCloseWarning=false"><p>Closing Publisher also closes its TikTok browser. Finish publishing there first. Its upload record will be kept to prevent another attachment.</p><div class="draft-actions"><UiButton @click="browserCloseWarning=false">Keep open</UiButton><UiButton variant="quiet" @click="finishClose">Close Publisher</UiButton></div></UiDialog>
  <TransferActivity v-if="uploadsOpen" :open="uploadsOpen" @close="uploadsOpen=false" @posts="showPosts"/>
  <DiagnosticsConsole v-if="consoleOpen" v-model:collapsed="consoleCollapsed" :service="diagnosticFilter.service" :language="diagnosticFilter.language" @close="consoleOpen=false"/>
  <div class="workspace">
    <aside class="sidebar" aria-label="Modules and episodes">
      <div class="module-label">Modules</div>
      <nav class="module-nav" aria-label="Modules">
        <button class="module-item" :aria-current="module==='publisher'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="navigate(async () => { module='publisher'; await refresh(); })">Publisher</button>
        <button class="module-item" :aria-current="module==='posts'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="showPosts()">Posts</button>
        <button class="module-item" :aria-current="module==='accounts'?'page':undefined" :disabled="navigating || accountBusy" @click="showAccounts">Accounts</button>
        <button class="module-item" :aria-current="module==='settings'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="showSettings()">Settings</button>
      </nav>
      <div class="library-label"><span>Pending episodes</span><span class="mono">{{ pendingEntries.length }}</span></div>
      <UiButton class="import-delivery" variant="secondary" :busy="importing" :disabled="busy || navigating || accountBusy || transferBusy" @click="navigate(importDelivery)">{{ importing ? 'Importing…' : 'Import delivery' }}</UiButton>
      <p v-if="importMessage" class="import-feedback" :class="{'import-error':importError}" :role="importError ? 'alert' : 'status'">{{ importMessage }}</p>
      <EpisodeBrowser :entries="pendingEntries" :completed="entries.length-pendingEntries.length" :selected="episode?.id" :disabled="busy || navigating || accountBusy" @select="id=>navigate(async()=>{module='publisher';await select(id)})"/>
      <div class="sidebar-tools"><button id="uploads-launcher" class="console-launcher" :aria-expanded="uploadsOpen" @click="uploadsOpen=true"><span>Uploads</span><span v-if="activeTransfers" role="status">{{activeTransfers}} active</span><span v-else-if="transferJobs.some(j=>j.state==='attention')">Needs attention</span><span v-else-if="transferJobs.length">Finished</span></button><button class="console-launcher" :aria-expanded="consoleOpen && !consoleCollapsed" aria-controls="diagnostics-console" aria-keyshortcuts="Control+Shift+D" title="Console (Ctrl+Shift+D)" @click="consoleOpen ? consoleCollapsed=false : openDiagnostics()"><span aria-hidden="true" class="console-symbol">&gt;_</span><span>Console</span><kbd>Ctrl ⇧ D</kbd></button></div>
    </aside>
    <AccountManager v-if="module==='accounts'" :initial-language="accountLanguage" @settings="showSettings" @language="accountLanguage=$event" @busy="accountBusy=$event" @diagnostics="openDiagnostics"/>
    <SettingsPage v-else-if="module==='settings'" :provider="settingsProvider" @select="showSettings" @busy="accountBusy=$event"/>
    <PostOverview v-else-if="module==='posts'" :entries="entries" :initial-episode="postsEpisode" :open-error="error" @open="openPost"/>
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
      <div v-if="busy" class="notice" role="status">{{ importing ? 'Importing and checking delivery files…' : 'Checking delivery files…' }}</div>
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
            :initial-platform="initialPlatform" :initial-publish="initialPublish" :episode="episode" :language="language" :video="video" :locked="navigating" :stage="stage" @busy="transferBusy=$event" @accounts="showAccounts" @prepare="navigate(()=>{stage='prepare'})" @publish="navigate(()=>{stage='publish'})" />
        </div>
      </section>
      <section v-if="!episode && !busy && !error" class="empty">
        <template v-if="entries.length&&!pendingEntries.length"><h2>All deliveries are complete</h2><UiButton variant="quiet" @click="showPosts()">View Posts</UiButton></template><template v-else><h2>No delivery to preview</h2><p>Choose Import delivery and select the folder containing manifest.json</p></template>
      </section>
    </main>
  </div>
</template>
