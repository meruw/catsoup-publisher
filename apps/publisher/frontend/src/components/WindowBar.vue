<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { desktopWindow } from '../services/window';

const maximised = ref(false);
const error = ref('');
let unsubscribe: (() => void) | undefined;
async function run(action: () => Promise<unknown>) {
  error.value = '';
  try { await action(); }
  catch { error.value = 'Window control failed. Try again.'; }
}
onMounted(() => {
  unsubscribe = desktopWindow.onMaximisedChange(value => { maximised.value = value; });
  void run(async () => { maximised.value = await desktopWindow.isMaximised(); });
});
onUnmounted(() => unsubscribe?.());
</script>

<template>
  <header class="window-bar" aria-label="Application window">
    <div class="window-identity"><img src="/brand/catsoup-rounded.png" width="24" height="24" alt=""><span>CATsoup</span></div>
    <span v-if="error" class="window-error" role="alert">{{ error }}</span>
    <div class="window-controls">
      <button aria-label="Minimize window" title="Minimize" @click="run(desktopWindow.minimise)"><svg viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8h10" /></svg></button>
      <button :aria-label="maximised ? 'Restore window' : 'Maximize window'" :title="maximised ? 'Restore' : 'Maximize'" @click="run(desktopWindow.toggleMaximise)">
        <svg v-if="maximised" viewBox="0 0 16 16" aria-hidden="true"><path d="M5 5V3h8v8h-2M3 5h8v8H3z" /></svg>
        <svg v-else viewBox="0 0 16 16" aria-hidden="true"><path d="M3 3h10v10H3z" /></svg>
      </button>
      <button class="window-close" aria-label="Close window" title="Close" @click="run(desktopWindow.close)"><svg viewBox="0 0 16 16" aria-hidden="true"><path d="m4 4 8 8M12 4l-8 8" /></svg></button>
    </div>
  </header>
</template>
