<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { diagnostics, type DiagnosticEvent, type DiagnosticQuery } from '../services/diagnostics';
import UiButton from './UiButton.vue';

const props = defineProps<{ service?: string; language?: string }>();
const collapsed = defineModel<boolean>('collapsed', { default: false });
const emit = defineEmits<{ close: [] }>();
const search = ref('');
const service = ref(props.service || '');
const language = ref(props.language || '');
const level = ref('');
const attempt = ref('');
const currentSession = ref(false);
const live = ref(true);
const before = ref(0);
const after = ref(0);
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
const copyMinutes = ref(0);
const height = ref(360);
const expanded = ref(false);
const input = ref<HTMLInputElement | null>(null);
const query = computed<DiagnosticQuery>(() => ({
  search: search.value, service: service.value, language: language.value,
  level: level.value, attempt: attempt.value, currentSession: currentSession.value, before: before.value, after: after.value,
}));
const filterCount = computed(() => [search.value, service.value, language.value, level.value, attempt.value, currentSession.value].filter(Boolean).length);
let sequence = 0;
let disposed = false;
let debounce: ReturnType<typeof setTimeout> | undefined;
let stop: (() => void) | undefined;
let previousFocus: HTMLElement | null = null;
let changedDuringRead = false;
let dragStart: { y: number; height: number } | null = null;
const serviceName: Record<string, string> = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram', x: 'X' };
const time = (value: string) => new Date(value).toLocaleTimeString([], { hour12: false });
watch(() => props.service, value => { service.value = value || ''; });
watch(() => props.language, value => { language.value = value || ''; });
watch(collapsed, async value => {
  if (value) expanded.value = false;
  try { await diagnostics.listen(!value); }
  catch { liveWarning.value = 'Live updates unavailable'; }
  if (!value && !disposed) { await refresh(); await nextTick(); input.value?.focus(); }
});
function expand() { collapsed.value = false; expanded.value = !expanded.value; }

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
    if (request === sequence && !disposed && changedDuringRead && live.value && !collapsed.value) {
      changedDuringRead = false;
      void refresh();
    }
  }
}
function resetPage() { before.value = 0; pages.value = []; }
function resetFilters() {
  search.value = ''; service.value = ''; language.value = ''; level.value = '';
  attempt.value = ''; currentSession.value = false; selected.value = null;
}
async function pasteSearch() {
  if (actionPending.value) return;
  actionPending.value = true;
  try {
    const text = await diagnostics.pasteText();
    if (disposed) return;
    search.value = text.slice(0, 200);
    feedback.value = text.length > 200 ? 'Pasted first 200 characters' : (text ? 'Pasted into search' : 'Clipboard is empty');
    input.value?.focus();
  } catch { error.value = 'Could not read clipboard. Paste into Search with Ctrl+V.'; }
  finally { actionPending.value = false; }
}
async function clearView() {
  if (actionPending.value) return;
  actionPending.value = true;
  try {
    // Clear through the latest event across all filters, without deleting logs.
    const latest = await diagnostics.query({search:'',level:'',service:'',language:'',attempt:'',currentSession:false,before:0,after:0});
    if (disposed) return;
    after.value = latest.events?.[0]?.id || 0;
    selected.value = null; resetPage();
    feedback.value = 'View cleared · saved logs retained';
    await refresh();
  } catch { error.value = 'Could not clear the view. Try again.'; }
  finally { actionPending.value = false; }
}
function showHistory() { after.value = 0; resetPage(); feedback.value = ''; void refresh(); }
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
async function copyResults() {
  if (actionPending.value) return;
  actionPending.value = true;
  error.value = ''; feedback.value = '';
  try {
    const count = await diagnostics.copyResults({...query.value, before:0}, copyMinutes.value);
    feedback.value = count ? `${count} events copied · oldest first` : 'No matching events · clipboard unchanged';
  } catch {
    error.value = 'Could not copy results. Narrow the filters or time range; large selections are limited to 10,000 events / 8 MB.';
  } finally { actionPending.value = false; }
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
    if (!live.value || collapsed.value) return;
    if (pending.value) changedDuringRead = true;
    else void refresh();
  });
  try { await diagnostics.listen(!collapsed.value); }
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
 <section id="diagnostics-console" class="diagnostics" aria-label="Diagnostics console" :class="{ expanded, collapsed }" :style="{height: height + 'px'}" @keydown="keydown">
  <div v-if="!collapsed && !expanded" class="resize-handle" role="separator" aria-label="Resize diagnostics" aria-orientation="horizontal" :aria-valuenow="height" aria-valuemin="300" tabindex="0"
   @pointerdown="beginResize" @pointermove="drag" @pointerup="dragStart=null" @lostpointercapture="dragStart=null"
   @keydown.up.prevent="resize(32)" @keydown.down.prevent="resize(-32)"></div>
  <header class="console-header">
   <div class="console-heading"><button class="console-title" :aria-label="collapsed ? 'Open console panel' : 'Collapse console'" :aria-expanded="!collapsed" @click="collapsed=!collapsed"><svg viewBox="0 0 16 16" aria-hidden="true" :class="{folded:collapsed}"><path d="m4 6 4 4 4-4"/></svg><span>Console</span></button>
    <span class="stream-state" :class="{paused:!live}" role="status" :title="!live ? 'Display paused; recording continues' : (collapsed ? 'Display folded; recording continues' : 'Events update automatically')">{{ !live ? 'Paused' : (collapsed ? 'Folded' : '● Live') }}</span>
    <span v-if="filterCount" class="filter-count">{{ filterCount }} {{ filterCount===1 ? 'filter' : 'filters' }}</span>
   </div>
   <div class="console-actions">
    <UiButton v-if="!collapsed" size="compact" variant="quiet" :aria-pressed="!live" @click="resume">{{ live ? 'Pause' : 'Resume live' }}</UiButton>
    <UiButton v-if="!collapsed" size="compact" variant="quiet" :busy="pending" @click="refresh">Refresh</UiButton>
    <UiButton size="compact" variant="quiet" :aria-pressed="expanded" @click="expand">{{ expanded ? 'Restore' : 'Expand' }}</UiButton>
    <UiButton size="compact" variant="quiet" @click="emit('close')">Close</UiButton>
   </div>
  </header>
  <template v-if="!collapsed">
  <div class="console-filters">
   <label class="search-field" :class="{filtered:search}">Search<span class="search-controls"><input ref="input" v-model="search" type="search" class="ui-field" aria-label="Search" maxlength="200" placeholder="Message, stage, code or attempt"><UiButton size="compact" variant="quiet" :disabled="actionPending" title="Paste clipboard text into Search" @click="pasteSearch">Paste</UiButton></span></label>
   <label :class="{filtered:level}">Level<select v-model="level" class="ui-field" aria-label="Level"><option value="">All levels</option><option>ERROR</option><option>WARN</option><option>INFO</option><option>DEBUG</option></select></label>
   <label :class="{filtered:service}">Source<select v-model="service" class="ui-field" aria-label="Source"><option value="">All sources</option><option v-for="(name,key) in serviceName" :key="key" :value="key">{{ name }}</option></select></label>
   <label :class="{filtered:language}">Context<select v-model="language" class="ui-field" aria-label="Context"><option value="">All</option><option value="es">ES</option><option value="en">EN</option></select></label>
   <label :class="{filtered:currentSession}">History<select v-model="currentSession" class="ui-field" aria-label="History"><option :value="false">Saved history</option><option :value="true">This session</option></select></label>
  </div>
  <div class="view-controls"><span v-if="filterCount">Filtered results</span><UiButton v-if="filterCount" size="compact" variant="quiet" @click="resetFilters">Reset filters</UiButton><span v-if="after">Since clear</span><UiButton v-if="after" size="compact" variant="quiet" @click="showHistory">Show saved history</UiButton>
   <div class="copy-controls" title="Copies all matching saved events, across pages. Active filters and Clear view apply; pausing does not stop recording.">
    <select v-model="copyMinutes" class="ui-field" aria-label="Copy time range"><option :value="0">All filtered events</option><option :value="1">Last minute</option><option :value="5">Last 5 minutes</option><option :value="10">Last 10 minutes</option><option :value="30">Last 30 minutes</option></select>
    <UiButton size="compact" variant="quiet" :disabled="actionPending" @click="copyResults">Copy results</UiButton>
   </div>
  </div>
  <div v-if="attempt" class="attempt-filter">Attempt <code>{{ attempt }}</code><UiButton size="compact" variant="quiet" @click="attempt=''">Clear attempt filter</UiButton></div>
  <p v-if="warning" class="console-warning" role="status">{{ warning }}</p>
  <p v-if="liveWarning" class="console-warning" role="status">{{ liveWarning }}</p>
  <p v-if="error" class="console-error" role="alert">{{ error }}</p>
  <div class="console-body" :class="{inspecting:selected}">
   <div class="log-scroll" tabindex="0" aria-label="Diagnostic events" :aria-busy="pending">
    <div v-if="!rows.length" class="log-empty">{{ search || level || service || language || attempt ? 'No matching events' : 'No events' }}</div>
    <table v-else class="log-table">
     <thead><tr><th scope="col">Time</th><th scope="col">Level</th><th scope="col">Source</th><th scope="col">Event</th></tr></thead>
     <tbody><tr v-for="row in rows" :key="row.id" :class="{selected:selected?.id===row.id}" @click="selected=row">
      <td><time :datetime="row.time" :title="new Date(row.time).toLocaleString()">{{ time(row.time) }}</time></td>
      <td><span class="log-level" :class="row.level.toLowerCase()">{{ row.level }}</span></td>
      <td>{{ serviceName[row.service] || row.service }} <span class="log-language">{{ row.language.toUpperCase() }}</span></td>
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
   <span>{{ rows.length }} {{ filterCount ? 'matches' : 'events' }} · Newest first</span>
   <div class="console-actions">
    <UiButton size="compact" variant="quiet" :disabled="pending || actionPending" title="Clear displayed events; keep saved log files" @click="clearView">Clear view</UiButton>
    <UiButton size="compact" variant="quiet" :disabled="!pages.length || pending" @click="newer">Newer</UiButton>
    <UiButton size="compact" variant="quiet" :disabled="!more || pending" @click="older">Older</UiButton>
    <UiButton size="compact" variant="quiet" :disabled="!rows.length || pending || actionPending" title="Export this page of up to 200 matching events" @click="exportPage">Export page</UiButton>
    <UiButton size="compact" variant="quiet" @click="action(diagnostics.openFolder,'Log folder opened.')">Log folder</UiButton>
   </div>
  </footer>
  <p v-if="feedback" class="console-feedback" role="status">{{ feedback }}</p>
  </template>
 </section>
