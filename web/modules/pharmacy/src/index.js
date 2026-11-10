import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'pharmacy',
  requires: ['core-checkout'],
  capabilities: ['batch_tracking', 'expiry_tracking', 'prescriptions', 'controlled_substances', 'multi_uom'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})