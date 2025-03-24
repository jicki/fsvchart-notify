import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import UserProfileView from '../views/UserProfileView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/profile',
      name: 'profile',
      component: UserProfileView,
      meta: { requiresAuth: true }
    },
    {
      path: '/send-records',
      name: 'sendRecords',
      component: () => import('../views/RunLogsView.vue'),
      meta: { requiresAuth: true }
    },
    // Add a catch-all route to redirect to login
    {
      path: '/:pathMatch(.*)*',
      redirect: '/login'
    }
  ]
})

// 全局前置守卫
router.beforeEach((to, from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const token = localStorage.getItem('token')
  
  if (requiresAuth && !token) {
    // 需要认证但未登录，重定向到登录页
    next({ name: 'login' })
  } else if (to.name === 'login' && token) {
    // 已登录用户访问登录页，重定向到首页
    next({ name: 'home' })
  } else {
    // 其他情况正常导航
    next()
  }
})

export default router 