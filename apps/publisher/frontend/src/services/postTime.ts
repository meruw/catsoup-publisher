import type {PostStatus} from './drafts';
export const localTimeZone=Intl.DateTimeFormat().resolvedOptions().timeZone;
export function postDate(value:string):string {
 const at=new Date(value);
 return value&&!Number.isNaN(at.getTime())?at.toLocaleString([], {year:'numeric',month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'}):'Unavailable';
}
export function postTime(post:PostStatus):{label:string;at:string} {
 if(post.publishedAt)return {label:'Published',at:post.publishedAt};
 if(post.confirmedAt){
  const label:Record<string,string>={published:'Publication confirmed',restricted:'Publication confirmed',private:'Private upload confirmed',inbox:'Sent to inbox',web_ready:'Prepared in browser'};
  if(label[post.receiptState])return {label:label[post.receiptState]!,at:post.confirmedAt};
 }
 if(['published','restricted'].includes(post.state))return {label:'Publication time unavailable',at:''};
 if(post.startedAt)return {label:'Latest attempt started',at:post.startedAt};
 return {label:'',at:''};
}
