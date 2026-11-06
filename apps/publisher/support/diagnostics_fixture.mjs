import assert from 'node:assert/strict';
import path from 'node:path';

export const diagnosticsFixture = `
let listener;
window.__diagnosticsListening=false;
const events=Array.from({length:450},(_,i)=>({
 id:i+1,time:new Date(Date.now()-(450-i)*1000).toISOString(),
 level:i%5===0?'ERROR':i%5===1?'WARN':i%5===2?'DEBUG':'INFO',
 msg:i%5===0?'Could not resolve the Instagram profile':i%5===1?'Rejected an unmatched browser response':'Received a page of Facebook results',
 session:'fixture',service:i%2===0?'instagram':'youtube',language:i%2===0?'es':'en',
 attempt:'attempt'+Math.floor(i/15),stage:i%5===0?'pages_failed':'pages_received',
 elapsed_ms:i*20,reason:i%5===0?'no_pages':'',http_status:i%5===0?403:200,
 provider_code:i%5===0?200:0,provider_subcode:0,pages:i%5===0?0:2,matches:0,
}));
window.__emitDiagnostic=()=>{
 const event={...events[events.length-1],id:events.length+1,msg:'Account assigned to the selected language',stage:'connection_finished',level:'INFO'};
 events.push(event);if(window.__diagnosticsListening)listener?.();
};
export const diagnostics={
 query:async q=>{
  const rows=events.filter(e=>(!q.before||e.id<q.before)&&(!q.level||e.level===q.level)&&(!q.service||e.service===q.service)&&(!q.language||e.language===q.language)&&(!q.attempt||e.attempt===q.attempt)&&(!q.search||JSON.stringify(e).toLowerCase().includes(q.search.toLowerCase()))).reverse();
  return {events:rows.slice(0,200),more:rows.length>200,warning:''};
 },
 listen:async value=>{window.__diagnosticsListening=value;},
 onChanged:callback=>{listener=callback;return ()=>{listener=undefined;};},
 copyEvent:async id=>{window.__diagnosticCopied=id;},
 export:async q=>{window.__diagnosticExport=q;return true;},
 openFolder:async()=>{window.__diagnosticFolder=true;},
};
`;

export async function checkDiagnostics(page,output) {
 await page.setViewportSize({width:1200,height:840});
 await page.getByRole('button',{name:'Settings',exact:true}).click();
 await page.getByRole('button',{name:'Open console',exact:true}).click();
 const panel=page.getByRole('region',{name:'Diagnostics console'});
 await panel.getByRole('button',{name:'Older',exact:true}).waitFor();
 await page.waitForFunction(()=>window.__diagnosticsListening);
 assert.equal(await panel.locator('tbody tr').count(),200);
 await panel.getByRole('button',{name:'Older',exact:true}).click();
 await panel.getByRole('button',{name:'Resume live',exact:true}).waitFor();
 await panel.getByRole('button',{name:'Newer',exact:true}).click();
 await panel.getByLabel('Service',{exact:true}).selectOption('instagram');
 await panel.getByLabel('Language',{exact:true}).selectOption('es');
 await panel.getByLabel('Level',{exact:true}).selectOption('ERROR');
 await panel.getByLabel('Search',{exact:true}).fill('no_pages');
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===45);
 await panel.locator('.event-button').first().click();
 await panel.getByRole('button',{name:'Copy event',exact:true}).click();
 assert((await page.evaluate(()=>window.__diagnosticCopied))>0);
 await page.screenshot({path:path.join(output,'diagnostics-desktop.png')});
 await panel.getByRole('button',{name:'Follow attempt',exact:true}).click();
 await panel.getByRole('button',{name:'Clear attempt filter'}).waitFor();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===7);
 await panel.getByRole('button',{name:'Export page',exact:true}).click();
 assert((await page.evaluate(()=>window.__diagnosticExport)).attempt.startsWith('attempt'));
 await panel.getByRole('button',{name:'Clear attempt filter'}).click();
 await panel.getByRole('button',{name:'Hide',exact:true}).click();
 await panel.getByLabel('Service',{exact:true}).selectOption('');
 await panel.getByLabel('Language',{exact:true}).selectOption('');
 await panel.getByRole('button',{name:'Resume live',exact:true}).click();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===200);
 await page.evaluate(()=>window.__emitDiagnostic());
 await panel.getByText('Account assigned to the selected language',{exact:true}).waitFor();
 await panel.getByRole('button',{name:'Pause',exact:true}).click();
 const first=await panel.locator('tbody tr').first().innerText();
 await page.evaluate(()=>window.__emitDiagnostic());
 assert.equal(await panel.locator('tbody tr').first().innerText(),first);
 await page.setViewportSize({width:820,height:700});
 await page.screenshot({path:path.join(output,'diagnostics-compact.png')});
 assert(await panel.evaluate(e=>e.scrollWidth<=e.clientWidth));
 const resizer=panel.getByRole('separator',{name:'Resize diagnostics'});
 const oldHeight=await panel.evaluate(e=>e.getBoundingClientRect().height);
 await resizer.focus();await page.keyboard.press('ArrowUp');
 assert((await panel.evaluate(e=>e.getBoundingClientRect().height))>oldHeight);
 await page.keyboard.press('Escape');
 await page.waitForFunction(()=>!window.__diagnosticsListening);
 await page.keyboard.press('Control+Shift+D');
 await panel.getByLabel('Search',{exact:true}).waitFor();
 await page.keyboard.press('Escape');
}
