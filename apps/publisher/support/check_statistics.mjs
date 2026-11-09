// Browser-only counters: no provider requests, credentials or publishing.
import assert from 'node:assert/strict';
import {mkdir} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
const root=fileURLToPath(new URL('../../../',import.meta.url));
const server=await createServer({configFile:false,root:path.join(root,'apps/publisher/frontend'),plugins:[{name:'statistics-fixture',enforce:'pre',resolveId(id){if(id==='/statistics.ts')return id},load(id){
 if(id==='/statistics.ts')return `import {createApp,h,ref} from 'vue';import PostStatistics from '/src/components/PostStatistics.vue';import UiDialog from '/src/components/UiDialog.vue';import '/src/style.css';
 createApp({setup(){const post=ref({key:{episode:'catsoup-demo',platform:'youtube',language:'es',account:'account:es'},reference:'abcdefghijk',videoId:'es_final',state:'published',connected:true});window.switchAccount=(language='en')=>{post.value={...post.value,key:{...post.value.key,language,account:'account:'+language}}};window.switchPlatform=platform=>{post.value={...post.value,key:{...post.value.key,platform}}};const opened=ref(true),busy=ref(false);window.reopen=()=>opened.value=true;window.lockDialog=value=>busy.value=value;return()=>opened.value?h(UiDialog,{title:'YouTube / CATsoup',busy:busy.value,onClose:()=>opened.value=false},()=>[h('p',{style:'color:var(--success)'},'Published'),h('p','The price of a light'),h(PostStatistics,{post:post.value})]):h('p','Closed')}}).mount('#app');`;
 const file=id.replaceAll('\\','/');
 if(file.endsWith('/services/uploads.ts'))return `export const uploads={statistics:async(key)=>{window.calls=(window.calls||0)+1;window.lastKey=key;return new Promise((resolve,reject)=>{window.finish=(items=[{label:'Views',value:'1248'},{label:'Likes',value:'0'},{label:'Comments',value:'6'}])=>resolve({items,updatedAt:'2026-11-07T18:00:00Z',note:'',history:window.historyFixture||[]});window.fail=()=>reject(Error('Provider temporarily unavailable'))})}};`;
 },configureServer(s){s.middlewares.use((req,res,next)=>{if(req.url!=='/statistics')return next();res.setHeader('Content-Type','text/html');res.end('<div id="app"></div><script type="module" src="/statistics.ts"></script>')})}},vue()],server:{host:'127.0.0.1',port:0}});
let browser;
try{
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:1100,height:780}});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
 await page.goto(server.resolvedUrls.local[0]+'statistics');await page.waitForFunction(()=>window.calls===1);assert.equal(await page.evaluate(()=>window.lastKey.language),'es');
 await page.evaluate(()=>{
 window.historyFixture=[150,400,880,1248].map((value,i)=>({updatedAt:new Date(Date.now()-(3-i)*86400000).toISOString(),items:[{label:'Views',value:String(value)},{label:'Likes',value:'0'},{label:'Comments',value:String(i*2)}]}));window.finish();
});await page.getByText('1,248',{exact:true}).waitFor();await page.locator('.metric-strip').getByText('0',{exact:true}).waitFor();
 await page.locator('.statistics-chart svg').waitFor();
 await page.locator('.statistics-chart svg').focus();await page.keyboard.press('Home');
 await page.getByText('150 views',{exact:true}).waitFor();
 await page.keyboard.press('End');await page.getByText('1,248 views',{exact:true}).waitFor();
 await page.getByRole('button',{name:'Likes',exact:true}).click();
 await page.getByText('0 likes',{exact:true}).waitFor();
 assert(!/NaN|Infinity/.test(await page.locator('.trend').getAttribute('d')),'zero series stays finite');
 await page.getByRole('button',{name:'Views',exact:true}).click();
 await page.getByLabel('Chart period').selectOption('7');
 await mkdir(path.join(root,'work/statistics-check'),{recursive:true});await page.screenshot({path:path.join(root,'work/statistics-check/statistics.png')});
 await page.getByRole('button',{name:'Refresh statistics'}).click();await page.evaluate(()=>window.fail());await page.getByRole('alert').waitFor();assert(await page.getByText(/Last successful update/).count());assert(await page.getByText('1,248',{exact:true}).count());
 await page.getByRole('button',{name:'Refresh statistics'}).click();await page.evaluate(()=>{const oldFinish=window.finish;window.historyFixture=[];window.switchAccount();oldFinish([{label:'Views',value:'999'}])});await page.waitForFunction(()=>window.calls===4);assert.equal(await page.locator('.metric-strip').count(),0);
 assert.equal(await page.evaluate(()=>window.lastKey.language),'en');await page.evaluate(()=>window.finish([{label:'Likes',value:'0'}]));await page.locator('.metric-strip').getByText('0',{exact:true}).waitFor();assert.equal(await page.getByText('Views',{exact:true}).count(),0);
 // A click in the content or a drag ending outside does not close the dialog.
 await page.getByRole('heading',{name:'Statistics',exact:true}).click();assert.equal(await page.getByRole('dialog').count(),1);
 const box=await page.getByRole('dialog').boundingBox();await page.mouse.move(box.x+20,box.y+20);await page.mouse.down();await page.mouse.move(2,2);await page.mouse.up();assert.equal(await page.getByRole('dialog').count(),1);
 await page.evaluate(()=>window.lockDialog(true));await page.mouse.click(2,2);assert.equal(await page.getByRole('dialog').count(),1);
 await page.evaluate(()=>window.lockDialog(false));await page.mouse.click(2,2);await page.getByText('Closed',{exact:true}).waitFor();
 await page.evaluate(()=>window.reopen());await page.locator('.metric-strip').getByText('0',{exact:true}).waitFor();assert.equal(await page.evaluate(()=>window.calls),4,'reopening reuses cached counts');
 for(const platform of ['tiktok','x']){await page.evaluate(p=>window.switchPlatform(p),platform);await page.waitForFunction(()=>!document.querySelector('.metric-strip'));assert.equal(await page.evaluate(()=>window.calls),4,'unsupported platforms never request counts');}
 await page.evaluate(()=>{window.switchPlatform('youtube');window.switchAccount('en')});await page.locator('.metric-strip').getByText('0',{exact:true}).waitFor();
 await page.setViewportSize({width:420,height:700});assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));assert.deepEqual(errors,[]);
 console.log('PASS: automatic reads, cache reuse, zero vs missing, stale response isolation, unsupported platforms, error retains dated counts, backdrop/busy/drag protection, responsive layout.');
}finally{await browser?.close();await server.close()}
