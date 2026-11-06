import {readFile} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createServer} from '../frontend/node_modules/vite/dist/node/index.js';
import {chromium} from '../../../.tools/visual-qa/node_modules/playwright/index.mjs';
const root=fileURLToPath(new URL('../../../',import.meta.url));
const server=await createServer({configFile:false,root:path.join(root,'apps/publisher/frontend'),server:{host:'127.0.0.1',port:0}});
let browser;
try {
 await server.listen();browser=await chromium.launch({channel:'chrome',headless:true});
 const page=await browser.newPage();
 const media=await readFile(path.join(root,'deliveries/catsoup-demo/video/es_final.mp4'));
 await page.route('**/sample.mp4',route=>{
 const range=/^bytes=(\d+)-(\d*)$/.exec(route.request().headers().range||'');
 if(!range)return route.fulfill({contentType:'video/mp4',headers:{'Accept-Ranges':'bytes'},body:media});
 const start=Number(range[1]),end=range[2]?Math.min(Number(range[2]),media.length-1):media.length-1;
 return route.fulfill({status:206,contentType:'video/mp4',headers:{'Accept-Ranges':'bytes','Content-Range':`bytes ${start}-${end}/${media.length}`},body:media.subarray(start,end+1)});
 });
 await page.route('**/thumbnail-check',route=>route.fulfill({contentType:'text/html',body:'<!doctype html><title>Thumbnail check</title>'}));
 await page.goto(server.resolvedUrls.local[0]+'thumbnail-check');
 const result=await page.evaluate(async()=>{
  const {thumbnail}=await import('/src/services/thumbnails.ts');
  const create=document.createElement.bind(document);let created=0,active=0,peak=0;
  document.createElement=function(name,...args){const el=create(name,...args);if(name==='video'){created++;active++;peak=Math.max(peak,active);const load=el.load.bind(el);el.load=()=>{if(!el.hasAttribute('src'))active--;return load();};}return el;};
  const source=await thumbnail('/sample.mp4','check',new AbortController().signal);
  const img=new Image();img.src=source;await img.decode();
  const canvas=document.createElement('canvas');canvas.width=img.width;canvas.height=img.height;
  const ctx=canvas.getContext('2d');ctx.drawImage(img,0,0);
  const pixels=ctx.getImageData(0,0,img.width,img.height).data;
  let colorful=0,opaque=0;for(let i=0;i<pixels.length;i+=4){if(pixels[i+3]>0)opaque++;if(Math.max(pixels[i],pixels[i+1],pixels[i+2])-Math.min(pixels[i],pixels[i+1],pixels[i+2])>50)colorful++;}
  const cached=await thumbnail('/sample.mp4','check',new AbortController().signal);
  const reused=cached===source&&created===1;
  await Promise.all([thumbnail('/sample.mp4','second',new AbortController().signal),thumbnail('/sample.mp4','third',new AbortController().signal)]);
  const cancelled=new AbortController();cancelled.abort();let rejected=false;
  try{await thumbnail('/sample.mp4','cancelled',cancelled.signal);}catch{rejected=true;}
  document.createElement=create;
  return {reused,rejected,created,active,peak,width:img.width,height:img.height,colorful,opaque,bytes:source.length};
 });
 console.log(result);assert(result.colorful>100,'thumbnail must contain decoded color-test pixels');assert(result.width<=192&&result.height<=192);assert(result.reused);assert(result.rejected);assert.equal(result.created,3);assert.equal(result.peak,1);assert.equal(result.active,0);
} finally {await browser?.close();await server.close();}
