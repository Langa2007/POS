<template>
  <section class="manager">
    <header><div><h3>Membership Plans</h3><p>Define billing cycles and plan pricing.</p></div><button class="primary" @click="open=!open">{{ open?'Cancel':'Add plan' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit">
      <label>Plan name<input v-model.trim="form.name" required /></label>
      <label>Price (KES)<input v-model.number="form.price" type="number" min="0" step="1" required /></label>
      <label>Billing period<select v-model="form.period"><option>monthly</option><option>quarterly</option><option>annual</option></select></label>
      <button class="primary">Save plan</button>
    </form>
    <div v-if="loading" class="empty">Loading plans…</div>
    <div v-else class="grid">
      <article v-for="plan in plans" :key="plan.id"><div class="plan-head"><span>{{ plan.period }}</span><button title="Remove plan" @click="$emit('remove',plan.id)">×</button></div><h4>{{ plan.name }}</h4><strong>KES {{ Number(plan.price||0).toLocaleString() }}</strong><small>Per {{ plan.period || 'month' }}</small></article>
      <div v-if="!plans.length" class="empty">No membership plans have been created.</div>
    </div>
  </section>
</template>
<script setup>
import { reactive, ref } from 'vue'
defineProps({ plans:{type:Array,default:()=>[]},loading:Boolean })
const emit=defineEmits(['create','remove']),open=ref(false),form=reactive({name:'',price:0,period:'monthly'})
function submit(){emit('create',{...form});Object.assign(form,{name:'',price:0,period:'monthly'});open.value=false}
</script>
<style scoped>
.manager{display:grid;gap:14px}header{display:flex;justify-content:space-between;align-items:center}h3{margin:0;font-size:17px}header p{margin:5px 0 0;color:#64748b;font-size:12px}.primary{padding:9px 12px;border:0;border-radius:7px;background:#7c3aed;color:#fff;font-weight:700;font-size:11px;cursor:pointer}.form{display:grid;grid-template-columns:2fr 1fr 1fr auto;gap:10px;padding:14px;border-radius:9px;background:#faf5ff}.form label{display:grid;gap:5px;color:#64748b;font-size:11px;font-weight:700}.form input,.form select{min-width:0;padding:9px;border:1px solid #ddd6fe;border-radius:6px}.form .primary{align-self:end}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(190px,1fr));gap:10px}.grid article{display:grid;gap:7px;padding:16px;border:1px solid #e9e4f1;border-radius:10px}.plan-head{display:flex;justify-content:space-between}.plan-head span{text-transform:uppercase;color:#7c3aed;font-size:9px;font-weight:800}.plan-head button{border:0;background:none;color:#94a3b8;cursor:pointer}.grid h4{margin:2px 0;color:#1e293b;font-size:14px}.grid strong{color:#6d28d9;font-size:20px}.grid small{color:#64748b}.empty{grid-column:1/-1;padding:24px;border:1px dashed #ddd6fe;border-radius:8px;color:#64748b;text-align:center;font-size:12px}@media(max-width:700px){.form{grid-template-columns:1fr}}
</style>
