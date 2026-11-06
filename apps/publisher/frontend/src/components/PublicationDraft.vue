<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import UiButton from './UiButton.vue';
import UiHelp from './UiHelp.vue';
import UiDialog from './UiDialog.vue';
import {uploads,type Profile} from '../services/uploads';
import DirectUpload from './DirectUpload.vue';
import BatchPublish from './BatchPublish.vue';
import { accountKey, modeOptions, modeLabel, postApproved, sendLabel, draftBlocker, type DeliveryMode } from '../services/publication';
import { drafts, type Draft, type DraftKey } from '../services/drafts';
import type { Episode, Video } from '../services/publisher';

const props = defineProps<{ episode: Episode; language: string; video: Video; locked: boolean; stage:'prepare'|'publish' }>();
const region = ref<HTMLElement | null>(null);
const platforms = { youtube: 'YouTube', tiktok: 'TikTok', instagram: 'Instagram', x: 'X' };
const emit=defineEmits<{accounts:[];prepare:[];publish:[];busy:[value:boolean]}>();
const profiles=ref<Profile[]>([]);
const batchOpen=ref(false);
const destinations=computed(()=>Object.entries(platforms).map(([id,label])=>{const profile=profiles.value.find(p=>p.platform===id && p.language===props.language);return {id,label,profile,draft:saved.value.find(d=>d.key.platform===id && d.key.account===accountKey(profile))};}));
const recoverable=computed(()=>saved.value.filter(d=>d.key.platform===form.value?.key.platform && !d.key.account.startsWith('account:') && d.key.account!==form.value?.key.account));
const recovery=ref('');
const restoreOpen=ref(false);
const approving=ref(false);
const destination=computed(()=>profiles.value.find(p=>p.platform===form.value?.key.platform && p.language===props.language && accountKey(p)===form.value?.key.account));
const platform = ref('youtube');
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
const name = (key: DraftKey) => `${platforms[key.platform as keyof typeof platforms]} / ${destination.value?.name || 'Unassigned'}`;
const stale = computed(() => !!form.value?.revision && (form.value.deliveryFingerprint !== props.episode.fingerprint ||
  form.value.videoHash !== props.episode.videos?.find(v => v.id === form.value?.videoId)?.sha256));
const differentVideo = computed(() => form.value?.videoId !== props.video.id);
const approved=computed(()=>!dirty.value && postApproved(form.value || undefined) && !stale.value && !differentVideo.value);
const approvalHint=computed(()=>!destination.value?'Connect this account first':stale.value || differentVideo.value || props.video.kind!=='final' || props.video.lang!==props.language?'Link the current final video first':!form.value || !(form.value.title+form.value.description+form.value.hashtags).trim()?'Add post text first':'');
const uploadHint=computed(()=>draftBlocker(form.value || undefined,destination.value,props.episode,props.video));
const batchBlocker=computed(()=>{
 if(loading.value || saving.value || dirty.value || platformBusy.value || approving.value)return 'Wait for your changes to finish saving.';
 const connected=destinations.value.filter(r=>r.profile);
 if(!connected.length)return 'Connect an account to publish.';
 const missing=connected.filter(r=>!postApproved(r.draft));
 if(missing.length)return 'Approval needed: '+missing.map(r=>r.label+' / '+r.profile!.name).join(', ')+'. Select each account and choose Approve post.';
 const blocked=connected.map(r=>({label:r.label,reason:draftBlocker(r.draft,r.profile,props.episode,props.video)})).filter(r=>r.reason);
 return blocked.map(r=>r.label+': '+r.reason).join('. ');
});
async function setApproval(value:boolean){
 if(approving.value || (value && approvalHint.value))return;
 approving.value=true;emit('busy',true);
 try{await flush();if(!form.value?.revision)throw new Error('Save the draft first');
 const result=await drafts.approve(form.value.key,form.value.revision,value);
 form.value=result;updateList(result);feedback.value=value?'Post approved for this account':'Approval removed';
 }catch(e){error.value=String(e)}finally{approving.value=false;emit('busy',false)}
}
const status = computed(() => saving.value ? 'Saving…' : dirty.value ? 'Unsaved changes' : form.value?.revision ? 'Saved locally' : 'Not saved yet');

