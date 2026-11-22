import { defineModule } from '@pos/module-sdk'
import routes from './routes'
import slots from './slots'
import nav from './nav'

export default defineModule({
  id: 'dealership',
  requires: ['core-checkout'],
  capabilities: ['serial_tracking', 'warranty', 'trade_ins', 'installments', 'deposits', 'document_packs'],
  routes,
  slots,
  nav,
})
