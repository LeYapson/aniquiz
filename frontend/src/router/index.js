import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import NotFound from '../components/NotFound.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: App },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFound },
  ],
})
