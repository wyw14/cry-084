import { createRouter, createWebHistory } from 'vue-router'
import DashboardPage from '../pages/DashboardPage.vue'
import AssetsPage from '../pages/AssetsPage.vue'
import RoutesPage from '../pages/RoutesPage.vue'
import InspectionsPage from '../pages/InspectionsPage.vue'
import HazardsPage from '../pages/HazardsPage.vue'
import ReportsPage from '../pages/ReportsPage.vue'
export const router=createRouter({history:createWebHistory(),routes:[{path:'/',component:DashboardPage},{path:'/assets',component:AssetsPage},{path:'/routes',component:RoutesPage},{path:'/inspections',component:InspectionsPage},{path:'/hazards',component:HazardsPage},{path:'/reports',component:ReportsPage}]})
