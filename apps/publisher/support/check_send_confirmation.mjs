// Real send components; fixture providers never access an external account.
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import {mkdir,readFile} from 'node:fs/promises';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const out=fileURLToPath(new URL('../../../work/send-confirmation/',import.meta.url));
await mkdir(out,{recursive:true});
const fixture={name:'send-confirmation-fixture',enforce:'pre',load(id){
 const name=id.replaceAll('\\','/');
 if(name.endsWith('/services/publisher.ts'))return `export const mediaURL=(episode,video)=>'/media?video='+video;`;
 if(name.endsWith('/services/drafts.ts'))return `export const drafts={get:async key=>{if(window.rejectApproval)return {...window.data.saved.find(d=>d.key.platform===key.platform),approvedRevision:0};return window.data.saved.find(d=>d.key.platform===key.platform)},copyUploadPath:async()=>{},openX:async()=>{}};`;
 if(name.endsWith('/services/uploads.ts'))return `export const uploads={accounts:async()=>window.data.profiles,receipt:async()=>({state:'',reference:'',retryAllowed:false}),sendMode:async(...args)=>{window.sends.push(args);return {state:'published'}},openX:async()=>{}};`;
 if(name.endsWith('/services/tiktokWeb.ts'))return `export const prepareTikTokWeb=async(d,id)=>{window.sends.push([d.key,id]);return 'Ready in browser'};`;
 if(name.endsWith('/components/TikTokBrowserSession.vue'))return `<script setup>import {onMounted} from 'vue';const emit=defineEmits(['verified']);onMounted(()=>emit('verified','fixture_en'));</script><template><p>Verified fixture_en</p></template>`;
},configureServer(server){
 server.middlewares.use('/media',async(req,res)=>{res.setHeader('Content-Type','video/mp4');res.end(await readFile(new URL('../../../deliveries/catsoup-demo/video/en_final.mp4',import.meta.url)));});
 server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">
import {createApp,h,reactive} from 'vue';import Batch from '/src/components/BatchPublish.vue';import Direct from '/src/components/DirectUpload.vue';import Web from '/src/components/TikTokWebPrepare.vue';import '/src/style.css';
const kind=new URLSearchParams(location.search).get('kind')||'batch';
const data=reactive({language:'en',episode:{id:'fixture',title:'A tower above the clouds',ready:true,fingerprint:'v1'},video:{id:'en_final',lang:'en',sha256:'hash-en',kind:'final'},profiles:['youtube','instagram','tiktok','facebook'].map(platform=>({platform,id:platform+'-en',language:'en',name:platform+' EN',api:false})),saved:[]});
data.saved=data.profiles.map(p=>({key:{episode:'fixture',language:'en',platform:p.platform,account:'account:'+p.id},title:'A tower above the clouds',description:'',hashtags:'',videoId:'en_final',videoHash:'hash-en',deliveryFingerprint:'v1',revision:1,approvedRevision:1,deliveryMode:p.platform==='tiktok'?'draft':'publish'}));
window.data=data;window.sends=[];window.queued=0;
createApp({render(){return h('main',{style:'width:480px;max-width:100%;padding:28px;margin:auto'},kind==='batch'?h(Batch,{...data,onQueued:()=>window.queued++}):kind==='web'?h(Web,{draft:data.saved[2],account:data.profiles[2],receipt:'',hint:'',flush:async()=>{}}):h(Direct,{draft:data.saved[1],hint:'',flush:async()=>{if(window.changeOnFlush)data.saved[1].revision++},onQueued:()=>window.queued++}));}}).mount('#app');
</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[fixture,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage();const errors=[];page.on('pageerror',e=>errors.push(String(e)));
 const go=async kind=>page.goto(server.resolvedUrls.local[0]+'fixture?kind='+kind);
 const check=()=>page.getByLabel('I confirm this is the English version for these accounts.');
 const final=()=>page.getByRole('button',{name:'Send English video',exact:true});
 const batchReview=async()=>{await page.getByLabel('YouTube audience').selectOption('general');await page.getByLabel('I confirm the selected accounts and delivery modes.').check();await page.getByRole('button',{name:'Send to 4 destinations',exact:true}).click();};
 await go('batch');await batchReview();assert.equal(await page.evaluate(()=>window.sends.length),0);assert(await final().isDisabled());
 assert.equal(await page.getByLabel('English video preview').getAttribute('src'),'/media?video=en_final');
 assert.equal(await page.getByLabel('Confirmed destinations').locator('li').count(),4);
 await check().check();await page.getByRole('button',{name:'Back',exact:true}).click();assert.equal(await page.evaluate(()=>window.sends.length),0);
 await page.getByRole('button',{name:'Send to 4 destinations',exact:true}).click();assert(await final().isDisabled());
 await page.setViewportSize({width:440,height:900});await page.screenshot({path:out+'batch.png',fullPage:true});
 await check().check();await final().click();await page.waitForFunction(()=>window.sends.length===4);
 assert((await page.evaluate(()=>window.sends)).every(args=>args[0].language==='en'));
 await go('batch');await batchReview();await check().check();await page.evaluate(()=>window.rejectApproval=true);await final().click();await page.getByRole('alert').waitFor();assert.equal(await page.evaluate(()=>window.sends.length),0);
 await go('batch');await batchReview();await check().check();await page.evaluate(()=>{window.data.video.id='new-video'});await page.getByRole('button',{name:'Send to 4 destinations',exact:true}).waitFor();assert.equal(await page.evaluate(()=>window.sends.length),0);
 await go('direct');await page.getByLabel('I confirm instagram EN and this public post.').check();await page.getByRole('button',{name:'Publish Reel on Instagram',exact:true}).click();assert.equal(await page.evaluate(()=>window.sends.length),0);assert(await final().isDisabled());await check().check();await final().click();await page.waitForFunction(()=>window.sends.length===1);assert.equal(await page.evaluate(()=>window.sends[0][0].language),'en');
 await go('direct');await page.getByLabel('I confirm instagram EN and this public post.').check();await page.getByRole('button',{name:'Publish Reel on Instagram',exact:true}).click();await check().check();await page.evaluate(()=>window.changeOnFlush=true);await final().click();await page.getByRole('alert').waitFor();assert.equal(await page.evaluate(()=>window.sends.length),0);
 await go('web');await page.getByRole('button',{name:'Prepare in TikTok',exact:true}).click();assert.equal(await page.evaluate(()=>window.sends.length),0);const prepare=page.getByRole('button',{name:'Prepare English video',exact:true});assert(await prepare.isDisabled());await check().check();await prepare.click();await page.waitForFunction(()=>window.sends.length===1);
 assert.deepEqual(errors,[]);console.log('PASS: batch, individual and browser final language gates; cancel/back, exact video, unchanged approvals, changed video/flush, EN routing; no provider calls before confirmation');
}finally{await browser?.close();await server.close();}
