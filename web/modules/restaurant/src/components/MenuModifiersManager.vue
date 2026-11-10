<template>
  <div class="menu-modifiers-manager">
    <div class="modifiers-header">
      <div>
        <h3>
          <PosIcon name="chef-hat" size="20" class="header-icon" />
          Menu Item Modifiers &amp; Combos
        </h3>
        <p class="subtitle">Customize cooking instructions, sides, upcharges, and allergen notes</p>
      </div>
    </div>

    <div class="builder-layout">
      <!-- Item Selection List -->
      <div class="item-selection-col">
        <h4>Select Dish to Configure:</h4>
        <div class="dishes-list">
          <div
            v-for="dish in baseDishes"
            :key="dish.id"
            class="dish-card"
            :class="{ 'active': selectedDish?.id === dish.id }"
            @click="selectDish(dish)"
          >
            <div class="dish-title">{{ dish.name }}</div>
            <div class="dish-cat">{{ dish.category }}</div>
            <div class="dish-base-price">KES {{ (dish.basePriceMinor / 100).toLocaleString() }}</div>
          </div>
        </div>
      </div>

      <!-- Modifier Options Column -->
      <div class="modifier-options-col" v-if="selectedDish">
        <div class="dish-config-card">
          <div class="config-header">
            <h4>{{ selectedDish.name }}</h4>
            <span class="base-badge">Base: KES {{ (selectedDish.basePriceMinor / 100).toLocaleString() }}</span>
          </div>

          <!-- Doneness Group -->
          <div v-if="selectedDish.donenessOptions" class="option-group">
            <span class="group-label">Cooking Doneness (Required):</span>
            <div class="pill-group">
              <button
                v-for="d in selectedDish.donenessOptions"
                :key="d"
                type="button"
                class="choice-pill"
                :class="{ 'selected': currentConfig.doneness === d }"
                @click="currentConfig.doneness = d"
              >
                {{ d }}
              </button>
            </div>
          </div>

          <!-- Sauce Selection -->
          <div v-if="selectedDish.sauceOptions" class="option-group">
            <span class="group-label">Sauce Selection:</span>
            <div class="pill-group">
              <button
                v-for="s in selectedDish.sauceOptions"
                :key="s.name"
                type="button"
                class="choice-pill"
                :class="{ 'selected': currentConfig.sauce === s.name }"
                @click="currentConfig.sauce = s.name"
              >
                {{ s.name }} <span v-if="s.extraMinor" class="extra-cost">+KES {{ s.extraMinor / 100 }}</span>
              </button>
            </div>
          </div>

          <!-- Side Choice -->
          <div v-if="selectedDish.sideOptions" class="option-group">
            <span class="group-label">Choice of Side:</span>
            <div class="pill-group">
              <button
                v-for="side in selectedDish.sideOptions"
                :key="side.name"
                type="button"
                class="choice-pill"
                :class="{ 'selected': currentConfig.side === side.name }"
                @click="currentConfig.side = side.name"
              >
                {{ side.name }} <span v-if="side.extraMinor" class="extra-cost">+KES {{ side.extraMinor / 100 }}</span>
              </button>
            </div>
          </div>

          <!-- Add-ons / Extras (Checkboxes) -->
          <div v-if="selectedDish.addonOptions" class="option-group">
            <span class="group-label">Add-ons &amp; Upgrades:</span>
            <div class="checkbox-grid">
              <label
                v-for="addon in selectedDish.addonOptions"
                :key="addon.id"
                class="addon-label"
                :class="{ 'checked': isAddonSelected(addon.id) }"
              >
                <input
                  type="checkbox"
                  :checked="isAddonSelected(addon.id)"
                  @change="toggleAddon(addon)"
                />
                <span>{{ addon.name }}</span>
                <strong class="addon-cost">+KES {{ addon.priceMinor / 100 }}</strong>
              </label>
            </div>
          </div>

          <!-- Kitchen Special Instructions -->
          <div class="option-group">
            <span class="group-label">Kitchen Preparation Notes:</span>
            <input
              type="text"
              v-model="currentConfig.specialNotes"
              placeholder="e.g. Extra spicy, sauce on side, gluten allergy"
              class="notes-input"
            />
          </div>

          <!-- Summary & Add to Cart -->
          <div class="config-summary-bar">
            <div class="total-breakdown">
              <span class="sub-label">Final Configured Price:</span>
              <strong class="total-kes">KES {{ (totalPriceMinor / 100).toLocaleString() }}</strong>
            </div>
            <button type="button" class="btn-add-cart" @click="addConfiguredItemToCart">
              <PosIcon name="plus" size="16" />
              Add to Active POS Cart
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { PosIcon } from '@pos/ui'

