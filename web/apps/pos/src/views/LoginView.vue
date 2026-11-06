<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { PosButton, PosInput } from '../../../../packages/ui/src/index.js'
import { useSessionStore } from '../stores/session.js'
import { useManifestStore } from '../stores/manifest.js'
import { useConnectivityStore } from '../stores/connectivity.js'
import { useToastStore } from '../stores/toast.js'

const { t } = useI18n()
const router = useRouter()
const route  = useRoute()
const session     = useSessionStore()
const manifest    = useManifestStore()
const connectivity= useConnectivityStore()
const toast       = useToastStore()

const tenantId = ref(session.tenantId)
const email    = ref('')
const password = ref('')
const busy     = ref(false)
const error    = ref('')

async function login() {
  error.value = ''
  busy.value  = true
  try {
    const res  = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tenant_id: tenantId.value, email: email.value, password: password.value }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error?.message ?? t('auth.invalidCredentials'))
    session.setTokens({ ...data, tenant_id: tenantId.value })
    await manifest.load()
    connectivity.startMonitor()
    await connectivity.refreshCounts()
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.push(redirect)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : t('errors.generic')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <main class="app-frame">
    <header class="topbar">
      <a class="brand" href="/" :aria-label="t('app.name')">
        <span class="brand-mark" aria-hidden="true">P</span>
        <span>platform<span class="brand-light">POS</span></span>
      </a>
      <span class="environment">{{ t('app.name').toUpperCase() }}</span>
    </header>

    <section class="auth-layout">
      <div class="intro">
        <p class="eyebrow">ONE PLATFORM. EVERY COUNTER.</p>
        <h1>{{ t('app.tagline').split(' ').slice(0, -1).join(' ') }}<br>
          <span>{{ t('app.tagline').split(' ').at(-1) }}</span>
        </h1>
        <p class="intro-copy">{{ t('app.intro') }}</p>
        <div class="intro-note">
          <span class="status-dot" aria-hidden="true" />
          Your workspace is ready when you are.
        </div>
      </div>

      <form class="login-card" @submit.prevent="login">
        <div class="card-heading">
          <span class="eyebrow">{{ t('auth.welcomeBack') }}</span>
          <h2>{{ t('auth.signInToWorkspace') }}</h2>
          <p>{{ t('auth.signInHelper') }}</p>
        </div>

        <PosInput
          id="tenant-id"
          v-model="tenantId"
          :label="t('auth.tenantId')"
          :placeholder="t('auth.tenantIdPlaceholder')"
          autocomplete="organization"
          required
        />
        <PosInput
          id="email"
          v-model="email"
          :label="t('auth.email')"
          type="email"
          :placeholder="t('auth.emailPlaceholder')"
          autocomplete="username"
          required
        />
        <PosInput
          id="password"
          v-model="password"
          :label="t('auth.password')"
          type="password"
          :placeholder="t('auth.passwordPlaceholder')"
          autocomplete="current-password"
          required
          :error="error"
        />

        <PosButton type="submit" :disabled="busy" style="width:100%;margin-top:8px">
          {{ busy ? t('auth.signingIn') : t('auth.signIn') }}
          <span aria-hidden="true">→</span>
        </PosButton>
        <p class="secure-note">
          <span aria-hidden="true">◈</span> {{ t('auth.protectedNote') }}
        </p>
      </form>
    </section>

    <footer>
      {{ t('app.name') }} <span>•</span> English <span>•</span> Kiswahili
    </footer>
  </main>
</template>

<style scoped>
/* Auth layout styles — scoped so they don't bleed into WorkspaceView */
.auth-layout {
  flex: 1; display: grid;
  grid-template-columns: 1fr 450px;
  align-items: center; gap: 70px; padding: 55px 8% 62px;
}
.intro { padding-bottom: 15px; }
.intro .eyebrow { color: #188269; }
.intro h1 { font: 800 52px/1.08 Inter, 'Segoe UI', sans-serif; letter-spacing: -2.8px; margin: 20px 0 18px; }
.intro h1 span { color: #14836b; }
.intro-copy { max-width: 390px; color: #6b7781; font-size: 15px; line-height: 1.8; }
.intro-note { margin-top: 37px; font-size: 12px; color: #63726f; display: flex; align-items: center; gap: 9px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; background: #39ac7b; box-shadow: 0 0 0 4px #e1f3eb; }
.login-card {
  background: white; padding: 34px; border: 1px solid #e8ebef;
  border-radius: 17px; box-shadow: 0 16px 44px #2436460b;
  display: grid; gap: 0;
}
.card-heading { margin-bottom: 20px; }
.card-heading .eyebrow { color: #188269; }
.card-heading h2 { font: 700 21px Inter, 'Segoe UI', sans-serif; letter-spacing: -.7px; margin: 11px 0 6px; }
.card-heading p { font-size: 12px; color: #84909a; margin: 0; }
.secure-note { text-align: center; font-size: 10px; color: #84908d; margin: 14px 0 0; }
.secure-note span { color: #168267; margin-right: 4px; }
@media (max-width: 700px) {
  .auth-layout { display: flex; flex-direction: column; padding: 42px 0; gap: 24px; }
  .intro h1 { font-size: 38px; }
  .intro-note { display: none; }
}
</style>

