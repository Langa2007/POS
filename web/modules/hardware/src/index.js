import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'hardware',
  requires: ['core-checkout'],
  capabilities: ['multi_uom', 'cut_to_length', 'credit_accounts', 'contract_pricing', 'quotes', 'delivery'],
  routes, slots, nav, i18n: { en, sw },
})