import {BrowserUsername,OpenBrowser,VerifyBrowser,Prepare,PreparationStatus,HasPreparedBrowser} from '../../bindings/catsoup/publisher/tiktokwebservice';
import {updateTransferProgress} from './transfers';
import type {Draft} from './drafts';
export const tikTokWeb={username:BrowserUsername,open:OpenBrowser,verify:VerifyBrowser,hasPreparedBrowser:HasPreparedBrowser};
export async function prepareTikTokWeb(draft:Draft,id:string){
 let polling=false;
 const timer=setInterval(async()=>{
  if(polling)return;polling=true;
  try{const text=await PreparationStatus(draft.key.language,id);if(text)updateTransferProgress(draft,text);}catch{/* The final call reports failure. */}finally{polling=false;}
 },1500);
 try{return await Prepare(draft.key,draft.revision,id);}finally{clearInterval(timer);}
}
