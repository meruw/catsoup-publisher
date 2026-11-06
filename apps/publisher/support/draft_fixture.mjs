// Test-only desktop transport. Delays/failures exercise real Vue save behavior.
// Production stores drafts in Go/SQLite, never localStorage.
export const draftFixture = `
 const storage = () => JSON.parse(localStorage.getItem('draft-fixture') || '{}');
 const keyOf = key => JSON.stringify(key);
 window.__draftWrites = [];
 window.__draftDelay = 0;
 window.__draftFail = false;
 window.__draftConflict = false;
 export const drafts = {
  review: async (episode,language,videoId) => JSON.parse(localStorage.getItem('review-'+episode+language+videoId) || 'null') || {episode,language,videoId,fingerprint:'',videoHash:'',state:'pending',notes:'',revision:0,updatedAt:''},
  saveReview: async (input,confirmed) => {if(input.state==='reviewed' && !confirmed)throw Error('Confirm review');const value={...input,revision:input.revision+1,updatedAt:new Date().toISOString()};localStorage.setItem('review-'+input.episode+input.language+input.videoId,JSON.stringify(value));return value;},
  list: async (episode, language) => Object.values(storage()).filter(d => d.key.episode === episode && d.key.language === language),
  get: async key => storage()[keyOf(key)] || {key,title:'',description:'',hashtags:'',videoId:'',videoHash:'',deliveryFingerprint:'',revision:0,updatedAt:''},
  save: async input => {
   window.__draftWrites.push(JSON.parse(JSON.stringify(input)));
   await new Promise(resolve => setTimeout(resolve, window.__draftDelay));
   if (window.__draftFail) throw new Error('Simulated disk failure');
   const all = storage(), old = all[keyOf(input.key)];
   if (window.__draftConflict || (old?.revision || 0) !== input.revision) throw new Error('draft changed in another window; reload the saved draft before editing');
   const attach = input.attachVideo || !old;
   const result = {...input, videoId: attach ? input.videoId : old.videoId,
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
