<template>
  <div class="loyalty-manager">
    <div class="manager-header">
      <div>
        <h3>⭐ Customer Loyalty & Rewards</h3>
        <p class="subtitle">Reward frequent shoppers, redeem points on checkout, and enroll new members</p>
      </div>
      <button class="btn-enroll" @click="showEnrollModal = true">
        ➕ Enroll New Member
      </button>
    </div>

    <!-- Customer Lookup Bar -->
    <div class="lookup-card">
      <label for="loyaltySearch">Search Member by Mobile Phone or Loyalty Card #:</label>
      <div class="search-input-group">
        <input
          id="loyaltySearch"
          type="text"
          v-model="searchQuery"
          placeholder="e.g. 0712345678, 0722001122, or LOYAL-1001"
          class="search-input"
          @keydown.enter.prevent="lookupCustomer"
        />
        <button type="button" class="btn-primary" @click="lookupCustomer">
          🔍 Find Member
        </button>
      </div>

      <!-- Quick sample suggestions -->
      <div class="quick-samples">
        <span>Try sample:</span>
        <button type="button" class="btn-tag" @click="quickSearch('0712345678')">0712345678 (Gold)</button>
        <button type="button" class="btn-tag" @click="quickSearch('0722001122')">0722001122 (Silver)</button>
        <button type="button" class="btn-tag" @click="quickSearch('0799887766')">0799887766 (Bronze)</button>
      </div>
    </div>

    <!-- Member Profile & Redemption Card -->
    <div v-if="activeMember" class="member-profile-card">
      <div class="profile-header">
        <div class="avatar-tier" :class="activeMember.tier.toLowerCase()">
          {{ activeMember.tier[0] }}
        </div>
        <div class="profile-details">
          <h4>{{ activeMember.name }}</h4>
          <div class="profile-meta">
            <span>📞 {{ activeMember.phone }}</span>
            <span>💳 Card: <code>{{ activeMember.cardNo }}</code></span>
            <span class="badge" :class="activeMember.tier.toLowerCase()">{{ activeMember.tier }} Member</span>
          </div>
        </div>
        <div class="points-hero">
          <div class="points-num">{{ activeMember.points.toLocaleString() }}</div>
          <div class="points-sub">Available Points (= KES {{ (activeMember.points / 10).toFixed(2) }} discount)</div>
        </div>
      </div>

      <!-- Point Redemption Control -->
      <div class="redemption-control">
        <div class="redeem-row">
          <div class="redeem-inputs">
            <label for="redeemInput">Points to Redeem:</label>
            <input
              id="redeemInput"
              type="number"
              min="100"
              :max="activeMember.points"
              step="50"
              v-model.number="pointsToRedeem"
              class="redeem-field"
            />
            <span class="equals-label">
              = KES <strong>{{ (pointsToRedeem / 10).toFixed(2) }}</strong> voucher
            </span>
          </div>
          <div class="redeem-actions">
            <button class="btn-secondary" @click="pointsToRedeem = activeMember.points">Max All</button>
            <button class="btn-redeem" :disabled="pointsToRedeem < 100 || pointsToRedeem > activeMember.points" @click="redeemPoints">
              🎁 Redeem on POS Cart
            </button>
          </div>
        </div>
        <small class="rate-note">Rule: 10 Loyalty Points = KES 1.00. Earn 1 Point for every KES 50 spent on retail sales.</small>
      </div>
    </div>

    <!-- Loyalty Member Directory -->
    <div class="members-directory">
      <h4>Recent Active Members:</h4>
      <div class="members-table-wrapper">
        <table class="members-table">
          <thead>
            <tr>
              <th>Member Name</th>
              <th>Phone</th>
              <th>Card No</th>
              <th>Tier</th>
              <th>Points</th>
              <th>Cash Value</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in members" :key="m.id">
              <td><strong>{{ m.name }}</strong></td>
              <td>{{ m.phone }}</td>
              <td><code>{{ m.cardNo }}</code></td>
              <td><span class="tier-pill" :class="m.tier.toLowerCase()">{{ m.tier }}</span></td>
              <td><strong>{{ m.points.toLocaleString() }}</strong> pts</td>
              <td>KES {{ (m.points / 10).toFixed(2) }}</td>
              <td>
                <button class="btn-select" @click="selectMember(m)">Select</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Enroll Modal -->
    <div v-if="showEnrollModal" class="modal-backdrop">
      <div class="modal-dialog">
        <div class="modal-header">
          <h4>Enroll New Loyalty Member</h4>
          <button class="close-btn" @click="showEnrollModal = false">✕</button>
        </div>
        <form @submit.prevent="enrollMember" class="modal-form">
          <label>Full Customer Name:
            <input type="text" v-model="newMember.name" required placeholder="e.g. Grace Njeri" />
          </label>
          <label>Mobile Phone Number:
            <input type="tel" v-model="newMember.phone" required placeholder="07XXXXXXXX" />
          </label>
          <label>Email Address (optional):
            <input type="email" v-model="newMember.email" placeholder="customer@email.com" />
          </label>
          <div class="enroll-perk">
            🎁 Welcome Perk: New member automatically receives <strong>100 bonus points (KES 10)</strong> on registration!
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" @click="showEnrollModal = false">Cancel</button>
            <button type="submit" class="btn-primary">Enroll Member</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['set-customer', 'apply-discount', 'notification'])

