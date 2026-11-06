// This endpoint has no backend. Only the originating browser tab may resume a login.
const redirect = location.origin + location.pathname;
const slot = 'catsoup-meta-attempt:' + location.pathname;
const message = document.getElementById('message');
function show(heading, detail) {
  const title = document.getElementById('heading');
  if (title) title.textContent = heading;
  message.textContent = detail;
}
const encode = bytes => btoa(String.fromCharCode(...bytes)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
const decode = text => Uint8Array.from(atob(text.replace(/-/g, '+').replace(/_/g, '/')), c => c.charCodeAt(0));
const paired = state => /^[a-zA-Z0-9_-]{26,128}$/.test(state || '');

async function run() {
  const start = new URLSearchParams(location.hash.slice(1));
  const result = new URLSearchParams(location.search);
  history.replaceState(null, '', location.pathname);
  if (location.protocol !== 'https:' || location.href !== redirect) throw Error('Unexpected sign-in address.');
  if (start.has('key') && !result.size) {
    if (start.has('return_uri') && start.get('return_uri') !== redirect) throw Error('The return page moved. Check the exact callback URL in Settings and Meta.');
    const app = start.get('app'), config = start.get('config');
    const state = start.get('state'), key = start.get('key'), port = Number(start.get('port'));
    if (!/^\d{1,24}$/.test(app || '') || !/^\d{1,24}$/.test(config || '') || !paired(state) ||
        !/^[A-Za-z0-9_-]{43}$/.test(key || '') || !Number.isInteger(port) || port < 1024 || port > 65535)
      throw Error('This sign-in link is incomplete. Start again from Publisher.');
    sessionStorage.setItem(slot, JSON.stringify({state,key,port,until:Date.now()+180_000}));
    const url = new URL('https://www.facebook.com/v25.0/dialog/oauth');
    url.search = new URLSearchParams({client_id:app,config_id:config,state,response_type:'code',
      override_default_response_type:'true',redirect_uri:redirect}).toString();
    show('Continue on Facebook', 'Choose the Pages linked to your Instagram accounts.');
    location.replace(url.href);
    return;
  }
  if (!result.size) {
    show('Start in Publisher', 'Open Accounts in the desktop app, choose ES or EN, then connect Instagram.');
    return;
  }
  const attempt = JSON.parse(sessionStorage.getItem(slot) || 'null');
  if (!attempt || attempt.until <= Date.now() || !paired(result.get('state')) || result.get('state')!==attempt.state)
    throw Error('There is no matching sign-in in this tab. Start again from Publisher.');
  const code = result.get('code');
  if ((!code && !result.has('error')) || (code && code.length > 8192)) throw Error('Invalid response. Start again from Publisher.');
  sessionStorage.removeItem(slot);
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const key = await crypto.subtle.importKey('raw',decode(attempt.key),'AES-GCM',false,['encrypt']);
  const bytes = await crypto.subtle.encrypt({name:'AES-GCM',iv,additionalData:new TextEncoder().encode(attempt.state)},
    key,new TextEncoder().encode(result.has('error')?'':code));
  const sealed = new Uint8Array(iv.length+bytes.byteLength);
  sealed.set(iv);sealed.set(new Uint8Array(bytes),iv.length);
  // Top-level navigation avoids cross-origin fetch and never delegates the host to input.
  const local = new URL('http://127.0.0.1:'+attempt.port+'/callback/');
  local.search = new URLSearchParams({state:attempt.state,reply:encode(sealed)}).toString();
  show('Returning to Publisher', 'The desktop app will check which Instagram profile belongs to the selected Page.');
  location.replace(local.href);
}
run().catch(error => show('Sign-in interrupted', error instanceof Error && !('code' in error) ? error.message : 'Open Accounts in Publisher and start sign-in again.'));
