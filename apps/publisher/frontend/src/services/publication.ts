import {uploads,type Profile} from './uploads';
import {drafts,type Draft} from './drafts';
import type {Episode,Video} from './publisher';

export type DeliveryMode='draft'|'publish';
export const accountKey=(profile?:Profile)=>profile ? `account:${profile.id}` : 'unassigned';
export const localDraft=(d:Draft)=>d.deliveryMode!=='publish' && !['youtube','tiktok'].includes(d.key.platform);
export function modeOptions(platform:string,profile?:Profile) {
  const draft=platform==='youtube'?'Private video':platform==='tiktok'?'TikTok inbox':'Local draft';
  return [{value:'draft' as DeliveryMode,label:draft},...(platform==='tiktok'?[]:[{value:'publish' as DeliveryMode,label:platform==='x'&&!profile?.api?'Finish in X':'Publish now'}])];
}
export function modeLabel(platform:string,mode:string,profile?:Profile){return modeOptions(platform,profile).find(o=>o.value===mode)?.label || 'Local draft';}
export function draftBlocker(d:Draft|undefined,p:Profile|undefined,episode:Episode,video:Video,reviewed:boolean):string {
  if(!p)return 'Connect account';
  if(!d?.revision)return 'Save a draft';
  if(d.key.account!==accountKey(p))return 'Account changed';
  if(!d.title.trim()&&!d.description.trim()&&!d.hashtags.trim())return 'Add post text';
  if(localDraft(d))return '';
  if(d.key.platform==='youtube'&&!d.title.trim())return 'Add a YouTube title';
  if(!episode.ready)return 'Delivery approval required';
  if(!reviewed)return 'Video review required';
  if(video.kind!=='final' || video.lang!==d.key.language || video.id!==d.videoId || video.sha256!==d.videoHash || episode.fingerprint!==d.deliveryFingerprint)return 'Link the current final video';
  return '';
}
export type TransferReceipt={state:string;reference:string;retryAllowed:boolean};
export function receiptBlocker(d:Draft,r:TransferReceipt):string {
  if(!r.state)return '';
  if(d.key.platform==='youtube' && ['private','sent'].includes(r.state))return d.deliveryMode==='publish'?'':'Already uploaded privately';
  if(['published','inbox','sent'].includes(r.state))return 'Already sent';
  return 'Check the previous attempt';
}
// Both entry points use the same provider actions and durable backend receipts.
export async function transfer(d:Draft,p:Profile,r:TransferReceipt,confirmed:boolean,paid:boolean,audience:string,synthetic:boolean,allowPromotionRecovery=false):Promise<string> {
  if(d.key.account!==accountKey(p))throw new Error('Destination account changed');
  if(!confirmed)throw new Error('Confirm the destination');
  if(localDraft(d))return 'Kept locally';
  const recovering=allowPromotionRecovery&&p.platform==='youtube'&&d.deliveryMode==='publish'&&r.state==='check_publish';
  const blocked=receiptBlocker(d,r);if(blocked&&!recovering)throw new Error(blocked);
  if(p.platform==='x'&&!p.api){await drafts.copyUploadPath(d.key,d.revision,true);await uploads.openX(d.key,d.revision);return 'Opened in X — finish there';}
  const result=p.platform==='youtube'&&(['private','sent'].includes(r.state)||recovering)
    ? await uploads.publishYouTube(d.key,d.revision,p.id,true,audience,synthetic)
    : await uploads.sendMode(d.key,d.revision,p.id,d.deliveryMode||'draft',true,paid,audience,synthetic);
  return result.state==='inbox'?'Sent to inbox':result.state==='private'?'Uploaded privately':'Published';
}
