<script setup lang="ts">
import { computed } from 'vue';
const props=defineProps<{state?:string;detail?:boolean}>();
const states = {
 missing: ['Not configured','Enter or import your credentials to enable sign-in.',''],
 invalid: ['Needs attention','The saved configuration is unreadable or incomplete. Replace it to continue.','warning'],
 saved: ['Saved · unverified','Saved locally. Sign in from Accounts to confirm these credentials.',''],
 confirmed: ['Confirmed at sign-in','The provider accepted these credentials at a previous sign-in.','success']
} as const;
const entry=computed(()=>states[props.state as keyof typeof states] ?? ['Status unavailable','','']);
</script>
<template>
 <span><span v-if="!detail || !entry[1]" class="ui-status" :class="entry[2] && 'ui-status--'+entry[2]"><span class="status-dot" aria-hidden="true"></span>{{ entry[0] }}</span>
  <span v-if="detail && entry[1]" class="state-hint">{{ entry[1] }}</span>
 </span>
</template>
<style scoped>
.state-hint {display:block;color:var(--secondary);margin-top:var(--space-1);font-size:var(--text-sm)}
</style>
