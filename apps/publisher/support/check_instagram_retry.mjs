import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import vm from 'node:vm';
import ts from '../frontend/node_modules/typescript/lib/typescript.js';

const source=(await readFile(new URL('../frontend/src/services/publication.ts',import.meta.url),'utf8')).replace(/^import .*;\r?$/gm,'').replace(/^export /gm,'');
const code=ts.transpileModule(source+'\nglobalThis.check={receiptBlocker,transfer};',{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.None}}).outputText;
const calls=[];
const context={receiptSummary:s=>s,uploads:{sendMode:async(...args)=>{calls.push(args);return {state:'published'}}},drafts:{}};
vm.createContext(context);vm.runInContext(code,context);
for(const platform of ['instagram','facebook','youtube','tiktok']){
 const draft={key:{platform,language:'es',account:'account:es'},revision:3,approvedRevision:3,deliveryMode:'publish'};
 const profile={platform,id:'es',language:'es'};
 for(const state of ['failed','sending','check','published','inbox']){
  const receipt={state,reference:'container',retryAllowed:false};
  if(platform==='instagram'&&['failed','check'].includes(state)){
   assert.equal(context.check.receiptBlocker(draft,receipt),'');
   assert.equal(await context.check.transfer(draft,profile,receipt,true,false,'',false),'published');
   await assert.rejects(()=>context.check.transfer({...draft,approvedRevision:2},profile,receipt,true,false,'',false),/Approve/);
   await assert.rejects(()=>context.check.transfer(draft,profile,receipt,false,false,'',false),/Confirm/);
   await assert.rejects(()=>context.check.transfer(draft,{...profile,id:'en'},receipt,true,false,'',false),/Destination/);
  }else{
   assert.notEqual(context.check.receiptBlocker(draft,receipt),'');
   await assert.rejects(()=>context.check.transfer(draft,profile,receipt,true,false,'',false));
  }
 }
}
assert.equal(calls.length,2,'only the approved, confirmed failed Instagram destination may retry');
console.log('PASS: failed Instagram can retry; uncertain/completed receipts, other platforms, stale approvals and changed accounts remain blocked.');
