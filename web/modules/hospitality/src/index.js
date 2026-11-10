import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'hospitality',
  requires: ['core-checkout'],
  capabilities: ['room_booking', 'rate_plans', 'guest_folio', 'housekeeping', 'night_audit', 'group_billing', 'deposits'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})