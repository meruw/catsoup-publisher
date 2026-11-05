// Local visual/regression harness. The bridge is a fixture; this is not a native Wails test.
// Uses generated demo MP4s without changing them. No social accounts or remote services.
import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import assert from 'node:assert/strict';
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
 export const publisher = {
  list: async () => window.__empty ? [] : [{id:episode.id,title:episode.title,valid:true,problem:''}],
  get: async () => episode,
 };
 export const mediaURL = (episode,video) => '/media?'+new URLSearchParams({episode,video});
`;
const server = await createServer({
  configFile: false, root: frontend,
  plugins: [{ name: 'test-only-bridge', enforce: 'pre', load(id) {
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
    if (!video || url.searchParams.get('episode') !== manifest.episode_id) return route.fulfill({ status: 404 });
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
  await page.evaluate(() => { window.__empty = true; });
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.getByRole('heading', { name: 'No delivery to preview' }).waitFor();
  assert.equal(await page.locator('video').count(), 0);
  assert.deepEqual(errors, []);
  const report = { passed: true, checks: ['ES/EN decode and playback', 'seek', 'variant selection', 'refresh keeps language', 'non-publishable state', 'empty state', '1200/820px layout', 'module navigation', 'window controls with fixture adapter', 'no JS errors'], native: false };
  await writeFile(path.join(output, 'report.json'), JSON.stringify(report, null, 2));
  console.log(JSON.stringify(report));
} finally {
  await browser?.close();
  await server.close();
}
