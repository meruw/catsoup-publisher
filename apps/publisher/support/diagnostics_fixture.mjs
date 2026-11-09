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
  const rows=events.filter(e=>(!q.before||e.id<q.before)&&(!q.after||e.id>q.after)&&(!q.level||e.level===q.level)&&(!q.service||e.service===q.service)&&(!q.language||e.language===q.language)&&(!q.attempt||e.attempt===q.attempt)&&(!q.search||JSON.stringify(e).toLowerCase().includes(q.search.toLowerCase()))).reverse();
  return {events:rows.slice(0,200),more:rows.length>200,warning:''};
 },
 listen:async value=>{window.__diagnosticsListening=value;},
 onChanged:callback=>{listener=callback;return ()=>{listener=undefined;};},
 copyEvent:async id=>{window.__diagnosticCopied=id;},
 copyResults:async(q,minutes)=>{window.__diagnosticBulkCopy={q,minutes};return 7;},
 pasteText:async()=>window.__diagnosticClipboard||'',
 export:async q=>{window.__diagnosticExport=q;return true;},
 openFolder:async()=>{window.__diagnosticFolder=true;},
};
`;

export async function checkDiagnostics(page,output) {
 await page.setViewportSize({width:1200,height:840});
 await page.getByRole('button',{name:'Settings',exact:true}).click();
 assert.equal(await page.getByRole('heading',{name:'Diagnostics',exact:true}).count(),0);
 await page.locator('.console-launcher[aria-controls="diagnostics-console"]').click();
 const panel=page.getByRole('region',{name:'Diagnostics console'});
 await panel.getByRole('button',{name:'Older',exact:true}).waitFor();
 await page.waitForFunction(()=>window.__diagnosticsListening);
 assert.equal(await panel.locator('tbody tr').count(),200);
 await panel.getByRole('button',{name:'Older',exact:true}).click();
 await panel.getByRole('button',{name:'Resume live',exact:true}).waitFor();
 await panel.getByRole('button',{name:'Newer',exact:true}).click();
 await panel.getByRole('button',{name:'Filters',exact:true}).click();
 await panel.getByLabel('Source',{exact:true}).selectOption('instagram');
 await panel.getByLabel('Context',{exact:true}).selectOption('es');
 await panel.getByLabel('Level',{exact:true}).selectOption('ERROR');
 await page.evaluate(()=>window.__diagnosticClipboard='no_pages');
 await panel.getByRole('button',{name:'Paste',exact:true}).click();
 assert.equal(await panel.getByLabel('Search',{exact:true}).inputValue(),'no_pages');
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===45);
 assert.equal(await panel.locator('.filter-count').innerText(),'4 filters');
 assert.equal(await panel.locator('.stream-state').innerText(),'Paused');
 await panel.locator('tbody tr').first().locator('td').first().click();
 await panel.getByRole('button',{name:'Copy event',exact:true}).click();
 assert((await page.evaluate(()=>window.__diagnosticCopied))>0);
 await panel.locator('.event-detail').evaluate(el=>el.scrollTop=0);
 await page.screenshot({path:path.join(output,'diagnostics-desktop.png')});
 await panel.getByRole('button',{name:'Follow attempt',exact:true}).click();
 await panel.getByRole('button',{name:'Clear attempt filter'}).waitFor();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===7);
 await panel.getByRole('button',{name:'Copy',exact:true}).click();
 await panel.getByRole('button',{name:'Copy results',exact:true}).click();
 const copiedAttempt=await page.evaluate(()=>window.__diagnosticBulkCopy);
 assert(copiedAttempt.q.attempt.startsWith('attempt'));
 assert.equal(copiedAttempt.q.before,0);
 assert.equal(copiedAttempt.minutes,0);
 for (const minutes of [1,5,10,30]) {
  await panel.getByLabel('Copy time range',{exact:true}).selectOption(String(minutes));
  await panel.getByRole('button',{name:'Copy results',exact:true}).click();
  assert.equal((await page.evaluate(()=>window.__diagnosticBulkCopy)).minutes,minutes);
 }
 await panel.getByLabel('Copy time range',{exact:true}).selectOption('0');
 await page.keyboard.press('Escape');
 await panel.getByRole('button',{name:'More',exact:true}).click();
 await panel.getByRole('button',{name:'Export page',exact:true}).click();
 assert((await page.evaluate(()=>window.__diagnosticExport)).attempt.startsWith('attempt'));
 await page.keyboard.press('Escape');
 await panel.getByRole('button',{name:'Clear attempt filter'}).click();
 await panel.getByRole('button',{name:'Hide',exact:true}).click();
 await panel.getByLabel('Source',{exact:true}).selectOption('');
 await panel.getByLabel('Context',{exact:true}).selectOption('');
 await panel.getByRole('button',{name:'Resume live',exact:true}).click();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===200);
 await page.evaluate(()=>window.__emitDiagnostic());
 await panel.getByText('Account assigned to the selected language',{exact:true}).waitFor();
 await panel.getByRole('button',{name:'Pause',exact:true}).click();
 const first=await panel.locator('tbody tr').first().innerText();
 await page.evaluate(()=>window.__emitDiagnostic());
 assert.equal(await panel.locator('tbody tr').first().innerText(),first);
 await panel.getByRole('button',{name:'Filters',exact:true}).click();
 await page.setViewportSize({width:820,height:700});
 await page.screenshot({path:path.join(output,'diagnostics-compact.png')});
 assert(await panel.evaluate(e=>e.scrollWidth<=e.clientWidth));
 const resizer=panel.getByRole('separator',{name:'Resize diagnostics'});
 const oldHeight=await panel.evaluate(e=>e.getBoundingClientRect().height);
 await resizer.focus();await page.keyboard.press('ArrowUp');
 assert((await panel.evaluate(e=>e.getBoundingClientRect().height))>oldHeight);
 await panel.getByLabel('Search',{exact:true}).fill('nothing-matches');
 await panel.getByText('No matching events',{exact:true}).waitFor();
 await panel.getByRole('button',{name:'Collapse console',exact:true}).click();
 await page.waitForFunction(()=>!window.__diagnosticsListening);
 assert.equal(await panel.evaluate(e=>e.getBoundingClientRect().height),36);
 assert.equal(await panel.getByLabel('Search',{exact:true}).count(),0);
 await page.screenshot({path:path.join(output,'diagnostics-collapsed.png')});
 await page.locator('.console-launcher[aria-controls="diagnostics-console"]').click();
 await page.waitForFunction(()=>window.__diagnosticsListening);
 assert.equal(await panel.getByLabel('Search',{exact:true}).inputValue(),'nothing-matches');
 await panel.getByRole('button',{name:'Expand',exact:true}).click();
 assert.equal(await panel.evaluate(e=>e.getBoundingClientRect().left),0);
 assert((await panel.evaluate(e=>e.getBoundingClientRect().height))>600);
 await panel.getByRole('button',{name:'Restore',exact:true}).click();
 assert.equal(await panel.evaluate(e=>e.getBoundingClientRect().left),200);
 await panel.getByRole('button',{name:/^Filters/}).click();
 await panel.getByRole('button',{name:'Reset filters',exact:true}).click();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===200);
 assert.equal(await panel.locator('.filter-count').count(),0);
 await panel.getByRole('button',{name:'More',exact:true}).click();
 await panel.getByRole('button',{name:'Hide existing events',exact:true}).click();
 await panel.getByText('No events',{exact:true}).waitFor();
 await panel.getByRole('button',{name:'Resume live',exact:true}).click();
 await page.evaluate(()=>window.__emitDiagnostic());
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===1);
 await panel.getByRole('button',{name:'More',exact:true}).click();
 await panel.getByRole('button',{name:'Export page',exact:true}).click();
 assert((await page.evaluate(()=>window.__diagnosticExport)).after>0);
 await panel.getByRole('button',{name:'Show saved history',exact:true}).click();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===200);
 // Clear restores every display control, including filters, range, history and pause.
 await panel.getByLabel('Level',{exact:true}).selectOption('WARN');
 await panel.getByRole('button',{name:'Pause',exact:true}).click();
 await panel.getByRole('button',{name:'Copy',exact:true}).click();
 await panel.getByLabel('Copy time range',{exact:true}).selectOption('30');
 await panel.getByRole('button',{name:'Clear',exact:true}).click();
 await page.waitForFunction(()=>document.querySelectorAll('.log-table tbody tr').length===200);
 assert.equal(await panel.locator('.stream-state').innerText(),'● Live');
 assert.equal(await panel.locator('.filter-count').count(),0);
 assert.equal(await panel.getByLabel('Level',{exact:true}).count(),0);
 assert.equal(await panel.locator('.event-detail').count(),0);
 assert.equal(await panel.evaluate(e=>e.getBoundingClientRect().height),360);
 await panel.getByRole('button',{name:'Copy',exact:true}).click();
 assert.equal(await panel.getByLabel('Copy time range',{exact:true}).inputValue(),'0');
 await page.keyboard.press('Escape');
 await panel.getByRole('button',{name:'Collapse console',exact:true}).click();
 await page.keyboard.press('Control+Shift+D');
 await panel.getByLabel('Search',{exact:true}).waitFor();
 await page.keyboard.press('Escape');
 await page.waitForFunction(()=>!window.__diagnosticsListening);
 await page.keyboard.press('Control+Shift+D');
 await panel.getByLabel('Search',{exact:true}).waitFor();
 await page.keyboard.press('Escape');
}
