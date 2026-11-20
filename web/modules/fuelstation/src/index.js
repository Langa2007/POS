import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'fuelstation',
  requires: ['core-checkout'],
  capabilities: ['pump_integration', 'tank_dips', 'meter_shifts', 'pump_reconciliation'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})
