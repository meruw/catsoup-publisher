<script setup lang="ts">
import UiIcon from './UiIcon.vue';
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
      <button aria-label="Minimize window" title="Minimize" @click="run(desktopWindow.minimise)"><UiIcon name="minus"/></button>
      <button :aria-label="maximised ? 'Restore window' : 'Maximize window'" :title="maximised ? 'Restore' : 'Maximize'" @click="run(desktopWindow.toggleMaximise)">
        <UiIcon :name="maximised ? 'copy' : 'square'"/>
      </button>
      <button class="window-close" aria-label="Close window" title="Close" @click="run(desktopWindow.close)"><UiIcon name="close"/></button>
    </div>
  </header>
</template>
