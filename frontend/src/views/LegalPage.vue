<template>
  <div class="legal-page">
    <div class="legal-inner">
      <button class="back-btn" @click="router.back()">← Retour</button>

      <h1>Mentions légales</h1>

      <p>
        <strong>Éditeur :</strong> Projet personnel AniQuiz — contact :
        <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a><br />
        <strong>Hébergement :</strong> GitHub / serveur dédié<br />
        <strong>Directeur de publication :</strong> Équipe AniQuiz
      </p>

      <p>
        Ce site est un projet indépendant sans but lucratif, développé par des passionnés
        d'anime. AniQuiz n'est affilié à aucun studio d'animation ni plateforme de streaming.
      </p>

      <h2>Propriété intellectuelle</h2>
      <p>
        Les noms et œuvres d'anime référencés appartiennent à leurs ayants droit respectifs.
        Les extraits musicaux sont utilisés dans un cadre éducatif et ludique.
      </p>

      <h2>Données personnelles</h2>
      <p>
        Voir notre <RouterLink to="/privacy">Politique de confidentialité</RouterLink>.
      </p>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
const router = useRouter();
</script>

<style scoped>
.legal-page {
  min-height: 100vh;
  background: #0f0f23;
  color: #94a3b8;
  padding: 60px 24px;
}
.legal-inner {
  max-width: 800px;
  margin: 0 auto;
}
.back-btn {
  background: none;
  border: none;
  color: #f97316;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
  padding: 0 0 24px;
}
.back-btn:hover { text-decoration: underline; }

h1 {
  font-size: 2rem;
  font-weight: 800;
  color: #f1f5f9;
  margin: 0 0 24px;
}
h2 {
  font-size: 1.2rem;
  font-weight: 700;
  color: #f97316;
  margin: 32px 0 12px;
}
p {
  line-height: 1.8;
  font-size: 0.95rem;
  margin: 0 0 16px;
}
a { color: #fb923c; }
</style>
