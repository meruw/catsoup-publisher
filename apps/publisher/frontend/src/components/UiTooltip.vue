<script setup lang="ts">
import {ref,useId} from 'vue';
const props=withDefaults(defineProps<{text:string;disabled?:boolean}>(),{disabled:false});
const id=useId(),anchor=ref<HTMLElement|null>(null),visible=ref(false),left=ref(0),top=ref(0),above=ref(false);
function show(){if(!props.text||!anchor.value)return;const r=anchor.value.getBoundingClientRect();left.value=Math.max(12,Math.min(r.left,window.innerWidth-292));above.value=r.bottom+160>window.innerHeight;top.value=above.value?r.top-8:r.bottom+8;visible.value=true;}
</script>
<template>
 <span ref="anchor" class="control-tip" :class="{'is-disabled':disabled}" :tabindex="disabled?0:undefined" :aria-label="disabled?text:undefined" :aria-describedby="text?id:undefined" @mouseenter="show" @mouseleave="visible=false" @focusin="show" @focusout="visible=false" @keydown.esc.stop="visible=false">
  <slot/>
  <Teleport to="body"><span v-if="visible && text" :id="id" role="tooltip" class="control-tooltip" :style="{left:left+'px',top:top+'px',transform:above?'translateY(-100%)':undefined}">{{text}}</span></Teleport>
 </span>
</template>
<style scoped>
.control-tip{display:inline-flex;min-width:0;vertical-align:middle}.control-tip.is-disabled>:deep(button),.control-tip.is-disabled>:deep(select){pointer-events:none}.control-tip:focus-visible{outline:2px solid var(--accent-ink);outline-offset:3px;border-radius:6px}
.control-tooltip{position:fixed;z-index:1000;pointer-events:none;max-width:min(280px,calc(100vw - 24px));padding:12px;background:var(--text);color:var(--page);border-radius:8px;font-size:var(--text-xs);line-height:1.5;font-weight:400;overflow-wrap:anywhere}
</style>
