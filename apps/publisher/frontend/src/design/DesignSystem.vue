<script setup lang="ts">
import { ref } from 'vue';
import UiButton from '../components/UiButton.vue';
import UiSegmented from '../components/UiSegmented.vue';
const language = ref('es');
const colors = [
  ['Canvas', '--page', '#FFFCF7'], ['Cream', '--sidebar', '#FBF3E8'],
  ['Cocoa', '--text', '#4F281B'], ['Orange', '--accent', '#FF793F'],
  ['Selection', '--accent-soft', '#F9E4D3'], ['Secondary', '--secondary', '#705446'],
];
</script>

<template>
  <div class="specimen">
    <header class="specimen-heading"><img src="/brand/catsoup-rounded.png" alt="" width="64" height="64"><div><h1>CATsoup foundations</h1><p class="type-subtitle">Warm, focused, open</p></div></header>
    <p class="specimen-intro">Space and typography organize the workspace. No card grids, colored selection rails or decorative shadows.</p>
    <section class="specimen-section"><h2>Color</h2><div class="swatches"><div v-for="[name, token, value] in colors" :key="token"><div class="swatch" :style="{ background: `var(${token})` }"></div><h3>{{ name }}</h3><p class="mono">{{ value }}</p></div></div></section>
    <section class="specimen-section specimen-columns">
      <div><h2>Typography</h2><p class="type-caption">Instrument Sans · bundled locally</p><div class="type-samples"><p class="type-title">Your next episode</p><p class="type-subtitle">Review the Spanish and English exports</p><p class="type-section">Delivery details</p><p class="type-label">Video version</p><p class="type-body">One episode, two languages, one clear workspace.</p><p class="type-caption">Metadata · 12 px / 16 px</p></div></div>
      <div><h2>Actions</h2><p class="type-caption">40 px default · 32 px compact · 8 px radius</p><div class="button-samples"><UiButton variant="primary">Primary</UiButton><UiButton>Secondary</UiButton><UiButton variant="quiet">Quiet</UiButton><UiButton variant="danger">Destructive</UiButton><UiButton disabled>Disabled</UiButton><UiButton busy>Working…</UiButton></div><p class="specimen-help">Orange uses dark text. One primary action per task. Keyboard focus stays visible.</p><UiSegmented v-model="language" label="Example language" :options="[{ value: 'es', label: 'ES', accessibleLabel: 'Spanish' }, { value: 'en', label: 'EN', accessibleLabel: 'English' }]" /></div>
    </section>
    <section class="specimen-section specimen-columns">
      <div><h2>Fields</h2><div class="field-example"><label for="example-title">Episode title</label><input id="example-title" class="ui-field" value="The streetlight"><p class="type-caption">Labels remain visible above the field</p></div><div class="field-example"><label for="example-invalid">Required title</label><input id="example-invalid" class="ui-field" aria-invalid="true" aria-describedby="title-error"><p id="title-error" class="ui-status ui-status--danger">Enter an episode title</p></div></div>
      <div><h2>Status</h2><div class="status-samples"><span class="ui-status ui-status--neutral">Draft · In progress</span><span class="ui-status ui-status--warning">Preview only · Approval needed</span><span class="ui-status ui-status--success">Approved · Owner review complete</span><span class="ui-status ui-status--danger">Error · File unavailable</span></div><p class="specimen-help">Meaning always appears in text. Orange is an interaction color, not an approval state.</p></div>
    </section>
    <section class="specimen-section"><h2>Spacing and surfaces</h2><p class="specimen-help">4 · 8 · 12 · 16 · 20 · 24 · 32 · 40 · 48 px. Group with space and occasional quiet rules. Reserve outlined surfaces for inputs and media.</p></section>
  </div>
</template>

<style scoped>
.specimen { max-width: 1120px; margin: 0 auto; padding: var(--space-12); }
.specimen-heading { display: flex; align-items: center; gap: var(--space-4); }
.specimen-intro { margin: var(--space-6) 0; color: var(--secondary); }
.specimen-section { border-top: 1px solid var(--line); padding: var(--space-8) 0; }
.specimen-section h2 { margin-bottom: var(--space-3); }
.swatches { display: grid; grid-template-columns: repeat(6, 1fr); gap: var(--space-4); margin-top: var(--space-6); }
.swatch { height: 56px; margin-bottom: var(--space-3); }
.swatches .mono { color: var(--muted); margin-top: var(--space-1); }
.specimen-columns { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-12); }
.type-samples, .status-samples { display: grid; gap: var(--space-3); margin-top: var(--space-6); }
.button-samples { display: flex; flex-wrap: wrap; gap: var(--space-3); margin-top: var(--space-6); }
.specimen-help { color: var(--secondary); margin: var(--space-4) 0; }
.field-example { display: grid; gap: var(--space-2); margin-top: var(--space-4); }
@media (max-width: 760px) { .specimen { padding: var(--space-6); } .specimen-columns { grid-template-columns: 1fr; } .swatches { grid-template-columns: repeat(3, 1fr); } }
</style>
