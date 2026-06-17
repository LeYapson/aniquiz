<template>
  <div class="reaction-root" aria-hidden="true">
    <!-- Particules flottantes -->
    <TransitionGroup name="float" tag="div" class="reaction-particles">
      <div
        v-for="p in particles"
        :key="p.id"
        class="reaction-particle"
        :style="{ left: p.x + '%', '--dur': p.dur + 'ms' }"
      >{{ p.emoji }}</div>
    </TransitionGroup>

    <!-- Barre de boutons -->
    <div class="reaction-bar">
      <button
        v-for="e in EMOJIS"
        :key="e"
        class="reaction-btn"
        :disabled="!connected"
        :aria-label="`Réagir ${e}`"
        @click="send(e)"
      >{{ e }}</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

const props = defineProps({
  connected: { type: Boolean, default: false },
});
const emit = defineEmits(['react']);

const EMOJIS = ['🔥', '🤔', '😱', '✅', '😭', '👏'];
const particles = ref([]);
let nextId = 0;

const send = (emoji) => {
  emit('react', emoji);
};

const addParticle = (emoji) => {
  const id = nextId++;
  particles.value.push({ id, emoji, x: 10 + Math.random() * 80, dur: 1800 + Math.random() * 600 });
  setTimeout(() => {
    particles.value = particles.value.filter(p => p.id !== id);
  }, 2500);
};

defineExpose({ addParticle });
</script>

<style scoped>
.reaction-root { position: relative; }

.reaction-particles {
  position: fixed;
  bottom: 80px;
  left: 0;
  right: 0;
  pointer-events: none;
  z-index: 300;
}
.reaction-particle {
  position: absolute;
  font-size: 2rem;
  bottom: 0;
  animation: floatUp var(--dur, 2000ms) ease-out forwards;
}
@keyframes floatUp {
  0%   { transform: translateY(0) scale(1); opacity: 1; }
  80%  { opacity: 0.8; }
  100% { transform: translateY(-220px) scale(1.4); opacity: 0; }
}
.float-enter-active { transition: none; }
.float-leave-active { transition: none; }

.reaction-bar {
  display: flex;
  gap: 6px;
  justify-content: center;
  margin: 12px 0 4px;
  flex-wrap: wrap;
}
.reaction-btn {
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 1.3rem;
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;
  line-height: 1;
}
.reaction-btn:hover:not(:disabled) {
  background: rgba(255,255,255,0.12);
  transform: scale(1.15);
}
.reaction-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.reaction-btn:focus-visible { outline: 2px solid #f97316; outline-offset: 2px; }
</style>
