import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'
import en from './locales/en.json'
import sw from './locales/sw.json'

export default defineModule({
  id: 'restaurant',
  requires: ['core-checkout'],
  capabilities: ['table_service', 'kitchen_display', 'menu_modifiers', 'split_bills', 'recipes'],
  routes,
  slots,
  nav,
  i18n: { en, sw },
})