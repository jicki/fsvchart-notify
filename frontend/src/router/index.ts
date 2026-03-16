import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../utils/storage'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue')
    },
    {
      path: '/',
      component: () => import('../components/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/push-tasks'
        },
        {
          path: 'datasources',
          name: 'datasources',
          component: () => import('../views/DataSourceView.vue')
        },
        {
          path: 'webhooks',
          name: 'webhooks',
          component: () => import('../views/WebhookView.vue')
        },
        {
          path: 'chart-templates',
          name: 'chartTemplates',
          component: () => import('../views/ChartTemplatesView.vue')
        },
        {
          path: 'promql',
          name: 'promql',
          component: () => import('../views/PromQLView.vue')
        },
        {
          path: 'push-tasks',
          name: 'pushTasks',
          component: () => import('../views/PushTasksView.vue')
        },
        {
          path: 'send-records',
          name: 'sendRecords',
          component: () => import('../views/RunLogsView.vue')
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('../views/UserProfileView.vue')
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/login'
    }
  ]
})

// 全局前置守卫
router.beforeEach((to, _from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const token = getToken()

  if (requiresAuth && !token) {
    next({ name: 'login' })
  } else if (to.name === 'login' && token) {
    next('/push-tasks')
  } else {
    next()
  }
})

export default router
