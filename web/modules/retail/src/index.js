import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'retail',
  requires: ['core-checkout'],
  capabilities: ['weighed_items', 'promotions', 'loyalty', 'perishables'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})