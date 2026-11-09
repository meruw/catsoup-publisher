import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import vm from 'node:vm';
import ts from '../frontend/node_modules/typescript/lib/typescript.js';
import {ref,computed} from '../frontend/node_modules/vue/dist/vue.runtime.esm-bundler.js';

const source=await readFile(new URL('../frontend/src/components/BatchPublish.vue',import.meta.url),'utf8');
const script=source.match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1].replace(/^import .*;\r?$/gm,'');
const code=ts.transpileModule(script+'\n globalThis.check={sendAll,rows,confirmed,audience,error,finished};',{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.None}}).outputText;
for(const fail of ['tiktok','instagram','facebook','']){
 const profiles=['tiktok','instagram','facebook','youtube'].map(platform=>({platform,id:platform,language:'en',name:platform,api:true}));
 const saved=profiles.map(p=>({key:{platform:p.platform,account:'account:'+p.id,language:'en'},revision:1,approvedRevision:1,deliveryMode:'draft'}));
 profiles.push({platform:'x',id:'manual-x',language:'en',name:'Manual X',api:false});
 const sent=[];const queued=[];let mounted;
 const context={ref,computed,batchAccount:p=>p.platform!=='x'||!!p.api,transferPending:()=>false,modeLabel:()=>'Publish now',enqueueTransfers:requests=>queued.push(...requests),defineProps:()=>({profiles,saved,language:'en',episode:{},video:{}}),defineEmits:()=>()=>{},onMounted:f=>{mounted=f},accountKey:p=>'account:'+p.id,postApproved:()=>true,draftBlocker:()=>'',localDraft:()=>false,receiptBlocker:()=>'',uploads:{accounts:async()=>profiles,receipt:async()=>({state:''})},drafts:{get:async key=>saved.find(d=>d.key.platform===key.platform&&d.key.account===key.account)},transfer:async(d,p)=>{sent.push(p.platform);if(p.platform===fail)throw Error('Provider rejected fixture');return 'Published'}};
 vm.createContext(context);vm.runInContext(code,context);await mounted();context.check.confirmed.value=true;context.check.audience.value='general';await context.check.sendAll();
 for(const request of queued){try{await request.send()}catch{}}
 assert.deepEqual(sent,['tiktok','instagram','facebook','youtube']);
 assert.equal(context.check.finished.value,true);
 assert.equal(queued.length,4);
 assert(queued.every(r=>r.draft!==saved.find(d=>d.key.platform===r.draft.key.platform)),'queue owns snapshots');
}
console.log('PASS: a failed destination does not stop remaining approved sends; per-account outcomes are retained.');
