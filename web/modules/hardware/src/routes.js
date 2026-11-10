export default [
  { path: '/hardware', name: 'hardware', component: () => import('./views/HardwareView.vue'), meta: { requiresAuth: true, permission: 'hardware.read' } },
]