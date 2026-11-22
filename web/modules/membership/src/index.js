import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import nav from './nav'

export default defineModule({
  id: 'membership',
  requires: ['core-checkout'],
  capabilities: ['recurring_billing', 'membership_plans', 'attendance', 'idempotent_billing'],
  routes,
  slots: {},
  nav,
})