const emit = defineEmits(['add-item', 'notification'])

const baseDishes = [
  {
    id: 'dish-1',
    name: 'Prime Ribeye Steak 350g',
    category: 'Grill & Steaks',
    basePriceMinor: 260000,
    donenessOptions: ['Rare', 'Medium Rare', 'Medium', 'Medium Well', 'Well Done'],
    sauceOptions: [
      { name: 'Mushroom Truffle', extraMinor: 20000 },
      { name: 'Madagascar Peppercorn', extraMinor: 15000 },
      { name: 'Garlic Herb Butter', extraMinor: 0 },
      { name: 'Red Wine Jus', extraMinor: 15000 },
    ],
    sideOptions: [
      { name: 'Rustic Potato Wedges', extraMinor: 0 },
      { name: 'Chips Masala', extraMinor: 5000 },
      { name: 'Creamed Garlic Spinach', extraMinor: 10000 },
      { name: 'Traditional Ugali & Sukuma', extraMinor: 0 },
    ],
    addonOptions: [
      { id: 'add-1', name: 'Jumbo Grilled Prawn (Surf & Turf)', priceMinor: 80000 },
      { id: 'add-2', name: 'Crispy Onion Rings', priceMinor: 25000 },
      { id: 'add-3', name: 'Fried Farm Egg', priceMinor: 15000 },
    ],
  },
  {
    id: 'dish-2',
    name: 'Craft Angus Smash Burger',
    category: 'Burgers & Sandwiches',
    basePriceMinor: 145000,
    donenessOptions: ['Medium Well', 'Well Done'],
    sauceOptions: [
      { name: 'Secret House Burger Sauce', extraMinor: 0 },
      { name: 'Spicy Sriracha Mayo', extraMinor: 0 },
      { name: 'Smoky BBQ Sauce', extraMinor: 0 },
    ],
    sideOptions: [
      { name: 'Skin-on French Fries', extraMinor: 0 },
      { name: 'Side Caesar Salad', extraMinor: 5000 },
      { name: 'Sweet Potato Fries', extraMinor: 10000 },
    ],
    addonOptions: [
      { id: 'add-4', name: 'Extra Cheddar Cheese Slice', priceMinor: 15000 },
      { id: 'add-5', name: 'Crispy Beef Bacon Strip', priceMinor: 25000 },
      { id: 'add-6', name: 'Fresh Guacamole Scoop', priceMinor: 20000 },
      { id: 'add-7', name: 'Extra 150g Angus Beef Patty', priceMinor: 50000 },
    ],
  },
  {
    id: 'dish-3',
    name: 'Whole Grilled Lake Naivasha Tilapia',
    category: 'Fresh Seafood',
    basePriceMinor: 175000,
    donenessOptions: null,
    sauceOptions: [
      { name: 'Lemon Herb Garlic Butter', extraMinor: 0 },
      { name: 'Spicy Coconut Curry Sauce', extraMinor: 15000 },
    ],
    sideOptions: [
      { name: 'Brown Ugali & Kachumbari', extraMinor: 0 },
      { name: 'White Ugali & Kunde', extraMinor: 0 },
      { name: 'French Fries & Salad', extraMinor: 5000 },
    ],
    addonOptions: [
      { id: 'add-8', name: 'Extra Bowl of Kachumbari', priceMinor: 15000 },
      { id: 'add-9', name: 'Avocado Slices', priceMinor: 15000 },
    ],
  },
]

const selectedDish = ref(baseDishes[0])
const currentConfig = ref({
  doneness: 'Medium Rare',
  sauce: 'Mushroom Truffle',
  side: 'Rustic Potato Wedges',
  selectedAddons: [],
  specialNotes: '',
})

function selectDish(dish) {
  selectedDish.value = dish
  currentConfig.value = {
    doneness: dish.donenessOptions ? dish.donenessOptions[1] || dish.donenessOptions[0] : null,
    sauce: dish.sauceOptions ? dish.sauceOptions[0].name : null,
    side: dish.sideOptions ? dish.sideOptions[0].name : null,
    selectedAddons: [],
    specialNotes: '',
  }
}

function isAddonSelected(addonId) {
  return currentConfig.value.selectedAddons.some(a => a.id === addonId)
}

function toggleAddon(addon) {
  const idx = currentConfig.value.selectedAddons.findIndex(a => a.id === addon.id)
  if (idx >= 0) {
    currentConfig.value.selectedAddons.splice(idx, 1)
  } else {
    currentConfig.value.selectedAddons.push(addon)
  }
}

