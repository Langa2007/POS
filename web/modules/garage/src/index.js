import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'garage',
  requires: ['core-checkout'],
  capabilities: ['job_cards', 'parts_and_labor', 'service_reminders', 'appointments'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})
