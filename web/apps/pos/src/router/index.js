import { createRouter, createWebHistory } from 'vue-router'
import { useSessionStore } from '../stores/session.js'
import { useManifestStore } from '../stores/manifest.js'

// Lazy-load every route so code-splitting keeps the initial bundle small
const LoginView    = () => import('../views/LoginView.vue')
const WorkspaceView= () => import('../views/WorkspaceView.vue')
const CheckoutView = () => import('../views/CheckoutView.vue')
const BackofficeView = () => import('../views/BackofficeView.vue')
const NotFoundView = () => import('../views/NotFoundView.vue')

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { public: true },
    },
    {
      path: '/',
      name: 'workspace',
      component: WorkspaceView,
      redirect: { name: 'checkout' },
      meta: { requiresAuth: true },
      children: [
        { path: 'checkout',  name: 'checkout',  component: CheckoutView },
        { path: 'backoffice',name: 'backoffice',component: BackofficeView },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: NotFoundView,
    },
  ],
})

export default router

// ── Navigation guards ──────────────────────────────────────────────────────
router.beforeEach(async (to) => {
  const session  = useSessionStore()
  const manifest = useManifestStore()

  // Allow public routes unconditionally
  if (to.meta.public) return true

  // Not authenticated → go to login
  if (!session.isAuthenticated) return { name: 'login', query: { redirect: to.fullPath } }

  // Authenticated but manifest not yet loaded → load it
  if (!manifest.manifest) {
    try {
      await manifest.load()
    } catch (err) {
      if (err.message === 'session-expired') {
        session.clear()
        return { name: 'login' }
      }
      // If we can't load the manifest (e.g. offline) but have a cached one, proceed
      if (!manifest.manifest) return { name: 'login' }
    }
  }

  return true
})
