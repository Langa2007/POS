export default [
  {
    path: '/pharmacy',
    name: 'pharmacy',
    component: () => import('./views/PharmacyView.vue'),
    meta: { requiresAuth: true, permission: 'pharmacy.read' },
  },
]