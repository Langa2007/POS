<template>
  <main class="membership">
    <header class="hero">
      <div><span class="eyebrow">MEMBERSHIP OPERATIONS</span><h1>Member services workspace</h1><p>Plans, member lifecycle, attendance, freezes and recurring billing.</p></div>
      <div class="today"><b>{{ today }}</b><span>Billing period {{ billingPeriod }}</span></div>
    </header>
    <section class="metrics">
      <article><span>Active members</span><strong>{{ dashboard.active_members }}</strong><small>Current and active accounts</small></article>
      <article><span>Membership plans</span><strong>{{ dashboard.plans }}</strong><small>Configured recurring plans</small></article>
      <article><span>Today's attendance</span><strong>{{ todayAttendance }}</strong><small>Recorded facility entries</small></article>
      <article><span>Pending billing</span><strong>{{ pendingBilling }}</strong><small>Charges awaiting payment</small></article>
    </section>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-if="billingMessage" class="success" role="status">{{ billingMessage }}</p>
    <nav class="tabs">
      <button v-for="tab in tabs" :key="tab.id" :class="{ active: active === tab.id }" @click="active = tab.id"><span>{{ tab.icon }}</span>{{ tab.label }}<b>{{ tab.count }}</b></button>
    </nav>
    <section class="content">
      <MemberManager v-if="active === 'members'" :members="members" :plans="plans" :loading="loading" @create="createRecord('members', $event)" @remove="removeRecord('members', $event)" />
      <PlanManager v-else-if="active === 'plans'" :plans="plans" :loading="loading" @create="createRecord('plans', $event)" @remove="removeRecord('plans', $event)" />
      <AttendanceManager v-else-if="active === 'attendance'" :attendance="attendance" :members="members" :loading="loading" @create="createRecord('attendance', $event)" @remove="removeRecord('attendance', $event)" />
      <BillingManager v-else-if="active === 'billing'" :billing="billing" :running="runningBilling" @run="runBilling" @mark-paid="markPaid" @remove="removeRecord('billing', $event)" />
      <FreezeManager v-else :freezes="freezes" :members="members" :loading="loading" @create="createRecord('freezes', $event)" @remove="removeRecord('freezes', $event)" />
    </section>
  </main>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useApi } from '../../../../apps/pos/src/composables/useApi.js'
