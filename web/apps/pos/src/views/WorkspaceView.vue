<script setup>
/**
 * WorkspaceView — authenticated shell.
 * Renders the topbar, active shop preset switcher, nav, sync badge, and <RouterView>.
 */
import { onMounted, onUnmounted } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { PosButton, PosSyncBadge } from '../../../../packages/ui/src/index.js'
import { useSessionStore } from '../stores/session.js'
import { useManifestStore } from '../stores/manifest.js'
import { useConnectivityStore } from '../stores/connectivity.js'
import { useToastStore } from '../stores/toast.js'
import PosToastStack from '../components/PosToastStack.vue'

const { t } = useI18n()
const router      = useRouter()
const session     = useSessionStore()
const manifest    = useManifestStore()
const connectivity= useConnectivityStore()
const toast       = useToastStore()

onMounted(() => {
  connectivity.startMonitor()
  connectivity.refreshCounts()
})
onUnmounted(() => connectivity.stopMonitor())

function signOut() {
  connectivity.stopMonitor()
  session.clear()
  manifest.clear()
  router.push({ name: 'login' })
}

function switchShop(shopId) {
  manifest.setActiveShop(shopId)
  const shop = manifest.availableShops.find((s) => s.id === shopId)
  toast.info(`Switched active shop till context to: ${shop ? shop.name : shopId}`)
}
</script>

<template>
  <div class="app-frame">
    <header class="topbar">
      <div class="brand-group">
        <a class="brand" href="/" :aria-label="t('app.name')">
          <span class="brand-mark" aria-hidden="true">P</span>
          <span>platform<span class="brand-light">POS</span></span>
        </a>

        <!-- Shop Switcher Dropdown -->
        <div class="shop-switcher-container">
          <label for="shopContextSelect" class="sr-only">Active Shop Context</label>
          <select
            id="shopContextSelect"
            :value="manifest.activeShop"
            @change="switchShop($event.target.value)"
            class="shop-context-select"
          >
            <option
              v-for="shop in manifest.availableShops"
              :key="shop.id"
              :value="shop.id"
            >
              {{ shop.badge }}
            </option>
          </select>
        </div>
      </div>

      <nav class="topbar-nav" aria-label="Main navigation">
        <RouterLink :to="{ name: 'checkout' }" class="nav-link">
          💳 {{ t('nav.checkout') }}
        </RouterLink>
        <RouterLink
          to="/retail"
          class="nav-link"
          :class="{ 'highlight': manifest.activeShop === 'retail' }"
        >
          🛒 Retail Store
        </RouterLink>
        <RouterLink
          to="/pharmacy"
          class="nav-link"
          :class="{ 'highlight': manifest.activeShop === 'pharmacy' }"
        >
          💊 Pharmacy Dispensary
        </RouterLink>
        <RouterLink
          to="/hardware"
          class="nav-link"
          :class="{ 'highlight': manifest.activeShop === 'hardware' }"
        >
          🧱 Hardware
        </RouterLink>
        <RouterLink
          to="/restaurant"
          class="nav-link"
          :class="{ 'highlight': manifest.activeShop === 'restaurant' }"
        >
          🍽️ Restaurant
        </RouterLink>
        <RouterLink
          to="/bookshop"
          class="nav-link"
          :class="{ 'highlight': manifest.activeShop === 'bookshop' }"
        >
          📚 Bookshop
        </RouterLink>
        <RouterLink
          to="/salon"
          class="nav-link"
          :class="{ 'highlight': manifest.activeShop === 'salon' }"
        >
          ✨ Salon
        </RouterLink>
        <RouterLink :to="{ name: 'backoffice' }" class="nav-link">
          📊 {{ t('nav.reports') }}
        </RouterLink>
      </nav>

      <div class="topbar-actions">
        <PosSyncBadge
          :online="connectivity.online"
          :pending="connectivity.pending"
          :rejected="connectivity.rejected.length"
        />
        <span class="topbar-user" :title="manifest.businessName">{{ manifest.businessName }}</span>
        <PosButton variant="secondary" @click="signOut">{{ t('auth.signOut') }}</PosButton>
      </div>
    </header>

    <main class="workspace-content">
      <RouterView />
    </main>

    <footer>
      {{ t('app.name') }} <span>•</span> Multi-Shop Single-Backend Architecture <span>•</span>
      <button class="lang-toggle" @click="$i18n.locale = $i18n.locale === 'en' ? 'sw' : 'en'">
        {{ $i18n.locale === 'en' ? 'Kiswahili' : 'English' }}
      </button>
    </footer>

    <PosToastStack />
  </div>
</template>

<style scoped>
.app-frame { max-width: 1360px; min-height: 100vh; margin: auto; padding: 0 40px; display: flex; flex-direction: column; }
.topbar {
  height: 68px; display: flex; align-items: center; justify-content: space-between;
  gap: 16px; border-bottom: 1px solid #e7eaf0; flex-shrink: 0;
}
.brand-group {
  display: flex;
  align-items: center;
  gap: 16px;
}
.shop-switcher-container {
  display: flex;
  align-items: center;
}
.shop-context-select {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 6px 10px;
  font-weight: 700;
  font-size: 12px;
  color: #0f172a;
  cursor: pointer;
  outline: none;
}
.shop-context-select:focus {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px #e0f2fe;
}
.topbar-nav { display: flex; gap: 4px; }
.nav-link {
  padding: 8px 14px; border-radius: 8px; font-size: 13px; font-weight: 650;
  color: #64748b; text-decoration: none; transition: background 0.15s, color 0.15s;
}
.nav-link:hover { background: #f0f4f7; color: #182230; }
.nav-link.router-link-active { background: #e0f2fe; color: #0369a1; }
.nav-link.highlight { border: 1px solid #bae6fd; }
.topbar-actions { display: flex; align-items: center; gap: 10px; }
.topbar-user { font-size: 12px; font-weight: 700; color: #475569; max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workspace-content { flex: 1; padding: 24px 0; }
footer { padding: 16px 0; border-top: 1px solid #e7eaf0; color: #89939b; font-size: 10px; text-align: center; }
footer span { padding: 0 8px; color: #c4cbc9; }
.lang-toggle { border: 0; background: none; font: inherit; font-size: 10px; color: #0284c7; cursor: pointer; padding: 0; font-weight: 700; }
.lang-toggle:hover { text-decoration: underline; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); border: 0; }
@media (max-width: 960px) {
  .app-frame { padding: 0 20px; }
  .topbar-user { display: none; }
}
@media (max-width: 768px) {
  .topbar-nav { display: none; }
}
</style>
