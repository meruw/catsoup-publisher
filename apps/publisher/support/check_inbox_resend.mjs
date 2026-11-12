// Real send components; fixture providers never access an external account.
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import {mkdir,readFile} from 'node:fs/promises';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const out=fileURLToPath(new URL('../../../work/inbox-resend/',import.meta.url));
await mkdir(out,{recursive:true});
const fixture={name:'send-confirmation-fixture',enforce:'pre',resolveId(id){if(id==='@wailsio/runtime')return '\0fixture-runtime'},load(id){if(id==='\0fixture-runtime')return 'export const Events={On:()=>()=>{}};';
 const name=id.replaceAll('\\','/');
 if(name.endsWith('/services/publisher.ts'))return `export const mediaURL=(episode,video)=>'/media?video='+video;`;
 if(name.endsWith('/services/drafts.ts'))return `export const drafts={get:async key=>{if(window.rejectApproval)return {...window.data.saved.find(d=>d.key.platform===key.platform),approvedRevision:0};return window.data.saved.find(d=>d.key.platform===key.platform)},copyUploadPath:async()=>{},openX:async()=>{}};`;
 if(name.endsWith('/services/uploads.ts'))return `export const uploads={accounts:async()=>window.data.profiles,receipt:async()=>window.inbox,allowInboxResend:async(...args)=>{window.resets.push(args);window.inbox={state:'',reference:'',retryAllowed:false,inboxHistory:[{reference:'old-inbox',confirmedAt:1794497377,archivedAt:1794497380}]};},sendMode:async(...args)=>{window.sends.push(args);return {state:'published'}},openX:async()=>{}};`;
 if(name.endsWith('/services/tiktokWeb.ts'))return `export const prepareTikTokWeb=async(d,id)=>{window.sends.push([d.key,id]);return 'Ready in browser'};`;
 if(name.endsWith('/components/TikTokBrowserSession.vue'))return `<script setup>import {onMounted} from 'vue';const emit=defineEmits(['verified']);onMounted(()=>emit('verified','fixture_en'));</script><template><p>Verified fixture_en</p></template>`;
},configureServer(server){
 server.middlewares.use('/media',async(req,res)=>{res.setHeader('Content-Type','video/mp4');res.end(await readFile(new URL('../../../deliveries/catsoup-demo/video/en_final.mp4',import.meta.url)));});
 server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">
import {createApp,h,reactive} from 'vue';import Batch from '/src/components/BatchPublish.vue';import Direct from '/src/components/DirectUpload.vue';import Web from '/src/components/TikTokWebPrepare.vue';import '/src/style.css';
const kind=new URLSearchParams(location.search).get('kind')||'batch';
const data=reactive({language:'en',episode:{id:'fixture',title:'A tower above the clouds',ready:true,fingerprint:'v1'},video:{id:'en_final',lang:'en',sha256:'hash-en',kind:'final'},profiles:['youtube','instagram','tiktok','facebook'].map(platform=>({platform,id:platform+'-en',language:'en',name:platform+' EN',api:false})),saved:[]});
data.saved=data.profiles.map(p=>({key:{episode:'fixture',language:'en',platform:p.platform,account:'account:'+p.id},title:'A tower above the clouds',description:'',hashtags:'',videoId:'en_final',videoHash:'hash-en',deliveryFingerprint:'v1',revision:1,approvedRevision:1,deliveryMode:p.platform==='tiktok'?'draft':'publish'}));
window.data=data;window.sends=[];window.resets=[];window.queued=0;window.inbox={state:'inbox',reference:'old-inbox',retryAllowed:false};
createApp({render(){return h('main',{style:'width:480px;max-width:100%;padding:28px;margin:auto'},kind==='batch'?h(Batch,{...data,onQueued:()=>window.queued++}):kind==='web'?h(Web,{draft:data.saved[2],account:data.profiles[2],receipt:'',hint:'',flush:async()=>{}}):h(Direct,{draft:data.saved[2],hint:'',flush:async()=>{if(window.changeOnFlush)data.saved[1].revision++},onQueued:()=>window.queued++}));}}).mount('#app');
</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[fixture,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage();const errors=[];page.on('pageerror',e=>errors.push(String(e)));

 await page.goto(server.resolvedUrls.local[0]+'fixture?kind=direct');
 await page.getByRole('button',{name:'Resend to inbox…',exact:true}).click();
 const enable=page.getByRole('button',{name:'Enable another send',exact:true});
 assert(await enable.isDisabled());assert.equal(await page.evaluate(()=>window.sends.length),0);
 await page.getByLabel(/I want to resend/).check();
 await page.getByRole('button',{name:'Cancel',exact:true}).click();
 assert.equal(await page.evaluate(()=>window.resets.length),0);
 await page.getByRole('button',{name:'Resend to inbox…',exact:true}).click();assert(await enable.isDisabled());
 await page.getByLabel(/I want to resend/).check();
 await page.setViewportSize({width:480,height:850});await page.screenshot({path:out+'confirm.png',fullPage:true});
 await enable.click();
 await page.getByText('Ready for another send. Confirm the account and language below.').waitFor();
 assert.equal(await page.evaluate(()=>window.resets.length),1);assert.equal(await page.evaluate(()=>window.sends.length),0);
 const args=await page.evaluate(()=>window.resets[0]);assert.equal(args[0].language,'en');assert.equal(args[1],'tiktok-en');assert.equal(args[3],'old-inbox');assert.equal(args[4],true);
 await page.getByText(/Previous attempts/).click();await page.getByText('old-inbox',{exact:true}).waitFor();
 const send=page.getByRole('button',{name:'Send to TikTok inbox',exact:true});assert(await send.isDisabled());
 await page.getByLabel('I confirm tiktok EN and this private / inbox transfer.').check();await send.click();
 const final=page.getByRole('button',{name:'Send English video',exact:true});assert(await final.isDisabled());
 assert.equal(await page.evaluate(()=>window.sends.length),0);
 await page.getByLabel('I confirm this is the English version for these accounts.').check();await final.click();
 await page.waitForFunction(()=>window.sends.length===1);
 assert.equal(await page.evaluate(()=>window.sends[0][0].language),'en');assert.deepEqual(errors,[]);
 console.log('PASS: inbox resend confirmation, cancellation, receipt history, no send during reset, final language gate and exact destination');
}finally{await browser?.close();await server.close();}
