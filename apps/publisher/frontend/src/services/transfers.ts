import {computed,reactive} from 'vue';
import type {Draft} from './drafts';
import type {Profile} from './uploads';

export type TransferJob={id:number;destination:string;episode:string;title:string;platform:string;language:string;account:string;mode:string;state:'queued'|'sending'|'done'|'attention'|'cancelled';result:string;started:number;finished:number};
export type TransferRequest={draft:Draft;profile:Profile;mode:string;send:()=>Promise<string>};
// Session-owned work survives component unmounts. Only receipts persist across restarts.
export const transferJobs=reactive<TransferJob[]>([]);
export const activeTransfers=computed(()=>transferJobs.filter(j=>j.state==='queued'||j.state==='sending').length);
export const transferUpdates=computed(()=>transferJobs.reduce((n,j)=>n+(j.finished?1:0),0));
let serial=0;
// One worker per provider: a processing wait must not hold other platforms.
// Serial sends within each provider also avoid concurrent token refreshes.
const runningPlatforms=new Set<string>();
const platforms=new Set(['youtube','instagram','facebook','tiktok','x']);
const actions=new Map<number,()=>Promise<string>>();
const destination=(d:Draft)=>JSON.stringify([d.key.platform,d.key.account,d.videoHash]);
export const transferPending=(d:Draft)=>transferJobs.some(j=>j.destination===destination(d)&&(j.state==='queued'||j.state==='sending'));
export function enqueueTransfers(requests:TransferRequest[]) {
 if(!requests.length)return;
 if(activeTransfers.value+requests.length>50)throw Error('The upload queue is full. Wait for a transfer to finish.');
 const keys=new Set<string>();
 for(const r of requests){if(!platforms.has(r.draft.key.platform))throw Error('Unsupported upload platform');const key=destination(r.draft);if(keys.has(key)||transferPending(r.draft))throw Error('This video is already queued for that account. Check Uploads.');keys.add(key);}
 // Validate the whole batch before appending anything.
 for(const {draft:d,profile:p,mode,send} of requests){
  const job:TransferJob={id:++serial,destination:destination(d),episode:d.key.episode,title:d.title||d.key.episode,platform:d.key.platform,language:d.key.language,account:p.name,mode,state:'queued',result:'',started:0,finished:0};
  transferJobs.push(job);actions.set(job.id,send);
 }
 trimHistory();for(const platform of new Set(requests.map(r=>r.draft.key.platform)))void drain(platform);
}
function trimHistory(){while(transferJobs.length>100){const i=transferJobs.findIndex(j=>j.finished);if(i<0)break;transferJobs.splice(i,1);}}
async function drain(platform:string){
 if(runningPlatforms.has(platform))return;runningPlatforms.add(platform);
 try{
  for(;;){const job=transferJobs.find(j=>j.platform===platform&&j.state==='queued');if(!job)break;
   job.state='sending';job.started=Date.now();
   try{job.result=await actions.get(job.id)!();job.state='done';}
   catch(e){job.state='attention';job.result=String(e).replace(/^(RuntimeError|Error):\s*/,'');}
   finally{job.finished=Date.now();actions.delete(job.id);}
  }
 }finally{runningPlatforms.delete(platform);trimHistory();}
}
export function cancelQueued(id:number){const j=transferJobs.find(j=>j.id===id&&j.state==='queued');if(j){j.state='cancelled';j.result='Not sent';j.finished=Date.now();actions.delete(id);}}
export function clearFinishedTransfers(){for(let i=transferJobs.length-1;i>=0;i--)if(transferJobs[i].finished)transferJobs.splice(i,1);}
export function receiptSummary(state:string){return ({inbox:'Sent to inbox',private:'Uploaded privately',restricted:'Posted with limited visibility',published:'Published'})[state]||'Check result in Posts';}
