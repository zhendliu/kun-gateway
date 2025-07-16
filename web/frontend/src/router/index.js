import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import Routes from '@/views/Routes.vue'
import Services from '@/views/Services.vue'
import Metrics from '@/views/Metrics.vue'
import Certificates from '@/views/Certificates.vue'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/routes',
    name: 'Routes',
    component: Routes
  },
  {
    path: '/services',
    name: 'Services',
    component: Services
  },
  {
    path: '/metrics',
    name: 'Metrics',
    component: Metrics
  },
  {
    path: '/certificates',
    name: 'Certificates',
    component: Certificates
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
