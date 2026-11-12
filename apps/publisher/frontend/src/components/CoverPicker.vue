<script setup lang="ts">
import {computed,ref,watch,onUnmounted} from 'vue';
import UiButton from './UiButton.vue';
import UiDialog from './UiDialog.vue';
import {drafts} from '../services/drafts';
import {uploads} from '../services/uploads';
import {mediaURL} from '../services/publisher';
import type {Draft, Cover} from '../../bindings/catsoup/publisher/internal/drafts/models';
const props=defineProps<{draft:Draft;disabled:boolean}>();
const emit=defineEmits<{change:[cover:Cover|null];busy:[value:boolean]}>();
const picker=ref<HTMLInputElement>(),player=ref<HTMLVideoElement>();
const remote=ref('');
const open=ref(false),busy=ref(false),image=ref(''),error=ref(''),result=ref('');
const supported=computed(()=>['youtube','instagram','tiktok'].includes(props.draft.key.platform));
const canImage=computed(()=>props.draft.key.platform==='youtube');
const hint=computed(()=>props.draft.key.platform==='tiktok'&&props.draft.deliveryMode!=='publish'
 ? 'TikTok Inbox does not receive a cover. Choose it in the TikTok mobile editor.'
 :props.draft.key.platform==='instagram'?'Instagram uses a video frame. Local PNG/JPG covers cannot be sent by this connector.'
 :props.draft.key.platform==='tiktok'?'Direct Post uses a video frame.'
 :props.draft.key.platform==='youtube'?'PNG/JPG or video frame · up to 2 MiB. Applied after YouTube processes the video.'
 :'This connector does not send a custom cover.');
