import type {Entry} from './publisher';
import type {Completion} from './drafts';
export function pendingDeliveries(entries:Entry[],completed:Completion[]):Entry[]{
 const done=new Set(completed.map(c=>JSON.stringify([c.episode,c.fingerprint,c.language])));
 return entries.filter(e=>!e.valid||!e.fingerprint||!e.languages?.length||!e.languages.every(lang=>done.has(JSON.stringify([e.id,e.fingerprint,lang]))));
}
