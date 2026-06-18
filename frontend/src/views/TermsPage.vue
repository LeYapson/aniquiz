<template>
  <div class="legal-page">
    <div class="legal-inner">
      <button class="back-btn" @click="router.back()">← Retour</button>

      <h1>Conditions Générales d'Utilisation</h1>
      <p class="updated">Dernière mise à jour : juin 2026</p>

      <h2>1. Acceptation</h2>
      <p>En utilisant AniQuiz, vous acceptez ces CGU.</p>

      <h2>2. Compte utilisateur</h2>
      <p>
        Vous êtes responsable de la confidentialité de vos identifiants. Tout comportement
        abusif (triche, harcèlement) entraîne la suspension du compte.
      </p>

      <h2>3. Contenu</h2>
      <p>
        AniQuiz est gratuit. Aucun achat n'est nécessaire pour accéder aux fonctionnalités
        principales. Les soutiens via Ko-fi sont volontaires et ne donnent pas accès à des
        avantages compétitifs.
      </p>

      <h2>4. Disponibilité</h2>
      <p>
        Le service est fourni « tel quel ». Nous ne garantissons pas une disponibilité de 100 %.
      </p>

      <h2>5. Modifications</h2>
      <p>
        Ces CGU peuvent être modifiées à tout moment. Les changements majeurs seront annoncés
        via le devlog.
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
  margin: 0 0 8px;
}
.updated {
  font-size: 0.82rem;
  color: #475569;
  margin: 0 0 24px;
}
h2 {
  font-size: 1.2rem;
  font-weight: 700;
  color: #f97316;
  margin: 28px 0 10px;
}
p {
  line-height: 1.8;
  font-size: 0.95rem;
  margin: 0 0 12px;
}
</style>
