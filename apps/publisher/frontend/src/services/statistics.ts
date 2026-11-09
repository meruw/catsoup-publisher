import {uploads} from './uploads';
import type {PostStatus} from './drafts';
type Counts=Awaited<ReturnType<typeof uploads.statistics>>;
const cache=new Map<string,{expires:number;value:Counts}>();
const pending=new Map<string,Promise<Counts>>();
export const statisticsKey=(p:PostStatus)=>JSON.stringify([p.key.episode,p.key.language,p.key.platform,p.key.account,p.reference,p.videoId,p.updatedAt,p.state,p.connected]);
export function statisticsNote(p:PostStatus):string {
 if(p.key.platform==='x')return 'X statistics are not requested because its API can incur charges.';
 if(p.key.platform==='tiktok')return 'TikTok statistics are not connected. An inbox receipt does not identify the final published video.';
 if(!p.connected)return 'Reconnect this account to read statistics.';
 if(!p.reference || !['published','private','sent'].includes(p.state))return 'Statistics are available after a confirmed upload.';
 return '';
}
export function readStatistics(p:PostStatus,force=false):Promise<Counts>{
 const key=statisticsKey(p),cached=cache.get(key);
 if(!force && cached && cached.expires>Date.now())return Promise.resolve(cached.value);
 const active=pending.get(key);if(active)return active;
 const request=uploads.statistics({...p.key}).then(value=>{
  cache.delete(key);cache.set(key,{expires:Date.now()+300_000,value});
  while(cache.size>100)cache.delete(cache.keys().next().value!);
  return value;
 }).finally(()=>pending.delete(key));
 pending.set(key,request);return request;
}
