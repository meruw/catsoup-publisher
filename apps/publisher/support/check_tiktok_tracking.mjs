import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import {mkdir} from 'node:fs/promises';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const out=fileURLToPath(new URL('../../../work/tiktok-tracking-ui/',import.meta.url));await mkdir(out,{recursive:true});
const fixture={name:'tracking-fixture',enforce:'pre',resolveId(id){if(id==='@wailsio/runtime')return '\0runtime'},load(id){
 if(id==='\0runtime')return `export const Events={On:(name,fn)=>{window.listener=fn;return ()=>{window.listener=null}}};`;
 if(id.replaceAll('\\','/').endsWith('/services/uploads.ts'))return `export const uploads={receipt:async()=>structuredClone(window.receipt)};`;
},configureServer(server){server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">
import {createApp,h} from 'vue';import Tracking from '/src/components/TikTokTracking.vue';import '/src/style.css';
window.receipt={state:'inbox',reference:'v_inbox_file~v2.222',retryAllowed:false,tracking:{status:'SEND_TO_USER_INBOX',checkedAt:1794515000,problem:''},inboxHistory:[{reference:'v_inbox_file~v2.111',tracking:{status:'PUBLISH_COMPLETE',checkedAt:1794514000}}]};
createApp({render:()=>h('main',{style:'max-width:480px;margin:auto;padding:24px'},h(Tracking,{postKey:{episode:'fixture',language:'es',platform:'tiktok',account:'account:es'},accountId:'es'}))}).mount('#app');</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[fixture,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:480,height:850}});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
await page.goto(server.resolvedUrls.local[0]+'fixture');await page.getByText('TikTok confirmed the inbox notification',{exact:true}).waitFor();await page.getByText('Previous attempts · 1').click();await page.getByText('Published from TikTok',{exact:true}).waitFor();
assert(await page.getByText(/cannot confirm that it is visible/).isVisible());assert.equal(await page.getByRole('button').count(),0);await page.screenshot({path:out+'inbox.png',fullPage:true});
await page.evaluate(()=>{window.receipt.tracking.problem='Could not check TikTok. Retrying automatically; last confirmed status retained.';window.listener()});await page.getByText(/Could not check TikTok/).waitFor();assert(await page.getByText('TikTok confirmed the inbox notification',{exact:true}).isVisible());
await page.evaluate(()=>{window.receipt.tracking={status:'PUBLISH_COMPLETE',checkedAt:1794516000,problem:''};window.listener()});await page.waitForFunction(()=>!document.body.innerText.includes('Could not check TikTok'));assert.equal(await page.getByText(/Checks automatically while/).count(),0);await page.screenshot({path:out+'published.png',fullPage:true});assert.deepEqual(errors,[]);console.log('PASS: live receipt updates, archive separation, last-known status on error, terminal display and narrow layout');
}finally{await browser?.close();await server.close()}

