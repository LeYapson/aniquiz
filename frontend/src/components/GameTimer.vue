<template>
  <div class="timer-container">
    <div class="timer-bar-bg">
      <div 
        class="timer-bar-fill" 
        :style="{ width: progressPercentage + '%', backgroundColor: barColor }"
      ></div>
    </div>
    <div class="timer-text" :class="{ 'timer-warn': timeLeft <= 5 }">
      ⏱️ Il reste : <strong>{{ timeLeft }}</strong> secondes
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onUnmounted, computed } from 'vue';

// On reçoit la durée dynamique depuis App.vue
const props = defineProps({
  duration: {
    type: Number,
    required: true
  }
});

const timeLeft = ref(props.duration);
let timerInterval = null;

// Calcul du pourcentage pour la barre de progression animée
const progressPercentage = computed(() => {
  return (timeLeft.value / props.duration) * 100;
});

// Changement de couleur de la barre quand le temps presse
const barColor = computed(() => {
  if (timeLeft.value <= 5) return '#ff4757'; // Rouge
  if (timeLeft.value <= 10) return '#ffa500'; // Orange
  return '#2ed573'; // Vert
});

// Logique du compte à rebours
const startCountdown = () => {
  // On nettoie l'ancien intervalle s'il y en a un
  if (timerInterval) clearInterval(timerInterval);
  
  timeLeft.value = props.duration;

  timerInterval = setInterval(() => {
    if (timeLeft.value > 0) {
      timeLeft.value--;
    } else {
      clearInterval(timerInterval); // Temps écoulé, on arrête
    }
  }, 1000);
};

// On surveille la prop duration : chaque fois qu'elle change (nouveau round), on relance
watch(() => props.duration, () => {
  startCountdown();
}, { immediate: true });

// Sécurité : Nettoyage du timer si le composant est retiré de l'écran
onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval);
});
</script>

<style scoped>
.timer-container {
  margin: 15px 0;
  text-align: center;
}
.timer-bar-bg {
  width: 100%;
  height: 12px;
  background-color: #e0e0e0;
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 8px;
}
.timer-bar-fill {
  height: 100%;
  transition: width 1s linear, background-color 0.5s ease;
}
.timer-text {
  font-size: 1.1rem;
  color: #333;
}
.timer-warn {
  color: #ff4757;
  font-weight: bold;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.03); }
  100% { transform: scale(1); }
}
</style>