</template>

<style scoped>
.diagnostics {
 --page:#111214; --sidebar:#191b1e; --raised:#25282c; --line:#2b2e33;
 --text:#e1e3e7; --secondary:#abb0b8; --muted:#959ba5; --control-border:#454b54;
 --accent-soft:#2c2825; --accent-ink:#ffad81; --danger:#ff9595; --warning:#e8bd76;
 --disabled-bg:transparent; --disabled-text:#787f8a;
 position:fixed;z-index:20;left:240px;right:0;bottom:0;display:flex;flex-direction:column;
 background:var(--page);color:var(--text);border-top:1px solid var(--control-border);
 max-height:calc(100vh - var(--window-bar-height));min-height:300px;
 font:12px/1.5 var(--font-mono);color-scheme:dark;
}
.diagnostics.expanded {left:0;height:calc(100vh - var(--window-bar-height))!important}
.diagnostics.collapsed {height:36px!important;min-height:36px}
.resize-handle {height:4px;flex-shrink:0;cursor:ns-resize;touch-action:none;background:var(--sidebar)}
.resize-handle:hover,.resize-handle:focus-visible {background:var(--control-border)}
.console-header,.console-footer {display:flex;align-items:center;justify-content:space-between;gap:8px;padding:4px 12px;flex-shrink:0}
.console-header {height:36px;border-bottom:1px solid var(--line);background:var(--sidebar)}
.collapsed .console-header {border-bottom:0}
.console-title {display:flex;align-items:center;gap:8px;background:transparent;border:0;padding:4px 0;font:inherit;color:var(--text)}
.console-title svg {width:16px;height:16px;fill:none;stroke:currentColor;stroke-width:1.5}
.console-title svg.folded {transform:rotate(-90deg)}
.console-actions {display:flex;gap:4px;align-items:center;flex-wrap:wrap}
.console-heading {display:flex;align-items:center;gap:12px;min-width:0}
.stream-state {color:#a7c5b0;white-space:nowrap}
.stream-state.paused,.filter-count {color:var(--warning)}
.diagnostics :deep(.ui-button) {font:inherit;min-height:28px;padding:0 8px;border-radius:4px;border-color:var(--control-border)}
.diagnostics :deep(.ui-button[aria-pressed='true']) {background:var(--accent-soft);border-color:var(--warning);color:var(--warning)}
.diagnostics :deep(.ui-button:disabled) {border-color:var(--line)}
.console-filters {display:flex;align-items:end;gap:8px;padding:8px 12px}
.console-filters label {display:flex;flex-direction:column;gap:4px;color:var(--secondary);font-size:11px}
.console-filters .ui-field {min-height:28px;height:28px;padding:4px 8px;font:12px var(--font-mono);border-radius:4px}
.search-field {flex:1;min-width:120px}
.search-field input {width:100%}
.search-controls {display:flex;gap:4px}
.search-controls input {min-width:0;flex:1}
.console-filters .filtered {color:var(--warning)}
.console-filters .filtered .ui-field {border-color:var(--warning)}
.view-controls {display:flex;align-items:center;gap:8px;padding:0 12px 8px;color:var(--secondary);flex-wrap:wrap}
.copy-controls {display:flex;align-items:center;gap:4px;margin-left:auto}
.copy-controls .ui-field {min-height:28px;height:28px;padding:4px 8px;font:inherit;border-radius:4px;max-width:185px}
.console-body {flex:1;min-height:0;display:grid;grid-template-columns:minmax(0,1fr);border-top:1px solid var(--line)}
.console-body.inspecting {grid-template-columns:minmax(0,1fr) 272px}
.log-scroll {min-height:0;overflow:auto;scrollbar-color:var(--control-border) var(--page)}
.log-table {width:100%;border-collapse:collapse;text-align:left;font:inherit}
.log-table th {position:sticky;top:0;background:var(--sidebar);font-weight:400;color:var(--secondary);z-index:1}
.log-table th,.log-table td {padding:4px 8px;vertical-align:top}
.log-table th {border-bottom:1px solid var(--line)}
.log-table th:first-child,.log-table td:first-child {padding-left:12px}
.log-table td:nth-child(-n+3) {white-space:nowrap}
.log-table time,.log-level {font:inherit;font-variant-numeric:tabular-nums}
.log-table time {color:var(--muted)}
.log-level.error,.console-error {color:var(--danger)}
.log-level.warn,.console-warning {color:var(--warning)}
.log-level.info {color:#9bbff1}
.log-level.debug {color:var(--muted)}
.log-language {margin-left:4px;color:var(--secondary)}
.event-button {display:block;border:0;background:transparent;color:inherit;text-align:left;padding:0;width:100%;font:inherit}
.event-button span {color:var(--secondary)}
.log-table tr:hover td {background:var(--sidebar)}
.log-table tbody tr {cursor:pointer}
.log-table tr.selected td {background:var(--accent-soft)}
.log-empty {padding:16px 12px;color:var(--muted)}
.event-detail {overflow:auto;border-left:1px solid var(--line);padding:8px 12px;background:var(--sidebar)}
.detail-header {display:flex;align-items:center;justify-content:space-between}
.detail-header h3 {font:inherit;color:var(--secondary)}
.detail-message {margin:8px 0}
.event-detail dl {display:grid;grid-template-columns:80px minmax(0,1fr);gap:8px;margin:12px 0}
.event-detail dt {color:var(--secondary)}
.event-detail dd {margin:0;overflow-wrap:anywhere}
.event-detail code,.attempt-filter code {font:inherit}
.detail-actions {display:flex;gap:4px;flex-wrap:wrap}
.console-footer {border-top:1px solid var(--line);flex-wrap:wrap;color:var(--secondary)}
.console-warning,.console-error,.console-feedback {padding:4px 12px;margin:0;font:inherit}
.attempt-filter {display:flex;align-items:center;gap:8px;padding:0 12px 8px}
@media(max-width:1050px) {
 .console-filters {flex-wrap:wrap}
 .search-field {flex-basis:100%}
 .console-body.inspecting {grid-template-columns:minmax(0,1fr) 224px}
 .log-table td:nth-child(3),.log-table th:nth-child(3) {display:none}
}
@media(max-width:900px) {.diagnostics {left:200px}}
</style>
