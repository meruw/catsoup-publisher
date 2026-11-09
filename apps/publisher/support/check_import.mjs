// Browser fixture check; native folder selection is checked separately.
import {readFile,mkdir} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
const root=fileURLToPath(new URL('../../../',import.meta.url));
const manifest=JSON.parse(await readFile(path.join(root,'deliveries/catsoup-demo/manifest.json'),'utf8'));
const episode={id:manifest.episode_id,title:manifest.internal_title,version:'v1',status:'exported',ready:false,reason:'Pending review',languages:manifest.languages,videos:manifest.videos,warnings:[]};
const server=await createServer({configFile:false,root:path.join(root,'apps/publisher/frontend'),plugins:[vue(),{name:'import-fixture',enforce:'pre',load(id){
 const file=id.replaceAll('\\','/');
 if(file.endsWith('/services/publisher.ts')) return `const episode=${JSON.stringify(episode)}; export const publisher={list:async()=>window.imported?[{id:episode.id,title:episode.title,valid:true}]:[],get:async()=>episode,importDelivery:async()=>{await new Promise(r=>setTimeout(r,200));if(window.mode==='cancel')return null;if(window.mode==='error')throw Error('Invalid delivery: file hash mismatch');const existing=!!window.imported;window.imported=true;return {id:episode.id,existing}}};export const mediaURL=()=>'';`;
 if(file.endsWith('/services/drafts.ts'))return `export const drafts={enableCloseGuard:async()=>{},onCloseRequested:f=>{window.requestClose=f;return()=>{}},confirmClose:async()=>{window.closedByPublisher=true},groups:async()=>({items:[],more:false}),list:async()=>[],get:async key=>({key,title:"",description:"",hashtags:"",revision:0,deliveryMode:"draft"})};`;
 if(file.endsWith('/services/uploads.ts'))return `export const uploads={accounts:async()=>[]};`;
 if(file.endsWith('/services/window.ts'))return `export const desktopWindow={onMaximisedChange:()=>()=>{},isMaximised:async()=>false};`;
}}],server:{host:'127.0.0.1',port:0}});
let browser;
try{
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:1200,height:840}});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
 await page.goto(server.resolvedUrls.local[0]);const button=page.getByRole('button',{name:'Import delivery',exact:true});await button.waitFor();
 await page.evaluate(()=>window.mode='cancel');await button.click();await page.getByRole('heading',{name:'No delivery to preview'}).waitFor();await button.waitFor();assert.equal(await page.locator('.import-feedback').count(),0);
 await page.evaluate(()=>window.mode='error');await button.click();await page.getByRole('alert').filter({hasText:'file hash mismatch'}).waitFor();
 await page.evaluate(()=>window.mode='success');await button.click();await page.getByRole('status').filter({hasText:'Delivery imported'}).waitFor();await page.getByRole('heading',{name:episode.title,exact:true}).waitFor();
 assert.equal(await page.locator('.episode-button').count(),1);
 await button.click();await page.getByRole('status').filter({hasText:'Already imported'}).waitFor();assert.equal(await page.locator('.episode-button').count(),1);
 await mkdir(path.join(root,'work/import-check'),{recursive:true});await page.screenshot({path:path.join(root,'work/import-check/import.png'),fullPage:true});
 await page.setViewportSize({width:820,height:650});assert(await button.isVisible());assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));assert.equal(await page.getByText('Could not open draft.',{exact:false}).count(),0);assert.deepEqual(errors,[]);
 await page.evaluate(async()=>{const q=await import('/src/services/transfers.ts');q.enqueueTransfers([{draft:{key:{platform:'facebook',language:'es',account:'account:page',episode:'fixture'},title:'A short story',videoHash:'fixture'},profile:{name:'CATsoup ESP'},mode:'Publish now',send:()=>new Promise(resolve=>{window.finishUpload=()=>resolve('Published')})}]);window.requestClose()});
 await page.getByRole('dialog',{name:'Uploads',exact:true}).waitFor();assert.equal(await page.evaluate(()=>!!window.closedByPublisher),false,'close guard protects active upload');await page.getByRole('button',{name:'Close uploads',exact:true}).click();
 await page.getByRole('button',{name:'Posts',exact:true}).click();await page.getByRole('heading',{name:'Posts',exact:true}).waitFor();
 await page.evaluate(()=>window.finishUpload());await page.locator('#uploads-launcher').getByText('Finished',{exact:true}).waitFor();await page.evaluate(()=>window.requestClose());await page.waitForFunction(()=>window.closedByPublisher===true);
 console.log('PASS: import states, compact layout, app navigation during upload and close protection until completion');
}finally{await browser?.close();await server.close()}
