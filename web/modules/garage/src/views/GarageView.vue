<template>
  <main class="garage">
    <header class="hero"><div><span class="eyebrow">WORKSHOP MANAGEMENT</span><h1>Garage service desk</h1><p>Coordinate vehicle intake, workshop jobs, technician capacity and planned maintenance.</p></div><div class="today"><b>{{ today }}</b><span>Workshop schedule</span></div></header>
    <section class="metrics"><article><span>Active job cards</span><strong>{{ jobs.filter(j=>!['Collected'].includes(j.status)).length }}</strong><small>Vehicles currently in the workshop pipeline</small></article><article><span>Awaiting parts</span><strong>{{ jobs.filter(j=>j.status==='Awaiting parts').length }}</strong><small>Parts procurement follow-up</small></article><article><span>Ready for collection</span><strong>{{ jobs.filter(j=>j.status==='Ready').length }}</strong><small>Customer handover pending</small></article><article><span>Workshop value</span><strong>KES {{ money(jobs.filter(j=>j.status!=='Collected').reduce((s,j)=>s+Number(j.parts||0)+Number(j.labor||0),0)) }}</strong><small>Parts and labor on active jobs</small></article></section>
    <nav class="tabs"><button v-for="tab in tabs" :key="tab.id" :class="{active:active===tab.id}" @click="active=tab.id"><span>{{ tab.icon }}</span>{{ tab.label }}<b>{{ tab.count }}</b></button></nav>
    <section class="content">
      <JobCardManager v-if="active==='jobs'" :jobs="jobs" :prefill="jobPrefill" @create="createJob" @update="updateJob" />
      <VehicleRegisterManager v-else-if="active==='vehicles'" :vehicles="vehicles" @create="addVehicle" @new-job="openJobForVehicle" />
      <WorkshopAppointmentManager v-else-if="active==='appointments'" :appointments="appointments" @create="addAppointment" />
      <ServiceReminderManager v-else :reminders="reminders" @create="addReminder" />
    </section>
    <footer>These workstation records save in this browser. Production multi-user operation needs tenant-scoped garage APIs, audit records, parts inventory linkage, and configured customer messaging.</footer>
  </main>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import JobCardManager from '../components/JobCardManager.vue'
import VehicleRegisterManager from '../components/VehicleRegisterManager.vue'
import WorkshopAppointmentManager from '../components/WorkshopAppointmentManager.vue'
import ServiceReminderManager from '../components/ServiceReminderManager.vue'

function read(key){try{return JSON.parse(localStorage.getItem(`pos.garage.${key}`)||'[]')}catch{return []}}
const jobs=ref(read('jobs')),vehicles=ref(read('vehicles')),appointments=ref(read('appointments')),reminders=ref(read('reminders'))
for(const [key,value] of Object.entries({jobs,vehicles,appointments,reminders}))watch(value,v=>localStorage.setItem(`pos.garage.${key}`,JSON.stringify(v)),{deep:true})
const active=ref('jobs'),jobPrefill=ref(null),today=new Intl.DateTimeFormat(undefined,{dateStyle:'full'}).format(new Date())
const tabs=computed(()=>[{id:'jobs',label:'Job cards',icon:'▤',count:jobs.value.length},{id:'vehicles',label:'Vehicle register',icon:'▰',count:vehicles.value.length},{id:'appointments',label:'Appointments',icon:'◷',count:appointments.value.filter(a=>['Booked','Arrived','In service'].includes(a.status)).length},{id:'reminders',label:'Service reminders',icon:'↻',count:reminders.value.filter(r=>r.status==='Due').length}])
function money(value){return Number(value||0).toLocaleString(undefined,{maximumFractionDigits:0})}
function createJob(data){const job={...data,id:crypto.randomUUID(),number:`JOB-${String(jobs.value.length+1).padStart(4,'0')}`,date:new Date().toLocaleDateString(),parts:0,labor:0,status:'Open'};jobs.value.unshift(job);if(!vehicles.value.some(v=>v.registration.toLowerCase()===job.registration.toLowerCase()))vehicles.value.unshift({id:crypto.randomUUID(),registration:job.registration,vehicle:job.vehicle,owner:job.customer,phone:job.phone,odometer:job.odometer,vin:''});jobPrefill.value=null}
function updateJob({id,field,value}){const job=jobs.value.find(j=>j.id===id);if(job)job[field]=value}
function addVehicle(vehicle){if(vehicles.value.some(v=>v.registration.toLowerCase()===vehicle.registration.toLowerCase()))return;vehicles.value.unshift({id:crypto.randomUUID(),...vehicle})}
function openJobForVehicle(vehicle){jobPrefill.value={customer:vehicle.owner,phone:vehicle.phone,registration:vehicle.registration,vehicle:vehicle.vehicle,odometer:vehicle.odometer};active.value='jobs'}
function addAppointment(item){appointments.value.unshift(item)}
function addReminder(item){reminders.value.unshift(item)}
</script>
<style scoped>
.garage{display:grid;gap:18px;padding:8px 0;color:#202531}.hero{display:flex;justify-content:space-between;align-items:center;padding:25px 28px;border-radius:14px;background:linear-gradient(120deg,#1c283d,#344967);color:#fff}.eyebrow{font-size:10px;font-weight:800;letter-spacing:1.5px;color:#b9c9e5}.hero h1{margin:7px 0;font-size:27px}.hero p{margin:0;color:#d4dceb;font-size:13px}.today{display:grid;gap:4px;padding-left:22px;border-left:1px solid #ffffff40;text-align:right}.today b{font-size:13px}.today span{color:#cbd5e5;font-size:10px}.metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:11px}.metrics article{padding:14px;border:1px solid #e1e5eb;border-radius:10px;background:#fff}.metrics span,.metrics small{display:block;color:#687385;font-size:11px}.metrics strong{display:block;margin:8px 0;font-size:22px;color:#202d43}.metrics small{font-size:10px}.tabs{display:flex;gap:4px;overflow:auto;border-bottom:1px solid #dfe4ea}.tabs button{display:flex;align-items:center;gap:7px;padding:11px 13px;border:0;border-bottom:2px solid transparent;background:none;color:#697589;font-size:12px;font-weight:700;white-space:nowrap;cursor:pointer}.tabs button.active{border-color:#426ba8;color:#31588f}.tabs button span{font-size:15px}.tabs button b{padding:2px 6px;border-radius:12px;background:#edf1f7;color:#526682;font-size:9px}.content{min-height:330px;padding:20px;border:1px solid #e1e5eb;border-radius:12px;background:#fff}footer{color:#8791a0;font-size:10px;line-height:1.5}@media(max-width:760px){.metrics{grid-template-columns:repeat(2,1fr)}.hero{align-items:flex-start;gap:15px;flex-direction:column}.today{padding:0;border:0;text-align:left}}@media(max-width:480px){.metrics{grid-template-columns:1fr}.content{padding:14px}}
</style>
