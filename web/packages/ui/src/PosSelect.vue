<script setup>
import { computed } from 'vue'
defineProps({
  modelValue: { type: String, default: '' },
  label:      { type: String, default: '' },
  options:    { type: Array, default: () => [] }, // [{ value, label }] or [string]
  placeholder:{ type: String, default: '' },
  disabled:   { type: Boolean, default: false },
  id:         { type: String, default: undefined },
})
defineEmits(['update:modelValue'])
</script>

<template>
  <div class="pos-field">
    <label v-if="label" :for="id" class="pos-field__label">{{ label }}</label>
    <select
      :id="id"
      class="pos-field__select"
      :value="modelValue"
      :disabled="disabled"
      @change="$emit('update:modelValue', $event.target.value)"
    >
      <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
      <template v-for="opt in options" :key="typeof opt === 'string' ? opt : opt.value">
        <option :value="typeof opt === 'string' ? opt : opt.value">
          {{ typeof opt === 'string' ? opt : opt.label }}
        </option>
      </template>
    </select>
  </div>
</template>

<style scoped>
.pos-field { display: grid; gap: 6px; }
.pos-field__label { font-size: 11px; font-weight: 650; color: #394852; }
.pos-field__select {
  height: 45px; border: 1px solid #dfe5e9; border-radius: 8px; padding: 0 12px;
  font: 400 13px inherit; background: white; color: var(--color-ink);
  outline: none; cursor: pointer; appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='8' fill='none'%3E%3Cpath d='M1 1l5 5 5-5' stroke='%2384909a' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat; background-position: right 12px center;
  padding-right: 36px;
  transition: border-color 0.15s;
}
.pos-field__select:focus { border-color: #31a88a; box-shadow: 0 0 0 3px #dff3ed; }
.pos-field__select:disabled { background-color: #f5f7f9; cursor: not-allowed; opacity: 0.7; }
</style>

