<script setup lang="ts">
import { ref, watch } from 'vue';

const props=defineProps<{ source: string; language: string; active?:boolean }>();
const player=ref<HTMLVideoElement|null>(null);
watch(()=>props.active,active=>{if(active===false)player.value?.pause()});
const state = ref<'loading' | 'ready' | 'error'>('loading');
</script>

<template>
  <div class="video-stage" :aria-busy="state === 'loading'">
    <video ref="player" :src="source" :aria-label="`${language} video preview`"
      controls preload="metadata" playsinline
      @loadeddata="state = 'ready'" @error="state = 'error'" />
    <p v-if="state === 'loading'" class="player-message" role="status">Loading preview…</p>
    <p v-if="state === 'error'" class="player-message" role="alert">Preview unavailable. Refresh to check the delivery.</p>
  </div>
</template>
