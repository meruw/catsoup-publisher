// Static, session-only previews. Decode one local video at a time, then release it.
const cache = new Map<string, {image:string; expires:number}>();
let queue:Promise<unknown> = Promise.resolve();

export function thumbnail(source:string, stamp:string, signal:AbortSignal):Promise<string> {
  const key=source+'|'+stamp;
  const job=queue.then(async()=>{
    if(signal.aborted) throw new DOMException('Cancelled','AbortError');
    const cached=cache.get(key);
    if(cached && cached.expires>Date.now()) {
      cache.delete(key);cache.set(key,cached);
      return cached.image;
    }
    const image=await frame(source,signal);
    cache.delete(key);cache.set(key,{image,expires:Date.now()+60_000});
    while(cache.size>64) cache.delete(cache.keys().next().value!);
    return image;
  });
  queue=job.catch(()=>{});
  return job;
}

function frame(source:string, signal:AbortSignal):Promise<string> {
  return new Promise((resolve,reject)=>{
    const video=document.createElement('video');
    video.muted=true;video.preload='auto';video.playsInline=true;
    let done=false,target=-1,frameRequest=0;
    const cleanup=()=>{
      clearTimeout(timeout);if(frameRequest)video.cancelVideoFrameCallback(frameRequest);signal.removeEventListener('abort',cancel);
      video.onloadedmetadata=null;video.onloadeddata=null;video.onseeked=null;video.onerror=null;
      video.pause();video.removeAttribute('src');video.load();
    };
    const finish=(image?:string)=>{
      if(done)return;done=true;cleanup();
      if(image)resolve(image);else reject(new Error('Thumbnail unavailable'));
    };
    const cancel=()=>finish();
    const timeout=setTimeout(()=>finish(),15_000);
    signal.addEventListener('abort',cancel,{once:true});
    const capture=()=>{
      if(done||target<0||video.seeking||video.readyState<2||Math.abs(video.currentTime-target)>.1)return;
      try {
        if(!video.videoWidth||!video.videoHeight)return finish();
        const scale=Math.min(1,192/Math.max(video.videoWidth,video.videoHeight));
        const canvas=document.createElement('canvas');
        canvas.width=Math.max(1,Math.round(video.videoWidth*scale));
        canvas.height=Math.max(1,Math.round(video.videoHeight*scale));
        const context=canvas.getContext('2d');
        if(!context)return finish();
        context.drawImage(video,0,0,canvas.width,canvas.height);
        finish(canvas.toDataURL('image/webp',.75));
      }catch{finish();}
    };
    video.onloadedmetadata=()=>{
      if(!Number.isFinite(video.duration)||video.duration<=0)return finish();
      target=Math.min(1,video.duration/4);
      const rendered:VideoFrameRequestCallback=(_,metadata)=>{
        if(done)return;
        if(metadata.mediaTime>=target-.1)capture();
        if(!done)frameRequest=video.requestVideoFrameCallback(rendered);
      };
      frameRequest=video.requestVideoFrameCallback(rendered);
      video.currentTime=target;
    };
    video.onerror=()=>finish();
    if(signal.aborted)return finish();
    video.src=source;
  });
}
