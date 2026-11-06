<!-- Accessible modal dialog using native <dialog> element -->
<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  open:      { type: Boolean, default: false },
  title:     { type: String, default: '' },
  size:      { type: String, default: 'md' }, // sm | md | lg
  closeOnBackdrop: { type: Boolean, default: true },
})
const emit = defineEmits(['close'])

const dialogEl = ref(null)

function close() { emit('close') }

function onBackdropClick(e) {
  if (props.closeOnBackdrop && e.target === dialogEl.value) close()
}

function onKeydown(e) {
  if (e.key === 'Escape') close()
}

watch(() => props.open, (val) => {
  if (!dialogEl.value) return
  if (val) {
    if (!dialogEl.value.open) dialogEl.value.showModal()
  } else {
    if (dialogEl.value.open) dialogEl.value.close()
  }
})

onMounted(() => {
  if (props.open && dialogEl.value && !dialogEl.value.open) dialogEl.value.showModal()
})
</script>

<template>
  <dialog ref="dialogEl" class="pos-dialog" :class="`pos-dialog--${size}`"
    @click="onBackdropClick" @keydown="onKeydown">
    <div class="pos-dialog__panel" role="document" @click.stop>
      <header v-if="title || $slots.header" class="pos-dialog__header">
        <slot name="header">
          <h2 class="pos-dialog__title">{{ title }}</h2>
        </slot>
        <button class="pos-dialog__close" aria-label="Close dialog" @click="close">✕</button>
      </header>
      <div class="pos-dialog__body">
        <slot />
      </div>
      <footer v-if="$slots.footer" class="pos-dialog__footer">
        <slot name="footer" />
      </footer>
    </div>
  </dialog>
</template>

<style scoped>
.pos-dialog {
  padding: 0; border: none; border-radius: 16px;
  box-shadow: 0 20px 60px rgba(24,34,48,0.18);
  max-height: 90vh; overflow: hidden; background: white;
}
.pos-dialog::backdrop { background: rgba(10,20,30,0.45); backdrop-filter: blur(2px); }
.pos-dialog--sm { width: min(400px, 94vw); }
.pos-dialog--md { width: min(560px, 94vw); }
.pos-dialog--lg { width: min(780px, 94vw); }
.pos-dialog__panel { display: flex; flex-direction: column; max-height: 90vh; }
.pos-dialog__header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px 16px; border-bottom: 1px solid #eef0f4; flex-shrink: 0;
}
.pos-dialog__title { font-size: 16px; font-weight: 700; margin: 0; letter-spacing: -0.4px; }
.pos-dialog__close {
  border: 0; background: #f0f3f7; color: #5d6b78;
  width: 30px; height: 30px; border-radius: 8px;
  font-size: 12px; cursor: pointer; display: grid; place-items: center;
}
.pos-dialog__close:hover { background: #e3e8ed; }
.pos-dialog__body { padding: 20px 24px; overflow-y: auto; flex: 1; }
.pos-dialog__footer {
  padding: 14px 24px 20px; border-top: 1px solid #eef0f4;
  display: flex; justify-content: flex-end; gap: 10px; flex-shrink: 0;
}
</style>

