<template>
  <section class="manager">
    <header><div><h3>Membership Freezes</h3><p>Pause memberships temporarily without cancellation.</p></div><button class="primary" @click="open=!open">{{ open?'Cancel':'Freeze member' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit">
      <label>Member<select v-model="form.member" required><option value="" disabled>Select member</option><option v-for="member in members" :key="member.id" :value="member.name">{{ member.name }}</option></select></label>
      <label>Start date<input v-model="form.start" type="date" required /></label>
      <label>End date<input v-model="form.end" type="date" required /></label>
      <label>Reason<input v-model.trim="form.reason" required /></label>
      <button class="primary">Create freeze</button>
    </form>
    <div v-if="loading" class="empty">Loading freezes…</div>
    <div v-else class="cards"><article v-for="freeze in freezes" :key="freeze.id"><div class="snow">❄</div><div class="info"><strong>{{ freeze.member }}</strong><span>{{ freeze.start }} → {{ freeze.end }}</span><small>{{ freeze.reason }}</small></div><button title="Remove freeze" @click="$emit('remove',freeze.id)">×</button></article><div v-if="!freezes.length" class="empty">No active freezes.</div></div>
  </section>
</template>
<script setup>
import { reactive, ref } from 'vue'
defineProps({ freezes:{type:Array,default:()=>[]},members:{type:Array,default:()=>[]},loading:Boolean })
const emit=defineEmits(['create','remove']),open=ref(false)
const form=reactive({member:'',start:new Date().toISOString().slice(0,10),end:new Date().toISOString().slice(0,10),reason:'Temporary pause'})
function submit(){emit('create',{...form});open.value=false}
</script>
<style scoped>
.manager{display:grid;gap:14px}header{display:flex;justify-content:space-between;align-items:center}h3{margin:0;font-size:17px}header p{margin:5px 0 0;color:#64748b;font-size:12px}.primary{padding:9px 12px;border:0;border-radius:7px;background:#7c3aed;color:#fff;font-weight:700;font-size:11px;cursor:pointer}.form{display:grid;grid-template-columns:1.5fr 1fr 1fr 1.5fr auto;gap:10px;padding:14px;border-radius:9px;background:#faf5ff}.form label{display:grid;gap:5px;color:#64748b;font-size:11px;font-weight:700}.form input,.form select{min-width:0;padding:9px;border:1px solid #ddd6fe;border-radius:6px}.form .primary{align-self:end}.cards{display:grid;gap:8px}.cards article{display:flex;align-items:center;gap:11px;padding:12px;border:1px solid #e9e4f1;border-radius:9px}.snow{display:grid;place-items:center;width:38px;height:38px;border-radius:9px;background:#eff6ff;color:#0284c7}.info{display:grid;gap:3px;flex:1}.info strong{font-size:12px}.info span,.info small{color:#64748b;font-size:10px}.cards article>button{border:0;background:none;color:#94a3b8;cursor:pointer}.empty{padding:24px;border:1px dashed #ddd6fe;border-radius:8px;text-align:center;color:#64748b;font-size:12px}@media(max-width:700px){.form{grid-template-columns:1fr}}
</style>
