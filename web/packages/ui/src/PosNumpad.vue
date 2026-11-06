<!-- Cash register numpad: emits digit/backspace/enter events -->
<script setup>
const emit = defineEmits(['digit', 'backspace', 'enter'])

const keys = [
  ['7', '8', '9'],
  ['4', '5', '6'],
  ['1', '2', '3'],
  ['00', '0', '.'],
]
</script>

<template>
  <div class="pos-numpad" role="group" aria-label="Number pad">
    <div v-for="(row, r) in keys" :key="r" class="pos-numpad__row">
      <button
        v-for="key in row" :key="key"
        class="pos-numpad__key"
        type="button"
        :aria-label="key"
        @click="emit('digit', key)"
      >{{ key }}</button>
    </div>
    <div class="pos-numpad__row pos-numpad__row--actions">
      <button class="pos-numpad__key pos-numpad__key--wide" type="button" aria-label="Backspace" @click="emit('backspace')">⌫</button>
      <button class="pos-numpad__key pos-numpad__key--enter" type="button" aria-label="Confirm" @click="emit('enter')">✓</button>
    </div>
  </div>
</template>

<style scoped>
.pos-numpad { display: grid; gap: 8px; }
.pos-numpad__row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.pos-numpad__row--actions { grid-template-columns: 2fr 1fr; }
.pos-numpad__key {
  min-height: 54px; border: 1px solid #e0e6ea; border-radius: 10px;
  font: 700 18px inherit; background: white; color: var(--color-ink);
  cursor: pointer; transition: background 0.1s, transform 0.08s;
  display: grid; place-items: center;
}
.pos-numpad__key:hover  { background: #f3fbf7; border-color: #35aa8d; }
.pos-numpad__key:active { transform: scale(0.95); background: #e5f7f0; }
.pos-numpad__key--enter {
  background: var(--color-brand); color: white; border-color: var(--color-brand); font-size: 20px;
}
.pos-numpad__key--enter:hover { background: #0f5e49; }
.pos-numpad__key:focus-visible { outline: 3px solid var(--color-focus); outline-offset: 2px; }
</style>

