// Local visual/regression harness. The bridge is a fixture; this is not a native Wails test.
// Uses generated demo MP4s without changing them. No social accounts or remote services.
import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import assert from 'node:assert/strict';
import { draftFixture } from './draft_fixture.mjs';
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
    if (id.replaceAll('\\', '/').endsWith('/src/services/uploads.ts')) return `
      const none = {state:'',reference:'',retryAllowed:false};
      let receipt = none;
      export const uploads = {
        connected: async () => ({id:'',name:''}),
        connect: async () => {receipt = none; return {id:'connected-id',name:'Fixture channel'};},
        receipt: async () => receipt,
        send: async (key, revision, id, confirmed) => {
          if (!confirmed || id !== 'connected-id' || key.episode !== 'catsoup-handoff-fixture' || revision < 1) throw Error('unsafe fixture upload');
          return receipt = {state:'sent',reference:'fixture-post',retryAllowed:false};
        },
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
  await page.goto(server.resolvedUrls.local[0]);
  await page.getByRole('heading', { name: manifest.internal_title }).waitFor();
  assert(await page.getByRole('navigation', { name: 'Modules', exact: true }).isVisible());
  assert(await page.getByRole('button', { name: 'Trends Planned' }).isDisabled());
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
  await page.getByRole('combobox', { name: 'Version' }).selectOption('en_clean');
  await page.waitForFunction(() => document.querySelector('video')?.currentSrc.includes('en_clean') && document.querySelector('video').readyState >= 2);
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.waitForFunction(() => document.querySelector('video')?.currentSrc.includes('en_final') && document.querySelector('video').readyState >= 2);
  await page.getByRole('button', { name: 'Spanish', exact: true }).click();
  await page.waitForFunction(() => document.querySelector('video')?.readyState >= 2);
  await page.locator('video').evaluate(v => { v.currentTime = 4; });
  await page.waitForFunction(() => !document.querySelector('video').seeking && Math.abs(document.querySelector('video').currentTime - 4) < .1);
  assert(await page.getByText('Not approved for publishing', { exact: true }).isVisible());
  for (const [width, height] of [[1200, 808], [820, 650]]) {
    await page.setViewportSize({ width, height });
    assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), 'horizontal overflow');
    await page.screenshot({ path: path.join(output, `publisher-${width}.png`), fullPage: true });
  }
  await page.setViewportSize({ width: 1200, height: 808 });
  await page.getByRole('button', { name: 'Publication', exact: true }).click();
  assert.equal(await page.evaluate(() => document.activeElement.id), 'draft-account');
  await page.getByLabel('Account label', { exact: true }).fill('@catsoup_es');
  await page.getByRole('button', { name: 'Open draft', exact: true }).click();
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
  await page.getByText('Available after the delivery is approved for publishing.').waitFor();
  assert.equal(await page.getByRole('button', { name: 'Copy upload file path' }).count(), 0);
  // Change language before the debounce fires: leaving must flush the ES draft.
  await page.getByLabel('Title', { exact: true }).fill('Guardado antes de cambiar');
  await page.getByRole('button', { name: 'English', exact: true }).click();
  await page.getByLabel('Account label', { exact: true }).fill('@catsoup_en');
  await page.getByRole('button', { name: 'Open draft', exact: true }).click();
  await page.getByLabel('Title', { exact: true }).fill('A tiny mystery');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  let persisted = await page.evaluate(() => Object.values(JSON.parse(localStorage.getItem('draft-fixture'))));
  assert.equal(persisted.length, 2);
  assert.equal(persisted.find(d => d.key.language === 'es').title, 'Guardado antes de cambiar');
  assert.equal(persisted.find(d => d.key.language === 'en').title, 'A tiny mystery');
  // A destination change also flushes; a separate account/platform stays isolated.
  await page.getByLabel('Platform', { exact: true }).selectOption('tiktok');
  await page.getByLabel('Account label', { exact: true }).fill('@catsoup_second');
  await page.getByRole('button', { name: 'Open draft', exact: true }).click();
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
  await page.getByLabel('Account label', { exact: true }).fill('@handoff_es');
  await page.getByRole('button', { name: 'Open draft', exact: true }).click();
  await page.getByLabel('Title', { exact: true }).fill('Ready for manual upload');
  await page.getByText('Saved locally', { exact: true }).waitFor();
  const upload = page.getByRole('button', { name: 'Copy upload file path' });
  assert(await upload.isDisabled());
  await page.getByRole('checkbox', { name: /@handoff_es/ }).check();
  await upload.click();
  await page.getByText(/Video path copied/).waitFor();
  assert.deepEqual(await page.evaluate(() => window.__uploadPathRequest), {
    key: { episode: 'catsoup-handoff-fixture', language: 'es', platform: 'youtube', account: '@handoff_es' }, revision: 1, confirmed: true,
  });
  assert(await upload.isDisabled());
  for (const [platform, login, action, result] of [
    ['youtube','Sign in with Google','Upload privately to YouTube','Private video uploaded'],
    ['tiktok','Sign in with TikTok','Send to TikTok inbox','Sent to TikTok inbox'],
    ['instagram','Connect account','Publish Reel on Instagram','Reel published'],
    ['x','','Open X composer',''],
  ]) {
    if (platform !== 'youtube') {
      await page.getByLabel('Platform', {exact:true}).selectOption(platform);
      await page.getByRole('button', {name:'Open draft',exact:true}).click();
      await page.getByLabel('Title',{exact:true}).fill('Platform fixture');
      await page.getByText('Saved locally', {exact:true}).waitFor();
    }
    if (platform === 'x') {
      await page.getByRole('button', {name:action,exact:true}).click();
      assert.equal(await page.evaluate(()=>window.__xComposer),'x');
      continue;
    }
    if (platform === 'instagram') await page.getByLabel('Facebook Page access token').fill('test-only');
    await page.getByRole('button',{name:login,exact:true}).click();
    const send=page.getByRole('button',{name:action,exact:true});
    await send.waitFor();
    assert(await send.isDisabled());
    await page.getByRole('checkbox',{name:'Send the ES video to Fixture channel.'}).check();
    await send.click();
    await page.getByText(result+' · fixture-post',{exact:true}).waitFor();
  }
  await page.evaluate(() => { window.__empty = true; });
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.getByRole('heading', { name: 'No delivery to preview' }).waitFor();
  assert.equal(await page.locator('video').count(), 0);
  assert.deepEqual(errors, []);
  const report = { passed: true, checks: ['ES/EN decode and playback', 'seek', 'variant selection', 'refresh keeps language', 'non-publishable state', 'empty state', '1200/820px layout', 'module navigation', 'window controls with fixture adapter', 'draft autosave and reload', 'language/platform/account isolation', 'typing during save', 'failed saves block navigation and close', 'conflict recovery', 'delivery change and explicit rebind', 'copy/reveal adapters', 'close flush', 'no JS errors'], native: false };
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
