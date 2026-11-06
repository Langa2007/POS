<script setup>
/**
 * WorkspaceView — authenticated shell.
 * Renders the topbar, nav, sync badge, and <RouterView> for child routes.
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

async function syncNow() {
  const result = await connectivity.sync()
  if (result?.accepted) toast.success(t('sync.syncSuccess', result.accepted))
  if (result?.rejected) toast.warning(t('sync.reviewErrors', result.rejected))
}
</script>

<template>
  <div class="app-frame">
    <header class="topbar">
      <a class="brand" href="/" :aria-label="t('app.name')">
        <span class="brand-mark" aria-hidden="true">P</span>
        <span>platform<span class="brand-light">POS</span></span>
      </a>

      <nav class="topbar-nav" aria-label="Main navigation">
        <RouterLink :to="{ name: 'checkout' }"  class="nav-link">{{ t('nav.checkout')  }}</RouterLink>
        <RouterLink :to="{ name: 'backoffice' }" class="nav-link">{{ t('nav.reports')   }}</RouterLink>
      </nav>

      <div class="topbar-actions">
        <PosSyncBadge
          :online="connectivity.online"
          :pending="connectivity.pending"
          :rejected="connectivity.rejected.length"
        />
        <span class="topbar-user">{{ manifest.businessName }}</span>
        <PosButton variant="secondary" @click="signOut">{{ t('auth.signOut') }}</PosButton>
      </div>
    </header>

    <main class="workspace-content">
      <RouterView />
    </main>

    <footer>
      {{ t('app.name') }} <span>•</span>
      <button class="lang-toggle" @click="$i18n.locale = $i18n.locale === 'en' ? 'sw' : 'en'">
        {{ $i18n.locale === 'en' ? 'Kiswahili' : 'English' }}
      </button>
    </footer>

    <PosToastStack />
  </div>
</template>

<style scoped>
.app-frame { max-width: 1280px; min-height: 100vh; margin: auto; padding: 0 56px; display: flex; flex-direction: column; }
.topbar {
  height: 64px; display: flex; align-items: center; justify-content: space-between;
  gap: 20px; border-bottom: 1px solid #e7eaf0; flex-shrink: 0;
}
.topbar-nav { display: flex; gap: 4px; }
.nav-link {
  padding: 7px 14px; border-radius: 8px; font-size: 13px; font-weight: 600;
  color: #6b7c88; text-decoration: none; transition: background 0.15s, color 0.15s;
}
.nav-link:hover { background: #f0f4f7; color: #182230; }
.nav-link.router-link-active { background: #e8f5f0; color: #126b55; }
.topbar-actions { display: flex; align-items: center; gap: 10px; }
.topbar-user { font-size: 12px; font-weight: 650; color: #5d6b78; max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workspace-content { flex: 1; padding: 32px 0; }
footer { padding: 16px 0; border-top: 1px solid #e7eaf0; color: #89939b; font-size: 10px; }
footer span { padding: 0 8px; color: #c4cbc9; }
.lang-toggle { border: 0; background: none; font: inherit; font-size: 10px; color: #89939b; cursor: pointer; padding: 0; }
.lang-toggle:hover { color: #126b55; }
@media (max-width: 900px) { .app-frame { padding: 0 24px; } }
@media (max-width: 600px) {
  .app-frame { padding: 0 16px; }
  .topbar-nav { display: none; }
}
</style>