function edited() {
  generation++;
  if(form.value) form.value.approvedRevision=0;
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
          hashtags: snapshot.hashtags, deliveryMode:snapshot.deliveryMode || 'draft', revision: snapshot.revision,
          videoId: attachVideo ? props.video.id : snapshot.videoId, attachVideo,
        });
        form.value = { ...form.value, revision: result.revision, updatedAt: result.updatedAt, approvedRevision: result.approvedRevision,
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
    platform.value = key.platform; recovery.value='';
    generation = 0; attachVideo = false; dirty.value = false;
    error.value = ''; feedback.value = ''; reloadPrompt.value = false;
  } catch (e) { error.value ||= `Could not open draft. ${String(e)}`; }
  finally { loading.value = false; }
}
async function choosePlatform(value:string) {
  if (loading.value || platformBusy.value || approving.value || props.locked) return;
  const profile=profiles.value.find(p=>p.platform===value && p.language===props.language);
  await open({episode:props.episode.id,language:props.language,platform:value,account:accountKey(profile)});
}
async function changeMode(value:string, mode:DeliveryMode) {
  if(loading.value || saving.value || platformBusy.value || approving.value || props.locked)return;
  loading.value=true;platformBusy.value=true;emit('busy',true);
  try {
    await flush();
    const profile=profiles.value.find(p=>p.platform===value && p.language===props.language);
    const key={episode:props.episode.id,language:props.language,platform:value,account:accountKey(profile)};
    const d=await drafts.get(key);
    const result=await drafts.save({key,title:d.title,description:d.description,hashtags:d.hashtags,deliveryMode:mode,
      revision:d.revision,videoId:d.revision?d.videoId:props.video.id,attachVideo:!d.revision});
    updateList(result);
    if(form.value?.key.platform===value) form.value=result;
  } catch(e) {error.value=String(e);}
  finally{loading.value=false;platformBusy.value=false;emit('busy',false);}
}
async function recoverDraft() {
  const old=recoverable.value.find(d=>d.key.account===recovery.value);
  if(!old || !form.value) return;
  form.value.title=old.title;form.value.description=old.description;form.value.hashtags=old.hashtags;
  form.value.deliveryMode='draft';edited();await saveNow();recovery.value='';if(!error.value)restoreOpen.value=false;
}
async function publishAll() {
  await saveNow();
  if(error.value || dirty.value || batchBlocker.value) return;
  batchOpen.value=true;emit('publish');
}
function closePublish(){batchOpen.value=false;emit('prepare');}
async function publishingOptions() {
  if (loading.value || props.locked || !form.value) return;
  await saveNow();
  if (!error.value && !dirty.value) {
    feedback.value='';uploadError.value='';accountConfirmed.value=false;
    region.value?.querySelector('.post-tools-menu')?.removeAttribute('open');
    batchOpen.value=false;emit('publish');
  }
}
async function saveNow() {
  if (form.value && !form.value.revision && !dirty.value) edited();
  try { await flush(); feedback.value='Draft saved · '+new Date().toLocaleTimeString([], {hour:'2-digit',minute:'2-digit',second:'2-digit'}); } catch { /* Keep edits on failure. */ }
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
    error.value = ''; reloadPrompt.value = false; restoreOpen.value=false; feedback.value='Saved draft restored'; updateList(result);
  } catch (e) { error.value = String(e); }
  finally { loading.value = false; }
}
onMounted(async () => {
  try {
    profiles.value = await uploads.accounts() || [];
    saved.value = await drafts.list(props.episode.id, props.language) || [];
    loading.value = false;
    const p=profiles.value.find(p=>p.language===props.language && p.platform==='youtube') || profiles.value.find(p=>p.language===props.language);
    await open({episode:props.episode.id,language:props.language,platform:p?.platform || 'youtube',account:accountKey(p)},true);
  } catch (e) { error.value = `Could not load saved destinations. ${String(e)}`; }
  finally { loading.value = false; }
});
onUnmounted(() => clearTimeout(timer));
function focusEditor() {
  region.value?.scrollIntoView({ block: 'start' });
  region.value?.querySelector<HTMLElement>(form.value ? '#draft-title' : '#draft-account')?.focus({ preventScroll: true });
}
async function beforeLeaving() {
  if (platformBusy.value || approving.value) throw new Error('Wait for the current account action to finish.');
  await flush();
}
defineExpose({ flush: beforeLeaving, focusEditor, publishAll, batchBlocker });
</script>

