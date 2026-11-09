import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import vm from 'node:vm';
import ts from '../frontend/node_modules/typescript/lib/typescript.js';
import {reactive,computed} from '../frontend/node_modules/vue/dist/vue.runtime.esm-bundler.js';
const src=(await readFile(new URL('../frontend/src/services/transfers.ts',import.meta.url),'utf8')).replace(/^import .*;\r?$/gm,'').replace(/^export /gm,'');
const js=ts.transpileModule(src+'\nglobalThis.queue={enqueueTransfers,transferJobs,activeTransfers,cancelQueued,clearFinishedTransfers};',{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.None}}).outputText;
const c={reactive,computed};vm.createContext(c);vm.runInContext(js,c);const q=c.queue;
const deferred=()=>{let resolve,reject;const promise=new Promise((r,j)=>{resolve=r;reject=j});return {promise,resolve,reject}};
const request=(id,send)=>({draft:{key:{platform:'facebook',account:'account:'+id,episode:'episode',language:'es'},videoHash:'hash',title:'A short story'},profile:{name:id},mode:'Publish now',send});
const tick=()=>new Promise(r=>setImmediate(r));
const a=deferred(),b=deferred();const called=[];
q.enqueueTransfers([request('ES',()=>{called.push('ES');return a.promise}),request('EN',()=>{called.push('EN');return b.promise}),request('cancel',async()=>{throw Error('cancelled job ran')})]);
assert.equal(q.activeTransfers.value,3);assert.equal(q.transferJobs[0].state,'sending');assert.equal(q.transferJobs[1].state,'queued');
assert.throws(()=>q.enqueueTransfers([request('new',async()=>''),request('ES',async()=>'')]),/already queued/);assert.equal(q.transferJobs.length,3,'batch enqueue is atomic');
q.cancelQueued(q.transferJobs[0].id);assert.equal(q.transferJobs[0].state,'sending','cannot cancel an active external request');
q.cancelQueued(q.transferJobs[2].id);q.clearFinishedTransfers();assert.equal(q.transferJobs.length,2,'clear retains active and queued work');
a.reject(Error('uncertain provider result'));await tick();assert.equal(q.transferJobs[0].state,'attention');assert.equal(q.transferJobs[1].state,'sending');assert.deepEqual(called,['ES','EN']);
b.resolve('Published');await tick();assert.equal(q.activeTransfers.value,0);assert.equal(q.transferJobs[1].result,'Published');
q.clearFinishedTransfers();assert.equal(q.transferJobs.length,0);
assert.throws(()=>q.enqueueTransfers(Array.from({length:51},(_,i)=>request(String(i),async()=>''))),/queue is full/);assert.equal(q.transferJobs.length,0);
const providerRequest=(platform,id,send)=>{const r=request(id,send);r.draft.key.platform=platform;return r};
const held=deferred(),later=deferred();const starts=[];
q.enqueueTransfers([
 providerRequest('instagram','ES',()=>{starts.push('instagram/ES');return held.promise}),
 providerRequest('instagram','EN',()=>{starts.push('instagram/EN');return later.promise}),
 ...['youtube','facebook','tiktok','x'].map(p=>providerRequest(p,'ES',async()=>{starts.push(p);return 'Published'})),
]);
assert.equal(q.transferJobs.filter(j=>j.state==='sending').length,5,'at most one worker per supported provider');
await tick();
assert.equal(q.activeTransfers.value,2,'all other providers finish while Instagram remains pending');
assert.equal(q.transferJobs.find(j=>j.platform==='instagram'&&j.account==='EN').state,'queued');
q.enqueueTransfers([providerRequest('youtube','EN',async()=>{starts.push('youtube/EN');return 'Published'})]);
await tick();assert(starts.includes('youtube/EN'),'a finished provider can start again while another is pending');
const count=q.transferJobs.length;
assert.throws(()=>q.enqueueTransfers([request('atomic',async()=>''),providerRequest('unknown','ES',async()=>'')]),/Unsupported/);
assert.equal(q.transferJobs.length,count,'unsupported providers reject the whole batch');
held.reject(Error('processing timeout'));await tick();assert(starts.includes('instagram/EN'),'same-provider FIFO resumes after failure');
later.resolve('Published');await tick();assert.equal(q.activeTransfers.value,0);
q.enqueueTransfers([request('sync-error',()=>{throw Error('sync failure')}),request('after-error',async()=>'Published')]);
await tick();assert.equal(q.transferJobs.at(-2).state,'attention');assert.equal(q.transferJobs.at(-1).state,'done');
q.clearFinishedTransfers();assert.equal(q.transferJobs.length,0);
console.log('PASS: independent bounded provider queues, per-provider FIFO, pending-provider isolation, atomic duplicate protection, cancellation, failure recovery and history clearing.');
