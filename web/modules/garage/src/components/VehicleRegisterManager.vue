<template>
  <section class="manager"><header><div><h3>Vehicle register</h3><p>Vehicle identity, ownership contact and last known workshop mileage.</p></div><button class="primary" @click="open = !open">{{ open ? 'Cancel' : 'Add vehicle' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit"><label>Registration<input v-model.trim="form.registration" required /></label><label>Make / model<input v-model.trim="form.vehicle" required /></label><label>VIN / chassis number<input v-model.trim="form.vin" /></label><label>Owner<input v-model.trim="form.owner" required /></label><label>Phone<input v-model.trim="form.phone" /></label><label>Current mileage (km)<input v-model.number="form.odometer" min="0" type="number" /></label><button class="primary">Save vehicle</button></form>
    <div class="cards"><article v-for="v in vehicles" :key="v.id"><div class="car">▰</div><div class="info"><strong>{{ v.registration }} · {{ v.vehicle }}</strong><span>{{ v.owner }} <template v-if="v.phone">· {{ v.phone }}</template></span><small>VIN {{ v.vin || 'Not recorded' }} · {{ Number(v.odometer).toLocaleString() }} km</small></div><button class="open-job" @click="$emit('new-job',v)">Open job</button></article><div v-if="!vehicles.length" class="empty">Vehicles registered on job intake are collected here.</div></div>
  </section>
</template>
<script setup>
import { reactive, ref } from 'vue'
defineProps({ vehicles: { type: Array, default: () => [] } })
const emit=defineEmits(['create','new-job']); const open=ref(false)
const form=reactive({registration:'',vehicle:'',vin:'',owner:'',phone:'',odometer:0})
function submit(){emit('create',{...form});Object.assign(form,{registration:'',vehicle:'',vin:'',owner:'',phone:'',odometer:0});open.value=false}
</script>
<style scoped>
.manager{display:grid;gap:14px}header{display:flex;align-items:center;justify-content:space-between}h3{margin:0;font-size:18px}header p{margin:5px 0 0;color:#6c7482;font-size:12px}.primary,.open-job{padding:9px 12px;border:0;border-radius:7px;background:#355fa0;color:white;font-weight:700;font-size:11px;cursor:pointer}.form{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;padding:15px;background:#f4f6f9;border-radius:9px}.form label{display:grid;gap:5px;color:#586376;font-size:11px;font-weight:700}.form input{min-width:0;padding:9px;border:1px solid #d5dbe4;border-radius:6px}.cards{display:grid;gap:9px}.cards article{display:flex;align-items:center;gap:12px;padding:12px;border:1px solid #e1e5eb;border-radius:9px}.car{display:grid;place-items:center;width:38px;height:38px;border-radius:9px;background:#eaf0fb;color:#355fa0;font-size:24px}.info{display:grid;gap:4px;flex:1}.info strong{font-size:12px}.info span,.info small{color:#6c7482;font-size:10px}.open-job{background:#eef2f8;color:#35517f}.empty{text-align:center;padding:24px;color:#8791a0}@media(max-width:600px){.form{grid-template-columns:1fr}.cards article{flex-wrap:wrap}}
</style>
