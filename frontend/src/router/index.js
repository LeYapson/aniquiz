import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import NotFound from '../components/NotFound.vue'
import LegalPage from '../views/LegalPage.vue'
import TermsPage from '../views/TermsPage.vue'
import PrivacyPage from '../views/PrivacyPage.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: App },
    { path: '/legal', component: LegalPage },
    { path: '/terms', component: TermsPage },
    { path: '/privacy', component: PrivacyPage },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFound },
  ],
})
