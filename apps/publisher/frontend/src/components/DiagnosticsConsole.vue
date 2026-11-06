<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { diagnostics, type DiagnosticEvent, type DiagnosticQuery } from '../services/diagnostics';
import UiButton from './UiButton.vue';

const props = defineProps<{ service?: string; language?: string }>();
const emit = defineEmits<{ close: [] }>();
const search = ref('');
const service = ref(props.service || '');
const language = ref(props.language || '');
const level = ref('');
const attempt = ref('');
const currentSession = ref(false);
const live = ref(true);
const before = ref(0);
const pages = ref<number[]>([]);
const rows = ref<DiagnosticEvent[]>([]);
const selected = ref<DiagnosticEvent | null>(null);
const more = ref(false);
const warning = ref('');
const liveWarning = ref('');
const error = ref('');
const feedback = ref('');
const pending = ref(false);
const actionPending = ref(false);
const height = ref(440);
const expanded = ref(false);
const input = ref<HTMLInputElement | null>(null);
const query = computed<DiagnosticQuery>(() => ({
  search: search.value, service: service.value, language: language.value,
  level: level.value, attempt: attempt.value, currentSession: currentSession.value, before: before.value,
}));
let sequence = 0;
let disposed = false;
let debounce: ReturnType<typeof setTimeout> | undefined;
let stop: (() => void) | undefined;
let previousFocus: HTMLElement | null = null;
let changedDuringRead = false;
let dragStart: { y: number; height: number } | null = null;
const serviceName: Record<string, string> = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram', x: 'X' };
const time = (value: string) => new Date(value).toLocaleTimeString([], { hour12: false });

