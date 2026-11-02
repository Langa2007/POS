export default [
  {
    path: '/demo',
    name: 'demo',
    component: () => import('./views/DemoView.vue'),
    meta: { requiresAuth: true },
  },
]
