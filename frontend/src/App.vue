<template>
  <div class="app-container">
    <h1>AniQuiz 🎵</h1>
    
    <!-- ÉTAPE 1 : Si pas connecté, on montre le formulaire -->
    <div v-if="!isConnected">
      <JoinRoom @join="setupWebSocket" />
    </div>

    <!-- ÉTAPE 2 : SINON (on est connecté), on montre l'interface de jeu -->
    <div v-else class="game-layout">
      
      <!-- Barre latérale avec les joueurs -->
      <aside class="sidebar">
        <h3>Joueurs ({{ players.length }})</h3>
        <ul>
          <li v-for="p in players" :key="p">
            {{ p }} <span v-if="p === user">⭐</span>
          </li>
        </ul>
      </aside>

      <!-- Zone principale -->
      <main class="game-area">
        <div class="status-bar">
          <p>Salon : <strong>{{ room }}</strong> | Joueur : <strong>{{ user }}</strong></p>
          <button @click="disconnect" class="btn-quit">Quitter</button>
        </div>

        <div class="quiz-box">
          <p v-if="players.length < 2">En attente d'autres joueurs pour commencer...</p>
          <p v-else>Le salon est prêt ! Le jeu va bientôt commencer.</p>
        </div>
      </main>

    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import JoinRoom from './components/JoinRoom.vue'

const isConnected = ref(false)
const user = ref('')
const room = ref('')
const players = ref([])
let socket = null

const setupWebSocket = ({ username, roomId }) => {
  user.value = username
  room.value = roomId

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${window.location.host}/ws?username=${username}&room=${roomId}`
  
  socket = new WebSocket(url)

  socket.onopen = () => {
    console.log("✅ Connecté au serveur Go !")
    isConnected.value = true
  }

  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      switch (data.type) {
        case 'UserList':
          players.value = data.payload
          break
        case 'NewQuestion':
          console.log("Nouvelle question reçue :", data.payload)
          break
      }
    } catch (err) {
      console.error("Erreur message:", err)
    }
  }

  socket.onclose = () => {
    isConnected.value = false
    players.value = []
  }
}

const disconnect = () => {
  if (socket) socket.close()
}
</script>

<style>.game-layout {
  display: flex;
  gap: 20px;
  max-width: 1000px;
  margin: 20px auto;
  text-align: left;
}

.sidebar {
  width: 200px;
  background: #f4f4f4;
  padding: 15px;
  border-radius: 8px;
}

.game-area {
  flex: 1;
  background: #fff;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #eee;
  padding-bottom: 10px;
  margin-bottom: 20px;
}

.btn-quit {
  background: #ff4757;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}
</style>