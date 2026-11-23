<template>
  <section class="manager">
    <header><div><h3>Recurring Billing</h3><p>Generate and review periodic member charges.</p></div><button class="primary" :disabled="running" @click="$emit('run')">{{ running?'Running…':'Run billing' }}</button></header>
    <div v-if="running" class="empty">Generating billing records…</div>
    <div v-else class="billing-grid">
      <article v-for="bill in billing" :key="bill.id"><div><span>{{ bill.period }}</span><small>{{ bill.status || 'Pending' }}</small></div><h4>{{ bill.member }}</h4><strong>KES {{ Number(bill.amount||0).toLocaleString() }}</strong><footer><button :disabled="bill.status==='Paid'" @click="$emit('mark-paid',bill)">{{ bill.status==='Paid'?'Paid':'Mark paid' }}</button><button class="remove" @click="$emit('remove',bill.id)">×</button></footer></article>
      <div v-if="!billing.length" class="empty">No billing records. Run a billing cycle to create charges.</div>
    </div>
  </section>
</template>
<script setup>
defineProps({ billing:{type:Array,default:()=>[]},running:Boolean })
defineEmits(['run','mark-paid','remove'])
</script>
<style scoped>
.manager{display:grid;gap:14px}header{display:flex;justify-content:space-between;align-items:center}h3{margin:0;font-size:17px}header p{margin:5px 0 0;color:#64748b;font-size:12px}.primary{padding:9px 12px;border:0;border-radius:7px;background:#7c3aed;color:#fff;font-weight:700;font-size:11px;cursor:pointer}.primary:disabled{opacity:.6}.billing-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(210px,1fr));gap:10px}.billing-grid article{display:grid;gap:8px;padding:14px;border:1px solid #e9e4f1;border-radius:9px}.billing-grid article>div{display:flex;justify-content:space-between}.billing-grid span{color:#7c3aed;font-size:10px;font-weight:800}.billing-grid small{padding:3px 6px;border-radius:9px;background:#fef3c7;color:#92400e;font-size:9px}.billing-grid h4{margin:2px 0;font-size:13px}.billing-grid strong{color:#4c1d95;font-size:18px}.billing-grid footer{display:flex;justify-content:flex-end;gap:6px}.billing-grid footer button{padding:6px 8px;border:0;border-radius:6px;background:#ede9fe;color:#6d28d9;font-size:10px;font-weight:700;cursor:pointer}.billing-grid footer .remove{background:none;color:#94a3b8}.empty{grid-column:1/-1;padding:24px;border:1px dashed #ddd6fe;border-radius:8px;color:#64748b;text-align:center;font-size:12px}
</style>
