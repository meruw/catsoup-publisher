import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import {readFile,mkdir} from 'node:fs/promises';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const out=fileURLToPath(new URL('../../../work/covers-qa/',import.meta.url));await mkdir(out,{recursive:true});
const plugin={name:'cover-fixture',enforce:'pre',load(id){
 const path=id.replaceAll('\\','/');
 if(path.endsWith('/services/drafts.ts'))return `export const drafts={importCover:async data=>{window.images.hash=data;return 'hash'},coverPreview:async hash=>window.images[hash]};`;
 if(path.endsWith('/services/uploads.ts'))return `export const uploads={receipt:async()=>({state:window.remote}),coverStatus:async()=>'',applyCover:async()=>{window.applied++;return 'Thumbnail applied'}};`;
 if(path.endsWith('/services/publisher.ts'))return `export const mediaURL=()=>'/sample.mp4';`;
},configureServer(server){
 server.middlewares.use('/sample.mp4',async(req,res)=>{res.setHeader('Content-Type','video/mp4');res.end(await readFile(new URL('../../../deliveries/catsoup-demo/video/es_final.mp4',import.meta.url)));});
 server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">
import {createApp,h,reactive} from 'vue';import Cover from '/src/components/CoverPicker.vue';import '/src/style.css';
window.images={};window.applied=0;window.remote='';window.changes=[];
const d=reactive({key:{episode:'catsoup-demo',language:'es',platform:'youtube',account:'account:fixture'},videoId:'es_final',revision:1,approvedRevision:1,deliveryMode:'publish'});window.draft=d;
createApp({render:()=>h('main',{style:'padding:28px;width:520px;margin:auto'},[h('h2','Post thumbnail'),h(Cover,{draft:d,disabled:false,onChange:c=>{d.cover=c;d.approvedRevision=0;window.changes.push(c)}})])}).mount('#app');
</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[plugin,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:1000,height:850}});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
 await page.goto(server.resolvedUrls.local[0]+'fixture');await page.locator('summary').click();
 await page.getByRole('button',{name:'Choose frame'}).click();await page.locator('video').evaluate(async v=>{await v.play();await new Promise(r=>setTimeout(r,300));v.pause()});await page.getByRole('button',{name:'Use this frame'}).click();
 await page.waitForFunction(()=>window.changes.length===1);assert.equal(await page.evaluate(()=>window.draft.cover.kind),'frame');assert.equal(await page.evaluate(()=>window.draft.approvedRevision),0);
 await page.getByAltText('Selected post thumbnail').waitFor();
 const img=await page.evaluate(()=>{const c=document.createElement('canvas');c.width=180;c.height=320;const ctx=c.getContext('2d');ctx.fillStyle='#ff7a32';ctx.fillRect(0,0,180,320);return c.toDataURL('image/png').split(',')[1]});
 await page.locator('input[type=file]').setInputFiles({name:'cover.png',mimeType:'image/png',buffer:Buffer.from(img,'base64')});await page.waitForFunction(()=>window.changes.length===2);assert.equal(await page.evaluate(()=>window.draft.cover.kind),'image');
 await page.screenshot({path:out+'image.png'});
 await page.getByRole('button',{name:'Remove',exact:true}).click();await page.waitForFunction(()=>window.draft.cover===null);
 await page.evaluate(()=>{window.draft.key.platform='instagram';window.draft.key.account='account:ig'});await page.getByRole('button',{name:'Choose frame'}).waitFor();assert.equal(await page.getByRole('button',{name:'Choose image'}).count(),0);
 await page.evaluate(()=>{window.draft.key.platform='tiktok';window.draft.deliveryMode='draft';});await page.getByText('TikTok Inbox does not receive a cover.',{exact:false}).waitFor();assert.equal(await page.getByRole('button',{name:'Choose frame'}).count(),0);
 assert.deepEqual(errors,[]);console.log('Cover UI: frame capture, image import, preview, removal and platform capabilities pass');
}finally{await browser?.close();await server.close()}
