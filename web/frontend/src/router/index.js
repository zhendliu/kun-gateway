import { createRouter, createWebHistory } from "vue-router";
import Dashboard from "../views/Dashboard.vue";
import Routes from "../views/Routes.vue";
import Services from "../views/Services.vue";
import Metrics from "../views/Metrics.vue";

const routes = [
  {
    path: "/",
    name: "Dashboard",
    component: Dashboard,
  },
  {
    path: "/routes",
    name: "Routes",
    component: Routes,
  },
  {
    path: "/services",
    name: "Services",
    component: Services,
  },
  {
    path: "/metrics",
    name: "Metrics",
    component: Metrics,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
