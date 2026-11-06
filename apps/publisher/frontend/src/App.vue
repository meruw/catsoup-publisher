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
import ReviewPanel from './components/ReviewPanel.vue';
import { drafts } from './services/drafts';
import { publisher, mediaURL, type Entry, type Episode } from './services/publisher';

const module = ref<'publisher'|'accounts'|'settings'>('publisher');
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
const stage = ref<'review'|'prepare'|'publish'>('review');
const reviewState = ref('pending');
const reviewPanel = ref<InstanceType<typeof ReviewPanel>|null>(null);
const stages = [{value:'review',label:'Review'},{value:'prepare',label:'Prepare'},{value:'publish',label:'Publish'}] as const;
const editor = ref<InstanceType<typeof PublicationDraft> | null>(null);
const navigating = ref(false);
let stopCloseListener: (() => void) | undefined;
const variants = computed(() => (episode.value?.videos || []).filter(v => v.lang === language.value || v.lang === null));
const video = computed(() => variants.value.find(v => v.id === videoID.value));
const source = computed(() => episode.value && video.value ? mediaURL(episode.value.id, video.value.id) : '');
const languageName = (lang: string) => lang === 'es' ? 'Spanish' : 'English';

function chooseLanguage(lang: string) {
  reviewState.value='pending';
  language.value = lang;
  videoID.value = variants.value.find(v => v.kind === 'final' && v.narration)?.id || variants.value[0]?.id || '';
}
async function select(id: string) {
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
  try { await editor.value?.flush(); await reviewPanel.value?.flush(); error.value = ''; await action(); }
  catch (e) { error.value = `Could not leave this step. ${String(e)}`; }
  finally { navigating.value = false; }
}
function switchVideo(event: Event) {
  const select = event.target as HTMLSelectElement;
  const next = select.value; select.value = videoID.value;
  void navigate(() => { videoID.value = next; reviewState.value='pending'; });
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
        <button class="module-item" :aria-current="module==='accounts'?'page':undefined" :disabled="navigating || accountBusy" @click="showAccounts">Accounts</button>
        <button class="module-item" :aria-current="module==='settings'?'page':undefined" :disabled="busy || navigating || accountBusy" @click="showSettings()">Settings</button>
        <button class="module-item" disabled title="Trends is not available yet">Trends<span>Planned</span></button>
      </nav>
      <div class="library-label"><span>Episodes</span><span class="mono">{{ entries.length }}</span></div>
      <nav aria-label="Publisher episodes">
        <button v-for="entry in entries" :key="entry.id" class="episode-button"
          :aria-current="episode?.id === entry.id ? 'true' : undefined"
          :disabled="busy || navigating || accountBusy" @click="navigate(async () => { module='publisher'; await select(entry.id); })">
          <span>{{ entry.title }}</span>
          <small>{{ entry.valid ? 'Available to preview' : 'Needs attention' }}</small>
        </button>
      </nav>
      <p v-if="!entries.length && !busy" class="muted">No deliveries yet</p>
      <div class="sidebar-tools"><button class="console-launcher" :aria-expanded="consoleOpen && !consoleCollapsed" aria-controls="diagnostics-console" aria-keyshortcuts="Control+Shift+D" title="Console (Ctrl+Shift+D)" @click="consoleOpen ? consoleCollapsed=false : openDiagnostics()"><span aria-hidden="true" class="console-symbol">&gt;_</span><span>Console</span><kbd>Ctrl ⇧ D</kbd></button></div>
    </aside>
    <AccountManager v-if="module==='accounts'" :initial-language="accountLanguage" @settings="showSettings" @language="accountLanguage=$event" @busy="accountBusy=$event" @diagnostics="openDiagnostics"/>
    <SettingsPage v-else-if="module==='settings'" :provider="settingsProvider" @select="showSettings" @busy="accountBusy=$event"/>
    <main v-else :aria-busy="busy">
      <header class="page-header">
        <h1>Publisher</h1>
        <div class="page-actions">
          <UiButton :busy="busy" :disabled="navigating" @click="navigate(refresh)">{{ busy ? 'Checking…' : 'Refresh' }}</UiButton>
        </div>
      </header>
      <div v-if="error" class="notice error" role="alert">{{ error }}</div>
      <div v-if="busy" class="notice" role="status">Checking delivery files…</div>
      <section v-if="episode && !busy" class="episode-workspace" aria-label="Selected delivery">
        <div class="episode-heading">
          <div><h2>{{ episode.title }}</h2><p class="mono episode-id">{{ episode.id }}</p></div>
          <span class="mono version">{{ episode.version }}</span>
        </div>
        <div class="approval-strip" :class="{ approved: episode.ready }">
          <span class="status-dot" aria-hidden="true"></span>
          <strong>{{ episode.ready ? 'Approved delivery' : 'Preview only' }}</strong>
          <span>{{ episode.ready ? 'Owner approval is present' : 'Not approved for publishing' }}</span>
        </div>
        <div class="workspace-toolbar">
          <nav class="workflow-steps" aria-label="Episode workflow">
            <button v-for="(step,index) in stages" :key="step.value" :aria-current="stage===step.value?'step':undefined" :disabled="navigating" @click="navigate(()=>{stage=step.value})"><span class="step-number mono">{{index+1}}</span><span>{{step.label}}</span></button>
          </nav>
          <UiSegmented :model-value="language" label="Workspace language" :disabled="navigating" :options="(episode.languages || []).map(lang => ({value:lang,label:lang.toUpperCase(),accessibleLabel:languageName(lang)}))" @update:model-value="lang=>navigate(()=>chooseLanguage(lang))"/>
        </div>
        <div v-show="stage==='review'" class="review-layout">
          <section class="preview-panel" aria-label="Video preview">
            <div class="section-header">
              <h3>Video preview</h3>
            </div>
            <VideoPreview v-if="source" :key="source" :source="source" :language="languageName(language)" :active="stage==='review'" />
            <div class="variant-control">
              <label for="variant">Version</label>
              <select id="variant" :value="videoID" :disabled="navigating" class="ui-field" @change="switchVideo">
                <option v-for="v in variants" :key="v.id" :value="v.id">{{ v.kind === 'final' ? 'Final' : 'Clean' }} · {{ v.narration ? 'with narration' : 'no narration' }}</option>
              </select>
            </div>
            <div v-if="video" class="video-specs mono" aria-label="Video specifications">
              <span>{{ video.measured.resolution }}</span><span>{{ video.measured.fps }} fps</span><span>{{ video.measured.duration_s }} s</span>
            </div>
          </section>
          <aside class="delivery-panel" aria-label="Delivery details">
            <ReviewPanel v-if="video" :key="episode.fingerprint+':'+language+':'+video.id" ref="reviewPanel" :episode="episode" :language="language" :video="video" :locked="navigating" @state="reviewState=$event"/>
            <details class="delivery-check"><summary>Delivery check · Valid</summary>
            <dl>
              <div><dt>Languages</dt><dd>{{ episode.languages?.map(languageName).join(' / ') }}</dd></div>
              <div><dt>Video versions</dt><dd class="mono">{{ episode.videos?.length }}</dd></div>
              <div><dt>Production status</dt><dd class="status-name">{{ episode.status.replaceAll('_', ' ') }}</dd></div>
              <div><dt>Owner approval</dt><dd>{{ episode.ready ? 'Present' : 'Not available' }}</dd></div>
            </dl>
            <p v-if="episode.reason" class="notice">{{ episode.reason }}</p>
            <p v-for="warning in episode.warnings" :key="warning" class="notice warning">{{ warning }}</p>
            </details>
          </aside>
        </div>
        <PublicationDraft v-if="video" v-show="stage!=='review'" :key="episode.id + ':' + language" ref="editor"
          :episode="episode" :language="language" :video="video" :locked="navigating" :stage="stage==='publish'?'publish':'prepare'" :reviewed="reviewState==='reviewed'" @accounts="showAccounts" @prepare="navigate(()=>{stage='prepare'})" />
        <div class="workflow-footer"><UiButton v-if="stage!=='publish'" @click="navigate(()=>{stage=stage==='review'?'prepare':'publish'})">{{stage==='review'?'Continue to Prepare':'Continue to Publish'}}</UiButton><UiButton v-else variant="quiet" @click="navigate(()=>{stage='prepare'})">Back to Prepare</UiButton></div>
      </section>
      <section v-if="!episode && !busy && !error" class="empty">
        <h2>No delivery to preview</h2><p>Add a production delivery, then refresh</p>
      </section>
      <footer><span>CATsoup Publisher</span><span>Your workspace stays on this PC</span></footer>
    </main>
  </div>
</template>
