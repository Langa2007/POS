<template>
  <main class="station">
    <header class="hero"><div><span class="eyebrow">FUEL OPERATIONS</span><h1>Station control room</h1><p>Manage pump closeouts, tank inventory, supplier receipts and equipment configuration.</p></div><div class="day"><b>{{ date }}</b><span>Local station time</span></div></header>
    <section class="metrics"><article><span>Registered nozzles</span><strong>{{ pumps.length }}</strong><small>Station equipment register</small></article><article><span>Meter volume reconciled</span><strong>{{ fmt(reconciledVolume) }} L</strong><small>{{ shifts.filter(s=>s.ok).length }} balanced closeouts</small></article><article><span>Variance requiring review</span><strong>{{ shifts.filter(s=>!s.ok).length }}</strong><small>Out-of-tolerance shift closes</small></article><article><span>Latest tank stock</span><strong>{{ fmt(latestStock) }} L</strong><small>Sum of latest readings by tank</small></article></section>
    <div class="notice"><span class="signal"></span><div><strong>Manual operations mode</strong><p>Records save to the authenticated station account. No live pump controller is paired with this station yet.</p></div><span class="mode">DEVICE LINK OFFLINE</span></div>
    <p v-if="error" class="error" role="alert">{{ error }}</p><p v-if="loading" class="loading">Loading station records…</p>
    <nav class="tabs"><button v-for="tab in tabs" :key="tab.id" :class="{active: active===tab.id}" @click="active=tab.id"><span>{{ tab.icon }}</span>{{ tab.label }}<b v-if="tab.id==='closeouts' && shifts.some(s=>!s.ok)">{{ shifts.filter(s=>!s.ok).length }}</b></button></nav>
    <section class="content">
      <MeterShiftManager v-if="active==='closeouts'" :shifts="shifts" @record="add('shifts',$event)" />
      <TankDipManager v-else-if="active==='tanks'" :dips="dips" @record="add('dips',$event)" />
      <FuelDeliveryManager v-else-if="active==='deliveries'" :deliveries="deliveries" @record="add('deliveries',$event)" />
      <PumpRegisterManager v-else :pumps="pumps" @add="add('pumps',$event)" @remove="removePump" />
    </section>
    <footer>Station records are tenant-scoped and permission-checked by the POS API. Live pump readings still require a supported controller and device integration.</footer>
  </main>
</template>
<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
import MeterShiftManager from '../components/MeterShiftManager.vue'
import TankDipManager from '../components/TankDipManager.vue'
import FuelDeliveryManager from '../components/FuelDeliveryManager.vue'
import PumpRegisterManager from '../components/PumpRegisterManager.vue'

const tabs=[{id:'closeouts',label:'Meter closeouts',icon:'↔'},{id:'tanks',label:'Tank inventory & dips',icon:'▤'},{id:'deliveries',label:'Supplier receipts',icon:'↓'},{id:'equipment',label:'Pump register',icon:'⛽'}]
const active=ref('closeouts'),date=new Intl.DateTimeFormat(undefined,{dateStyle:'full'}).format(new Date())
const api=useApi(),loading=ref(false),error=ref('')
function read(key){try{return JSON.parse(localStorage.getItem(`pos.fuelstation.${key}`)||'[]')}catch{return []}}
const shifts=ref(read('shifts')),dips=ref(read('dips')),deliveries=ref(read('deliveries')),pumps=ref(read('pumps'))
for(const [key,value] of Object.entries({shifts,dips,deliveries,pumps}))watch(value,v=>localStorage.setItem(`pos.fuelstation.${key}`,JSON.stringify(v)),{deep:true})
onMounted(loadRecords)
async function loadRecords(){loading.value=true;error.value='';try{for(const key of ['pumps','shifts','dips','deliveries']){const result=await api.get(`/api/v1/fuelstation/${key}`);({pumps,shifts,dips,deliveries}[key]).value=result[key]??[]}}catch(err){error.value=`Station records could not be loaded from the server: ${err.message}`}finally{loading.value=false}}
const reconciledVolume=computed(()=>shifts.value.filter(s=>s.ok).reduce((total,s)=>total+Number(s.volume||0),0))
const latestStock=computed(()=>{const seen=new Set();let total=0;for(const dip of dips.value){if(!seen.has(dip.tank)){seen.add(dip.tank);total+=Number(dip.volume||0)}}return total})
function fmt(value){return Number(value||0).toLocaleString(undefined,{maximumFractionDigits:1})}
async function add(key,item){error.value='';try{const saved=await api.post(`/api/v1/fuelstation/${key}`,item);({shifts,dips,deliveries,pumps}[key]).value.unshift(saved)}catch(err){error.value=`Could not save ${key.slice(0,-1)}: ${err.message}`}}
async function removePump(id){error.value='';try{await api.delete(`/api/v1/fuelstation/pumps/${id}`);pumps.value=pumps.value.filter(p=>p.id!==id)}catch(err){error.value=`Could not remove pump: ${err.message}`}}
</script>
<style scoped>
.station{display:grid;gap:18px;padding:8px 0;color:#17211d}.hero{display:flex;justify-content:space-between;align-items:center;padding:25px 28px;border-radius:14px;background:linear-gradient(120deg,#123c30,#176c4d);color:#fff}.eyebrow{font-size:10px;font-weight:800;letter-spacing:1.5px;color:#9ce0bc}.hero h1{margin:7px 0;font-size:27px}.hero p{margin:0;color:#d0e5da;font-size:13px}.day{display:grid;gap:4px;padding-left:22px;border-left:1px solid #ffffff40;text-align:right}.day b{font-size:13px}.day span{color:#c1dfce;font-size:10px}.metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:11px}.metrics article{padding:14px;border:1px solid #e2e8e0;border-radius:10px;background:#fff}.metrics span,.metrics small{display:block;color:#69766f;font-size:11px}.metrics strong{display:block;margin:8px 0;font-size:22px;color:#173d2d}.metrics small{font-size:10px}.notice{display:flex;align-items:center;gap:11px;padding:12px 15px;border:1px solid #f1dcaa;border-radius:9px;background:#fff9ea}.notice p{margin:3px 0 0;color:#766b53;font-size:11px}.notice strong{color:#655022;font-size:12px}.signal{width:9px;height:9px;border-radius:50%;background:#d49425}.mode{margin-left:auto;color:#987124;font-size:9px;font-weight:800;white-space:nowrap}.tabs{display:flex;gap:4px;overflow:auto;border-bottom:1px solid #dce4de}.tabs button{display:flex;align-items:center;gap:7px;padding:11px 13px;border:0;border-bottom:2px solid transparent;background:none;color:#66736c;font-size:12px;font-weight:700;white-space:nowrap;cursor:pointer}.tabs button.active{border-color:#1b8c5c;color:#14734d}.tabs button span{font-size:15px}.tabs button b{padding:2px 6px;border-radius:12px;background:#fff0e4;color:#9a4e14;font-size:9px}.content{min-height:330px;padding:20px;border:1px solid #e2e8e0;border-radius:12px;background:#fff}.error{margin:0;padding:10px 12px;border-radius:7px;background:#fff0e4;color:#9a3412;font-size:12px}.loading{margin:0;color:#68766d;font-size:12px}footer{color:#839087;font-size:10px;line-height:1.5}@media(max-width:760px){.metrics{grid-template-columns:repeat(2,1fr)}.hero{align-items:flex-start;gap:15px;flex-direction:column}.day{padding:0;border:0;text-align:left}}@media(max-width:480px){.metrics{grid-template-columns:1fr}.content{padding:14px}.mode{display:none}}
</style>
