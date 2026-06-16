<template>
  <div class="not-found">
    <img :src="koraLost" alt="Kora perdue" class="nf-mascot" />
    <h1 class="nf-code">404</h1>
    <p class="nf-title">Cette page n'existe pas...</p>
    <p class="nf-sub">Kora a cherché partout, mais elle ne trouve rien ici !</p>
    <RouterLink to="/" class="nf-btn">Retour à l'accueil</RouterLink>
  </div>
</template>

<script setup>
const koraLost = '/kora_lost.png';
</script>

<style scoped>
.not-found {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  background: #0f0f23;
  padding: 40px 24px;
  text-align: center;
}

.nf-mascot {
  width: 220px;
  max-width: 60vw;
  filter: drop-shadow(0 8px 24px rgba(249, 115, 22, 0.25));
  animation: float 4s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50%       { transform: translateY(-12px); }
}

.nf-code {
  font-size: 6rem;
  font-weight: 900;
  line-height: 1;
  background: linear-gradient(135deg, #f97316, #ea580c);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.nf-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.nf-sub {
  color: #94a3b8;
  font-size: 1rem;
  margin: 0;
  max-width: 360px;
}

.nf-btn {
  margin-top: 8px;
  display: inline-block;
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: #fff;
  font-weight: 700;
  font-size: 1rem;
  padding: 12px 32px;
  border-radius: 8px;
  text-decoration: none;
  box-shadow: 0 4px 14px rgba(249, 115, 22, 0.35);
  transition: transform 0.15s, box-shadow 0.15s;
}

.nf-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(249, 115, 22, 0.45);
  text-decoration: none;
}

.nf-btn:focus-visible {
  outline: 2px solid #f97316;
  outline-offset: 3px;
}
</style>
