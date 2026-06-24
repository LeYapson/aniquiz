<template>
  <div class="legal-page">
    <div class="legal-inner">
      <button class="back-btn" @click="router.back()">← Retour</button>

      <h1>Conditions Générales d'Utilisation</h1>
      <p class="updated">Dernière mise à jour : juin 2026</p>

      <h2>1. Objet et acceptation</h2>
      <p>
        Les présentes Conditions Générales d'Utilisation (« CGU ») régissent l'accès et
        l'utilisation du jeu en ligne AniQuiz (le « Service »). En créant un compte ou en
        utilisant le Service, vous reconnaissez avoir lu et accepté l'intégralité des CGU. Si
        vous n'êtes pas d'accord, n'utilisez pas le Service.
      </p>

      <h2>2. Description du Service</h2>
      <p>
        AniQuiz est un blind-test musical d'anime : les joueurs écoutent des extraits de
        génériques (openings / endings) et tentent d'identifier l'anime, le titre ou l'artiste,
        en solo, en multijoueur ou en mode speedrun. Le Service est fourni gratuitement.
      </p>

      <h2>3. Accès et compte</h2>
      <ul>
        <li>La création d'un compte requiert un pseudo, une adresse e-mail et un mot de passe.</li>
        <li>Vous vous engagez à fournir des informations exactes et à garder vos identifiants confidentiels.</li>
        <li>Vous êtes responsable de toute activité réalisée depuis votre compte.</li>
        <li>Un seul compte par personne ; les comptes multiples destinés à fausser le classement sont interdits.</li>
      </ul>

      <h2>4. Âge minimum</h2>
      <p>
        Le Service est destiné aux personnes âgées d'au moins 15 ans. En dessous de cet âge,
        l'utilisation nécessite l'accord d'un titulaire de l'autorité parentale.
      </p>

      <h2>5. Règles de conduite</h2>
      <p>Sont notamment interdits :</p>
      <ul>
        <li>la triche, l'automatisation des réponses, les bots, ou tout outil de reconnaissance audio ;</li>
        <li>le harcèlement, les propos haineux, discriminatoires ou illicites dans le chat ;</li>
        <li>les pseudos offensants, trompeurs ou usurpant l'identité d'autrui ;</li>
        <li>toute tentative de perturber le Service (spam, surcharge, exploitation de failles) ;</li>
        <li>l'extraction massive (scraping) ou la réutilisation non autorisée des contenus.</li>
      </ul>

      <h2>6. Contenu publié par les utilisateurs</h2>
      <p>
        Vous restez responsable des contenus que vous diffusez (messages de chat, pseudo,
        réactions). Nous pouvons modérer, masquer ou supprimer tout contenu contraire aux CGU et
        suspendre les comptes concernés, sans préavis en cas d'abus grave.
      </p>

      <h2>7. Propriété intellectuelle</h2>
      <p>
        La marque, le logo, la mascotte et le code d'AniQuiz appartiennent à l'éditeur. Les
        anime, visuels et musiques appartiennent à leurs ayants droit ; voir les
        <RouterLink to="/legal">Mentions légales</RouterLink> pour la procédure de signalement.
      </p>

      <h2>8. Services tiers</h2>
      <p>
        Vous pouvez lier votre compte AniList ou MyAnimeList pour jouer à partir de votre liste
        personnelle. Ces services sont régis par leurs propres conditions et politiques de
        confidentialité ; AniQuiz n'en est pas responsable.
      </p>

      <h2>9. Gratuité et soutien</h2>
      <p>
        Le Service est gratuit et aucun achat n'est requis pour accéder aux fonctionnalités
        principales. D'éventuels éléments cosmétiques ou un soutien volontaire (don) pourront
        être proposés à l'avenir : ils ne confèreront <strong>aucun avantage compétitif</strong>
        (pas de « pay-to-win »).
      </p>

      <h2>10. Disponibilité et garanties</h2>
      <p>
        Le Service est fourni « en l'état » et « selon disponibilité », sans garantie de
        continuité, d'absence d'erreurs ni d'adéquation à un usage particulier. Il s'agit d'un
        projet en développement : des interruptions, pertes de données ou modifications de
        fonctionnalités peuvent survenir.
      </p>

      <h2>11. Limitation de responsabilité</h2>
      <p>
        Dans les limites permises par la loi, l'éditeur ne saurait être tenu responsable des
        dommages indirects résultant de l'utilisation ou de l'impossibilité d'utiliser le
        Service.
      </p>

      <h2>12. Suspension et résiliation</h2>
      <p>
        Vous pouvez supprimer votre compte à tout moment (voir la
        <RouterLink to="/privacy">Politique de confidentialité</RouterLink>). Nous pouvons
        suspendre ou clôturer un compte en cas de violation des CGU.
      </p>

      <h2>13. Modification des CGU</h2>
      <p>
        Ces CGU peuvent évoluer. Les changements importants seront annoncés via le devlog
        in-app. L'utilisation continue du Service après mise à jour vaut acceptation.
      </p>

      <h2>14. Droit applicable</h2>
      <p>
        Les présentes CGU sont régies par le droit français. En cas de litige, une solution
        amiable sera recherchée en priorité ; à défaut, les tribunaux français seront
        compétents. Conformément à la réglementation, vous pouvez recourir gratuitement à un
        médiateur de la consommation.
      </p>

      <h2>15. Contact</h2>
      <p>
        <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>
      </p>

      <p class="note">
        ⚖️ Modèle de bonne foi adapté au Service, ne constituant pas un avis juridique.
        À faire relire avant une mise en production publique.
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
.note {
  margin-top: 40px;
  padding: 14px 16px;
  background: rgba(249, 115, 22, 0.06);
  border: 1px solid rgba(249, 115, 22, 0.2);
  border-radius: 10px;
  font-size: 0.85rem;
}
</style>
