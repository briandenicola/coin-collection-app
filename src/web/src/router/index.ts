import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/RegisterPage.vue'),
    },
    {
      path: '/',
      name: 'collection',
      component: () => import('@/pages/CollectionPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/coin/:id',
      name: 'coin-detail',
      component: () => import('@/pages/CoinDetailPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/add',
      name: 'add-coin',
      component: () => import('@/pages/AddCoinPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/edit/:id',
      name: 'edit-coin',
      component: () => import('@/pages/EditCoinPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/wishlist',
      name: 'wishlist',
      component: () => import('@/pages/WishlistPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/stats',
      name: 'stats',
      component: () => import('@/pages/StatsPage.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login' }
  }
})

export default router
