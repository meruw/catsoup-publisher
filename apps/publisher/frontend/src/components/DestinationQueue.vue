<script setup lang="ts">
import {onMounted,ref} from 'vue';
import {uploads} from '../services/uploads';
import type {Draft} from '../services/drafts';
const props=defineProps<{drafts:Draft[];selected:string;locked:boolean}>();
const emit=defineEmits<{select:[draft:Draft]}>();
const states=ref<Record<string,string>>({});
const key=(d:Draft)=>JSON.stringify(d.key);
const names:Record<string,string>={youtube:'YouTube',tiktok:'TikTok',instagram:'Instagram',x:'X'};
onMounted(async()=>{
 try{const profiles=await uploads.accounts()||[];
 await Promise.all(props.drafts.map(async d=>{
 const p=profiles.find(p=>p.platform===d.key.platform && p.language===d.key.language);
 if(!p){states.value[key(d)]='Connect account';return;}
 if(d.key.platform==='x' && !p.api){states.value[key(d)]='Local draft · Browser';return;}
 try{const r=await uploads.receipt(d.key,p.id);states.value[key(d)]=r.state==='sent'?(d.key.platform==='youtube'?'Private upload':d.key.platform==='tiktok'?'In inbox':'Published'):({private:'Private upload',published:'Published',inbox:'In inbox',sending:'Sending',publishing:'Publishing',check:'Check destination',check_publish:'Check visibility'} as Record<string,string>)[r.state] || 'Local draft';}catch{states.value[key(d)]='Status unavailable'}
 }));
 }catch{for(const d of props.drafts)states.value[key(d)]='Status unavailable'}
});
</script>
<template><div class="destination-queue" aria-label="Saved destination status"><table><thead><tr><th>Destination</th><th>Status</th><th><span class="sr-only">Open</span></th></tr></thead><tbody><tr v-for="d in drafts" :key="key(d)" :class="{selected:key(d)===selected}"><td><strong>{{names[d.key.platform]}}</strong><span>{{d.key.account}}</span></td><td>{{states[key(d)] || 'Checking…'}}</td><td><button :disabled="locked" :aria-label="`Open ${names[d.key.platform]} draft for ${d.key.account}`" :aria-current="key(d)===selected?'true':undefined" @click="emit('select',d)">{{key(d)===selected?'Selected':'Open'}}</button></td></tr></tbody></table></div></template>
<style scoped>
.destination-queue{overflow-x:auto;margin-bottom:24px}table{width:100%;border-collapse:collapse;font-size:var(--text-sm)}th{text-align:left;color:var(--muted);font-weight:400;padding:8px 12px;border-bottom:1px solid var(--line)}td{padding:12px;border-bottom:1px solid var(--line)}td strong{display:block;font-weight:500}td span{display:block;color:var(--secondary);font-size:var(--text-xs);margin-top:4px;overflow-wrap:anywhere}td:last-child{text-align:right}button{border:0;background:transparent;padding:8px;color:var(--accent-ink)}button:hover{background:var(--raised);border-radius:4px}.selected{background:var(--sidebar)}
</style>
