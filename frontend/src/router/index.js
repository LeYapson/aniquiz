import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import NotFound from '../components/NotFound.vue'
import LegalPage from '../views/LegalPage.vue'
import TermsPage from '../views/TermsPage.vue'
import PrivacyPage from '../views/PrivacyPage.vue'
import SpeedrunGame from '../views/SpeedrunGame.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: App,
      meta: {
        title: "AniQuiz — Blindtest d'anime en ligne gratuit",
        description: "AniQuiz est le blindtest d'anime en ligne gratuit et multijoueur. Reconnais les openings et endings d'anime, affronte tes amis en temps réel. 500+ animes. Compte gratuit en 30 secondes.",
        canonical: 'https://aniquiz.fr/',
      },
    },
    {
      path: '/speedrun',
      component: SpeedrunGame,
      meta: {
        title: "Speed Run — Blindtest anime 5 minutes | AniQuiz",
        description: "Mode Speed Run d'AniQuiz : reconnais un maximum d'animes en 5 minutes. Enchaîne les ouvertures, maintiens ta série et bats ton record.",
        canonical: 'https://aniquiz.fr/speedrun',
      },
    },
    {
      path: '/legal',
      component: LegalPage,
      meta: {
        title: "Mentions légales | AniQuiz",
        description: "Mentions légales du site AniQuiz, blindtest d'anime en ligne.",
        canonical: 'https://aniquiz.fr/legal',
        noindex: true,
      },
    },
    {
      path: '/terms',
      component: TermsPage,
      meta: {
        title: "Conditions générales d'utilisation | AniQuiz",
        description: "Conditions générales d'utilisation du site AniQuiz.",
        canonical: 'https://aniquiz.fr/terms',
        noindex: true,
      },
    },
    {
      path: '/privacy',
      component: PrivacyPage,
      meta: {
        title: "Politique de confidentialité | AniQuiz",
        description: "Politique de confidentialité et traitement des données personnelles sur AniQuiz.",
        canonical: 'https://aniquiz.fr/privacy',
        noindex: true,
      },
    },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFound },
  ],
})

router.afterEach((to) => {
  const { title, description, canonical, noindex } = to.meta ?? {}

  if (title) document.title = title

  let metaDesc = document.querySelector('meta[name="description"]')
  if (metaDesc && description) metaDesc.setAttribute('content', description)

  let ogTitle = document.querySelector('meta[property="og:title"]')
  if (ogTitle && title) ogTitle.setAttribute('content', title)

  let ogDesc = document.querySelector('meta[property="og:description"]')
  if (ogDesc && description) ogDesc.setAttribute('content', description)

  let ogUrl = document.querySelector('meta[property="og:url"]')
  if (ogUrl && canonical) ogUrl.setAttribute('content', canonical)

  let canonicalTag = document.querySelector('link[rel="canonical"]')
  if (canonicalTag && canonical) canonicalTag.setAttribute('href', canonical)

  let robotsMeta = document.querySelector('meta[name="robots"]')
  if (robotsMeta) robotsMeta.setAttribute('content', noindex ? 'noindex, nofollow' : 'index, follow')
})

export default router
