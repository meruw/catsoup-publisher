// Exercise the shipped HTTPS page in Chrome without contacting Meta or using credentials.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { createServer } from 'node:http';
import { createDecipheriv, randomBytes } from 'node:crypto';
import { chromium } from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';

const redirect = process.argv[2] || 'https://publisher.meruw.dev/auth/meta/';
const site=new URL(redirect);
const root = new URL('../../../docs/auth/meta/', import.meta.url);
const key = randomBytes(32), state = randomBytes(24).toString('base64url');
let reply;
const local = createServer((req,res) => {
  const url = new URL(req.url,'http://127.0.0.1');
  if (url.pathname === '/callback/') reply=url;
  res.end('Return to Publisher');
});
await new Promise(resolve=>local.listen(0,'127.0.0.1',resolve));
const browser = await chromium.launch({channel:'chrome',headless:true});
try {
  const page=await browser.newPage();
  await page.route(site.origin+'/**', async route=>{
    const path = new URL(route.request().url()).pathname;
    if(path===site.pathname || path===site.pathname+'return.js'){
      await route.fulfill({contentType:path.endsWith('.js')?'application/javascript':'text/html',body:await readFile(new URL(path.endsWith('.js')?'return.js':'index.html',root))});
    } else await route.fulfill({contentType:'text/css',body:''});
  });
  await page.route('https://www.facebook.com/**', async route=>{
    const s=new URL(route.request().url()).searchParams;
    assert.equal(s.get('state'),state);
    assert.equal(s.get('config_id'),'456');
    assert.equal(s.get('client_id'),'123');
    assert.equal(s.get('response_type'),'code');
    assert.equal(s.get('redirect_uri'),redirect);
    assert(!route.request().url().includes(key.toString('base64url')));
    await route.fulfill({contentType:'text/html',body:'Simulated consent'});
  });
  const start = new URLSearchParams({return_uri:redirect,key:key.toString('base64url'),state,app:'123',config:'456',port:String(local.address().port)});
  await page.route('https://publisher.meruw.dev/**',async route => {
    if(site.hostname!=='publisher.meruw.dev') throw Error('Custom callback contacted the official site');
    await route.fallback();
  });
  await page.goto(redirect+'#'+start);
  await page.waitForURL('https://www.facebook.com/**');
  // A returned URL alone, pasted into an unrelated tab, cannot start a handoff.
  const stranger=await browser.newPage();
  await stranger.route(site.origin+'/**', async route=>{
    const js=new URL(route.request().url()).pathname.endsWith('.js');
    await route.fulfill({contentType:js?'application/javascript':'text/html',body:await readFile(new URL(js?'return.js':'index.html',root))});
  });
  const result=redirect+'?'+new URLSearchParams({code:'synthetic-code',state});
  await stranger.goto(result);
  await stranger.getByText(/no matching sign-in/i).waitFor();
  assert.equal(stranger.url(),redirect);
  assert.equal(reply,undefined);
  await page.goto(result);
  await page.waitForURL('http://127.0.0.1:'+local.address().port+'/callback/**');
  assert.equal(reply.searchParams.get('state'),state);
  assert(!reply.href.includes('synthetic-code'));
  const raw=Buffer.from(reply.searchParams.get('reply'),'base64url');
  const decipher=createDecipheriv('aes-256-gcm',key,raw.subarray(0,12));
  decipher.setAAD(Buffer.from(state));decipher.setAuthTag(raw.subarray(-16));
  assert.equal(Buffer.concat([decipher.update(raw.subarray(12,-16)),decipher.final()]).toString(),'synthetic-code');
  console.log('Browser HTTPS callback and encrypted loopback return passed.');
} finally {
  await browser.close();local.close();
}
