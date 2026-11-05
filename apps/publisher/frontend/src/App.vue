<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import VideoPreview from './components/VideoPreview.vue';
import UiButton from './components/UiButton.vue';
import UiSegmented from './components/UiSegmented.vue';
import WindowBar from './components/WindowBar.vue';
import { publisher, mediaURL, type Entry, type Episode } from './services/publisher';

const entries = ref<Entry[]>([]);
const episode = ref<Episode | null>(null);
const busy = ref(false);
const error = ref('');
const language = ref('es');
const videoID = ref('');
const variants = computed(() => (episode.value?.videos || []).filter(v => v.lang === language.value || v.lang === null));
const video = computed(() => variants.value.find(v => v.id === videoID.value));
const source = computed(() => episode.value && video.value ? mediaURL(episode.value.id, video.value.id) : '');
const languageName = (lang: string) => lang === 'es' ? 'Spanish' : 'English';

function chooseLanguage(lang: string) {
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
onMounted(refresh);
</script>

<template>
  <WindowBar />
  <div class="workspace">
    <aside class="sidebar" aria-label="Modules and episodes">
      <div class="module-label">Modules</div>
      <nav class="module-nav" aria-label="Modules">
        <button class="module-item" aria-current="page" @click="refresh">Publisher</button>
        <button class="module-item" disabled title="Trends is not available yet">Trends<span>Planned</span></button>
      </nav>
      <div class="library-label"><span>Episodes</span><span class="mono">{{ entries.length }}</span></div>
      <nav aria-label="Publisher episodes">
        <button v-for="entry in entries" :key="entry.id" class="episode-button"
          :aria-current="episode?.id === entry.id ? 'true' : undefined"
          :disabled="busy" @click="select(entry.id)">
          <span>{{ entry.title }}</span>
          <small>{{ entry.valid ? 'Available to preview' : 'Needs attention' }}</small>
        </button>
      </nav>
      <p v-if="!entries.length && !busy" class="muted">No deliveries yet</p>
      <div class="workspace-note"><span>Local workspace</span><small>Spanish + English</small></div>
    </aside>
    <main :aria-busy="busy">
      <header class="page-header">
        <div><h1>Publisher</h1><p>Review deliveries in Spanish and English</p></div>
        <UiButton :busy="busy" @click="refresh">{{ busy ? 'Checking…' : 'Refresh' }}</UiButton>
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
        <div class="review-layout">
          <section class="preview-panel" aria-label="Video preview">
            <div class="section-header">
              <h3>Video preview</h3>
              <UiSegmented :model-value="language" label="Preview language"
                :options="(episode.languages || []).map(lang => ({ value: lang, label: lang.toUpperCase(), accessibleLabel: languageName(lang) }))"
                @update:model-value="chooseLanguage" />
            </div>
            <VideoPreview v-if="source" :key="source" :source="source" :language="languageName(language)" />
            <div class="variant-control">
              <label for="variant">Version</label>
              <select id="variant" v-model="videoID" class="ui-field">
                <option v-for="v in variants" :key="v.id" :value="v.id">{{ v.kind === 'final' ? 'Final' : 'Clean' }} · {{ v.narration ? 'with narration' : 'no narration' }}</option>
              </select>
            </div>
            <div v-if="video" class="video-specs mono" aria-label="Video specifications">
              <span>{{ video.measured.resolution }}</span><span>{{ video.measured.fps }} fps</span><span>{{ video.measured.duration_s }} s</span>
            </div>
          </section>
          <aside class="delivery-panel" aria-label="Delivery details">
            <div class="section-header"><h3>Delivery check</h3><span class="valid-label">Valid</span></div>
            <p class="section-description">Files match the delivery manifest</p>
            <dl>
              <div><dt>Languages</dt><dd>{{ episode.languages?.map(languageName).join(' / ') }}</dd></div>
              <div><dt>Video versions</dt><dd class="mono">{{ episode.videos?.length }}</dd></div>
              <div><dt>Production status</dt><dd class="status-name">{{ episode.status.replaceAll('_', ' ') }}</dd></div>
              <div><dt>Owner approval</dt><dd>{{ episode.ready ? 'Present' : 'Not available' }}</dd></div>
            </dl>
            <section class="review-note">
              <h4>{{ episode.ready ? 'Review notes' : 'Before publishing' }}</h4>
              <p>{{ episode.reason || (episode.ready ? 'Approval is tied to the current delivery files.' : 'Review the video and obtain owner approval for the current files.') }}</p>
            </section>
            <p v-for="warning in episode.warnings" :key="warning" class="notice warning">{{ warning }}</p>
            <p class="read-only-note">Previewing leaves the review status unchanged</p>
          </aside>
        </div>
      </section>
      <section v-if="!episode && !busy && !error" class="empty">
        <h2>No delivery to preview</h2><p>Add a production delivery, then refresh</p>
      </section>
      <footer><span>CATsoup Publisher</span><span>Delivery review · Local only</span></footer>
    </main>
  </div>
</template>
