<script setup lang="ts">
import {computed,onMounted,ref} from 'vue';
import UiButton from './UiButton.vue';
import UiHelp from './UiHelp.vue';
import {drafts,type Review} from '../services/drafts';
import type {Episode,Video} from '../services/publisher';
const props=defineProps<{episode:Episode;language:string;video:Video;locked:boolean}>();
const emit=defineEmits<{state:[value:string]}>();
const review=ref<Review|null>(null),notes=ref(''),confirmed=ref(false),busy=ref(false),error=ref('');
const panel=ref<HTMLDetailsElement|null>(null);
const current=computed(()=>review.value?.fingerprint===props.episode.fingerprint && review.value?.videoHash===props.video.sha256);
const decision=computed(()=>current.value?review.value?.state:'pending');
const final=computed(()=>props.video.kind==='final' && props.video.lang===props.language);
const dirty=computed(()=>notes.value!==(review.value?.notes||''));
const status=computed(()=>dirty.value?'Unsaved notes':decision.value==='reviewed'?'Reviewed':decision.value==='changes_requested'?'Changes requested':'Review needed');
onMounted(async()=>{busy.value=true;try{review.value=await drafts.review(props.episode.id,props.language,props.video.id);notes.value=review.value.notes;emit('state',decision.value||'pending');}catch(e){error.value=String(e)}finally{busy.value=false}});
async function save(state:string){
 if(!review.value || busy.value || !final.value) return;
 busy.value=true;error.value='';
 try{
 review.value=await drafts.saveReview({...review.value,fingerprint:props.episode.fingerprint,videoHash:props.video.sha256,state,notes:notes.value},confirmed.value);
 confirmed.value=false;emit('state',state);if(state==='reviewed' && panel.value)panel.value.open=false;
 }catch(e){error.value=String(e);throw e}finally{busy.value=false}
}
function decide(state:string){void save(state).catch(()=>{});}
async function flush(){if(busy.value)throw Error('Wait for the review to finish saving.');if(dirty.value)await save('pending');}
function focusReview(){if(panel.value){panel.value.open=true;panel.value.scrollIntoView({block:'nearest'});panel.value.querySelector<HTMLInputElement>('input[type=checkbox]')?.focus();}}
defineExpose({flush,focusReview});
</script>
<template>
 <section class="review-decision" aria-label="Review decision">
  <details ref="panel" class="review-details">
  <summary>Video review <span class="review-state" :class="{'valid-label':decision==='reviewed' && !dirty}" role="status">{{status}}</span></summary>
  <div class="review-help"><UiHelp text="Records that you checked this exact video and audio. Replacing the files requires another review. Production approval is checked separately before sending."/></div>
  <p v-if="!current && review?.revision" class="notice warning">The delivery changed. Review this version again; your notes are kept.</p>
  <fieldset :disabled="locked || busy || !review || !final" class="draft-fields review-controls">
   <label v-if="decision!=='reviewed' || dirty" class="review-confirm"><input v-model="confirmed" type="checkbox"> I watched and listened to this final version</label>
   <div class="draft-actions">
    <UiButton v-if="decision!=='reviewed' || dirty" size="compact" :disabled="!confirmed" :busy="busy" @click="decide('reviewed')">Mark reviewed</UiButton>
   </div>
   <details class="review-notes">
    <summary>Notes<span v-if="notes.trim()" class="notes-present">{{dirty?'Unsaved':'Saved note'}}</span></summary>
    <label for="review-notes" class="sr-only">Review notes</label>
    <textarea id="review-notes" v-model="notes" class="ui-field" rows="3" maxlength="4000" @input="confirmed=false;emit('state','pending')" placeholder="Changes to make or details to keep"></textarea>
    <div class="draft-actions"><UiButton :disabled="!notes.trim()" size="compact" title="Record needed changes locally" @click="decide('changes_requested')">Request changes</UiButton><UiButton v-if="dirty" size="compact" variant="quiet" @click="decide('pending')">Save notes</UiButton></div>
   </details>
  </fieldset>
  <p v-if="!final" class="publication-hint">Select the final version to record a review</p>
  <p v-if="error" class="notice error" role="alert">{{error}}</p>
  </details>
 </section>
</template>
<style scoped>
.review-decision{margin-top:20px}
.review-details>summary{cursor:pointer;font-size:var(--text-sm);color:var(--secondary)}
.review-details>summary .review-state{float:right;font-size:var(--text-xs)}
.review-help{display:flex;justify-content:flex-end;margin-top:8px}
.review-decision h3{font-size:var(--text-md)}
.review-state{margin-left:auto;white-space:nowrap}
.review-decision .section-header{gap:8px;margin-bottom:8px}
.review-notes{margin-top:16px;font-size:var(--text-sm);color:var(--secondary)}
.review-notes summary{cursor:pointer}
.notes-present{font-size:var(--text-xs);color:var(--muted);margin-left:12px}
.review-notes .draft-actions{margin-top:12px}
textarea{display:block;width:100%;margin-top:8px;resize:vertical}
.review-controls{margin-top:16px}.review-confirm{display:flex;align-items:flex-start;gap:8px;margin-bottom:16px}.review-state{font-size:var(--text-xs)}
</style>
