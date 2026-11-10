export default [
  {
    path: '/restaurant',
    name: 'restaurant',
    component: () => import('./views/RestaurantView.vue'),
    meta: { requiresAuth: true, permission: 'restaurant.read' },
  },
]