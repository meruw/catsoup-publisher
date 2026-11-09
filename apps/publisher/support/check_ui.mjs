// Local visual/regression harness. The bridge is a fixture; this is not a native Wails test.
// Uses generated demo MP4s without changing them. No social accounts or remote services.
import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import assert from 'node:assert/strict';
import { draftFixture } from './draft_fixture.mjs';
import { diagnosticsFixture, checkDiagnostics } from './diagnostics_fixture.mjs';
import { createServer } from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import { chromium } from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';

const root = fileURLToPath(new URL('../../../', import.meta.url));
const frontend = path.join(root, 'apps/publisher/frontend');
const output = path.join(root, 'work/publisher-ui-check');
await mkdir(output, { recursive: true });
const deliveryFolder = path.join(root, 'deliveries/catsoup-demo');
const manifest = JSON.parse(await readFile(path.join(deliveryFolder, 'manifest.json'), 'utf8'));
const episode = {
  id: manifest.episode_id, title: manifest.internal_title, version: manifest.version,
  status: manifest.status, reason: manifest.not_publishable_reason, ready: false,
  languages: manifest.languages, videos: manifest.videos, fingerprint: 'visual-fixture', warnings: [],
};
const fixture = `
 const episode = ${JSON.stringify(episode)};
 window.__fixtureEpisode = episode;
 export const publisher = {
  list: async () => window.__empty ? [] : [{id:episode.id,title:episode.title,valid:true,problem:''}],
  get: async () => episode,
 };
 export const mediaURL = (episode,video) => '/media?'+new URLSearchParams({episode,video});
`;
const server = await createServer({
  configFile: false, root: frontend,
  plugins: [{ name: 'test-only-bridge', enforce: 'pre', load(id) {
    if (id.replaceAll('\\', '/').endsWith('/src/services/drafts.ts')) return draftFixture;
    if (id.replaceAll('\\', '/').endsWith('/src/services/diagnostics.ts')) return diagnosticsFixture;
    if (id.replaceAll('\\', '/').endsWith('/src/services/uploads.ts')) return `
      const none = {state:'',reference:'',retryAllowed:false};
      const receipts = new Map(); window.__clearReceipts=()=>receipts.clear();window.__setReceipt=(id,r)=>receipts.set(id,r);window.__sendCalls=[]; const profiles = () => JSON.parse(localStorage.getItem('account-fixture') || '[]');
      export const uploads = {
        statistics:async key=>{window.__statisticsReads=(window.__statisticsReads||0)+1;if(window.__statisticsDelay){window.__statisticsActive=(window.__statisticsActive||0)+1;window.__statisticsPeak=Math.max(window.__statisticsPeak||0,window.__statisticsActive);await new Promise(r=>setTimeout(r,50));window.__statisticsActive--;}if(key.account===window.__statisticsFail)throw Error('Fixture read failed');return {items:[{label:'Likes',value:key.language==='es'?'42':'18'},{label:'Comments',value:'3'}],updatedAt:new Date().toISOString(),note:''}},
        tikTokDirectEnabled:async()=>!!window.__tikTokDirect,
        setTikTokDirectEnabled:async enabled=>{window.__tikTokDirect=enabled},
        tikTokPolicy:async policy=>{window.__tikTokPolicy=policy},
        tikTokCreator:async()=>({creator_username:'catsoup_es',creator_nickname:'CATsoup ES',privacy_level_options:['SELF_ONLY','PUBLIC_TO_EVERYONE'],comment_disabled:false,duet_disabled:true,stitch_disabled:false,max_video_post_duration_sec:300}),
        sendTikTok:async(key,revision,id,confirmed,options)=>{window.__sendCalls.push({key,revision,id,confirmed,options});const receipt={state:options.privacy==='SELF_ONLY'?'private':'published',reference:'direct-fixture',retryAllowed:false};receipts.set(id,receipt);return receipt;},
        metaCallback: async () => localStorage.getItem('meta-callback-fixture') || '',
        enableHostedMetaCallback: async () => {const url='https://publisher.meruw.dev/auth/meta/';localStorage.setItem('meta-callback-fixture',url);return url;},
        setMetaCallback: async url => {localStorage.setItem('meta-callback-fixture',url);return url;},
        credentialStatuses: async () => Object.fromEntries(['youtube','tiktok','instagram','x'].map(p=>[p,localStorage.getItem(p+'-client-fixture')==='ready'?'saved':'missing'])),
        clientReady: async p => localStorage.getItem(p+'-client-fixture')==='ready',
        googleSetup: async step => { window.__googleSetupStep=step; },
        xSetup: async step => { window.__xSetupStep=step; },
        configure: async (platform,raw) => {
          if(platform==='x') {
            const data=JSON.parse(raw);
            if(!data.client_id || data.redirect_uri!=='http://127.0.0.1:48173/callback/' || data.client_secret) throw Error('Invalid X Native App fixture');
            localStorage.setItem('x-client-fixture','ready');return;
          }
          if(platform==='tiktok') {
            const data=JSON.parse(raw);
            if(data.client_key!=='fixture-key' || data.client_secret!=='fixture-secret') throw Error('Invalid fixture credentials');
            localStorage.setItem('tiktok-client-fixture','ready');
            return;
          }
          if(platform==='instagram') {
            const data=JSON.parse(raw);
            if(!data.app_id || !data.app_secret || !data.config_id) throw Error('Invalid Meta fixture');
            localStorage.setItem('instagram-client-fixture','ready');
            return;
          }
          const client=JSON.parse(raw).installed;
          if(platform!=='youtube'||!client?.client_id||!client.client_secret) throw Error('select Google Desktop app JSON');
          localStorage.setItem('youtube-client-fixture','ready');
        },
        accounts: async () => profiles(),
        linkX: async (language,expected,handle,paid) => {
          if(!paid || !handle) throw Error('Missing paid consent or username');
          const p={platform:'x',language,id:'x-'+language,name:handle,api:true};const rows=profiles().filter(row=>row.platform!=='x'||row.language!==language);localStorage.setItem('account-fixture',JSON.stringify([...rows,p]));return p;
        },
        link: async (platform,language) => {const p={platform,language,id:platform+'-'+language,name:'Fixture '+language};const rows=profiles().filter(row=>row.platform!==platform||row.language!==language);localStorage.setItem('account-fixture',JSON.stringify([...rows,p]));return p;},
        unlink: async (platform,language) => localStorage.setItem('account-fixture',JSON.stringify(profiles().filter(row=>row.platform!==platform||row.language!==language))),
        receipt: async (_,id) => receipts.get(id)||none,
        sendX: async (key,revision,id,confirmed,paid) => {
          if(!confirmed || !paid || key.platform!=='x' || id!=='x-'+key.language || revision<1) throw Error('Unsafe X upload');
          window.__xPaidRequest={key,revision,id,confirmed,paid};const receipt={state:'sent',reference:'https://x.com/i/status/123',retryAllowed:false};receipts.set(id,receipt);return receipt;
        },
        sendMode: async (key, revision, id, mode, confirmed, paid, audience, synthetic) => {
          window.__sendCalls.push({key,mode,id});
          await new Promise(resolve=>setTimeout(resolve,window.__sendDelay||0));
          if(window.__failPlatform===key.platform)throw Error('Simulated transfer failure');
          if(key.platform==='x') {if(!paid)throw Error('Missing paid consent');window.__xPaidRequest={key,revision,id,confirmed,paid};}
          if(key.platform==='youtube' && !audience)throw Error('Missing audience');
          if (!confirmed || id !== key.platform+'-'+key.language || key.episode !== 'catsoup-handoff-fixture' || revision < 1) throw Error('unsafe fixture upload');
          const receipt = {state:key.platform==='youtube'&&mode==='draft'?'private':key.platform==='tiktok'?'inbox':'published',reference:key.platform==='x'?'https://x.com/i/status/123':'fixture-post',retryAllowed:false};receipts.set(id,receipt);return receipt;
        },
        publishYouTube:async(key,revision,id,confirmed,audience)=>{if(!confirmed || !audience)throw Error('Unsafe promotion');window.__promotion={id,reference:receipts.get(id).reference};const r={state:'published',reference:receipts.get(id).reference,retryAllowed:false};receipts.set(id,r);return r;},
        openX: async key => {window.__xComposer=key.platform;},
      };`;
    if (id.replaceAll('\\', '/').endsWith('/src/services/publisher.ts')) return fixture;
    if (id.replaceAll('\\', '/').endsWith('/src/services/window.ts')) return `
      let max = false, listener;
      window.__windowCalls = [];
      export const desktopWindow = {
        minimise: async () => { window.__windowCalls.push('minimise'); },
        close: async () => { window.__windowCalls.push('close'); },
        toggleMaximise: async () => { max = !max; window.__windowCalls.push('toggle'); listener?.(max); },
        isMaximised: async () => max,
        onMaximisedChange: callback => { listener = callback; return () => { listener = undefined; }; },
      };
    `;
  } }, vue()],
  server: { host: '127.0.0.1', port: 0 },
});
let browser;
try {
  await server.listen();
  browser = await chromium.launch({ channel: 'chrome', headless: true });
  const page = await browser.newPage({ viewport: { width: 1200, height: 808 } });
  const errors = [];
  page.on('pageerror', e => errors.push(String(e)));
  await page.route('**/media?**', async route => {
    const url = new URL(route.request().url());
    const video = manifest.videos.find(v => v.id === url.searchParams.get('video'));
    if (!video || ![manifest.episode_id, 'catsoup-handoff-fixture'].includes(url.searchParams.get('episode'))) return route.fulfill({ status: 404 });
    const data = await readFile(path.resolve(deliveryFolder, video.path));
    const range = /^bytes=(\d+)-(\d*)$/.exec(route.request().headers().range || '');
    if (range) {
      const start = Number(range[1]);
      const end = range[2] ? Math.min(Number(range[2]), data.length - 1) : data.length - 1;
      await route.fulfill({ status: 206, contentType: 'video/mp4',
        headers: { 'Accept-Ranges': 'bytes', 'Content-Range': `bytes ${start}-${end}/${data.length}` },
        body: data.subarray(start, end + 1) });
    } else {
      await route.fulfill({ contentType: 'video/mp4', headers: { 'Accept-Ranges': 'bytes' }, body: data });
    }
  });
  const waitUploads=async()=>{await page.getByRole('dialog').waitFor({state:'detached'});await page.waitForFunction(()=>!document.querySelector('#uploads-launcher')?.textContent?.includes('active'));};
  const closePublishing=async()=>{const button=page.getByRole('button',{name:'Close publishing options',exact:true});if(await button.isVisible())await button.click();};
  const moreTools=async()=>{if(!await page.locator('.post-tools-menu').evaluate(el=>el.open))await page.locator('.post-tools-menu>summary').click();};
  const step=async value=>{
    if(value==='Publish') await page.locator('#open-publishing').click();
    else {await closePublishing();if(value==='Prepare')await moreTools();}
  };

  const review=async()=>{await closePublishing();await page.getByRole('button',{name:'Save draft',exact:true}).click();const button=page.getByRole('button',{name:'Approve post',exact:true});if(await button.isVisible())await button.click();await page.getByRole('button',{name:'Undo approval',exact:true}).waitFor();};
  await page.addInitScript(()=>{
    if(!localStorage.getItem('initial-accounts')){
      localStorage.setItem('initial-accounts','yes');
      localStorage.setItem('account-fixture',JSON.stringify(['es','en'].flatMap(language=>['youtube','tiktok'].map(platform=>({platform,language,id:platform+'-'+language,name:'Fixture '+language,api:false})))));
    }
  });
  await page.goto(server.resolvedUrls.local[0]);
  await page.getByRole('heading', { name: manifest.internal_title }).waitFor();
  assert(await page.getByRole('navigation', { name: 'Modules', exact: true }).isVisible());
  assert.equal(await page.getByRole('button', { name: 'Trends Planned' }).count(),0);
  await page.getByRole('button', { name: 'Maximize window', exact: true }).click();
  await page.getByRole('button', { name: 'Restore window', exact: true }).click();
  await page.getByRole('button', { name: 'Minimize window', exact: true }).click();
  await page.getByRole('button', { name: 'Close window', exact: true }).click();
  assert.deepEqual(await page.evaluate(() => window.__windowCalls), ['toggle', 'toggle', 'minimise', 'close']);
  for (const lang of ['Spanish', 'English']) {
    await page.getByRole('button', { name: lang, exact: true }).click();
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 2);
    const initial = await page.locator('video').evaluate(v => ({ duration: v.duration, width: v.videoWidth, height: v.videoHeight, src: v.currentSrc }));
    assert.equal(initial.duration, 12);
    assert.equal(initial.width, 540);
    assert.equal(initial.height, 960);
    assert(initial.src.includes(lang === 'Spanish' ? 'es_final' : 'en_final'));
    await page.locator('video').evaluate(async v => { v.muted = true; await v.play(); });
    await page.waitForFunction(() => document.querySelector('video').currentTime > 0.25);
    await page.locator('video').evaluate(v => { v.pause(); v.currentTime = 8; });
    await page.waitForFunction(() => !document.querySelector('video').seeking && Math.abs(document.querySelector('video').currentTime - 8) < .1);
  }
  await page.locator('.delivery-check>summary').click();
  await page.getByRole('combobox', { name: 'Version' }).selectOption('en_clean');
  await page.waitForFunction(() => document.querySelector('video')?.currentSrc.includes('en_clean') && document.querySelector('video').readyState >= 2);
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.waitForFunction(() => document.querySelector('video')?.currentSrc.includes('en_final') && document.querySelector('video').readyState >= 2);
  await page.getByRole('button', { name: 'Spanish', exact: true }).click();
  await page.waitForFunction(() => document.querySelector('video')?.readyState >= 2);
  await page.locator('video').evaluate(v => { v.currentTime = 4; });
  await page.waitForFunction(() => !document.querySelector('video').seeking && Math.abs(document.querySelector('video').currentTime - 4) < .1);
  assert(await page.getByLabel('Version',{exact:true}).isVisible(),'version is always visible');
  assert.equal(await page.locator('.review-details').count(),0);
  assert.equal(await page.locator('#open-publishing').count(),1,'only selected account has an individual action');
  assert.equal(await page.locator('.destination-plan-row .ui-button').count(),0);
  await page.locator('.publisher-main').evaluate(el=>el.scrollTop=el.scrollHeight);
  const pinned=await page.locator('#publish-all').boundingBox();
  assert(pinned.y>=40 && pinned.y+pinned.height<650,'Publish all remains in view while scrolling');
  await page.locator('.publisher-main').evaluate(el=>el.scrollTop=0);
  assert(await page.locator('#publish-all').isDisabled());
  for (const [width, height] of [[1200, 808], [820, 650]]) {
    await page.setViewportSize({ width, height });
    assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), 'horizontal overflow');
    await page.screenshot({ path: path.join(output, `publisher-${width}.png`), fullPage: true });
  }
  await page.setViewportSize({ width: 1200, height: 808 });
  assert.equal(await page.getByRole('navigation',{name:'Episode workflow'}).count(),0,'no numbered workflow navigation');
  assert(await page.locator('video').isVisible());
  assert(await page.getByLabel('Title',{exact:true}).isVisible(),'video and editor share one workspace');
  const platformPicker=page.getByRole('combobox',{name:'Post platform',exact:true});
  await platformPicker.selectOption('tiktok');
  await page.waitForFunction(()=>document.querySelector('#post-platform')?.value==='tiktok');
  assert(await page.getByLabel('Title',{exact:true}).isVisible());
  assert(await page.getByLabel('TikTok delivery mode',{exact:true}).isDisabled());
  await page.locator('.post-mode').hover();
  await page.getByRole('tooltip').filter({hasText:'Inbox only. Enable Direct Post'}).waitFor();
  await page.keyboard.press('Escape');
  await platformPicker.selectOption('youtube');
  const tip=page.locator('.approval-control');
  await tip.focus();
  assert(await page.getByRole('tooltip').filter({hasText:'Add post text first'}).isVisible());
  await page.keyboard.press('Escape');
  assert(!await page.getByRole('tooltip').filter({hasText:'Add post text first'}).isVisible());
  await page.locator('video').evaluate(async v=>{v.muted=true;await v.play()});
  await step('Prepare');
  assert(!await page.locator('video').evaluate(v=>v.paused),'editing must keep the visible preview playable');
  await page.getByLabel('Title', { exact: true }).fill('Un pequeño misterio');
  await page.getByLabel('Description', { exact: true }).fill('La luz se apagó. ¿Qué pasó después?');
  await page.getByLabel('Hashtags', { exact: true }).fill('#catsoup #animación');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  await page.getByRole('button', { name: 'Copy title', exact: true }).click();
  assert.equal(await page.evaluate(() => window.__copied), 'Un pequeño misterio');
  await page.getByRole('button', { name: 'Copy caption', exact: true }).click();
  assert.equal(await page.evaluate(() => window.__copied), 'La luz se apagó. ¿Qué pasó después?\n\n#catsoup #animación');
  await page.getByRole('button', { name: 'Open video folder', exact: true }).click();
  assert.equal((await page.evaluate(() => window.__revealed)).episode, 'catsoup-demo');
  await review();
  assert(await page.locator('#open-publishing').isDisabled(),'PoC remains non-publicable even when locally approved');
  assert(await page.locator('#publish-all').isDisabled(),'other accounts still need approval');
  await page.locator('.selected-send').focus();
  await page.getByRole('tooltip').filter({hasText:'production manifest'}).waitFor();
  await page.screenshot({path:path.join(output,'publish-blocked.png')});
  await page.keyboard.press('Escape');
  await page.getByRole('button',{name:'Save draft',exact:true}).click();
  await page.getByText(/Draft saved ·/).waitFor();
  assert(await page.getByRole('button',{name:'Undo approval',exact:true}).isVisible(),'no-op save preserves approval');
  // Change language before the debounce fires: leaving must flush the ES draft.
  await page.getByLabel('Title', { exact: true }).fill('Guardado antes de cambiar');
  assert(await page.getByRole('button',{name:'Approve post',exact:true}).isVisible(),'editing invalidates selected approval');
  await page.getByRole('button', { name: 'English', exact: true }).click();
  await page.getByLabel('Title', { exact: true }).fill('A tiny mystery');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  let persisted = await page.evaluate(() => Object.values(JSON.parse(localStorage.getItem('draft-fixture'))));
  assert.equal(persisted.length, 2);
  assert.equal(persisted.find(d => d.key.language === 'es').title, 'Guardado antes de cambiar');
  assert.equal(persisted.find(d => d.key.language === 'en').title, 'A tiny mystery');
  // A destination change also flushes; a separate account/platform stays isolated.
  await page.getByLabel('Post platform', { exact: true }).selectOption('tiktok');
  await page.getByLabel('Title', { exact: true }).fill('TikTok draft');
  await page.getByRole('button', { name: 'Spanish', exact: true }).click();
  await page.getByText('Saved locally', { exact: true }).waitFor();
  assert.equal(await page.getByLabel('Title', { exact: true }).inputValue(), 'Guardado antes de cambiar');
  // Type while a write is in flight: the response must not replace the newer text.
  await page.evaluate(() => { window.__draftDelay = 1000; });
  await page.getByLabel('Description', { exact: true }).fill('First edit');
  await page.getByText('Saving…', { exact: true }).waitFor();
  await page.getByLabel('Description', { exact: true }).fill('Latest edit while saving');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  assert.equal(await page.getByLabel('Description', { exact: true }).inputValue(), 'Latest edit while saving');
  await page.evaluate(() => { window.__draftDelay = 0; window.__draftFail = true; });
  await page.getByLabel('Title', { exact: true }).fill('Keep this unsaved edit');
  await page.getByRole('button', { name: 'English', exact: true }).click();
  await page.getByText(/Could not save\. Your edits are still here/).waitFor();
  assert.equal(await page.getByRole('button', { name: 'Spanish', exact: true }).getAttribute('aria-pressed'), 'true');
  assert.equal(await page.getByLabel('Title', { exact: true }).inputValue(), 'Keep this unsaved edit');
  await page.evaluate(() => window.__requestClose());
  await page.waitForTimeout(100);
  assert.equal(await page.evaluate(() => !!window.__closeConfirmed), false);
  await page.evaluate(() => { window.__draftFail = false; });
  await moreTools();
  await page.getByRole('button', { name: 'Save draft', exact: true }).click();
  await page.getByText('Saved locally', { exact: true }).waitFor();
  // A stale revision preserves edits until an explicit reload decision.
  await page.evaluate(() => { window.__draftConflict = true; });
  await page.getByLabel('Description', { exact: true }).fill('My conflicting edit');
  await page.getByText(/draft changed in another window/).waitFor();
  await page.getByRole('button', { name: 'Reload saved draft', exact: true }).click();
  await page.getByRole('button', { name: 'Keep editing', exact: true }).click();
  assert.equal(await page.getByLabel('Description', { exact: true }).inputValue(), 'My conflicting edit');
  await page.evaluate(() => { window.__draftConflict = false; });
  await page.getByRole('button', { name: 'Reload saved draft', exact: true }).click();
  await page.getByRole('button', { name: 'Reload and replace edits', exact: true }).click();
  assert.equal(await page.getByLabel('Description', { exact: true }).inputValue(), 'Latest edit while saving');
  await page.evaluate(() => { window.__fixtureEpisode.fingerprint = 'changed-fixture'; });
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.getByText(/The delivery changed since this draft was linked/).waitFor();
  await page.getByRole('button', { name: 'Use selected video', exact: true }).click();
  await page.getByText('Saved locally', { exact: true }).waitFor();
  assert.equal(await page.getByText(/The delivery changed since this draft was linked/).count(), 0);
  // Reload the page: the fixture persists data, while production persistence is tested in Go.
  await page.reload();
  await page.getByRole('button',{name:'Approve post',exact:true}).waitFor();
  await step('Prepare');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  assert.equal(await page.getByLabel('Title', { exact: true }).inputValue(), 'Keep this unsaved edit');
  await page.getByLabel('Title', { exact: true }).fill('Final saved draft');
  await page.evaluate(() => window.__requestClose());
  await page.waitForFunction(() => window.__closeConfirmed === true);
  persisted = await page.evaluate(() => Object.values(JSON.parse(localStorage.getItem('draft-fixture'))));
  assert.equal(persisted.find(d => d.key.language === 'es').title, 'Final saved draft');
  assert.equal(persisted.length, 3);
  await page.getByRole('button', { name: 'Use selected video', exact: true }).click();
  await page.getByText('Saved locally', { exact: true }).waitFor();
  // Capture the open, populated section separately from the video review.
  await page.getByRole('region', { name: 'Publication', exact: true }).screenshot({ path: path.join(output, 'publication-1200.png') });
  await page.setViewportSize({ width: 820, height: 650 });
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), 'draft horizontal overflow');
  await page.getByRole('region', { name: 'Publication', exact: true }).screenshot({ path: path.join(output, 'publication-820.png') });
  // Exercise the account confirmation using a separate in-memory episode.
  await page.evaluate(() => Object.assign(window.__fixtureEpisode, {
    id: 'catsoup-handoff-fixture', ready: true, status: 'approved_for_publishing', reason: '',
  }));
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.getByLabel('Title', { exact: true }).fill('Ready for manual upload');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  assert(await page.locator('#open-publishing').isDisabled());
  await review();
  await step('Publish');
  await page.getByText('Upload manually',{exact:true}).click();
  const upload = page.getByRole('button', { name: 'Copy upload file path' });
  assert(await upload.isDisabled());
  await page.getByRole('checkbox', { name: /is open in my browser/ }).check();
  await upload.click();
  await page.getByText(/Video path copied/).waitFor();
  assert.deepEqual(await page.evaluate(() => window.__uploadPathRequest), {
    key: { episode: 'catsoup-handoff-fixture', language: 'es', platform: 'youtube', account: 'account:youtube-es' }, revision: 1, confirmed: true,
  });
  assert(await upload.isDisabled());
  await closePublishing();
  await page.evaluate(()=>localStorage.removeItem('account-fixture'));
  await page.getByRole('button',{name:'Accounts',exact:true}).click();
  await page.getByRole('button',{name:'Settings',exact:true}).click();
  await page.getByRole('heading',{name:'Settings',exact:true}).waitFor();
  assert.equal(await page.getByText('Not configured',{exact:true}).count(),0);
  const defaults=page.getByRole('region',{name:'Publishing defaults',exact:true});
  assert.equal(await defaults.count(),0);
  await page.getByRole('button',{name:'Manage publishing',exact:true}).click();
  assert(await defaults.isVisible());
  assert(await defaults.getByLabel('YouTube default mode',{exact:true}).isDisabled(),'unconnected account cannot set defaults');
  await page.getByRole('button',{name:'Back to Settings',exact:true}).click();
  const overview=page.getByRole('region',{name:'Developer credentials',exact:true});
  const headerBox=await page.getByRole('heading',{name:'Settings',exact:true}).locator('..').boundingBox();
  const titleBox=await overview.getByRole('heading').boundingBox();
  assert(titleBox.y-(headerBox.y+headerBox.height)>=20,'settings section needs space below divider');
  await page.screenshot({path:path.join(output,'settings-credentials.png')});
  assert.equal(await page.getByLabel('Client secret',{exact:true}).count(),0);
  await page.getByRole('button',{name:'Manage credentials'}).click();
  await page.getByRole('group',{name:'Provider settings'}).getByText('Not configured',{exact:true}).first().waitFor();
  await page.getByRole('button',{name:'Back to Settings'}).click();
  await closePublishing();
  await page.getByRole('button',{name:'Accounts',exact:true}).click();
  for (const [lang,platform,signIn] of [['es','YouTube','Sign in with Google'],['es','TikTok','Sign in with TikTok'],['es','Instagram','Sign in with Facebook'],['es','X','Save handle'],['en','YouTube','Sign in with Google'],['en','TikTok','Sign in with TikTok']]) {
    await page.getByRole('button',{name:lang==='es'?'Spanish accounts':'English accounts',exact:true}).click();
    const row=page.getByRole('region',{name:platform,exact:true});
    if(lang==='es' && platform==='YouTube'){
      const divider=await page.locator('.account-manager .page-header').boundingBox();
      const toggle=await page.getByRole('button',{name:'Spanish accounts',exact:true}).boundingBox();
      assert(toggle.y-(divider.y+divider.height)>=16,'language toggle needs space below divider');
      await page.screenshot({path:path.join(output,'accounts-spacing.png')});
    }
    if (lang==='es' && platform!=='X') {
      if(platform==='Instagram') {
        await row.getByText('Other sign-in options',{exact:true}).click();
        await row.getByRole('button',{name:'Use Page token'}).click();
        await row.getByLabel('Facebook Page access token').waitFor();
        await row.getByRole('button',{name:'Cancel'}).click();
      }
      await row.getByRole('button',{name:'Set up credentials'}).click();
      await page.getByRole('heading',{name:'Developer credentials',exact:true}).waitFor();
      const setup=page.getByRole('region',{name:(platform==='Instagram'?'Meta':platform)+' settings'});
      if(platform==='YouTube') {
        await setup.getByText('Google Cloud setup').click();
        await setup.getByRole('button',{name:'Open Google clients'}).click();
        assert.equal(await page.evaluate(()=>window.__googleSetupStep),'client');
        const picker=setup.getByLabel('Import YouTube client JSON');
        const chooserEvent=page.waitForEvent('filechooser');
        await setup.getByRole('button',{name:'Import JSON…'}).focus();
        await page.keyboard.press('Enter');
        const chooser=await chooserEvent;
        await chooser.setFiles({name:'wrong.json',mimeType:'application/json',buffer:Buffer.from('{}')});
        await page.getByRole('alert').filter({hasText:'select Google Desktop app JSON'}).waitFor();
        await picker.setInputFiles({name:'download.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify({installed:{client_id:'fixture',client_secret:'fixture'}}))});
      } else if(platform==='TikTok') {
        const save=setup.getByRole('button',{name:'Save TikTok credentials'});
        await setup.getByLabel('Client key',{exact:true}).fill(' fixture-key ');
        assert(await save.isDisabled());
        const secret=setup.getByLabel('Client secret',{exact:true});
        assert.equal(await secret.getAttribute('type'),'password');
        await secret.fill(' fixture-secret ');
        await save.click();
        await setup.getByRole('status').waitFor();
        assert.equal(await secret.inputValue(),'');
      } else {
        await setup.getByText('Callback hosting').click();
        assert(await setup.getByLabel('Off',{exact:true}).isChecked());
        await setup.getByLabel('My HTTPS domain').check();
        const field=setup.getByLabel('HTTPS callback URL');
        assert.equal(await field.inputValue(),'');
        await field.fill('https://reels.example.org/team/meta/');
        await setup.getByRole('button',{name:'Save callback'}).click();
        await setup.getByText('https://reels.example.org/team/meta/',{exact:true}).waitFor();
        await setup.getByLabel('CATsoup-hosted callback',{exact:true}).check();
        assert(await setup.getByRole('button',{name:'Save callback'}).isDisabled());
        await setup.getByLabel('I agree to use the CATsoup-hosted callback.').check();
        await setup.getByRole('button',{name:'Save callback'}).click();
        await setup.getByText('CATsoup hosting is on.',{exact:false}).waitFor();
        await setup.getByLabel('App ID',{exact:true}).fill('123');
        await setup.getByLabel('App secret',{exact:true}).fill('fixture');
        const save=setup.getByRole('button',{name:'Save Meta credentials'});
        assert(await save.isDisabled());
        await setup.getByLabel('Configuration ID',{exact:true}).fill('456');
        await save.click();
      }
      await setup.getByRole('status').waitFor();
      await page.getByRole('group',{name:'Provider settings'}).getByRole('button',{pressed:true}).getByText('Saved · unverified',{exact:true}).waitFor();
      if(platform==='YouTube') await page.screenshot({path:path.join(output,'developer-credentials.png')});
      if(platform==='Instagram') {
        await setup.getByLabel('Off',{exact:true}).check();
        await setup.getByRole('button',{name:'Save callback'}).click();
        await setup.getByText('Meta browser sign-in is off.').waitFor();
        await closePublishing();
  await page.getByRole('button',{name:'Accounts',exact:true}).click();
        await row.getByRole('button',{name:'Set up callback'}).click();
        await setup.getByText('Callback hosting').click();
        await setup.getByLabel('CATsoup-hosted callback',{exact:true}).check();
        const agree=setup.getByLabel('I agree to use the CATsoup-hosted callback.');
        assert(!await agree.isChecked());
        await agree.check();
        await setup.getByRole('button',{name:'Save callback'}).click();
        await setup.getByText('CATsoup hosting is on.',{exact:false}).waitFor();
      }
      await closePublishing();
  await page.getByRole('button',{name:'Accounts',exact:true}).click();
    }
    await row.getByRole('button',{name:platform==='X'?'Set handle':'Connect',exact:true}).click();
    if(platform==='Instagram'||platform==='X') await row.getByLabel(platform+' username').fill('fixture');
    await row.getByRole('button',{name:signIn,exact:true}).click();
    await row.getByText('Fixture '+lang,{exact:true}).waitFor();
    if(platform==='YouTube' && lang==='es') {
      await page.screenshot({path:path.join(output,'accounts-connected.png')});
      await row.getByText('Account details').click();
      await row.getByRole('button',{name:'Remove',exact:true}).waitFor();
      await row.getByText('Account details').click();
    }
  }
  await page.getByRole('button',{name:'Settings',exact:true}).click();
  assert.equal(await page.getByRole('region',{name:'Publishing defaults',exact:true}).count(),0);
  await page.getByRole('button',{name:'Manage publishing',exact:true}).click();
  const prefs=page.getByRole('region',{name:'Publishing defaults',exact:true});
  await prefs.getByLabel('YouTube default mode',{exact:true}).selectOption('publish');
  await prefs.getByText('Default saved for Fixture es',{exact:true}).waitFor();
  const beforeDefaults=await page.evaluate(()=>JSON.parse(localStorage.getItem('draft-fixture')));
  await prefs.getByRole('button',{name:'English defaults',exact:true}).click();
  assert.equal(await prefs.getByLabel('YouTube default mode',{exact:true}).inputValue(),'draft');
  await prefs.getByRole('button',{name:'Spanish defaults',exact:true}).click();
  assert.equal(await prefs.getByLabel('YouTube default mode',{exact:true}).inputValue(),'publish');
  await page.evaluate(()=>window.__defaultFail=true);
  await prefs.getByLabel('YouTube default mode',{exact:true}).selectOption('draft');
  await prefs.getByText(/Simulated preference failure/).waitFor();
  assert.equal(await page.getByLabel('YouTube default mode',{exact:true}).inputValue(),'publish','failed preference save restores the stored value');
  await page.evaluate(()=>window.__defaultFail=false);
  await prefs.getByLabel('YouTube default mode',{exact:true}).selectOption('draft');
  await prefs.getByText('Default saved for Fixture es',{exact:true}).waitFor();
  assert.deepEqual(await page.evaluate(()=>JSON.parse(localStorage.getItem('draft-fixture'))),beforeDefaults,'preferences never change saved posts');
  await page.screenshot({path:path.join(output,'settings-publishing-defaults.png')});
  await page.getByRole('button',{name:'Publisher',exact:true}).click();
  await page.getByRole('button',{name:'English',exact:true}).click();
  await step('Prepare');
  await page.getByLabel('Title',{exact:true}).waitFor();
  await page.getByRole('button',{name:'Spanish',exact:true}).click();
  for (const [platform, login, action, result] of [
    ['youtube','Sign in with Google','Upload privately to YouTube','Private video uploaded'],
    ['tiktok','Sign in with TikTok','Send to TikTok inbox','Sent to inbox'],
    ['instagram','Connect account','Publish Reel on Instagram','Published'],
    ['x','','Copy video path and open X',''],
  ]) {
    await step('Prepare');
    if (platform !== 'youtube') {
  await page.getByLabel('Post platform', {exact:true}).selectOption(platform);
      await page.getByLabel('Title',{exact:true}).fill('Platform fixture');
      await page.getByText('Saved locally', {exact:true}).waitFor();
    }
    if(['instagram','x'].includes(platform))await page.getByLabel((platform==='instagram'?'Instagram':'X')+' delivery mode',{exact:true}).selectOption('publish');
    await review();
    await step('Publish');
    if(platform==='tiktok'){assert.equal(await page.getByRole('dialog').getByRole('radio').count(),0);await page.screenshot({path:path.join(output,'publish-tiktok.png')});}
    if(platform==='youtube')await page.getByLabel('Audience',{exact:true}).selectOption('general');
    if (platform === 'x') {
      assert(await page.getByRole('button', {name:action,exact:true}).isDisabled());
      await page.getByLabel('I checked the active account in X.').check();
      await page.getByRole('button', {name:action,exact:true}).click();
      assert.equal(await page.evaluate(()=>window.__xComposer),'x');
      assert.equal(await page.evaluate(()=>window.__uploadPathRequest.key.platform),'x');
      continue;
    }
    const send=page.getByRole('button',{name:action,exact:true});
    await send.waitFor();
    assert(await send.isDisabled());
    await page.getByRole('checkbox',{name:/I confirm Fixture es and/}).check();
    if(platform==='youtube')await page.evaluate(()=>window.__sendDelay=300);
    await send.click();
    await waitUploads();
    if(platform==='youtube'){await closePublishing();await page.getByLabel('YouTube delivery mode',{exact:true}).selectOption('publish');await review();await step('Publish');await page.getByLabel('Audience',{exact:true}).selectOption('general');await page.getByLabel('I confirm Fixture es and this public post.').check();await page.getByRole('button',{name:'Publish existing YouTube video'}).click();await waitUploads();assert.equal((await page.evaluate(()=>window.__promotion)).reference,'fixture-post');}
  }
  // X stays free until the user explicitly connects API mode. All calls below
  // are local fixture adapters, including paid consent and the post receipt.
  await closePublishing();
  await page.getByRole('button',{name:'Accounts',exact:true}).click();
  await page.getByRole('button',{name:'Spanish accounts',exact:true}).click();
  const xRow=page.getByRole('region',{name:'X',exact:true});
  await xRow.getByText('Browser · Free',{exact:true}).waitFor();
  await xRow.getByRole('button',{name:'Change',exact:true}).click();
  await xRow.getByRole('button',{name:'API · Paid',exact:true}).click();
  await xRow.getByRole('button',{name:'Set up X credentials'}).click();
  const xSetup=page.getByRole('region',{name:'X settings'});
  assert.equal(await xSetup.getByLabel('Client secret',{exact:true}).count(),0);
  await xSetup.getByText('X Developer setup',{exact:true}).click();
  await xSetup.getByRole('button',{name:'X pricing'}).click();
  assert.equal(await page.evaluate(()=>window.__xSetupStep),'pricing');
  await xSetup.getByLabel('Client ID',{exact:true}).fill('fixture-client');
  await xSetup.getByRole('button',{name:'Save X credentials'}).click();
  await xSetup.getByText('X credentials saved for both languages.').waitFor();
  await page.screenshot({path:path.join(output,'x-developer-settings.png')});
  await closePublishing();
  await page.getByRole('button',{name:'Accounts',exact:true}).click();
  await xRow.getByRole('button',{name:'Change',exact:true}).click();
  assert.equal(await xRow.getByRole('button',{name:'Browser · Free',exact:true}).getAttribute('aria-pressed'),'true');
  await xRow.getByRole('button',{name:'API · Paid',exact:true}).click();
  await xRow.getByLabel('X username').fill('@cats_es');
  assert(await xRow.getByRole('button',{name:'Sign in with X',exact:true}).isDisabled());
  await xRow.getByLabel('I understand X may charge my developer project for API usage.').check();
  await page.screenshot({path:path.join(output,'x-account-api.png')});
  await xRow.getByRole('button',{name:'Sign in with X',exact:true}).click();
  await xRow.getByText('@cats_es',{exact:true}).waitFor();
  await xRow.getByText('API · Paid',{exact:true}).waitFor();
  await page.getByRole('button',{name:'Publisher',exact:true}).click();
  await step('Prepare');
  await page.getByLabel('Post platform',{exact:true}).selectOption('x');
  await page.getByLabel('Title',{exact:true}).fill('Approved X fixture');
  await page.getByText('Saved locally',{exact:true}).waitFor();
  await page.getByLabel('X delivery mode',{exact:true}).selectOption('publish');
  await review();
  await step('Publish');
  const paidSend=page.getByRole('button',{name:'Publish on X · Paid API',exact:true});
  await paidSend.waitFor();assert(await paidSend.isDisabled());
  await page.getByLabel('I confirm @cats_es and this public post.').check();
  assert(await paidSend.isDisabled());
  await page.getByLabel('I accept X API usage charges for this attempt.').check();
  await paidSend.scrollIntoViewIfNeeded();
  await page.screenshot({path:path.join(output,'x-publish-paid.png')});
  await paidSend.click();
  await waitUploads();
  assert.equal(await page.evaluate(()=>window.__xPaidRequest.key.language),'es');
  assert.equal(await paidSend.count(),0,'successful receipt prevents another post');
  await closePublishing();
  // Unassigned text recovery is explicit, platform-scoped and leaves its source intact.
  await page.evaluate(()=>{
    const all=JSON.parse(localStorage.getItem('draft-fixture'));
    const source=Object.values(all).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.platform==='youtube'&&d.key.language==='es');
    const legacy={...source,key:{...source.key,account:'Unassigned fixture'},title:'Recovered text'};
    all[JSON.stringify(legacy.key)]=legacy;localStorage.setItem('draft-fixture',JSON.stringify(all));
  });
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await moreTools();
  await page.getByRole('button',{name:'Restore draft',exact:true}).click();
  await page.locator('#recover-draft').selectOption('Unassigned fixture');
  await page.getByRole('button',{name:'Replace current text',exact:true}).click();
  await page.waitForFunction(()=>document.querySelector('#draft-title')?.value==='Recovered text');
  await page.getByText('Saved locally',{exact:true}).waitFor();
  assert.equal(await page.evaluate(()=>Object.values(JSON.parse(localStorage.getItem('draft-fixture'))).find(d=>d.key.account==='Unassigned fixture').title),'Recovered text');
  // Identical visible names must not share drafts after an account replacement.
  await page.getByLabel('Post platform',{exact:true}).selectOption('youtube');
  const oldTitle=await page.getByLabel('Title',{exact:true}).inputValue();
  await page.evaluate(()=>{
    const rows=JSON.parse(localStorage.getItem('account-fixture'));
    rows.find(p=>p.platform==='youtube'&&p.language==='es').id='replacement-id';
    localStorage.setItem('account-fixture',JSON.stringify(rows));
  });
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.waitForFunction(()=>document.querySelector('#draft-title')?.value==='');
  assert.equal(await page.getByLabel('Title',{exact:true}).inputValue(),'');
  await page.evaluate(()=>{
    const rows=JSON.parse(localStorage.getItem('account-fixture'));
    rows.find(p=>p.platform==='youtube'&&p.language==='es').id='youtube-es';
    localStorage.setItem('account-fixture',JSON.stringify(rows));
  });
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.waitForFunction(title=>document.querySelector('#draft-title')?.value===title,oldTitle);
  assert.equal(await page.locator('#saved-destination').count(),0,'no combined draft selector');
  // Reuse only text from a connected account for the same episode and language.
  await page.getByLabel('Post platform',{exact:true}).selectOption('youtube');
  await page.getByLabel('Description',{exact:true}).fill('Una historia para compartir');
  await page.getByLabel('Hashtags',{exact:true}).fill('#catsoup #animación');
  await page.getByRole('button',{name:'Save draft',exact:true}).click();
  await page.getByLabel('Post platform',{exact:true}).selectOption('instagram');
  await review();
  const beforeCopy=await page.evaluate(()=>Object.values(JSON.parse(localStorage.getItem('draft-fixture'))).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.language==='es'&&d.key.platform==='instagram'));
  const sourceCopy=await page.evaluate(()=>Object.values(JSON.parse(localStorage.getItem('draft-fixture'))).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.language==='es'&&d.key.platform==='youtube'&&d.key.account==='account:youtube-es'));
  await page.getByRole('button',{name:'Copy from…',exact:true}).click();
  const copyDialog=page.getByRole('dialog',{name:'Copy post text'});
  assert.equal(await copyDialog.locator('#copy-source option[value="instagram"]').count(),0);
  await copyDialog.locator('#copy-source').selectOption('youtube');
  await page.screenshot({path:path.join(output,'copy-post-text.png')});
  await copyDialog.getByRole('button',{name:'Replace text',exact:true}).click();
  await page.getByText('Text copied from YouTube · draft saved',{exact:true}).waitFor();
  const afterCopy=await page.evaluate(()=>Object.values(JSON.parse(localStorage.getItem('draft-fixture'))).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.language==='es'&&d.key.platform==='instagram'));
  for(const field of ['title','description','hashtags'])assert.equal(afterCopy[field],sourceCopy[field]);
  for(const field of ['key','videoId','videoHash','deliveryMode'])assert.deepEqual(afterCopy[field],beforeCopy[field]);
  assert.equal(afterCopy.approvedRevision,0,'copied text requires its own approval');
  assert.deepEqual(await page.evaluate(()=>Object.values(JSON.parse(localStorage.getItem('draft-fixture'))).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.language==='es'&&d.key.platform==='youtube'&&d.key.account==='account:youtube-es')),sourceCopy,'source unchanged');
  // Batch preparation and mode persistence across reload of the workspace.
  for(const [platform,label,mode] of [['youtube','YouTube','draft'],['tiktok','TikTok','draft'],['instagram','Instagram','publish'],['x','X','draft']]){
    await page.getByLabel('Post platform',{exact:true}).selectOption(platform);
    await page.getByLabel('Title',{exact:true}).fill('Batch '+platform);
    await page.getByRole('button',{name:'Save draft',exact:true}).click();
    if(platform!=='tiktok')await page.getByLabel(label+' delivery mode',{exact:true}).selectOption(mode);
    await page.getByText('Saved locally',{exact:true}).waitFor();
    await review();
  }
  await page.getByRole('button',{name:'Undo approval',exact:true}).click();
  assert(await page.locator('#publish-all').isDisabled(),'one missing approval blocks the entire batch');
  await page.locator('.batch-header-action').focus();
  await page.getByRole('tooltip').filter({hasText:'Approval needed: X'}).waitFor();
  await page.keyboard.press('Escape');
  // Manual X is excluded even when its draft has no approval. Paid X still blocks above.
  await page.evaluate(()=>{const p=JSON.parse(localStorage.getItem('account-fixture'));p.find(p=>p.platform==='x'&&p.language==='es').api=false;localStorage.setItem('account-fixture',JSON.stringify(p));});
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.locator('#publish-all:not([disabled])').waitFor();
  await page.locator('#publish-all').click();
  await page.getByText('X is handled separately in the browser.',{exact:true}).waitFor();
  assert.equal(await page.getByRole('dialog').locator('.batch-row').filter({hasText:'X /'}).count(),0);
  await closePublishing();
  await page.evaluate(()=>{const p=JSON.parse(localStorage.getItem('account-fixture'));p.find(p=>p.platform==='x'&&p.language==='es').api=true;localStorage.setItem('account-fixture',JSON.stringify(p));});
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.getByLabel('Post platform',{exact:true}).selectOption('x');
  await review();
  await page.getByRole('button',{name:'Restore draft',exact:true}).click();
  await page.getByRole('dialog').getByRole('button',{name:'Reload saved draft',exact:true}).click();
  await page.getByText('Saved draft restored',{exact:true}).waitFor();
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.getByLabel('Post platform',{exact:true}).selectOption('instagram');
  assert.equal(await page.getByLabel('Instagram delivery mode',{exact:true}).inputValue(),'publish');
  await page.getByLabel('Post platform',{exact:true}).selectOption('x');
  assert.equal(await page.getByLabel('X delivery mode',{exact:true}).inputValue(),'draft');
  await page.evaluate(()=>{window.__clearReceipts();window.__sendCalls=[];window.__sendDelay=150;});
  await page.getByRole('button',{name:'Publish all…',exact:true}).click();
  const batch=page.getByRole('dialog');
  await batch.getByLabel('YouTube audience').selectOption('general');
  await batch.getByLabel('I confirm the selected accounts and delivery modes.').check();
  await page.screenshot({path:path.join(output,'publish-all.png')});
  await batch.getByRole('button',{name:'Send to 4 destinations',exact:true}).click();
  await waitUploads();
  assert.deepEqual((await page.evaluate(()=>window.__sendCalls)).map(c=>[c.key.platform,c.mode]),[['youtube','draft'],['tiktok','draft'],['instagram','publish']]);
  assert((await page.evaluate(()=>window.__sendCalls)).every(c=>c.key.language==='es'));
  await closePublishing();
  // Previously completed transfers are excluded rather than repeated.
  await page.getByRole('button',{name:'Publish all…',exact:true}).click();
  await batch.getByText('Already uploaded privately',{exact:true}).waitFor();
  assert.equal(await batch.locator('.batch-row input:checked').count(),1,'only the local X draft remains');
  await closePublishing();
  // Revoking any account while confirmation is open stops every send.
  await page.evaluate(()=>{window.__clearReceipts();window.__sendCalls=[];});
  await page.locator('#publish-all').click();
  await batch.getByLabel('YouTube audience').selectOption('general');
  await batch.getByLabel('I confirm the selected accounts and delivery modes.').check();
  await page.evaluate(()=>{const all=JSON.parse(localStorage.getItem('draft-fixture'));const d=Object.values(all).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.platform==='x'&&d.key.language==='es');d.approvedRevision=0;localStorage.setItem('draft-fixture',JSON.stringify(all));});
  await batch.getByRole('button',{name:'Send to 4 destinations',exact:true}).click();
  await batch.getByText('Every connected account needs a current approval.').waitFor();
  assert.equal((await page.evaluate(()=>window.__sendCalls)).length,0);
  await closePublishing();
  await page.getByLabel('Post platform',{exact:true}).selectOption('x');
  await review();
  // A partial failure is retained while later approved accounts continue; no automatic retries.
  await page.evaluate(()=>{window.__clearReceipts();window.__sendCalls=[];window.__failPlatform='tiktok';});
  await page.getByRole('button',{name:'Publish all…',exact:true}).click();
  await batch.getByLabel('YouTube audience').selectOption('general');
  await batch.getByLabel('I confirm the selected accounts and delivery modes.').check();
  await batch.getByRole('button',{name:'Send to 4 destinations',exact:true}).click();
  await waitUploads();
  assert.deepEqual((await page.evaluate(()=>window.__sendCalls)).map(c=>c.key.platform),['youtube','tiktok','instagram']);
  assert(await page.locator('#uploads-launcher').getByText('Needs attention').isVisible());
  await closePublishing();
  // A changed account invalidates the whole selection before its first request.
  await page.evaluate(()=>{window.__clearReceipts();window.__sendCalls=[];window.__failPlatform='';});
  await page.getByRole('button',{name:'Publish all…',exact:true}).click();
  await batch.getByLabel('YouTube audience').selectOption('general');
  await batch.getByLabel('I confirm the selected accounts and delivery modes.').check();
  await page.evaluate(()=>{
    const rows=JSON.parse(localStorage.getItem('account-fixture'));
    rows.find(p=>p.platform==='instagram'&&p.language==='es').id='changed-during-confirmation';
    localStorage.setItem('account-fixture',JSON.stringify(rows));
  });
  await batch.getByRole('button',{name:'Send to 4 destinations',exact:true}).click();
  await batch.getByText(/Every connected account needs a current approval|A draft or account changed/).waitFor();
  assert.equal((await page.evaluate(()=>window.__sendCalls)).length,0);
  await closePublishing();
  // Direct Post is offered only after the account grants video.publish.
  await page.evaluate(()=>{
    window.__clearReceipts();window.__sendCalls=[];
    const rows=JSON.parse(localStorage.getItem('account-fixture'));
    rows.find(p=>p.platform==='tiktok'&&p.language==='es').directPost=true;
    localStorage.setItem('account-fixture',JSON.stringify(rows));
  });
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.getByLabel('Post platform',{exact:true}).selectOption('tiktok');
  await page.getByLabel('TikTok delivery mode',{exact:true}).selectOption('publish');
  await review();await step('Publish');
  const direct=page.getByRole('dialog');
  await direct.getByLabel('Who can watch').waitFor();
  assert.equal(await direct.getByLabel('Who can watch').inputValue(),'');
  for(const name of ['Allow comments','Allow Stitch'])assert.equal(await direct.getByLabel(name,{exact:true}).isChecked(),false);
  assert(await direct.getByLabel('Allow Duet',{exact:true}).isDisabled());
  assert(await direct.getByRole('button',{name:'Post on TikTok',exact:true}).isDisabled());
  await direct.getByLabel('Who can watch').selectOption('PUBLIC_TO_EVERYONE');
  await direct.getByLabel('This post promotes a brand, product or service').check();
  await direct.getByRole('checkbox',{name:/By posting/}).check();
  await direct.getByRole('checkbox',{name:/I confirm Fixture es and/}).check();
  assert(await direct.getByRole('button',{name:'Post on TikTok',exact:true}).isDisabled());
  await direct.getByLabel('Branded content · Paid partnership').check();
  assert(await direct.locator('#tiktok-privacy option[value="SELF_ONLY"]').isDisabled());
  await direct.getByRole('button',{name:'Branded Content Policy',exact:true}).click();
  assert.equal(await page.evaluate(()=>window.__tikTokPolicy),'branded');
  await direct.getByLabel('This post promotes a brand, product or service').uncheck();
  await direct.getByLabel('Who can watch').selectOption('SELF_ONLY');
  assert(await direct.getByLabel('Branded content · Paid partnership').count()===0);
  await direct.getByRole('checkbox',{name:/By posting/}).check();
  await direct.getByRole('checkbox',{name:/I confirm Fixture es and/}).check();
  await page.screenshot({path:path.join(output,'publish-tiktok-direct.png')});
  await direct.getByRole('button',{name:'Post on TikTok',exact:true}).click();
  await waitUploads();
  const directCall=(await page.evaluate(()=>window.__sendCalls)).at(-1);
  assert.equal(directCall.options.privacy,'SELF_ONLY');assert.equal(directCall.options.branded,false);
  await closePublishing();
  await page.evaluate(()=>{const rows=JSON.parse(localStorage.getItem('account-fixture'));rows.find(p=>p.platform==='tiktok'&&p.language==='es').directPost=false;localStorage.setItem('account-fixture',JSON.stringify(rows));});
  await page.getByRole('button',{name:'Refresh',exact:true}).click();
  await page.getByLabel('Post platform',{exact:true}).selectOption('tiktok');
  assert.equal(await page.getByLabel('TikTok delivery mode',{exact:true}).inputValue(),'publish');
  await page.getByLabel('TikTok delivery mode',{exact:true}).selectOption('draft');
  await page.getByText('Saved locally',{exact:true}).waitFor();
  assert(await page.getByLabel('TikTok delivery mode',{exact:true}).isDisabled());
  await page.getByRole('button',{name:'Posts',exact:true}).click();
  await page.getByRole('heading',{name:'Posts',exact:true}).waitFor();
  await page.locator('.video-row').first().waitFor();
  const unique=await page.evaluate(()=>new Set(Object.values(JSON.parse(localStorage.getItem('draft-fixture'))).map(d=>d.key.episode)).size);
  assert.equal(await page.locator('.video-row').count(),unique,'one row per video, across all languages and platforms');
  await page.getByRole('button',{name:'Filters',exact:true}).click();
  assert.equal(await page.getByRole('dialog').getByRole('radio').count(),5);
  await page.getByRole('radio',{name:'Sent',exact:true}).check();
  await page.getByText('Includes public posts, private uploads and TikTok inbox deliveries.',{exact:true}).waitFor();
  await page.getByRole('radio',{name:'All',exact:true}).check();
  await page.screenshot({path:path.join(output,'posts-filters.png')});
  await page.getByLabel('Filter language').selectOption('en');
  await page.getByRole('button',{name:'Apply filters',exact:true}).click();
  await page.locator('.active-filters').getByText('English',{exact:true}).waitFor();
  await page.getByRole('button',{name:'Clear filters',exact:true}).click();
  await page.getByLabel('Search posts').fill('no-such-video');
  await page.getByText('No matching videos',{exact:true}).waitFor();
  await page.getByLabel('Search posts').fill('');
  await page.locator('.video-row').first().waitFor();
  await page.getByLabel('Search posts').focus();
  await page.locator('.video-row .video-thumbnail img').first().waitFor();
  assert.equal(await page.locator('.posts-main video').count(),0,'list retains static images, not video players');
  await page.screenshot({path:path.join(output,'posts-overview.png')});
  await page.locator('.video-row').first().click();
  await page.locator('.destination-row').first().waitFor();
  await page.getByRole('navigation',{name:'Video languages'}).getByRole('button',{name:'English',exact:true}).click();
  await page.waitForTimeout(300);
  assert.equal(await page.locator('.destination-row').filter({hasText:' · ES'}).count(),0);
  await page.getByRole('navigation',{name:'Video languages'}).getByRole('button',{name:'All languages',exact:true}).click();
  await page.locator('.destination-row').filter({hasText:'· ES'}).first().waitFor();
  await page.locator('.destination-row').first().waitFor();
  await page.locator('.video-detail-heading .video-thumbnail img').waitFor();
  await page.screenshot({path:path.join(output,'posts-video-detail.png')});
  await page.locator('.destination-row').first().click();
  await page.getByRole('dialog').waitFor();
  await page.keyboard.press('Escape');
  assert.equal(await page.getByRole('dialog').count(),0);
  await page.evaluate(()=>{const d=Object.values(JSON.parse(localStorage.getItem('draft-fixture')))[0];window.__statusRows=['youtube','instagram','facebook'].flatMap(platform=>['es','en'].map(language=>({key:{...d.key,platform,language,account:'account:'+platform+'-'+language},title:'The price of a light',videoId:d.videoId,accountName:'CATsoup '+language.toUpperCase(),connected:true,state:'published',reference:platform==='youtube'?'abcdefghijk':'123456',updatedAt:'2026-11-07T18:00:00Z'})));});
  await page.getByRole('button',{name:'Refresh posts',exact:true}).click();
  await page.getByRole('navigation',{name:'Video details'}).getByRole('button',{name:'Statistics',exact:true}).click();
  await page.waitForFunction(()=>document.querySelectorAll('.metric-strip').length===6);
  assert.equal(await page.locator('.statistics-platform').count(),3);
  for(const platform of ['YouTube','Instagram','Facebook'])assert.equal(await page.locator('.statistics-platform').filter({has:page.getByRole('heading',{name:platform,exact:true})}).locator('.post-statistics').count(),2);
  await page.screenshot({path:path.join(output,'statistics-platforms.png')});
  const reads=await page.evaluate(()=>window.__statisticsReads);
  await page.getByRole('navigation',{name:'Video details'}).getByRole('button',{name:'Posts',exact:true}).click();
  await page.getByRole('navigation',{name:'Video details'}).getByRole('button',{name:'Statistics',exact:true}).click();
  await page.waitForFunction(()=>document.querySelectorAll('.metric-strip').length===6);
  assert.equal(await page.evaluate(()=>window.__statisticsReads),reads,'tab switching reuses cached counts');
  await page.getByRole('button',{name:'Refresh all',exact:true}).click();
  await page.getByText('6 accounts updated',{exact:true}).waitFor();
  assert.equal(await page.evaluate(()=>window.__statisticsReads),reads+6,'refresh all bypasses cache for each supported account');
  await page.evaluate(()=>{window.__statisticsDelay=true;window.__statisticsFail='account:youtube-es'});
  await page.getByRole('button',{name:'Refresh all',exact:true}).click();
  await page.getByText('5 updated · 1 failed',{exact:true}).waitFor();
  assert.equal(await page.evaluate(()=>window.__statisticsReads),reads+12,'failure does not stop other accounts');
  assert.equal(await page.evaluate(()=>window.__statisticsPeak),2,'bulk reads have bounded concurrency');
  assert.equal(await page.locator('.metric-strip').count(),6,'failure retains dated counts');
  await page.evaluate(()=>{window.__statisticsDelay=false;window.__statisticsFail=''});


  await page.getByRole('button',{name:'Back to Posts',exact:true}).click();
  // Saved Instagram receipt opens the confirmation, never an automatic send.
  await page.evaluate(()=>{
    const stored=JSON.parse(localStorage.getItem('draft-fixture'));
    const d=Object.values(stored).find(d=>d.key.episode==='catsoup-handoff-fixture'&&d.key.platform==='instagram'&&d.key.language==='es');
    const profiles=JSON.parse(localStorage.getItem('account-fixture'));
    profiles.find(p=>p.platform==='instagram'&&p.language==='es').id=d.key.account.slice(8);
    localStorage.setItem('account-fixture',JSON.stringify(profiles));
    window.__statusRows=[{key:d.key,title:d.title,videoId:d.videoId,accountName:'Instagram ES',connected:true,state:'check',reference:'saved-container',updatedAt:d.updatedAt}];
    window.__setReceipt(d.key.account.slice(8),{state:'check',reference:'saved-container',retryAllowed:false});
  });
  await page.getByRole('button',{name:'Refresh posts',exact:true}).click();
  await page.locator('.video-row').first().click();
  await page.locator('.destination-row').first().click();
  // A changed account must show its error inside the detail dialog.
  await page.evaluate(()=>{const p=JSON.parse(localStorage.getItem('account-fixture'));window.__restoreAccount=p;p.find(r=>r.platform==='instagram'&&r.language==='es').id='replacement';localStorage.setItem('account-fixture',JSON.stringify(p));});
  await page.getByRole('dialog').getByRole('button',{name:'Review Instagram upload',exact:true}).click();
  await page.getByRole('dialog').getByRole('alert').filter({hasText:'This account changed'}).waitFor();
  await page.evaluate(()=>{const d=window.__statusRows[0];window.__restoreAccount.find(r=>r.platform==='instagram'&&r.language==='es').id=d.key.account.slice(8);localStorage.setItem('account-fixture',JSON.stringify(window.__restoreAccount));});
  const sendsBeforeReview=await page.evaluate(()=>window.__sendCalls.length);
  await page.getByRole('dialog').getByRole('button',{name:'Review Instagram upload',exact:true}).click();
  await page.getByRole('button',{name:'Check and finish on Instagram',exact:true}).waitFor();
  assert.equal(await page.getByLabel('Post platform',{exact:true}).inputValue(),'instagram');
  assert.equal(await page.evaluate(()=>window.__sendCalls.length),sendsBeforeReview,'opening recovery never publishes');
  await closePublishing();
  await page.getByRole('button',{name:'Posts',exact:true}).click();
  await checkDiagnostics(page,output);
  await page.getByRole('button',{name:'Publisher',exact:true}).click();
  await page.evaluate(() => { window.__empty = true; });
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.getByRole('heading', { name: 'No delivery to preview' }).waitFor();
  assert.equal(await page.locator('video').count(), 0);
  assert.deepEqual(errors, []);
  const report = { passed: true, checks: ['post overview search, filters and detail dialog', 'TikTok Direct Post explicit privacy, disabled interactions, commercial disclosures and private receipts', 'per-account publishing defaults, language isolation and failed-save recovery', 'hover and keyboard tooltips without help buttons', 'account ID isolation with identical names', 'saved per-account delivery modes', 'batch language routing, duplicate protection, preflight and independent failure handling', 'visible save action', 'dialog focus, Escape and background transfer handoff', 'minimal composer and folded secondary actions', 'per-account approval persistence and edit invalidation', 'review and approval gates', 'private/public modes and same-ID promotion', 'per-destination receipts and duplicate prevention', 'ES/EN decode and playback', 'seek', 'variant selection', 'refresh keeps language', 'non-publishable state', 'empty state', '1200/820px layout', 'module navigation', 'window controls with fixture adapter', 'draft autosave and reload', 'language/platform/account isolation', 'typing during save', 'failed saves block navigation and close', 'conflict recovery', 'delivery change and explicit rebind', 'copy/reveal adapters', 'close flush', 'diagnostics search and filters', 'attempt grouping and export', 'history pagination and pause', 'live diagnostics and keyboard access', 'console collapse and expansion preserve filters', 'no JS errors'], native: false };
  await writeFile(path.join(output, 'report.json'), JSON.stringify(report, null, 2));
  console.log(JSON.stringify(report));
} catch (error) {
  const page = browser?.contexts()[0]?.pages()[0];
  if (page) console.error((await page.locator('body').innerText()).slice(-4500));
  throw error;
} finally {
  await browser?.close();
  await server.close();
}
