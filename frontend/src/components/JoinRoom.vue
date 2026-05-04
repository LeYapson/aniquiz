<template>
  <div class="join-container">
    <h2>Rejoindre une partie</h2>
    <div class="form-group">
      <input v-model="username" placeholder="Ton pseudo" />
      <input v-model="roomId" placeholder="Nom du salon (ex: general)" />
      <button @click="connect">Rejoindre</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const username = ref('')
const roomId = ref('')

const emit = defineEmits(['join'])

const connect = () => {
  if (username.value && roomId.value) {
    emit('join', { username: username.value, roomId: roomId.value })
  }
}
</script>

<style scoped>
.join-container {
  border: 1px solid #ccc;
  padding: 2rem;
  border-radius: 8px;
  max-width: 300px;
  margin: 2rem auto;
}
input { display: block; width: 100%; margin-bottom: 1rem; padding: 0.5rem; }
button { width: 100%; background: #42b983; color: white; border: none; padding: 0.5rem; cursor: pointer; }
</style>