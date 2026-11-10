export default [
  {
    path: '/retail',
    name: 'retail',
    component: () => import('./views/RetailView.vue'),
    meta: { requiresAuth: true, permission: 'retail.read' },
  },
]