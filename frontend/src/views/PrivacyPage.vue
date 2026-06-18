<template>
  <div class="legal-page">
    <div class="legal-inner">
      <button class="back-btn" @click="router.back()">← Retour</button>

      <h1>Politique de confidentialité</h1>
      <p class="updated">Dernière mise à jour : juin 2026</p>

      <h2>Données collectées</h2>
      <ul>
        <li>Pseudo et mot de passe hashé (bcrypt) pour l'authentification</li>
        <li>Scores et historique de parties pour le classement et le profil</li>
        <li>Tokens OAuth AniList/MAL (optionnels, stockés de manière sécurisée)</li>
      </ul>

      <h2>Utilisation</h2>
      <p>
        Ces données servent uniquement au fonctionnement du jeu. Elles ne sont ni vendues ni
        partagées avec des tiers.
      </p>

      <h2>Durée de conservation</h2>
      <p>
        Les données sont conservées tant que le compte est actif. Suppression sur demande à
        <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>.
      </p>

      <h2>Vos droits (RGPD)</h2>
      <p>
        Accès, rectification, suppression : contactez-nous à
        <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>.
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
ul {
  line-height: 1.8;
  font-size: 0.95rem;
  padding-left: 22px;
  margin: 0 0 12px;
}
a { color: #fb923c; }
</style>
