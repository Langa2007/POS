<script setup>
defineProps({
  modelValue: { type: [String, Number], default: '' },
  label:      { type: String, default: '' },
  type:       { type: String, default: 'text' },
  placeholder:{ type: String, default: '' },
  autocomplete:{ type: String, default: 'off' },
  required:   { type: Boolean, default: false },
  disabled:   { type: Boolean, default: false },
  error:      { type: String, default: '' },
  hint:       { type: String, default: '' },
  id:         { type: String, default: undefined },
})
defineEmits(['update:modelValue'])
</script>

<template>
  <div class="pos-field">
    <label v-if="label" :for="id" class="pos-field__label">{{ label }}</label>
    <input
      :id="id"
      class="pos-field__input"
      :class="{ 'pos-field__input--error': error }"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :required="required"
      :disabled="disabled"
      :aria-describedby="error ? `${id}-error` : hint ? `${id}-hint` : undefined"
      :aria-invalid="error ? 'true' : undefined"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <p v-if="error" :id="`${id}-error`" class="pos-field__error" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="`${id}-hint`" class="pos-field__hint">{{ hint }}</p>
  </div>
</template>

<style scoped>
.pos-field { display: grid; gap: 6px; }
.pos-field__label { font-size: 11px; font-weight: 650; color: #394852; }
.pos-field__input {
  height: 45px; border: 1px solid #dfe5e9; border-radius: 8px; padding: 0 12px;
  font: 400 13px inherit; outline: none; background: white; color: var(--color-ink);
  transition: border-color 0.15s, box-shadow 0.15s;
}
.pos-field__input:focus { border-color: #31a88a; box-shadow: 0 0 0 3px #dff3ed; }
.pos-field__input::placeholder { color: #aab3b9; }
.pos-field__input:disabled { background: #f5f7f9; cursor: not-allowed; opacity: 0.7; }
.pos-field__input--error { border-color: #d32f2f; }
.pos-field__input--error:focus { box-shadow: 0 0 0 3px #fde8e8; }
.pos-field__error { font-size: 11px; color: #b42318; margin: 0; }
.pos-field__hint  { font-size: 11px; color: #84909a; margin: 0; }
</style>

