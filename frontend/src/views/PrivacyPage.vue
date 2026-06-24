<template>
  <div class="legal-page">
    <div class="legal-inner">
      <button class="back-btn" @click="router.back()">← Retour</button>

      <h1>Politique de confidentialité</h1>
      <p class="updated">Dernière mise à jour : juin 2026</p>

      <p>
        Cette politique explique quelles données personnelles AniQuiz (le « Service ») collecte,
        pourquoi, combien de temps elles sont conservées, et quels sont vos droits. Nous
        appliquons une logique de <strong>minimisation</strong> : nous ne collectons que ce qui
        est nécessaire au fonctionnement du jeu.
      </p>

      <h2>1. Responsable du traitement</h2>
      <p>
        Le responsable du traitement est l'éditeur d'AniQuiz
        (<span>yatokishi</span>). Contact :
        <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>.
      </p>

      <h2>2. Données collectées et finalités</h2>
      <table class="data-table">
        <thead>
          <tr><th>Donnée</th><th>Finalité</th><th>Base légale</th></tr>
        </thead>
        <tbody>
          <tr>
            <td>Pseudo, e-mail, mot de passe (haché via bcrypt)</td>
            <td>Création et sécurisation du compte, connexion</td>
            <td>Exécution du contrat</td>
          </tr>
          <tr>
            <td>XP, niveau, scores, historique de parties, scores speedrun</td>
            <td>Progression, profil, classements</td>
            <td>Exécution du contrat</td>
          </tr>
          <tr>
            <td>Liste d'amis, invitations en salon</td>
            <td>Fonctionnalités sociales</td>
            <td>Exécution du contrat</td>
          </tr>
          <tr>
            <td>Cadre d'avatar (cosmétique)</td>
            <td>Personnalisation du profil</td>
            <td>Exécution du contrat</td>
          </tr>
          <tr>
            <td>Jetons et identifiants AniList / MyAnimeList (optionnels)</td>
            <td>Jouer à partir de votre liste d'anime personnelle</td>
            <td>Consentement (liaison volontaire)</td>
          </tr>
          <tr>
            <td>Adresse IP</td>
            <td>Limitation de débit anti-abus, sécurité</td>
            <td>Intérêt légitime</td>
          </tr>
        </tbody>
      </table>
      <p>
        Les <strong>messages de chat</strong> en partie sont diffusés en temps réel et ne sont
        pas conservés de manière durable en base de données.
      </p>

      <h2>3. Cookies et stockage local</h2>
      <p>
        AniQuiz n'utilise <strong>pas de cookies publicitaires ni de traceurs tiers</strong>.
        Un jeton d'authentification (JWT) est enregistré dans le stockage local
        (<em>localStorage</em>) de votre navigateur pour vous garder connecté ; il est supprimé
        à la déconnexion.
      </p>

      <h2>4. Partage des données</h2>
      <p>
        Vos données ne sont <strong>ni vendues ni louées</strong>. Elles peuvent être traitées par :
      </p>
      <ul>
        <li>notre hébergeur, pour faire fonctionner le service ;</li>
        <li>AniList / MyAnimeList, uniquement si vous liez votre compte (selon leurs propres politiques) ;</li>
        <li>l'API Jikan, pour récupérer des métadonnées d'anime (sans donnée personnelle).</li>
      </ul>

      <h2>5. Durée de conservation</h2>
      <ul>
        <li>Données de compte et de jeu : tant que le compte est actif.</li>
        <li>Jetons OAuth : jusqu'à la dissociation du compte tiers ou la suppression du compte.</li>
        <li>Adresse IP (anti-abus) : conservée de manière transitoire, puis supprimée.</li>
        <li>Après suppression du compte, les données associées sont effacées (sauf obligation légale de conservation).</li>
      </ul>

      <h2>6. Sécurité</h2>
      <p>
        Les mots de passe sont hachés (bcrypt) et jamais stockés en clair. Les échanges sont
        chiffrés (HTTPS/TLS) et l'accès aux fonctionnalités sensibles est protégé par jeton.
        Aucune mesure n'étant infaillible, nous ne pouvons garantir une sécurité absolue.
      </p>

      <h2>7. Vos droits (RGPD)</h2>
      <p>
        Vous disposez des droits d'<strong>accès</strong>, de <strong>rectification</strong>,
        d'<strong>effacement</strong>, de <strong>limitation</strong>, de
        <strong>portabilité</strong> et d'<strong>opposition</strong>. Vous pouvez aussi retirer
        votre consentement (dissociation AniList/MAL) à tout moment.
      </p>
      <p>
        Pour exercer ces droits : <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>.
        Vous pouvez également introduire une réclamation auprès de la
        <a href="https://www.cnil.fr" target="_blank" rel="noopener">CNIL</a>.
      </p>

      <h2>8. Mineurs</h2>
      <p>
        Le Service s'adresse aux personnes d'au moins 15 ans. Nous ne collectons pas sciemment de
        données concernant des enfants en dessous de cet âge sans accord parental.
      </p>

      <h2>9. Modifications</h2>
      <p>
        Cette politique peut être mise à jour ; la date en haut de page indique la dernière
        révision. Les changements importants seront signalés via le devlog in-app.
      </p>

      <h2>10. Contact</h2>
      <p>
        Toute question relative à vos données : <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>.
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
em { color: #cbd5e1; font-style: italic; }

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin: 0 0 16px;
  font-size: 0.88rem;
  background: #16213e;
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 10px;
  overflow: hidden;
}
.data-table th {
  background: #0f0f23;
  color: #f97316;
  text-align: left;
  padding: 10px 12px;
  font-size: 0.74rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.data-table td {
  padding: 10px 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  color: #cbd5e1;
  vertical-align: top;
}

.todo {
  background: rgba(251, 191, 36, 0.12);
  color: #fbbf24;
  font-style: italic;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 0.88rem;
}
.note {
  margin-top: 40px;
  padding: 14px 16px;
  background: rgba(249, 115, 22, 0.06);
  border: 1px solid rgba(249, 115, 22, 0.2);
  border-radius: 10px;
  font-size: 0.85rem;
}

@media (max-width: 600px) {
  .data-table { display: block; overflow-x: auto; -webkit-overflow-scrolling: touch; }
}
</style>