const searchQuery = ref('')
const activeMember = ref(null)
const pointsToRedeem = ref(500)
const showEnrollModal = ref(false)

const members = ref([
  { id: 'mem-1', name: 'John Kamau', phone: '0712345678', cardNo: 'LOYAL-1001', tier: 'Gold', points: 3450 },
  { id: 'mem-2', name: 'Mary Wanjiku', phone: '0722001122', cardNo: 'LOYAL-1002', tier: 'Silver', points: 1200 },
  { id: 'mem-3', name: 'Peter Omondi', phone: '0799887766', cardNo: 'LOYAL-1003', tier: 'Bronze', points: 450 },
  { id: 'mem-4', name: 'Amina Mohamed', phone: '0733445566', cardNo: 'LOYAL-1004', tier: 'Gold', points: 4800 },
])

const newMember = ref({ name: '', phone: '', email: '' })

// Auto-select first member
activeMember.value = members.value[0]

function lookupCustomer() {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return
  const match = members.value.find((m) =>
    m.phone.includes(q) || m.cardNo.toLowerCase().includes(q) || m.name.toLowerCase().includes(q)
  )
  if (match) {
    selectMember(match)
    emit('notification', { message: `Loaded loyalty profile for ${match.name} (${match.points} pts)`, type: 'info' })
  } else {
    emit('notification', { message: `No member found matching "${searchQuery.value}". Click Enroll to add.`, type: 'warning' })
  }
}

function quickSearch(phone) {
  searchQuery.value = phone
  lookupCustomer()
}

function selectMember(m) {
  activeMember.value = m
  pointsToRedeem.value = Math.min(500, m.points)
  emit('set-customer', {
    id: m.id,
    name: m.name,
    phone: m.phone,
    loyaltyCard: m.cardNo,
    points: m.points,
  })
}

function redeemPoints() {
  if (!activeMember.value || pointsToRedeem.value < 100) return
  const cashDiscountKes = pointsToRedeem.value / 10
  const discountMinor = Math.round(cashDiscountKes * 100)

  // Deduct from profile
  activeMember.value.points -= pointsToRedeem.value

  emit('apply-discount', {
    code: `LOYALTY-${activeMember.value.cardNo}`,
    description: `Loyalty Redemption (${pointsToRedeem.value} pts - ${activeMember.value.name})`,
    amountMinor: discountMinor,
  })

  emit('notification', {
    message: `Redeemed ${pointsToRedeem.value} points for KES ${cashDiscountKes.toFixed(2)} discount on cart!`,
    type: 'success',
  })
}

function enrollMember() {
  if (!newMember.value.name || !newMember.value.phone) return
  const cardNo = `LOYAL-${1000 + members.value.length + 1}`
  const created = {
    id: `mem-${Date.now()}`,
    name: newMember.value.name,
    phone: newMember.value.phone,
    cardNo,
    tier: 'Bronze',
    points: 100, // 100 bonus points
  }
  members.value.unshift(created)
  selectMember(created)
  showEnrollModal.value = false
  emit('notification', {
    message: `Enrolled ${created.name} with card ${created.cardNo} + 100 welcome points!`,
    type: 'success',
  })
  newMember.value = { name: '', phone: '', email: '' }
}
</script>

