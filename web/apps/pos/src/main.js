import { createApp, h } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { RouterView } from 'vue-router'
import './style.css'
import router from './router/index.js'
import en from './locales/en.json'
import sw from './locales/sw.json'

const i18n = createI18n({ legacy: false, locale: 'en', fallbackLocale: 'en', messages: { en, sw } })
createApp({ render: () => h(RouterView) }).use(createPinia()).use(router).use(i18n).mount('#app')
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => navigator.serviceWorker.register('/sw.js').catch(() => undefined))
}
