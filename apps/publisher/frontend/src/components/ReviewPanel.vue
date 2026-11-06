<script setup lang="ts">
import {computed,onMounted,ref} from 'vue';
import UiButton from './UiButton.vue';
import {drafts,type Review} from '../services/drafts';
import type {Episode,Video} from '../services/publisher';
const props=defineProps<{episode:Episode;language:string;video:Video;locked:boolean}>();
const emit=defineEmits<{state:[value:string]}>();
const review=ref<Review|null>(null),notes=ref(''),confirmed=ref(false),busy=ref(false),error=ref('');
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
 confirmed.value=false;emit('state',state);
 }catch(e){error.value=String(e);throw e}finally{busy.value=false}
}
function decide(state:string){void save(state).catch(()=>{});}
async function flush(){if(busy.value)throw Error('Wait for the review to finish saving.');if(dirty.value)await save('pending');}
defineExpose({flush});
</script>
<template>
 <section class="review-decision" aria-label="Review decision">
  <div class="section-header"><h3>Editorial review</h3><span class="review-state" :class="{'valid-label':decision==='reviewed' && !dirty}" role="status">{{status}}</span></div>
  <p v-if="!current && review?.revision" class="notice warning">The delivery changed. Review this version again; your notes are kept.</p>
  <label for="review-notes">Review notes</label>
  <textarea id="review-notes" v-model="notes" class="ui-field" rows="3" maxlength="4000" :disabled="locked || busy || !final" @input="confirmed=false;emit('state','pending')" placeholder="Optional"></textarea>
  <fieldset :disabled="locked || busy || !review || !final" class="draft-fields review-controls">
   <label v-if="decision!=='reviewed' || dirty" class="review-confirm"><input v-model="confirmed" type="checkbox"> I watched and listened to this final version</label>
   <div class="draft-actions">
    <UiButton v-if="decision!=='reviewed' || dirty" variant="primary" :disabled="!confirmed" :busy="busy" @click="decide('reviewed')">Mark reviewed</UiButton>
    <UiButton :disabled="!notes.trim()" title="Record needed changes locally" @click="decide('changes_requested')">Request changes</UiButton>
    <UiButton v-if="dirty" variant="quiet" @click="decide('pending')">Save notes</UiButton>
   </div>
  </fieldset>
  <p v-if="!final" class="publication-hint">Select the final version to record a review</p>
  <p v-if="error" class="notice error" role="alert">{{error}}</p>
 </section>
</template>
<style scoped>
.review-decision{padding-top:0}
textarea{display:block;width:100%;margin-top:8px;resize:vertical}
.review-controls{margin-top:16px}.review-confirm{display:flex;align-items:flex-start;gap:8px;margin-bottom:16px}.review-state{font-size:var(--text-xs)}
</style>