<style scoped>
.loyalty-manager {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
}
.manager-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}
.subtitle {
  margin: 0.25rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}
.btn-enroll {
  background: #f59e0b;
  color: #1e293b;
  border: none;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.btn-enroll:hover {
  background: #d97706;
  color: #fff;
}
.lookup-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}
.lookup-card label {
  font-weight: 600;
  font-size: 0.9rem;
  color: #334155;
}
.search-input-group {
  display: flex;
  gap: 0.5rem;
  max-width: 520px;
}
.search-input {
  flex: 1;
  padding: 0.6rem 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.95rem;
}
.btn-primary {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.quick-samples {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: #64748b;
  flex-wrap: wrap;
}
.btn-tag {
  background: #e2e8f0;
  border: none;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.75rem;
  font-weight: 600;
}
.member-profile-card {
  background: #fff;
  border: 2px solid #38bdf8;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  box-shadow: 0 4px 12px rgba(2, 132, 199, 0.08);
}
.profile-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.avatar-tier {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-weight: 900;
  font-size: 1.5rem;
  color: #fff;
}
.avatar-tier.gold { background: linear-gradient(135deg, #f59e0b, #b45309); }
.avatar-tier.silver { background: linear-gradient(135deg, #94a3b8, #475569); }
.avatar-tier.bronze { background: linear-gradient(135deg, #d97706, #78350f); }
.profile-details {
  flex-grow: 1;
}
.profile-details h4 {
  margin: 0 0 0.35rem;
  font-size: 1.2rem;
}
.profile-meta {
  display: flex;
  gap: 1rem;
  font-size: 0.85rem;
  color: #64748b;
  flex-wrap: wrap;
}
.badge {
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
}
.badge.gold { background: #fef3c7; color: #b45309; }
.badge.silver { background: #f1f5f9; color: #334155; }
.badge.bronze { background: #ffedd5; color: #9a3412; }
.points-hero {
  text-align: right;
  border-left: 2px solid #e2e8f0;
  padding-left: 1.25rem;
}
.points-num {
  font-size: 2rem;
  font-weight: 900;
  color: #0284c7;
  line-height: 1;
}
.points-sub {
  font-size: 0.8rem;
  color: #64748b;
  margin-top: 0.25rem;
}
.redemption-control {
  background: #f8fafc;
  padding: 1rem;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.redeem-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
}
.redeem-inputs {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  font-size: 0.9rem;
}
.redeem-field {
  width: 90px;
  padding: 0.4rem 0.5rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-weight: bold;
}
.equals-label strong {
  color: #166534;
}
.redeem-actions {
  display: flex;
  gap: 0.5rem;
}
.btn-secondary {
  background: #fff;
  border: 1px solid #cbd5e1;
  padding: 0.45rem 0.85rem;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.btn-redeem {
  background: #166534;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 700;
  cursor: pointer;
}
.btn-redeem:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.rate-note {
  color: #64748b;
  font-size: 0.75rem;
}
.members-directory h4 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.members-table-wrapper {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}
.members-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.85rem;
}
.members-table th, .members-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}
.members-table th {
  background: #f8fafc;
  font-weight: 700;
  color: #475569;
}
.tier-pill {
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
}
.tier-pill.gold { background: #fef3c7; color: #92400e; }
.tier-pill.silver { background: #e2e8f0; color: #334155; }
.tier-pill.bronze { background: #ffedd5; color: #9a3412; }
.btn-select {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.25rem 0.65rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.75rem;
  font-weight: 600;
}
.btn-select:hover {
  background: #0284c7;
  color: white;
}
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: grid;
  place-items: center;
  z-index: 1000;
}
.modal-dialog {
  background: white;
  width: min(480px, 94vw);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.close-btn {
  background: none;
  border: none;
  font-size: 1.25rem;
  cursor: pointer;
}
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.modal-form label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.modal-form input {
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
}
.enroll-perk {
  background: #fef3c7;
  color: #92400e;
  padding: 0.75rem;
  border-radius: 8px;
  font-size: 0.85rem;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
</style>
