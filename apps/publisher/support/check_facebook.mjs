// Isolated browser fixture: no native bridge, credentials or provider traffic.
import assert from 'node:assert/strict';
import {mkdir} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import vue from '../frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
const root=fileURLToPath(new URL('../../../',import.meta.url));
const output=path.join(root,'work/publisher-ui-check');await mkdir(output,{recursive:true});
const server=await createServer({configFile:false,root:path.join(root,'apps/publisher/frontend'),plugins:[{name:'facebook-fixture',enforce:'pre',resolveId(id){if(id==='/facebook-fixture.ts')return id;},load(id){
 if(id==='/facebook-fixture.ts')return `import {createApp,h} from 'vue';import AccountManager from '/src/components/AccountManager.vue';import '/src/style.css';createApp({render:()=>h(AccountManager,{initialLanguage:'es',onSettings:(p)=>window.__settings=p})}).mount('#app');`;
 if(id.replaceAll('\\','/').endsWith('/src/services/uploads.ts'))return `
 const rows=[{platform:'instagram',language:'es',id:'ig-es',name:'catsoup_es',api:true},{platform:'instagram',language:'en',id:'ig-en',name:'catsoup_en',api:true}];
 export const uploads={accounts:async()=>rows,clientReady:async()=>!window.__missing,metaCallback:async()=> 'https://example.com/meta/',link:async(platform,language,expected,handle)=>{window.__linked={platform,language,expected,handle};const p={platform,language,id:language==='es'?'22':'33',name:language==='es'?'CATsoup ESP':'CATsoup EN',api:true};rows.push(p);return p}};`;
},configureServer(s){s.middlewares.use((req,res,next)=>{if(req.url!=='/facebook-fixture')return next();res.setHeader('Content-Type','text/html');res.end('<div id="app"></div><script type="module" src="/facebook-fixture.ts"></script>')});}},vue()],server:{host:'127.0.0.1',port:0}});
let browser;
try {
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});const page=await browser.newPage({viewport:{width:1120,height:850}});const errors=[];page.on('pageerror',e=>errors.push(String(e)));
 await page.goto(server.resolvedUrls.local[0]+'facebook-fixture');
 const facebook=page.getByRole('region',{name:'Facebook',exact:true});
 await facebook.getByRole('button',{name:'Connect',exact:true}).click();
 await page.getByLabel('Page ID or linked Instagram username').waitFor();
 assert.equal(await page.getByLabel('Page ID or linked Instagram username').inputValue(),'catsoup_es');
 await page.screenshot({path:path.join(output,'facebook-connect.png'),fullPage:true});
 await facebook.getByRole('button',{name:'Sign in with Facebook',exact:true}).click();
 await page.getByText('CATsoup ESP assigned to Spanish.').waitFor();
 assert.deepEqual(await page.evaluate(()=>window.__linked),{platform:'facebook',language:'es',expected:'',handle:'catsoup_es'});
 await facebook.getByRole('button',{name:'Change',exact:true}).click();assert.equal(await page.getByLabel('Page ID or linked Instagram username').inputValue(),'22');
 await facebook.getByRole('button',{name:'Cancel',exact:true}).click();
 await page.getByRole('button',{name:'English accounts',exact:true}).click();
 await facebook.getByRole('button',{name:'Connect',exact:true}).click();assert.equal(await page.getByLabel('Page ID or linked Instagram username').inputValue(),'catsoup_en');
 await facebook.getByRole('button',{name:'Sign in with Facebook',exact:true}).click();await page.getByText('CATsoup EN assigned to English.').waitFor();
 assert.equal((await page.evaluate(()=>window.__linked)).language,'en');
 await page.addInitScript(()=>window.__missing=true);await page.reload();
 await facebook.getByRole('button',{name:'Set up credentials',exact:true}).click();assert.equal(await page.evaluate(()=>window.__settings),'instagram');
 assert.deepEqual(errors,[]);console.log('PASS: Facebook connects independently per language, preselects the linked profile, confirms the Page and routes setup to shared Meta credentials.');
}finally{await browser?.close();await server.close()}
