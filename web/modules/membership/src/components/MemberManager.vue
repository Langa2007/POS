<template>
  <section class="manager">
    <header><div><h3>Member Directory</h3><p>Enroll, search and maintain member records.</p></div><button class="primary" @click="open = !open">{{ open ? 'Cancel' : 'Add member' }}</button></header>
    <form v-if="open" class="form" @submit.prevent="submit">
      <label>Full name<input v-model.trim="form.name" required /></label>
      <label>Email<input v-model.trim="form.email" type="email" required /></label>
      <label>Phone<input v-model.trim="form.phone" required /></label>
      <label>Plan<select v-model="form.plan" required><option value="" disabled>Select a plan</option><option v-for="plan in plans" :key="plan.id" :value="plan.name">{{ plan.name }}</option><option value="Standard">Standard</option></select></label>
      <label>Status<select v-model="form.status"><option>Active</option><option>Paused</option><option>Cancelled</option></select></label>
      <button class="primary">Enroll member</button>
    </form>
    <div class="lookup"><input v-model.trim="query" placeholder="Search name, email or phone" @input="filter" /></div>
    <div v-if="loading" class="empty">Loading members…</div>
    <div v-else class="cards">
      <article v-for="member in visible" :key="member.id">
        <div class="avatar">{{ initials(member.name) }}</div>
        <div class="info"><strong>{{ member.name }}</strong><span>{{ member.email || 'No email' }} · {{ member.phone || 'No phone' }}</span><small>{{ member.plan || 'No plan' }}</small></div>
        <span class="status" :class="member.status?.toLowerCase()">{{ member.status || 'Active' }}</span>
        <button class="icon-button" title="Remove member" @click="$emit('remove', member.id)">×</button>
      </article>
      <div v-if="!visible.length" class="empty">No members match your search.</div>
    </div>
  </section>
</template>
<script setup>
import { computed, reactive, ref } from 'vue'
const props = defineProps({ members: { type: Array, default: () => [] }, plans: { type: Array, default: () => [] }, loading: Boolean })
const emit = defineEmits(['create', 'remove'])
const open = ref(false), query = ref('')
const form = reactive({ name: '', email: '', phone: '', plan: '', status: 'Active' })
const visible = computed(() => {
  const needle = query.value.toLowerCase()
  if (!needle) return props.members
  return props.members.filter((member) => [member.name, member.email, member.phone, member.plan].some((value) => String(value || '').toLowerCase().includes(needle)))
})
function initials(name) { return String(name || '?').split(/\s+/).map((part) => part[0]).join('').slice(0, 2).toUpperCase() }
function submit() { emit('create', { ...form }); Object.assign(form, { name: '', email: '', phone: '', plan: '', status: 'Active' }); open.value = false }
function filter() {}
</script>
<style scoped>
.manager{display:grid;gap:14px}header{display:flex;justify-content:space-between;align-items:center}h3{margin:0;font-size:17px}header p{margin:5px 0 0;color:#64748b;font-size:12px}.primary{padding:9px 12px;border:0;border-radius:7px;background:#7c3aed;color:#fff;font-weight:700;font-size:11px;cursor:pointer}.form{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;padding:14px;border-radius:9px;background:#faf5ff}.form label{display:grid;gap:5px;color:#64748b;font-size:11px;font-weight:700}.form input,.form select,.lookup input{min-width:0;padding:9px;border:1px solid #ddd6fe;border-radius:6px}.form .primary{align-self:end}.lookup input{width:100%;box-sizing:border-box}.cards{display:grid;gap:8px}.cards article{display:flex;align-items:center;gap:11px;padding:12px;border:1px solid #e9e4f1;border-radius:9px}.avatar{display:grid;place-items:center;width:38px;height:38px;border-radius:50%;background:#ede9fe;color:#6d28d9;font-weight:800}.info{display:grid;gap:3px;flex:1}.info strong{font-size:12px}.info span,.info small{color:#64748b;font-size:10px}.status{padding:4px 7px;border-radius:10px;background:#dcfce7;color:#166534;font-size:9px}.status.paused{background:#fef3c7;color:#92400e}.status.cancelled{background:#fee2e2;color:#991b1b}.icon-button{border:0;background:none;color:#94a3b8;font-size:18px;cursor:pointer}.empty{padding:24px;border:1px dashed #ddd6fe;border-radius:8px;color:#64748b;text-align:center;font-size:12px}@media(max-width:700px){.form{grid-template-columns:1fr}}
</style>
