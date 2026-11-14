// Synthetic browser fixture: no provider calls or native bridge.
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import {mkdir} from 'node:fs/promises';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const out=fileURLToPath(new URL('../../../work/instagram-hosting-ui/',import.meta.url));await mkdir(out,{recursive:true});
const fixture={name:'hosting-fixture',enforce:'pre',load(id){
 const name=id.replaceAll('\\','/');
 if(name.endsWith('/services/uploads.ts'))return `export const uploads={accounts:async()=>[],instagramHosting:async()=>({enabled:localStorage.enabled==='true',available:!window.missing}),setInstagramHosting:async(enabled,consent)=>{window.calls.push({enabled,consent});if(window.fail)throw Error('Save failed');localStorage.enabled=String(enabled)}};`;
 if(name.endsWith('/services/drafts.ts'))return `export const drafts={publishingDefaults:async()=>[]};`;
},configureServer(server){server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">
import {createApp,h} from 'vue';import Settings from '/src/components/PublishingSettings.vue';import '/src/style.css';
window.calls=[];window.missing=location.search.includes('missing');createApp({render:()=>h('main',{style:'max-width:768px;margin:auto;padding:24px'},h(Settings))}).mount('#app');</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[fixture,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:920,height:850}});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
await page.goto(server.resolvedUrls.local[0]+'fixture');const toggle=page.getByRole('checkbox',{name:'Use temporary HTTPS for Instagram'});await toggle.waitFor();assert(!await toggle.isChecked());await toggle.check();await page.getByRole('status').waitFor();assert.deepEqual(await page.evaluate(()=>window.calls),[{enabled:true,consent:true}]);assert(await page.getByText(/Anyone with the temporary link/).isVisible());
await page.reload();await toggle.waitFor();await page.waitForFunction(()=>document.querySelector('input[type=checkbox]').checked);await page.evaluate(()=>{window.fail=true});await toggle.click();await page.getByRole('alert').waitFor();assert(await toggle.isChecked());
await page.evaluate(()=>{window.fail=false});await toggle.uncheck();await page.getByRole('status').waitFor();await page.screenshot({path:out+'settings.png',fullPage:true});
await page.setViewportSize({width:480,height:900});await page.goto(server.resolvedUrls.local[0]+'fixture?missing');await toggle.waitFor();assert(await toggle.isDisabled());assert(await page.getByText(/Install cloudflared/).isVisible());await page.screenshot({path:out+'narrow.png',fullPage:true});assert.deepEqual(errors,[]);console.log('PASS: opt-in, consent, persistence, save-error rollback, unavailable helper, responsive settings');
}finally{await browser?.close();await server.close()}
