<template>
  <div class="dispensing-label-generator">
    <div class="manager-header">
      <div>
        <h3>🏷️ Clinical Prescription & Dispensing Labels</h3>
        <p class="subtitle">Generate standard clinical sticker labels for medication bottles, boxes, and blister strips</p>
      </div>
      <button class="btn-primary" @click="printLabel">
        🖨️ Print Dispensing Label
      </button>
    </div>

    <div class="label-builder-layout">
      <!-- Edit Controls Form -->
      <div class="editor-column">
        <div class="edit-card">
          <div class="card-title">Label Details Editor</div>

          <label>Pharmacy Name & Header:
            <input type="text" v-model="labelData.pharmacyName" />
          </label>
          <label>Pharmacy Contact / Location:
            <input type="text" v-model="labelData.pharmacyContact" />
          </label>

          <div class="form-row">
            <label>Patient Full Name:
              <input type="text" v-model="labelData.patientName" />
            </label>
            <label>Date Dispensed:
              <input type="date" v-model="labelData.dispensedDate" />
            </label>
          </div>

          <div class="form-row">
            <label>Drug Name & Strength:
              <input type="text" v-model="labelData.drugName" />
            </label>
            <label>Quantity Dispensed:
              <input type="text" v-model="labelData.quantity" />
            </label>
          </div>

          <label>Patient Directions for Use (Sig):
            <textarea v-model="labelData.directions" rows="2"></textarea>
          </label>

          <div class="quick-directions">
            <span>Quick Presets:</span>
            <button type="button" class="btn-dir" @click="setDir('Take ONE tablet THREE times daily after meals')">TDS After Meals</button>
            <button type="button" class="btn-dir" @click="setDir('Take ONE tablet TWICE daily every 12 hours with water')">BD 12-Hourly</button>
            <button type="button" class="btn-dir" @click="setDir('Take ONE tablet ONCE daily at night before bed')">OD At Bedtime</button>
            <button type="button" class="btn-dir" @click="setDir('Take 1-2 tablets as needed for pain. Max 8 tabs per 24 hours')">PRN For Pain</button>
          </div>

          <div class="cautionary-advisories">
            <label class="section-label">Cautionary Advisory Stickers:</label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="labelData.advisories.completeCourse" />
              ⚠️ Complete the full prescribed course of treatment
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="labelData.advisories.drowsiness" />
              🚗 May cause drowsiness - do not drive or operate machinery
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="labelData.advisories.withFood" />
              🍽️ Take with or immediately after food
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="labelData.advisories.noAlcohol" />
              🚫 Avoid alcoholic beverages while taking this medication
            </label>
          </div>
        </div>
      </div>

      <!-- Live Clinical Label Preview -->
      <div class="preview-column">
        <div class="preview-card">
          <div class="preview-title">Label Preview (Thermal 80mm Sticker)</div>
          
          <div class="clinical-sticker">
            <div class="sticker-top">
              <div class="pharmacy-brand">{{ labelData.pharmacyName }}</div>
              <div class="pharmacy-sub">{{ labelData.pharmacyContact }} · Lic: PPB-49210</div>
            </div>

            <div class="sticker-patient-line">
              <span>Patient: <strong>{{ labelData.patientName }}</strong></span>
              <span>Date: {{ labelData.dispensedDate }}</span>
            </div>

            <div class="sticker-drug-box">
              <div class="drug-main">{{ labelData.drugName }}</div>
              <div class="drug-qty">Total Qty: <strong>{{ labelData.quantity }}</strong></div>
            </div>

            <div class="sticker-directions">
              <div class="sig-label">DIRECTIONS:</div>
              <div class="sig-text">{{ labelData.directions }}</div>
            </div>

            <div class="sticker-warnings" v-if="hasActiveAdvisories">
              <div v-if="labelData.advisories.completeCourse" class="warning-tag">
                ⚠️ FINISH COMPLETE COURSE
              </div>
              <div v-if="labelData.advisories.drowsiness" class="warning-tag">
                🚗 MAY CAUSE DROWSINESS
              </div>
              <div v-if="labelData.advisories.withFood" class="warning-tag">
                🍽️ TAKE AFTER MEALS
              </div>
              <div v-if="labelData.advisories.noAlcohol" class="warning-tag">
                🚫 NO ALCOHOL
              </div>
            </div>

            <div class="sticker-foot">
              <span>KEEP OUT OF REACH OF CHILDREN</span>
              <span>Dispensed by: R.Ph Faith Waweru</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['notification'])

