<template>
  <div class="legal-page">
    <div class="legal-inner">
      <button class="back-btn" @click="router.back()">← Retour</button>

      <h1>Mentions légales</h1>
      <p class="updated">Dernière mise à jour : juin 2026</p>

      <h2>1. Éditeur du site</h2>
      <p>
        AniQuiz est un projet indépendant et <strong>non lucratif</strong>, développé par des
        passionné·e·s d'anime dans un cadre éducatif et associatif.
      </p>
      <p>
        <strong>Éditeur / responsable de publication :</strong>
        <span class="todo">[à compléter : nom et prénom ou raison sociale]</span><br />
        <strong>Statut :</strong> projet personnel / étudiant sans but lucratif<br />
        <strong>Contact :</strong> <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>
        <span class="todo">[adapter l'adresse si besoin]</span>
      </p>

      <h2>2. Hébergement</h2>
      <p>
        Le code source et les images du service sont hébergés via GitHub. L'application et la
        base de données sont hébergées sur :
        <span class="todo">[à compléter : nom de l'hébergeur, raison sociale et adresse — ex. OVH, Scaleway, etc.]</span>.
      </p>

      <h2>3. Propriété intellectuelle</h2>
      <p>
        La marque « AniQuiz », son logo, sa mascotte (Kora) ainsi que le code source du jeu sont
        la propriété de l'éditeur. Toute reproduction sans autorisation est interdite.
      </p>
      <p>
        En revanche, <strong>les titres d'anime, les visuels, les musiques (openings / endings)
        et tout contenu audio référencés appartiennent à leurs ayants droit respectifs</strong>
        (studios, labels, artistes, éditeurs). AniQuiz n'est affilié, sponsorisé ni approuvé par
        aucun studio d'animation, label ni plateforme de streaming.
      </p>
      <p>
        Les extraits sont proposés à des fins ludiques et culturelles, dans une logique de
        découverte. Les métadonnées proviennent de l'API Jikan (MyAnimeList) et les liens audio
        de sources tierces (AnimeThemes / themes.moe).
      </p>

      <h2>4. Signalement de contenu (procédure de retrait)</h2>
      <p>
        Vous êtes ayant droit et souhaitez qu'un contenu soit retiré ? Écrivez à
        <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a> en précisant le contenu
        concerné et le titre de l'œuvre. Nous nous engageons à retirer dans les meilleurs délais
        tout contenu signalé par son titulaire de droits.
      </p>

      <h2>5. Données personnelles</h2>
      <p>
        Le traitement de vos données est détaillé dans notre
        <RouterLink to="/privacy">Politique de confidentialité</RouterLink>. L'utilisation du
        service est régie par nos <RouterLink to="/terms">Conditions Générales d'Utilisation</RouterLink>.
      </p>

      <h2>6. Contact</h2>
      <p>
        Pour toute question : <a href="mailto:contact@aniquiz.fr">contact@aniquiz.fr</a>.
      </p>

      <p class="note">
        ⚖️ Ce document est un modèle de bonne foi adapté au service ; il ne constitue pas un
        avis juridique. Complétez les champs <span class="todo">[à compléter]</span> et faites-le
        relire avant une mise en production publique.
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
  margin: 32px 0 12px;
}
p {
  line-height: 1.8;
  font-size: 0.95rem;
  margin: 0 0 16px;
}
a { color: #fb923c; }
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
</style>