const enabled=computed(()=>supported.value&&!(props.draft.key.platform==='tiktok'&&props.draft.deliveryMode!=='publish'));
async function refreshStatus(event:Event){
 if(!(event.target as HTMLDetailsElement).open)return;
 if(props.draft.key.account.startsWith('account:')){try{const r=await uploads.receipt(props.draft.key,props.draft.key.account.slice(8));remote.value=r.state;}catch{}}
 if(props.draft.key.platform!=='youtube'||!props.draft.cover)return;
 const key=JSON.stringify(props.draft.key);const hash=props.draft.cover.hash;
 try{const status=await uploads.coverStatus(props.draft.key);if(JSON.stringify(props.draft.key)===key&&props.draft.cover?.hash===hash)result.value=status==='applied'?'Thumbnail applied':status==='pending'||status==='needs_attention'?'Video kept · thumbnail needs attention':'';}catch(e){error.value=String(e);}
}
let disposed=false;onUnmounted(()=>{disposed=true;generation++;});
let generation=0;
watch(()=>[JSON.stringify(props.draft.key),props.draft.cover?.hash],async()=>{
 const ticket=++generation;image.value='';error.value='';result.value='';remote.value='';open.value=false;
 try{
  if(props.draft.key.account.startsWith('account:')){const r=await uploads.receipt(props.draft.key,props.draft.key.account.slice(8));if(ticket===generation)remote.value=r.state;}
  if(props.draft.cover){const src=await drafts.coverPreview(props.draft.cover.hash);if(ticket===generation)image.value=src;}
  if(props.draft.key.platform==='youtube'&&props.draft.cover){const status=await uploads.coverStatus(props.draft.key);if(ticket===generation)result.value=status==='applied'?'Thumbnail applied':status==='needs_attention'||status==='pending'?'Video kept · thumbnail needs attention':'';}
 }catch(e){if(ticket===generation)error.value=String(e);}
},{immediate:true});
async function save(data:string,kind:'image'|'frame',timeMs=0){
 busy.value=true;emit('busy',true);error.value='';
 try{const hash=await drafts.importCover(data);if(disposed)return;emit('change',{kind,hash,timeMs});open.value=false;}
 catch(e){error.value=String(e);}finally{busy.value=false;emit('busy',false);}
}
async function fileChanged(e:Event){
 const input=e.target as HTMLInputElement;const file=input.files?.[0];input.value='';if(!file)return;
 if(!['image/jpeg','image/png'].includes(file.type)||file.size>2*1024*1024){error.value='Choose a PNG or JPEG up to 2 MiB';return;}
 busy.value=true;emit('busy',true);
 try{const data=await new Promise<string>((resolve,reject)=>{const r=new FileReader();r.onload=()=>resolve(String(r.result));r.onerror=reject;r.readAsDataURL(file);});await save(data,'image');}catch{error.value='Could not read the image';}finally{busy.value=false;emit('busy',false);}
}
async function capture(){
 const video=player.value;if(!video||video.readyState<2||video.seeking)return;
 video.pause();
 const canvas=document.createElement('canvas');const scale=Math.min(1,1920/Math.max(video.videoWidth,video.videoHeight));
 canvas.width=Math.round(video.videoWidth*scale);canvas.height=Math.round(video.videoHeight*scale);
 try{const ctx=canvas.getContext('2d');if(!ctx)throw new Error('Frame capture unavailable');ctx.drawImage(video,0,0,canvas.width,canvas.height);await save(canvas.toDataURL('image/jpeg',.88),'frame',Math.round(video.currentTime*1000));}catch(e){error.value=String(e);}
}
async function apply(){busy.value=true;emit('busy',true);error.value='';try{result.value=await uploads.applyCover(props.draft.key,props.draft.revision);}catch(e){error.value=String(e);}finally{busy.value=false;emit('busy',false);}}
</script>
<template>
 <details class="cover-picker" @toggle="refreshStatus">
  <summary>Thumbnail <span>{{draft.cover?(draft.cover.kind==='frame'?'Video frame':'Custom image'):'Platform default'}}</span></summary>
  <div class="cover-body">
   <img v-if="image" :src="image" alt="Selected post thumbnail" class="cover-image">
   <div class="cover-options">
    <p class="muted">{{hint}}</p><p v-if="remote && !canImage" class="muted">An upload already exists. Its cover is fixed.</p>
    <div class="cover-actions" v-if="enabled && (!remote || canImage)">
     <UiButton size="compact" :disabled="disabled||busy" @click="open=true">Choose frame</UiButton>
     <UiButton v-if="canImage" size="compact" :disabled="disabled||busy" @click="picker?.click()">Choose image</UiButton>
     <input ref="picker" type="file" accept="image/png,image/jpeg" hidden @change="fileChanged">
     <UiButton v-if="draft.cover" size="compact" variant="quiet" :disabled="disabled||busy" @click="emit('change',null)">Remove</UiButton>
    </div>
    <p v-if="result" role="status">{{result}}</p>
    <UiButton v-if="canImage&&draft.cover&&['published','private'].includes(remote)&&result!=='Thumbnail applied'" size="compact" :busy="busy" title="Approve the saved post to apply its thumbnail without resending the video" :disabled="disabled||draft.approvedRevision!==draft.revision" @click="apply">Apply thumbnail</UiButton>
   </div>
  </div>
  <p v-if="error&&!open" class="notice error" role="alert">{{error}}</p>
 </details>
 <UiDialog v-if="open" title="Choose a video frame" :busy="busy" @close="open=false">
  <video ref="player" class="cover-video" :src="mediaURL(draft.key.episode,draft.videoId)" preload="metadata" controls playsinline></video>
  <p class="muted">Pause on the frame you want to use.</p>
  <UiButton variant="primary" :busy="busy" @click="capture">Use this frame</UiButton>
  <p v-if="error" role="alert" class="notice error">{{error}}</p>
 </UiDialog>
</template>
<style scoped>
.cover-picker{border-top:1px solid var(--border);padding:16px 0;margin:8px 0}.cover-picker summary{cursor:pointer;line-height:1.6}.cover-picker summary span{font-size:12px;color:var(--muted);margin-left:12px}.cover-body{display:flex;gap:20px;padding-top:16px;align-items:flex-start}.cover-image{width:80px;height:108px;object-fit:contain;background:#211a16;border-radius:4px;flex-shrink:0}.cover-options{min-width:0}.cover-options p{margin:0 0 12px;font-size:12px;line-height:1.6;max-width:46ch}.cover-actions{display:flex;gap:8px;flex-wrap:wrap}.cover-video{display:block;width:100%;height:min(48vh,380px);background:#211a16;object-fit:contain;border-radius:4px}.cover-picker .notice{margin-top:12px}
</style>