<template>
  <section ref="region" class="publication" aria-label="Publication">
    <fieldset :disabled="locked || loading || preparingUpload || platformBusy || approving" class="draft-fields">
      <div class="compose-destination">
        <label for="post-platform" class="sr-only">Post platform</label>
        <select id="post-platform" :value="form?.key.platform || platform" class="ui-field" @change="choosePlatform(($event.target as HTMLSelectElement).value)">
          <option v-for="(label,value) in platforms" :key="value" :value="value">{{label}}{{profiles.find(p=>p.platform===value && p.language===language)?.name ? ' / '+profiles.find(p=>p.platform===value && p.language===language)?.name : ''}}</option>
        </select>
        <UiButton :disabled="!form" :busy="saving" @click="saveNow">Save draft</UiButton>
        <UiButton id="restore-draft" variant="quiet" :disabled="saving || (!form?.revision && !recoverable.length)" @click="restoreOpen=true">Restore draft</UiButton>
        <span v-if="feedback && stage==='prepare'" class="save-feedback" role="status">{{feedback}}</span>
      </div>
      <div v-if="!destination && form" class="compose-connect"><UiButton variant="quiet" size="compact" @click="emit('accounts')">Connect {{platforms[form.key.platform as keyof typeof platforms]}}</UiButton></div>
      <template v-if="form">
        <div class="account-post-actions">
          <div class="approval-control">
            <UiButton v-if="!approved" :disabled="!!approvalHint || saving" :busy="approving" @click="setApproval(true)">Approve post</UiButton>
            <template v-else><span class="post-approved">Approved</span><UiButton variant="quiet" size="compact" @click="setApproval(false)">Undo approval</UiButton></template>
            <UiHelp symbol="i" :text="approvalHint || 'Approve after watching the video and checking the text, account and delivery mode. This approval belongs only to this account. Editing the post or linking another video removes it. Nothing is sent until you confirm publishing.'"/>
          </div>
          <div class="selected-send"><UiButton id="open-publishing" variant="primary" :disabled="!!uploadHint || dirty || saving" @click="publishingOptions">{{sendLabel(form,destination)}}</UiButton><UiHelp v-if="uploadHint" symbol="i" :text="uploadHint==='Delivery approval required'?'This file is blocked by the production manifest. Approve the delivery in the production project and refresh Publisher. Test deliveries cannot be sent.':uploadHint"/></div>
        </div>
        <div v-if="differentVideo || stale" class="draft-video">
          <span>{{stale?'The video files changed':'A different video is selected'}}</span>
          <UiButton size="compact" :disabled="saving" @click="useSelectedVideo">Use selected video</UiButton>
        </div>
        <p v-if="stale" class="notice warning">The delivery changed since this draft was linked. Your text is preserved.</p>
        <div class="field-group"><label for="draft-title">Title</label><input id="draft-title" v-model="form.title" class="ui-field" maxlength="500" @input="edited"></div>
        <div class="field-group"><label for="draft-description">Description</label><textarea id="draft-description" v-model="form.description" class="ui-field" rows="4" maxlength="10000" @input="edited"></textarea></div>
        <div class="field-group"><label for="draft-hashtags">Hashtags</label><input id="draft-hashtags" v-model="form.hashtags" class="ui-field" maxlength="2000" placeholder="#catsoup" @input="edited"></div>
        <div class="compose-footer">
          <span role="status" :class="{warning:dirty}" title="Your edits save automatically on this PC">{{status}}</span>
          <details class="post-tools-menu">
            <summary>More</summary>
            <div class="post-tools-content">
              <div class="draft-actions">

                <UiButton variant="quiet" size="compact" :disabled="saving || !form.title.trim()" @click="copy('title')">Copy title</UiButton>
                <UiButton variant="quiet" size="compact" :disabled="saving || !(form.description.trim() || form.hashtags.trim())" title="Copy description and hashtags" @click="copy('caption')">Copy caption</UiButton>
                <UiButton variant="quiet" size="compact" @click="reveal">Open video folder</UiButton>
              </div>

            </div>
          </details>
        </div>
      </template>
    </fieldset>
    <section class="destination-plan" aria-label="Publishing destinations">
      <header><span>Delivery modes / {{language.toUpperCase()}}</span></header>
      <div v-for="row in destinations" :key="row.id" class="destination-plan-row">
        <button class="destination-name" :aria-current="form?.key.platform===row.id?'true':undefined" :disabled="locked || loading || platformBusy || approving" @click="choosePlatform(row.id)"><strong>{{row.label}}</strong><small>{{row.profile?.name || 'Not connected'}}</small></button>
        <div class="destination-mode"><select :aria-label="row.label+' delivery mode'" class="ui-field" :value="row.draft?.deliveryMode || 'draft'" :disabled="locked || loading || saving || platformBusy || approving || !row.profile || row.id==='tiktok'" :title="modeLabel(row.id,row.draft?.deliveryMode || 'draft',row.profile)" @change="changeMode(row.id,($event.target as HTMLSelectElement).value as DeliveryMode)">
          <option v-for="option in modeOptions(row.id,row.profile)" :key="option.value" :value="option.value">{{option.label}}</option>
        </select><UiHelp v-if="row.id==='tiktok'" symbol="i" text="This connector sends videos to your TikTok inbox. You finish editing and posting in TikTok. Direct Post is not implemented here, so the mode cannot be changed."/></div>
        <span class="destination-approval" :class="{'post-approved':postApproved(row.draft)}">{{!row.profile?'—':postApproved(row.draft)?'Approved':'Needs approval'}}</span>
      </div>
    </section>
    <UiDialog v-if="restoreOpen" title="Restore draft" return-focus-id="restore-draft" @close="restoreOpen=false">
      <p class="restore-account">{{platforms[platform as keyof typeof platforms]}} / {{destination?.name || 'Unassigned'}}</p>
      <p class="muted">Reload the last saved text for this account. Edits save automatically; this is not a version history.</p>
      <UiButton :disabled="!form?.revision || saving" @click="reloadSaved">Reload saved draft</UiButton>
      <div v-if="recoverable.length" class="restore-legacy">
        <label for="recover-draft">Recover unassigned text for this platform</label>
        <select id="recover-draft" v-model="recovery" class="ui-field"><option value="">Choose text to copy</option><option v-for="d in recoverable" :key="d.key.account" :value="d.key.account">{{d.key.account}} / {{d.title || 'Untitled'}}</option></select>
        <p class="muted">Replaces this account's text. The original is kept.</p>
        <UiButton :disabled="!recovery || saving" @click="recoverDraft">Replace current text</UiButton>
      </div>
      <p v-if="error" role="alert" class="notice error">{{error}}</p>
    </UiDialog>
    <UiDialog v-if="stage==='publish' && batchOpen" title="Publish all" return-focus-id="publish-all" :busy="platformBusy" @close="closePublish">
      <BatchPublish :episode="episode" :language="language" :video="video" :profiles="profiles" :saved="saved" @busy="platformBusy=$event;emit('busy',$event)"/>
    </UiDialog>
    <UiDialog v-if="stage==='publish' && !batchOpen && form" :title="'Publish to '+platforms[form.key.platform as keyof typeof platforms]" return-focus-id="open-publishing" :busy="platformBusy || preparingUpload" @close="closePublish">
      <DirectUpload :key="JSON.stringify([form.key,form.videoHash])" :draft="form" :hint="uploadHint" :flush="flush" @busy="platformBusy=$event;emit('busy',$event)" @accounts="emit('accounts')"/>
      <details v-if="!uploadHint && form.key.platform!=='x'" class="manual-handoff">
        <summary>Upload manually</summary>
        <fieldset :disabled="platformBusy || preparingUpload" class="draft-fields">
          <div class="draft-actions"><UiButton variant="quiet" size="compact" @click="copy('title')">Copy title</UiButton><UiButton variant="quiet" size="compact" @click="copy('caption')">Copy caption</UiButton></div>
          <label><input v-model="accountConfirmed" type="checkbox"> I checked that {{name(form.key)}} is open in my browser.</label>
          <div class="draft-actions"><UiButton :disabled="!accountConfirmed || saving" :busy="preparingUpload" @click="copyUploadPath">Copy upload file path</UiButton></div>
        </fieldset>
        <p v-if="uploadError" class="notice error" role="alert">{{uploadError}}</p>
      </details>
      <p v-if="feedback" class="ui-status ui-status--success" role="status">{{feedback}}</p>
      <p v-if="error" class="notice error" role="alert">{{error}}</p>
    </UiDialog>
    <div v-if="error && stage==='prepare'" class="notice error" role="alert">
      <p>{{error}}</p>
      <UiButton v-if="form?.revision && !reloadPrompt" variant="quiet" size="compact" :disabled="saving || locked" @click="reloadPrompt=true">Reload saved draft</UiButton>
      <div v-if="reloadPrompt" class="reload-prompt"><p>Replace your unsaved changes with the saved version?</p><UiButton size="compact" :disabled="loading || saving || locked" @click="reloadSaved">Reload and replace edits</UiButton><UiButton size="compact" variant="quiet" @click="reloadPrompt=false">Keep editing</UiButton></div>
    </div>
  </section>
</template>