import MemberManager from '../components/MemberManager.vue'
import PlanManager from '../components/PlanManager.vue'
import AttendanceManager from '../components/AttendanceManager.vue'
import BillingManager from '../components/BillingManager.vue'
import FreezeManager from '../components/FreezeManager.vue'
const api=useApi(),active=ref('members'),loading=ref(false),runningBilling=ref(false),error=ref(''),billingMessage=ref('')
const today=new Intl.DateTimeFormat(undefined,{dateStyle:'full'}).format(new Date()),billingPeriod=new Date().toISOString().slice(0,7)
const members=ref([]),plans=ref([]),attendance=ref([]),billing=ref([]),freezes=ref([]),dashboard=ref({active_members:0,plans:0})
const tabs=computed(()=>[{id:'members',label:'Members',icon:'♟',count:members.value.length},{id:'plans',label:'Plans',icon:'▤',count:plans.value.length},{id:'attendance',label:'Attendance',icon:'◷',count:attendance.value.length},{id:'billing',label:'Recurring billing',icon:'$',count:billing.value.length},{id:'freezes',label:'Freezes',icon:'❄',count:freezes.value.length}])
const todayAttendance=computed(()=>{const key=new Date().toISOString().slice(0,10);return attendance.value.filter(item=>String(item.entered||'').slice(0,10)==key).length})
const pendingBilling=computed(()=>billing.value.filter(item=>(item.status||'Pending').toLowerCase()!=='paid').length)
onMounted(loadRecords)
async function loadRecords(){loading.value=true;error.value='';try{for(const resource of ['members','plans','attendance','billing','freezes']){const data=await api.get(`/api/v1/membership/${resource}`);({members,plans,attendance,billing,freezes})[resource].value=data[resource]??[]}await refreshDashboard()}catch(err){error.value=`Membership records could not be loaded: ${err.message}`}finally{loading.value=false}}
async function createRecord(resource,item){error.value='';try{const saved=await api.post(`/api/v1/membership/${resource}`,item);const collection=({members,plans,attendance,billing,freezes})[resource];if(collection)collection.value.unshift(saved);await refreshDashboard()}catch(err){error.value=`Could not save ${resource.slice(0,-1)}: ${err.message}`}}
async function removeRecord(resource,id){error.value='';try{await api.delete(`/api/v1/membership/${resource}/${id}`);const collection=({members,plans,attendance,billing,freezes})[resource];if(collection)collection.value=collection.value.filter(item=>item.id!==id);await refreshDashboard()}catch(err){error.value=`Could not remove record: ${err.message}`}}
async function markPaid(bill){try{const updated=await api.patch(`/api/v1/membership/billing/${bill.id}`,{status:'Paid'});Object.assign(bill,updated)}catch(err){error.value=`Could not update billing record: ${err.message}`}}
async function runBilling(){runningBilling.value=true;billingMessage.value='';error.value='';try{const result=await api.post('/api/v1/membership/billing/run',{period:billingPeriod});billingMessage.value=result.duplicate?`Billing for ${billingPeriod} has already been run.`:`Billing run complete: ${result.processed} charge(s) created for ${billingPeriod}.`;await loadRecords()}catch(err){error.value=`Billing run failed: ${err.message}`}finally{runningBilling.value=false}}
async function refreshDashboard(){try{dashboard.value=await api.get('/api/v1/membership/dashboard')}catch{dashboard.value={active_members:members.value.filter(member=>['active','current'].includes(String(member.status||'active').toLowerCase())).length,plans:plans.value.length}}}
</script>
<style scoped>
.membership{display:grid;gap:18px;padding:8px 0}.hero{display:flex;justify-content:space-between;align-items:center;padding:25px 28px;border-radius:14px;background:linear-gradient(120deg,#4c1d95,#7c3aed);color:#fff}.eyebrow{font-size:10px;font-weight:800;letter-spacing:1.5px;color:#ddd6fe}.hero h1{margin:7px 0;font-size:27px}.hero p{margin:0;color:#ede9fe;font-size:13px}.today{display:grid;gap:4px;padding-left:22px;border-left:1px solid #ffffff40;text-align:right}.today b{font-size:13px}.today span{color:#ddd6fe;font-size:10px}.metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:11px}.metrics article{padding:14px;border:1px solid #e9e4f1;border-radius:10px;background:#fff}.metrics span,.metrics small{display:block;color:#64748b;font-size:11px}.metrics strong{display:block;margin:8px 0;font-size:22px;color:#4c1d95}.metrics small{font-size:10px}.tabs{display:flex;gap:4px;overflow:auto;border-bottom:1px solid #e9e4f1}.tabs button{display:flex;align-items:center;gap:7px;padding:11px 13px;border:0;border-bottom:2px solid transparent;background:none;color:#64748b;font-size:12px;font-weight:700;white-space:nowrap;cursor:pointer}.tabs button.active{border-color:#7c3aed;color:#6d28d9}.tabs button span{font-size:15px}.tabs button b{padding:2px 6px;border-radius:12px;background:#ede9fe;color:#6d28d9;font-size:9px}.content{min-height:330px;padding:20px;border:1px solid #e9e4f1;border-radius:12px;background:#fff}.error,.success{margin:0;padding:10px 12px;border-radius:7px;font-size:12px}.error{background:#fff1f2;color:#9f1239}.success{background:#f0fdf4;color:#166534}@media(max-width:760px){.metrics{grid-template-columns:repeat(2,1fr)}.hero{align-items:flex-start;gap:15px;flex-direction:column}.today{padding:0;border:0;text-align:left}}@media(max-width:480px){.metrics{grid-template-columns:1fr}.content{padding:14px}}
</style>
