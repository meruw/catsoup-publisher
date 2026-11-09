import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
const root=fileURLToPath(new URL('../frontend/',import.meta.url));
const rows=['youtube','instagram','facebook','tiktok','x'].map((platform,i)=>({key:{episode:'fixture',platform,language:'es',account:'account:'+platform},title:'Choose your world',videoId:'es_final',accountName:'catsoup_es',connected:true,mode:'public',reference:'123',state:i<3?'published':i===3?'inbox':'check',receiptState:i<3?'published':i===3?'inbox':'web_ready',updatedAt:'2026-11-08T10:00:00Z',startedAt:'2026-11-08T09:00:00Z',publishedAt:platform==='youtube'?'2026-11-08T09:30:00Z':'',confirmedAt:platform==='instagram'?'':'2026-11-08T09:35:00Z'}));
const group={episode:'fixture',title:'Choose your world',videoId:'es_final',languages:'es',platforms:5,total:5,counts:{published:3,inbox:1,check:1},updatedAt:'2026-11-08T10:00:00Z'};
const mocks={
 '/services/drafts.ts':`export const drafts={groups:async()=>({items:[${JSON.stringify(group)}],more:false}),statuses:async()=>({items:${JSON.stringify(rows)},more:false})}`,
 '/services/statistics.ts':`export const statisticsKey=p=>JSON.stringify(p.key);export const statisticsNote=p=>p.key.platform==='tiktok'||p.key.platform==='x'?'Not available':'';export const readStatistics=async()=>({publishedAt:'2026-11-08T09:30:00Z',items:[{label:'Likes',value:'2'}],updatedAt:'2026-11-09T09:30:00Z',note:'',history:[]});`,
 '/services/publisher.ts':`export const mediaURL=()=>'';`,
 '/services/thumbnails.ts':`export const thumbnail=async()=>'';`
};
const fixture={name:'times-fixture',enforce:'pre',transform(code,id){for(const [suffix,text] of Object.entries(mocks))if(id.replaceAll('\\','/').endsWith(suffix))return text;},configureServer(server){server.middlewares.use('/fixture',async(req,res)=>{res.setHeader('Content-Type','text/html');res.end(await server.transformIndexHtml('/fixture',`<div id="app"></div><script type="module">import {createApp,h} from 'vue';import View from '/src/components/PostOverview.vue';import '/src/style.css';createApp({render(){return h(View,{entries:[{id:'fixture',title:'Choose your world',valid:true}],openError:'',initialEpisode:'fixture'})}}).mount('#app');</script>`));});}};
const server=await createServer({configFile:false,root,plugins:[fixture,vue()],server:{host:'127.0.0.1',port:0}});let browser;
try{
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:1000,height:900},timezoneId:'America/Mexico_City',locale:'en-US'});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
 await page.goto(server.resolvedUrls.local[0]+'fixture');await page.locator('.destination-row').first().waitFor();
 assert.match(await page.locator('.destination-row').nth(0).innerText(),/Published.*Nov 8, 2026, 03:30 AM/s);
 assert.match(await page.locator('.destination-row').nth(1).innerText(),/Publication time unavailable/);
 assert.match(await page.locator('.destination-row').nth(2).innerText(),/Publication confirmed/);
 assert.match(await page.locator('.destination-row').nth(3).innerText(),/Sent to inbox/);
 assert.match(await page.locator('.destination-row').nth(4).innerText(),/Prepared in browser/);
 await page.screenshot({path:fileURLToPath(new URL('../../../work/post-times-list.png',import.meta.url))});
 await page.locator('.destination-row').nth(1).click();await page.locator('.post-details').getByText('Nov 8, 2026, 03:30 AM',{exact:true}).waitFor();
 assert.match(await page.locator('.post-details').innerText(),/Latest attempt started/);
 await page.screenshot({path:fileURLToPath(new URL('../../../work/post-times-detail.png',import.meta.url))});
 await page.getByRole('button',{name:'Close post details'}).click();assert.match(await page.locator('.destination-row').nth(1).innerText(),/Published.*Nov 8, 2026, 03:30 AM/s);
 await page.setViewportSize({width:620,height:800});assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
 assert.deepEqual(errors,[]);console.log('PASS: local zone, publication vs confirmation/inbox/browser, remote date refresh, modal, narrow layout');
}finally{await browser?.close();await server.close();}
