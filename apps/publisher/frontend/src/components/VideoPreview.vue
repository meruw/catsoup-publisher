<script setup lang="ts">
import { ref } from 'vue';

defineProps<{ source: string; language: string }>();
const state = ref<'loading' | 'ready' | 'error'>('loading');
</script>

<template>
  <div class="video-stage" :aria-busy="state === 'loading'">
    <video :src="source" :aria-label="`${language} video preview`"
      controls preload="metadata" playsinline
      @loadeddata="state = 'ready'" @error="state = 'error'" />
    <p v-if="state === 'loading'" class="player-message" role="status">Loading preview…</p>
    <p v-if="state === 'error'" class="player-message" role="alert">Preview unavailable. Refresh to check the delivery.</p>
  </div>
</template>
