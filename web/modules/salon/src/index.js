import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'salon',
  requires: ['core-checkout'],
  capabilities: ['appointments', 'commissions', 'packages', 'loyalty'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})