const totalPriceMinor = computed(() => {
  if (!selectedDish.value) return 0
  let total = selectedDish.value.basePriceMinor

  // Sauce extra
  if (selectedDish.value.sauceOptions && currentConfig.value.sauce) {
    const s = selectedDish.value.sauceOptions.find(o => o.name === currentConfig.value.sauce)
    if (s && s.extraMinor) total += s.extraMinor
  }

  // Side extra
  if (selectedDish.value.sideOptions && currentConfig.value.side) {
    const side = selectedDish.value.sideOptions.find(o => o.name === currentConfig.value.side)
    if (side && side.extraMinor) total += side.extraMinor
  }

  // Addons
  for (const a of currentConfig.value.selectedAddons) {
    total += a.priceMinor
  }

  return total
})

function addConfiguredItemToCart() {
  if (!selectedDish.value) return

  const mods = []
  if (currentConfig.value.doneness) mods.push(currentConfig.value.doneness)
  if (currentConfig.value.sauce) mods.push(currentConfig.value.sauce)
  if (currentConfig.value.side) mods.push(`Side: ${currentConfig.value.side}`)
  if (currentConfig.value.selectedAddons.length) {
    mods.push(`Add: ${currentConfig.value.selectedAddons.map(a => a.name).join(', ')}`)
  }
  if (currentConfig.value.specialNotes) {
    mods.push(`Note: ${currentConfig.value.specialNotes}`)
  }

  const modifierString = mods.join(' · ')

  const cartItem = {
    id: `dish-${selectedDish.value.id}-${Date.now()}`,
    name: `${selectedDish.value.name} (${modifierString})`,
    sku: `MOD-${selectedDish.value.id}`,
    price_minor: totalPriceMinor.value,
    unit_price_minor: totalPriceMinor.value,
    quantity: 1,
    kind: 'service',
  }

  // Store in pending queue and emit
  try {
    const existing = JSON.parse(localStorage.getItem('pos.retailCartPending') || '[]')
    localStorage.setItem('pos.retailCartPending', JSON.stringify([...existing, cartItem]))
  } catch (e) {}

  window.dispatchEvent(new CustomEvent('pos:add-item', { detail: cartItem }))
  emit('add-item', cartItem)

  emit('notification', {
    message: `Added custom ${selectedDish.value.name} (KES ${(totalPriceMinor.value / 100).toLocaleString()}) to POS Cart`,
    type: 'success',
  })
}
</script>

<style scoped>
.menu-modifiers-manager {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modifiers-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: #0284c7;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #64748b;
}

.builder-layout {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) minmax(320px, 1.8fr);
  gap: 18px;
  align-items: start;
}

.item-selection-col h4 {
  margin: 0 0 10px;
  font-size: 14px;
  color: #475569;
}

.dishes-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dish-card {
  background: white;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.dish-card:hover {
  background: #f8fafc;
  border-color: #94a3b8;
}

.dish-card.active {
  border-color: #0284c7;
  background: #f0f9ff;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
}

.dish-title {
  font-weight: 700;
  font-size: 14px;
  color: #1e293b;
}

.dish-cat {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}

.dish-base-price {
  font-size: 13px;
  font-weight: 700;
  color: #0284c7;
  margin-top: 6px;
}

/* Config Card */
.dish-config-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 10px;
}

.config-header h4 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}

.base-badge {
  background: #e2e8f0;
  color: #334155;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 6px;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-label {
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.choice-pill {
  padding: 6px 12px;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
  transition: all 0.15s ease;
}

.choice-pill:hover {
  background: #f1f5f9;
}

.choice-pill.selected {
  background: #0284c7;
  color: white;
  border-color: #0284c7;
}

.extra-cost {
  font-size: 10px;
  opacity: 0.85;
}

.checkbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.addon-label {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: #f8fafc;
  cursor: pointer;
  font-size: 12px;
  color: #334155;
}

.addon-label.checked {
  background: #eff6ff;
  border-color: #93c5fd;
}

.addon-cost {
  margin-left: auto;
  color: #0284c7;
  font-size: 11px;
}

.notes-input {
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  outline: none;
}

.notes-input:focus {
  border-color: #0284c7;
}

.config-summary-bar {
  margin-top: 10px;
  padding-top: 14px;
  border-top: 2px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.total-breakdown {
  display: flex;
  flex-direction: column;
}

.sub-label {
  font-size: 11px;
  color: #64748b;
}

.total-kes {
  font-size: 18px;
  color: #0f172a;
}

.btn-add-cart {
  background: #0284c7;
  color: white;
  border: none;
  padding: 10px 18px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: background 0.15s ease;
}

.btn-add-cart:hover {
  background: #0369a1;
}

@media (max-width: 768px) {
  .builder-layout {
    grid-template-columns: 1fr;
  }
}
</style>
