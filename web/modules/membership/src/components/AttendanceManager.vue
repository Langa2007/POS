<template>
  <section class="manager">
    <header><div><h3>Attendance</h3><p>Log member visits and facility access.</p></div><button class="primary" @click="open=!open">{{ open?'Cancel':'Record attendance' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit">
      <label>Member<select v-model="form.member" required><option value="" disabled>Select member</option><option v-for="member in members" :key="member.id" :value="member.name">{{ member.name }}</option></select></label>
      <label>Location<input v-model.trim="form.location" required /></label>
      <label>Time<input v-model="form.entered" type="datetime-local" required /></label>
      <button class="primary">Record entry</button>
    </form>
    <div v-if="loading" class="empty">Loading attendance…</div>
    <table v-else><thead><tr><th>Member</th><th>Location</th><th>Time</th><th></th></tr></thead><tbody><tr v-for="item in attendance" :key="item.id"><td>{{ item.member }}</td><td>{{ item.location }}</td><td>{{ formatTime(item.entered) }}</td><td><button title="Remove entry" @click="$emit('remove',item.id)">×</button></td></tr><tr v-if="!attendance.length"><td colspan="4" class="empty">No attendance has been recorded.</td></tr></tbody></table>
  </section>
</template>
<script setup>
import { reactive, ref } from 'vue'
defineProps({ attendance:{type:Array,default:()=>[]},members:{type:Array,default:()=>[]},loading:Boolean })
const emit=defineEmits(['create','remove']),open=ref(false)
const form=reactive({member:'',location:'Main facility',entered:new Date().toISOString().slice(0,16)})
function submit(){emit('create',{...form,entered:new Date(form.entered).toISOString()});Object.assign(form,{member:'',location:'Main facility',entered:new Date().toISOString().slice(0,16)});open.value=false}
function formatTime(value){const date=new Date(value);return Number.isNaN(date.getTime())?value:date.toLocaleString()}
</script>
<style scoped>
.manager{display:grid;gap:14px}header{display:flex;justify-content:space-between;align-items:center}h3{margin:0;font-size:17px}header p{margin:5px 0 0;color:#64748b;font-size:12px}.primary{padding:9px 12px;border:0;border-radius:7px;background:#7c3aed;color:#fff;font-weight:700;font-size:11px;cursor:pointer}.form{display:grid;grid-template-columns:2fr 1fr 1fr auto;gap:10px;padding:14px;border-radius:9px;background:#faf5ff}.form label{display:grid;gap:5px;color:#64748b;font-size:11px;font-weight:700}.form input,.form select{min-width:0;padding:9px;border:1px solid #ddd6fe;border-radius:6px}.form .primary{align-self:end}table{width:100%;border-collapse:collapse}th,td{padding:11px;border-bottom:1px solid #ede9fe;text-align:left;font-size:12px}th{color:#64748b}td button{border:0;background:none;color:#94a3b8;cursor:pointer}.empty{padding:24px;text-align:center;color:#64748b}@media(max-width:700px){.form{grid-template-columns:1fr}}
</style>
