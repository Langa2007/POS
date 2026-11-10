export default [
  {
    path: '/hospitality',
    name: 'hospitality',
    component: () => import('./views/HospitalityView.vue'),
    meta: { requiresAuth: true, permission: 'hospitality.read' },
  },
]