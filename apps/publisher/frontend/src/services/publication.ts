import {receiptSummary} from './transfers';
import {uploads,type Profile} from './uploads';
import {drafts,type Draft} from './drafts';
import type {Episode,Video} from './publisher';

// Browser-only X is an individual handoff, not a batch destination.
export const batchAccount=(p:Profile)=>p.platform!=='x' || !!p.api;
export type DeliveryMode='draft'|'publish';
export const accountKey=(profile?:Profile)=>profile ? `account:${profile.id}` : 'unassigned';
export const postApproved=(d?:Draft)=>!!d?.revision && d.approvedRevision===d.revision;
export function sendLabel(d:Draft|undefined,p?:Profile){
 if(!d)return 'Select a platform';
  if(d.key.platform==='tiktok')return 'Send to TikTok…';
 if(d.deliveryMode!=='publish')return d.key.platform==='youtube'?'Upload privately…':d.key.platform==='tiktok'?'Send to TikTok inbox…':'Keep local draft';
 return d.key.platform==='x'&&!p?.api?'Finish in X…':`Publish on ${d.key.platform==='youtube'?'YouTube':d.key.platform==='instagram'?'Instagram':d.key.platform==='facebook'?'Facebook':'X'}…`;
}
export const localDraft=(d:Draft)=>d.deliveryMode!=='publish' && !['youtube','tiktok'].includes(d.key.platform);
export function modeOptions(platform:string,profile?:Profile) {
  const draft=platform==='youtube'?'Private video':platform==='tiktok'?'TikTok inbox':'Local draft';
  return [{value:'draft' as DeliveryMode,label:draft},...(platform==='tiktok'&&!profile?.directPost?[]:[{value:'publish' as DeliveryMode,label:platform==='tiktok'?'Direct Post':platform==='x'&&!profile?.api?'Finish in X':'Publish now'}])];
}
export function modeLabel(platform:string,mode:string,profile?:Profile){if(platform==='tiktok'&&mode==='publish')return 'Direct Post';return modeOptions(platform,profile).find(o=>o.value===mode)?.label || 'Local draft';}
export function draftBlocker(d:Draft|undefined,p:Profile|undefined,episode:Episode,video:Video):string {
  if(!p)return 'Connect account';
  if(!d?.revision)return 'Save a draft';
  if(d.key.account!==accountKey(p))return 'Account changed';
  if(!d.title.trim()&&!d.description.trim()&&!d.hashtags.trim())return 'Add post text';
  if(d.key.platform==='tiktok'&&d.deliveryMode==='publish'&&!p.directPost)return 'Enable Direct Post and reconnect this account with video.publish';
  if(!postApproved(d))return 'Approve this account’s post';
  if(localDraft(d))return '';
  if(d.key.platform==='youtube'&&!d.title.trim())return 'Add a YouTube title';
  if(!episode.ready)return 'Delivery approval required';
  if(video.kind!=='final' || video.lang!==d.key.language || video.id!==d.videoId || video.sha256!==d.videoHash || episode.fingerprint!==d.deliveryFingerprint)return 'Link the current final video';
  return '';
}
export type TransferReceipt={state:string;reference:string;retryAllowed:boolean};
export function receiptBlocker(d:Draft,r:TransferReceipt):string {
  if(!r.state)return '';
  if(d.key.platform==='instagram' && ['failed','check'].includes(r.state))return '';
  if(d.key.platform==='youtube' && ['private','sent'].includes(r.state))return d.deliveryMode==='publish'?'':'Already uploaded privately';
  if(['published','inbox','sent','private','restricted'].includes(r.state))return 'Already sent';
  return 'Check the previous attempt';
}
// Both entry points use the same provider actions and durable backend receipts.
export async function transfer(d:Draft,p:Profile,r:TransferReceipt,confirmed:boolean,paid:boolean,audience:string,synthetic:boolean,allowPromotionRecovery=false):Promise<string> {
  if(d.key.account!==accountKey(p))throw new Error('Destination account changed');
  if(!confirmed)throw new Error('Confirm the destination');
  if(!postApproved(d))throw new Error('Approve this account’s current post');
  if(localDraft(d))return 'Kept locally';
  const recovering=allowPromotionRecovery&&p.platform==='youtube'&&d.deliveryMode==='publish'&&r.state==='check_publish';
  const blocked=receiptBlocker(d,r);if(blocked&&!recovering)throw new Error(blocked);
  if(p.platform==='x'&&!p.api){await drafts.copyUploadPath(d.key,d.revision,true);await uploads.openX(d.key,d.revision);return 'Opened in X — finish there';}
  const result=p.platform==='youtube'&&(['private','sent'].includes(r.state)||recovering)
    ? await uploads.publishYouTube(d.key,d.revision,p.id,true,audience,synthetic)
    : await uploads.sendMode(d.key,d.revision,p.id,d.deliveryMode||'draft',true,paid,audience,synthetic);
  return receiptSummary(result.state);
}

export function modeHint(platform:string,mode:string,p?:Profile){
 if(platform==='tiktok')return mode==='publish'?'Post directly with the privacy you choose in the confirmation dialog. TikTok approval and video.publish are required.':p?.directPost?'Send to the TikTok inbox and finish posting there.':'Inbox only. Enable Direct Post in Settings and reconnect this account with video.publish to unlock direct posting.';
 if(mode!=='publish')return platform==='youtube'?'Uploads a private YouTube video. It will not be public.':'Keeps the post on this PC. No video is sent to this platform.';
 if(platform==='x')return p?.api?'Publishes through your paid X API connection. You must confirm API charges before sending.':'Opens X with your text and copies the video path. You attach the video and finish posting in X.';
 return platform==='youtube'?'Makes the video public on this YouTube channel.':platform==='facebook'?'Publishes a Reel on this Facebook Page.':'Publishes a Reel on this Instagram account.';
}
