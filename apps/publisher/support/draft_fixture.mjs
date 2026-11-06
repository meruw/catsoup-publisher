// Test-only desktop transport. Delays/failures exercise real Vue save behavior.
// Production stores drafts in Go/SQLite, never localStorage.
export const draftFixture = `
 const storage = () => JSON.parse(localStorage.getItem('draft-fixture') || '{}');
 const keyOf = key => JSON.stringify(key);
 window.__draftWrites = [];
 window.__draftDelay = 0;
 window.__draftFail = false;
 window.__draftConflict = false;
 const prefKey=p=>p.platform+'/'+p.language+'/'+p.accountId;
 const getDefault=key=>JSON.parse(localStorage.getItem('defaults-fixture')||'{}')[prefKey({platform:key.platform,language:key.language,accountId:key.account.replace('account:','')})]?.mode||'draft';
 export const drafts = {
  groups:async f=>{const all=await drafts.statuses({episode:'',search:'',platform:'',language:'',state:'',offset:0});const matches=await drafts.statuses({...f,offset:0});const ids=[...new Set(matches.items.map(r=>r.key.episode))];const items=ids.map(episode=>{const rows=all.items.filter(r=>r.key.episode===episode);const counts={};for(const r of rows)counts[r.state]=(counts[r.state]||0)+1;return {episode,videoId:rows.map(r=>r.videoId).sort()[0],title:rows[0].title,languages:[...new Set(rows.map(r=>r.key.language))].join(','),platforms:new Set(rows.map(r=>r.key.platform)).size,total:rows.length,counts,updatedAt:rows[0].updatedAt}});return {items:items.slice(f.offset,f.offset+20),more:items.length>f.offset+20}},
  statuses:async f=>{const profiles=JSON.parse(localStorage.getItem('account-fixture')||'[]');const all=(window.__statusRows||Object.values(storage()).map(d=>({key:d.key,title:d.title,videoId:d.videoId,mode:d.deliveryMode,accountName:profiles.find(p=>'account:'+p.id===d.key.account&&p.platform===d.key.platform)?.name||d.key.account,connected:profiles.some(p=>'account:'+p.id===d.key.account&&p.platform===d.key.platform),state:d.approvedRevision===d.revision?'approved':'draft',reference:'',updatedAt:d.updatedAt}))).filter(r=>(!f.episode||r.key.episode===f.episode)&&(!f.platform||r.key.platform===f.platform)&&(!f.language||r.key.language===f.language)&&(!f.state||({unsent:["draft","approved"],active:["sending"],delivered:["published","private","restricted","inbox"],attention:["failed","check"]}[f.state]||[f.state]).includes(r.state))&&(!f.search||f.episodeMatches?.includes(r.key.episode)||(r.title+' '+r.accountName).toLowerCase().includes(f.search.toLowerCase())));return {items:all.slice(f.offset,f.offset+50),more:all.length>f.offset+50};},
  publishingDefaults:async()=>Object.values(JSON.parse(localStorage.getItem('defaults-fixture')||'{}')),
  setPublishingDefault:async p=>{if(window.__defaultFail)throw Error('Simulated preference failure');const all=JSON.parse(localStorage.getItem('defaults-fixture')||'{}');all[prefKey(p)]=p;localStorage.setItem('defaults-fixture',JSON.stringify(all));},
  review: async (episode,language,videoId) => JSON.parse(localStorage.getItem('review-'+episode+language+videoId) || 'null') || {episode,language,videoId,fingerprint:'',videoHash:'',state:'pending',notes:'',revision:0,updatedAt:''},
  saveReview: async (input,confirmed) => {if(input.state==='reviewed' && !confirmed)throw Error('Confirm review');const value={...input,revision:input.revision+1,updatedAt:new Date().toISOString()};localStorage.setItem('review-'+input.episode+input.language+input.videoId,JSON.stringify(value));return value;},
  approve: async (key,revision,approved)=>{const all=storage(),d=all[keyOf(key)];if(!d||d.revision!==revision||!key.account.startsWith('account:'))throw Error('Approval conflict');d.approvedRevision=approved?revision:0;localStorage.setItem('draft-fixture',JSON.stringify(all));return d;},
  list: async (episode, language) => Object.values(storage()).filter(d => d.key.episode === episode && d.key.language === language),
  get: async key => storage()[keyOf(key)] || {key,approvedRevision:0,deliveryMode:getDefault(key),title:'',description:'',hashtags:'',videoId:'',videoHash:'',deliveryFingerprint:'',revision:0,updatedAt:''},
  save: async input => {
   window.__draftWrites.push(JSON.parse(JSON.stringify(input)));
   await new Promise(resolve => setTimeout(resolve, window.__draftDelay));
   if (window.__draftFail) throw new Error('Simulated disk failure');
   const all = storage(), old = all[keyOf(input.key)];
   if (window.__draftConflict || (old?.revision || 0) !== input.revision) throw new Error('draft changed in another window; reload the saved draft before editing');
   const attach = input.attachVideo || !old;
   const result = {...input, approvedRevision:0, videoId: attach ? input.videoId : old.videoId,
    videoHash: attach ? window.__fixtureEpisode.videos.find(v=>v.id===input.videoId).sha256 : old.videoHash,
    deliveryFingerprint: attach ? window.__fixtureEpisode.fingerprint : old.deliveryFingerprint,
    revision: input.revision + 1, updatedAt: new Date().toISOString()};
   all[keyOf(input.key)] = result; localStorage.setItem('draft-fixture',JSON.stringify(all));return result;
  },
  copy: async (key,part) => {const d=storage()[keyOf(key)];window.__copied=part==='title'?d.title:(d.description+'\\n\\n'+d.hashtags).trim();},
  reveal: async (episode,video) => {window.__revealed = {episode,video};},
  copyUploadPath: async (key,revision,confirmed) => {window.__uploadPathRequest = {key,revision,confirmed};},
  onCloseRequested: callback => {window.__requestClose=callback;return()=>{delete window.__requestClose;};},
  enableCloseGuard: async () => {},
  confirmClose: async () => {window.__closeConfirmed=true;},
 };
`;
