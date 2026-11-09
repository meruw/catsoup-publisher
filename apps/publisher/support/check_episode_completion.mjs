import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const entries=['done','partial','new-version','invalid','new'].map(id=>({id,title:id,valid:id!=='invalid',fingerprint:id==='new-version'?'v2':'v1',languages:['es','en'],problem:''}));
const completed=['done','partial','new-version','invalid'].flatMap(episode=>(episode==='partial'?['es']:['es','en']).map(language=>({episode,language,fingerprint:'v1'})));
const fixture={name:'completion-fixture',configureServer(server){server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">import {createApp,h,ref,computed} from 'vue';import Browser from '/src/components/EpisodeBrowser.vue';import {pendingDeliveries} from '/src/services/completion';import '/src/style.css';const entries=${JSON.stringify(entries)},completed=ref(${JSON.stringify(completed)});const pending=computed(()=>pendingDeliveries(entries,completed.value));window.pending=()=>pending.value.map(e=>e.id);window.finish=()=>completed.value.push({episode:'partial',language:'en',fingerprint:'v1'});window.reset=()=>completed.value=[];createApp({render(){return h('aside',{class:'sidebar',style:'width:280px;height:700px'},[h(Browser,{entries:pending.value,completed:entries.length-pending.value.length,disabled:false,selected:'done',onSelect:id=>window.selected=id})])}}).mount('#app');</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[fixture,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage();const errors=[];page.on('pageerror',e=>errors.push(String(e)));await page.goto(server.resolvedUrls.local[0]+'fixture');await page.locator('.episode-button').first().waitFor();
assert.deepEqual(await page.evaluate(()=>window.pending()),['partial','new-version','invalid','new']);
await page.getByRole('button',{name:'partial',exact:true}).click();assert.equal(await page.evaluate(()=>window.selected),'partial');
await page.evaluate(()=>window.finish());await page.waitForFunction(()=>!window.pending().includes('partial'));
await page.getByLabel('Search episodes').fill('done');await page.getByText('No matching episodes',{exact:true}).waitFor();
await page.getByLabel('Search episodes').fill('');await page.evaluate(()=>window.reset());await page.getByRole('button',{name:'done',exact:true}).waitFor();
assert.deepEqual(errors,[]);console.log('PASS: bilingual completion, partial/version/invalid retained, live completion removes row, failed status refresh restores visibility');
}finally{await browser?.close();await server.close();}
