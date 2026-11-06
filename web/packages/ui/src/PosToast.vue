<!-- Toast/notification system: use the useToast composable to push messages -->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  toasts: { type: Array, default: () => [] }, // [{ id, message, type, duration? }]
})
const emit = defineEmits(['dismiss'])

const iconMap = { success: '✓', error: '✕', warning: '⚠', info: 'ℹ' }
</script>

<template>
  <Teleport to="body">
    <div class="pos-toast-stack" role="region" aria-label="Notifications" aria-live="polite">
      <Transition v-for="toast in toasts" :key="toast.id" name="toast">
        <div class="pos-toast" :class="`pos-toast--${toast.type ?? 'info'}`"
          role="alert" @click="emit('dismiss', toast.id)">
          <span class="pos-toast__icon" aria-hidden="true">{{ iconMap[toast.type] ?? iconMap.info }}</span>
          <span class="pos-toast__message">{{ toast.message }}</span>
          <button class="pos-toast__close" aria-label="Dismiss" @click.stop="emit('dismiss', toast.id)">✕</button>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<style scoped>
.pos-toast-stack {
  position: fixed; bottom: 24px; right: 24px; z-index: 9000;
  display: flex; flex-direction: column; gap: 10px; max-width: 380px;
}
.pos-toast {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 14px 16px; border-radius: 12px; cursor: pointer;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12); font-size: 13px;
  backdrop-filter: blur(4px);
}
.pos-toast--success { background: #edfaf4; border: 1px solid #a7f3d0; color: #0a6640; }
.pos-toast--error   { background: #fff1f0; border: 1px solid #fbb; color: #8b1a1a; }
.pos-toast--warning { background: #fffbea; border: 1px solid #fcd34d; color: #7a4b08; }
.pos-toast--info    { background: #f0f6ff; border: 1px solid #bfdbfe; color: #1e3a6e; }
.pos-toast__icon { font-weight: 700; font-size: 14px; flex-shrink: 0; margin-top: 1px; }
.pos-toast__message { flex: 1; line-height: 1.5; }
.pos-toast__close {
  border: 0; background: transparent; cursor: pointer;
  font-size: 11px; opacity: 0.6; padding: 0 2px; color: inherit;
}
.pos-toast__close:hover { opacity: 1; }
/* transitions */
.toast-enter-active { transition: all 0.25s ease; }
.toast-leave-active { transition: all 0.2s ease; }
.toast-enter-from { opacity: 0; transform: translateY(12px) scale(0.97); }
.toast-leave-to   { opacity: 0; transform: translateX(20px); }
</style>

