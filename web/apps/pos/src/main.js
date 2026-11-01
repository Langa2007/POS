import { createApp, h } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { createRouter, createWebHistory } from 'vue-router'
import { RouterView } from 'vue-router'
import App from './App.vue'
import './style.css'

const i18n = createI18n({ legacy: false, locale: 'en', fallbackLocale: 'en', messages: {
  en: { language: 'English' },
  sw: { language: 'Kiswahili' },
} })
const router = createRouter({ history: createWebHistory(), routes: [{ path: '/:pathMatch(.*)*', component: App }] })
createApp({ render: () => h(RouterView) }).use(createPinia()).use(router).use(i18n).mount('#app')
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => navigator.serviceWorker.register('/sw.js').catch(() => undefined))
}