async function refresh() {
  const request = ++sequence;
  pending.value = true;
  try {
    const result = await diagnostics.query(query.value);
    if (request !== sequence || disposed) return;
    rows.value = result.events || [];
    more.value = result.more;
    warning.value = result.warning;
    error.value = '';
  } catch {
    if (request === sequence && !disposed) error.value = 'Could not read diagnostics. Try Refresh.';
  } finally {
    if (request === sequence && !disposed) pending.value = false;
    if (request === sequence && !disposed && changedDuringRead && live.value) {
      changedDuringRead = false;
      void refresh();
    }
  }
}
function resetPage() { before.value = 0; pages.value = []; }
watch([search, service, language, level, attempt, currentSession], () => {
  resetPage();
  clearTimeout(debounce);
  debounce = setTimeout(() => void refresh(), 180);
});
async function action(task: () => Promise<unknown>, message: string) {
  if (actionPending.value) return;
  actionPending.value = true;
  feedback.value = '';
  try {
    const result = await task();
    if (result !== false) feedback.value = message;
  } catch { error.value = 'The diagnostic action could not finish. Please try again.'; }
  finally { actionPending.value = false; }
}
function older() {
  if (!rows.value.length) return;
  live.value = false;
  pages.value.push(before.value);
  before.value = rows.value[rows.value.length - 1].id;
  void refresh();
}
function newer() { before.value = pages.value.pop() || 0; void refresh(); }
function resume() {
  live.value = !live.value;
  if (live.value) { resetPage(); void refresh(); }
}
function followAttempt() {
  if (!selected.value) return;
  attempt.value = selected.value.attempt;
  level.value = '';
  search.value = '';
}
function copySelected() {
  const id = selected.value?.id;
  if (id) void action(() => diagnostics.copyEvent(id), 'Event copied.');
}
function exportPage() { void action(() => diagnostics.export(query.value), 'Diagnostic page exported.'); }
function keydown(event: KeyboardEvent) {
  if (event.key === 'Escape') { event.preventDefault(); emit('close'); }
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'f') {
    event.preventDefault(); input.value?.focus(); input.value?.select();
  }
}
function resize(delta: number) {height.value = Math.max(300, Math.min(window.innerHeight - 72, height.value + delta));}
function beginResize(event: PointerEvent) {
  expanded.value = false;
  dragStart = { y: event.clientY, height: height.value };
  (event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
}
function drag(event: PointerEvent) {
  if (!dragStart) return;
  height.value = Math.max(300, Math.min(window.innerHeight - 72, dragStart.height + dragStart.y - event.clientY));
}
onMounted(async () => {
  previousFocus = document.activeElement as HTMLElement | null;
  stop = diagnostics.onChanged(() => {
    if (!live.value) return;
    if (pending.value) changedDuringRead = true;
    else void refresh();
  });
  try { await diagnostics.listen(true); }
  catch { liveWarning.value = 'Live updates are unavailable. Use Refresh to read saved events.'; }
  if (disposed) { void diagnostics.listen(false); return; }
  await refresh();
  await nextTick();
  if (!disposed) input.value?.focus();
});
onUnmounted(() => {
  disposed = true;
  sequence++;
  clearTimeout(debounce);
  stop?.();
  void diagnostics.listen(false).catch(() => {});
  if (previousFocus?.isConnected) previousFocus.focus();
});
</script>

<template>
 <section class="diagnostics" aria-label="Diagnostics console" :class="{ expanded }" :style="{height: height + 'px'}" @keydown="keydown">
  <div class="resize-handle" role="separator" aria-label="Resize diagnostics" aria-orientation="horizontal" :aria-valuenow="height" aria-valuemin="300" tabindex="0"
   @pointerdown="beginResize" @pointermove="drag" @pointerup="dragStart=null" @lostpointercapture="dragStart=null"
   @keydown.up.prevent="resize(32)" @keydown.down.prevent="resize(-32)"></div>
  <header class="console-header">
   <div><h2>Diagnostics</h2><span class="console-caption">Account connections</span></div>
   <div class="console-actions">
    <UiButton size="compact" variant="quiet" :aria-pressed="!live" @click="resume">{{ live ? 'Pause' : 'Resume live' }}</UiButton>
    <UiButton size="compact" variant="quiet" :busy="pending" @click="refresh">Refresh</UiButton>
    <UiButton size="compact" variant="quiet" :aria-pressed="expanded" @click="expanded=!expanded">{{ expanded ? 'Restore' : 'Expand' }}</UiButton>
    <UiButton size="compact" variant="quiet" @click="emit('close')">Close</UiButton>
   </div>
  </header>
  <div class="console-filters">
   <label class="search-field">Search<input ref="input" v-model="search" type="search" class="ui-field" aria-label="Search" maxlength="200" placeholder="Message, stage, code or attempt"></label>
   <label>Level<select v-model="level" class="ui-field" aria-label="Level"><option value="">All levels</option><option>ERROR</option><option>WARN</option><option>INFO</option><option>DEBUG</option></select></label>
   <label>Service<select v-model="service" class="ui-field" aria-label="Service"><option value="">All services</option><option v-for="(name,key) in serviceName" :key="key" :value="key">{{ name }}</option></select></label>
   <label>Language<select v-model="language" class="ui-field" aria-label="Language"><option value="">ES + EN</option><option value="es">ES</option><option value="en">EN</option></select></label>
   <label>History<select v-model="currentSession" class="ui-field" aria-label="History"><option :value="false">Saved history</option><option :value="true">This session</option></select></label>
  </div>
  <div v-if="attempt" class="attempt-filter">Attempt <code>{{ attempt }}</code><UiButton size="compact" variant="quiet" @click="attempt=''">Clear attempt filter</UiButton></div>
  <p v-if="warning" class="console-warning" role="status">{{ warning }}</p>
  <p v-if="liveWarning" class="console-warning" role="status">{{ liveWarning }}</p>
  <p v-if="error" class="console-error" role="alert">{{ error }}</p>
  <div class="console-body" :class="{inspecting:selected}">
   <div class="log-scroll" tabindex="0" aria-label="Diagnostic events" :aria-busy="pending">
    <div v-if="!rows.length" class="log-empty"><h3>{{ search || level || service || language || attempt ? 'No matching events' : 'Ready for your next connection' }}</h3><p>{{ search || level || service || language || attempt ? 'Adjust the filters or start a new connection in Accounts.' : 'Connect an account to see its sign-in stages here.' }}</p></div>
    <table v-else class="log-table">
     <thead><tr><th scope="col">Time</th><th scope="col">Level</th><th scope="col">Account</th><th scope="col">Event</th></tr></thead>
     <tbody><tr v-for="row in rows" :key="row.id" :class="{selected:selected?.id===row.id}">
      <td><time :datetime="row.time" :title="new Date(row.time).toLocaleString()">{{ time(row.time) }}</time></td>
      <td><span class="log-level" :class="row.level.toLowerCase()">{{ row.level }}</span></td>
      <td>{{ serviceName[row.service] }} <span class="log-language">{{ row.language.toUpperCase() }}</span></td>
      <td><button class="event-button" :aria-pressed="selected?.id===row.id" @click="selected=row">{{ row.msg }}<span v-if="row.reason"> · {{ row.reason.replaceAll('_',' ') }}</span></button></td>
     </tr></tbody>
    </table>
   </div>
   <aside v-if="selected" class="event-detail" aria-label="Event details">
    <div class="detail-header"><h3>Event details</h3><UiButton size="compact" variant="quiet" @click="selected=null">Hide</UiButton></div>
    <p class="detail-message">{{ selected.msg }}</p>
    <dl>
     <dt>Time</dt><dd>{{ new Date(selected.time).toLocaleString() }}</dd>
     <dt>Stage</dt><dd><code>{{ selected.stage }}</code></dd>
     <dt>Elapsed</dt><dd>{{ selected.elapsed_ms.toLocaleString() }} ms</dd>
     <template v-if="selected.reason"><dt>Reason</dt><dd>{{ selected.reason.replaceAll('_',' ') }}</dd></template>
     <template v-if="selected.http_status"><dt>HTTP</dt><dd>{{ selected.http_status }}</dd></template>
     <template v-if="selected.provider_code"><dt>Provider code</dt><dd>{{ selected.provider_code }}<span v-if="selected.provider_subcode"> / {{ selected.provider_subcode }}</span></dd></template>
     <template v-if="selected.stage==='pages_received'"><dt>Pages</dt><dd>{{ selected.pages }}</dd></template>
     <dt>Attempt</dt><dd><code>{{ selected.attempt }}</code></dd>
    </dl>
    <div class="detail-actions">
     <UiButton size="compact" @click="followAttempt">Follow attempt</UiButton>
     <UiButton size="compact" variant="quiet" @click="copySelected">Copy event</UiButton>
    </div>
   </aside>
  </div>
  <footer class="console-footer">
   <span>{{ rows.length }} events · {{ live ? 'Live' : 'Paused' }} · Newest first</span>
   <div class="console-actions">
    <UiButton size="compact" variant="quiet" :disabled="!pages.length || pending" @click="newer">Newer</UiButton>
    <UiButton size="compact" variant="quiet" :disabled="!more || pending" @click="older">Older</UiButton>
    <UiButton size="compact" variant="quiet" :disabled="!rows.length" title="Export this page of up to 200 matching events" @click="exportPage">Export page</UiButton>
    <UiButton size="compact" variant="quiet" @click="action(diagnostics.openFolder,'Log folder opened.')">Log folder</UiButton>
   </div>
  </footer>
  <p v-if="feedback" class="console-feedback" role="status">{{ feedback }}</p>
 </section>
</template>

<style scoped>
.diagnostics {position:fixed;z-index:20;left:16px;right:16px;bottom:16px;display:flex;flex-direction:column;background:var(--page);border:1px solid var(--control-border);max-height:calc(100vh - 72px);min-height:300px}
.diagnostics.expanded {height:calc(100vh - 72px)!important}
.resize-handle {height:8px;flex-shrink:0;cursor:ns-resize;touch-action:none;background:var(--sidebar)}
.console-header,.console-footer {display:flex;align-items:center;justify-content:space-between;gap:16px;padding:12px 20px}
.console-header {border-bottom:1px solid var(--line)}
.console-header>div:first-child {display:flex;align-items:baseline;gap:12px}
.console-header h2 {font-size:20px;margin:0}
.console-caption,.console-footer,.console-filters label {font-size:var(--text-sm);color:var(--secondary)}
.console-actions {display:flex;gap:4px;align-items:center;flex-wrap:wrap}
.console-filters {display:flex;align-items:end;gap:12px;padding:12px 20px}
.console-filters label {display:flex;flex-direction:column;gap:4px}
.console-filters .ui-field {min-height:32px;height:32px;font-size:13px}
.search-field {flex:1;min-width:140px}
.search-field input {width:100%}
.console-body {flex:1;min-height:0;display:grid;grid-template-columns:minmax(0,1fr);border-top:1px solid var(--line)}
.console-body.inspecting {grid-template-columns:minmax(0,1fr) 300px}
.log-scroll {min-height:0;overflow:auto}
.log-table {width:100%;border-collapse:collapse;text-align:left;font-size:13px}
.log-table th {position:sticky;top:0;background:var(--sidebar);font-weight:500;color:var(--secondary);z-index:1}
.log-table th,.log-table td {padding:8px 12px;border-bottom:1px solid var(--line);vertical-align:top}
.log-table th:first-child,.log-table td:first-child {padding-left:20px}
.log-table td:nth-child(-n+3) {white-space:nowrap}
.log-table time,.log-level {font:12px/1.5 var(--font-mono);font-variant-numeric:tabular-nums}
.log-level.error,.console-error {color:var(--danger)}
.log-level.warn,.console-warning {color:var(--warning)}
.log-level.info {color:var(--text)}
.log-level.debug {color:var(--muted)}
.log-language {margin-left:4px;color:var(--secondary);font-size:12px}
.event-button {display:block;border:0;background:transparent;text-align:left;padding:0;width:100%;line-height:1.5}
.event-button span {color:var(--secondary)}
.log-table tr:hover td {background:var(--sidebar)}
.log-table tr.selected td {background:var(--accent-soft)}
.log-empty {padding:32px 20px;color:var(--secondary)}
.log-empty h3 {font-size:16px;font-weight:500;color:var(--text)}
.log-empty p {margin-top:8px}
.event-detail {overflow:auto;border-left:1px solid var(--line);padding:12px 16px;background:var(--sidebar)}
.detail-header {display:flex;align-items:center;justify-content:space-between}
.detail-header h3 {font-size:14px}
.detail-message {margin:12px 0}
.event-detail dl {display:grid;grid-template-columns:72px minmax(0,1fr);gap:8px;margin:16px 0}
.event-detail dt {color:var(--secondary);font-size:12px}
.event-detail dd {margin:0;font-size:12px;overflow-wrap:anywhere}
.event-detail code,.attempt-filter code {font-family:var(--font-mono);font-size:12px}
.detail-actions {display:flex;gap:4px;flex-wrap:wrap}
.console-footer {border-top:1px solid var(--line);padding-block:8px;flex-wrap:wrap}
.console-warning,.console-error,.console-feedback {padding:4px 20px;margin:0;font-size:13px}
.attempt-filter {display:flex;align-items:center;gap:8px;padding:0 20px 8px;font-size:12px}
@media(max-width:950px) {
 .console-caption {display:none}
 .console-body.inspecting {grid-template-columns:minmax(0,1fr) 240px}
 .console-filters {gap:8px;flex-wrap:wrap}
 .search-field {flex-basis:100%}
 .log-table td:nth-child(3),.log-table th:nth-child(3) {display:none}
}
</style>
