import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'bookshop',
  requires: ['core-checkout'],
  capabilities: ['isbn_catalog', 'consignment', 'publisher_returns', 'special_orders', 'second_hand', 'school_orders'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})
