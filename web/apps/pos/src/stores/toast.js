/**
 * Toast store — global notification queue.
 *
 * Usage:  const toast = useToastStore()
 *         toast.success('Sale completed!')
 *         toast.error('Something went wrong')
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'

let _id = 0

export const useToastStore = defineStore('toast', () => {
  const toasts = ref([])

  function push(message, type = 'info', duration = 4000) {
    const id = ++_id
    toasts.value.push({ id, message, type })
    if (duration > 0) setTimeout(() => dismiss(id), duration)
    return id
  }

  function dismiss(id) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  const success = (msg, dur) => push(msg, 'success', dur)
  const error   = (msg, dur) => push(msg, 'error',   dur ?? 6000)
  const warning = (msg, dur) => push(msg, 'warning', dur)
  const info    = (msg, dur) => push(msg, 'info',    dur)

  return { toasts, push, dismiss, success, error, warning, info }
})

