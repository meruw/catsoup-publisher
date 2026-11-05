// Checks the real shared controls and token contrasts; no production bridge or data.
import { mkdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import assert from 'node:assert/strict';
import { createServer } from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import { chromium } from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';

const root = fileURLToPath(new URL('../../../', import.meta.url));
const output = path.join(root, 'work/publisher-ui-check');
await mkdir(output, { recursive: true });
const server = await createServer({ configFile: false, root: path.join(root, 'apps/publisher/frontend'), plugins: [vue()], server: { host: '127.0.0.1', port: 0 } });
let browser;
try {
  await server.listen();
  browser = await chromium.launch({ channel: 'chrome', headless: true });
  const page = await browser.newPage({ viewport: { width: 1120, height: 1000 } });
  const errors = [];
  page.on('pageerror', error => errors.push(String(error)));
  await page.goto(new URL('design-system.html', server.resolvedUrls.local[0]).href);
  await page.evaluate(() => document.fonts.ready);
  assert(await page.evaluate(() => document.fonts.check('14px "Instrument Sans"')));
  const ratios = await page.evaluate(() => {
    const tokens = getComputedStyle(document.documentElement);
    const luminance = token => {
      const hex = tokens.getPropertyValue(token).trim().slice(1);
      const rgb = [0, 2, 4].map(i => parseInt(hex.slice(i, i + 2), 16) / 255).map(c => c <= .04045 ? c / 12.92 : ((c + .055) / 1.055) ** 2.4);
      return rgb[0] * .2126 + rgb[1] * .7152 + rgb[2] * .0722;
    };
    return [['--text','--page'],['--secondary','--page'],['--muted','--sidebar'],['--muted','--accent-soft'],['--on-accent','--accent'],['--on-accent','--accent-hover'],['--on-accent','--accent-pressed'],['--success','--page'],['--warning','--page'],['--danger','--page']].map(([fg,bg]) => {
      const a=luminance(fg), b=luminance(bg);
      return { pair: `${fg}/${bg}`, ratio: (Math.max(a,b)+.05)/(Math.min(a,b)+.05) };
    });
  });
  for (const item of ratios) assert(item.ratio >= 4.5, `${item.pair}: ${item.ratio}`);
  assert(await page.getByRole('button', { name: 'Disabled', exact: true }).isDisabled());
  assert(await page.getByRole('button', { name: 'Working…', exact: true }).isDisabled());
  await page.getByRole('button', { name: 'English', exact: true }).click();
  assert.equal(await page.getByRole('button', { name: 'English', exact: true }).getAttribute('aria-pressed'), 'true');
  await page.getByRole('button', { name: 'Primary', exact: true }).focus();
  await page.keyboard.press('Tab');
  assert(await page.evaluate(() => getComputedStyle(document.activeElement).outlineStyle === 'solid'));
  await page.screenshot({ path: path.join(output, 'design-system.png'), fullPage: true });
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ passed: true, contrastRatios: ratios.map(x => ({ ...x, ratio: +x.ratio.toFixed(2) })) }));
} finally {
  await browser?.close();
  await server.close();
}