const labelData = ref({
  pharmacyName: 'GoodHealth Community Pharmacy',
  pharmacyContact: 'Westlands Mall, Ground Floor · Tel: 0722-123456',
  patientName: 'Sarah Mwangi',
  dispensedDate: new Date().toISOString().split('T')[0],
  drugName: 'Amoxicillin Capsules 500mg',
  quantity: '21 Capsules (7-day course)',
  directions: 'Take ONE capsule THREE times daily after meals with a full glass of water.',
  advisories: {
    completeCourse: true,
    drowsiness: false,
    withFood: true,
    noAlcohol: false,
  },
})

const hasActiveAdvisories = computed(() => {
  const a = labelData.value.advisories
  return a.completeCourse || a.drowsiness || a.withFood || a.noAlcohol
})

function setDir(text) {
  labelData.value.directions = text
}

function printLabel() {
  window.print()
  emit('notification', { message: `Dispensing label sent to printer for ${labelData.value.patientName}`, type: 'success' })
}
</script>

<style scoped>
.dispensing-label-generator {
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
.btn-primary {
  background: #0284c7;
  color: white;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.label-builder-layout {
  display: grid;
  grid-template-columns: minmax(320px, 1.2fr) minmax(320px, 1fr);
  gap: 1.5rem;
}
.edit-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.card-title {
  font-weight: 800;
  color: #0f172a;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 0.35rem;
}
.edit-card label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}
.edit-card input, .edit-card textarea {
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: white;
  font-family: inherit;
}
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.quick-directions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
  font-size: 0.75rem;
  color: #64748b;
}
.btn-dir {
  background: #e2e8f0;
  border: none;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-dir:hover {
  background: #cbd5e1;
}
.cautionary-advisories {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  padding: 0.75rem;
  border-radius: 8px;
}
.section-label {
  font-size: 0.8rem;
  font-weight: 700;
  color: #0f172a;
}
.checkbox-label {
  flex-direction: row !important;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  cursor: pointer;
}
.preview-card {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}
.preview-title {
  font-size: 0.85rem;
  font-weight: 700;
  color: #64748b;
}
.clinical-sticker {
  background: white;
  border: 2px solid #0f172a;
  border-radius: 6px;
  width: 100%;
  max-width: 360px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  box-shadow: 0 4px 15px rgba(0,0,0,0.08);
}
.sticker-top {
  text-align: center;
  border-bottom: 2px solid #0f172a;
  padding-bottom: 0.4rem;
}
.pharmacy-brand {
  font-weight: 900;
  font-size: 1rem;
  color: #0f172a;
  text-transform: uppercase;
}
.pharmacy-sub {
  font-size: 0.7rem;
  color: #475569;
}
.sticker-patient-line {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  border-bottom: 1px dashed #cbd5e1;
  padding-bottom: 0.35rem;
}
.sticker-drug-box {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 0.5rem;
  border-radius: 4px;
}
.drug-main {
  font-size: 1.05rem;
  font-weight: 900;
  color: #0f172a;
}
.drug-qty {
  font-size: 0.8rem;
  color: #334155;
  margin-top: 0.15rem;
}
.sticker-directions {
  padding: 0.4rem 0;
}
.sig-label {
  font-size: 0.75rem;
  font-weight: 800;
  color: #0f172a;
}
.sig-text {
  font-size: 0.95rem;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.3;
  margin-top: 0.15rem;
}
.sticker-warnings {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.35rem 0;
}
.warning-tag {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
  font-size: 0.7rem;
  font-weight: 800;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
}
.sticker-foot {
  display: flex;
  flex-direction: column;
  align-items: center;
  border-top: 1px solid #0f172a;
  padding-top: 0.35rem;
  font-size: 0.65rem;
  font-weight: 700;
  color: #64748b;
  text-align: center;
}

@media (max-width: 768px) {
  .label-builder-layout {
    grid-template-columns: 1fr;
  }
}
@media print {
  body * {
    visibility: hidden;
  }
  .clinical-sticker, .clinical-sticker * {
    visibility: visible;
  }
  .clinical-sticker {
    position: absolute;
    top: 0;
    left: 0;
    width: 72mm;
    border: none;
  }
}
</style>
