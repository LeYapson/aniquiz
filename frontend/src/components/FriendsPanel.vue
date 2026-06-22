<template>
  <div class="friends-panel">
    <!-- Ajouter un ami -->
    <div class="add-row">
      <input
        v-model="newFriend"
        type="text"
        placeholder="Pseudo d'un joueur à ajouter…"
        class="friend-input"
        @keyup.enter="sendRequest"
        :disabled="sending"
      />
      <button @click="sendRequest" :disabled="sending || newFriend.trim().length < 1" class="btn-add-friend">
        {{ sending ? '…' : '➕ Ajouter' }}
      </button>
    </div>

    <!-- Demandes reçues -->
    <div v-if="requests.length" class="friend-block">
      <h4>Demandes reçues ({{ requests.length }})</h4>
      <div v-for="r in requests" :key="r.request_id" class="friend-row request">
        <span class="friend-id">{{ r.username }} <small>niv. {{ r.level }}</small></span>
        <div class="friend-actions">
          <button class="btn-accept" @click="respond(r.request_id, true)">Accepter</button>
          <button class="btn-decline" @click="respond(r.request_id, false)">Refuser</button>
        </div>
      </div>
    </div>

    <!-- Liste d'amis -->
    <div class="friend-block">
      <h4>Mes amis ({{ friends.length }})</h4>
      <div v-if="loading" class="friends-empty">Chargement…</div>
      <div v-else-if="friends.length === 0" class="friends-empty">
        Aucun ami pour l'instant. Ajoute un joueur par son pseudo !
      </div>
      <div v-else>
        <div v-for="f in friends" :key="f.user_id" class="friend-row">
          <span class="friend-id">
            <span class="friend-dot">{{ f.username.charAt(0).toUpperCase() }}</span>
            {{ f.username }} <small>niv. {{ f.level }}</small>
          </span>
          <button class="btn-remove" @click="remove(f)" :title="`Retirer ${f.username}`">✕</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { authStore } from "../authStore";
import { API_URL } from "../config";
import { useToast } from "../composables/useToast";

const toast = useToast();

const friends = ref([]);
const requests = ref([]);
const loading = ref(true);
const newFriend = ref("");
const sending = ref(false);

const loadFriends = async () => {
  try {
    const [fRes, rRes] = await Promise.all([
      fetch(`${API_URL}/api/friends`, { headers: authStore.authHeaders() }),
      fetch(`${API_URL}/api/friends/requests`, { headers: authStore.authHeaders() }),
    ]);
    if (fRes.ok) friends.value = await fRes.json();
    if (rRes.ok) requests.value = await rRes.json();
  } catch (e) {
    console.error("Erreur chargement amis :", e);
  } finally {
    loading.value = false;
  }
};

const sendRequest = async () => {
  const username = newFriend.value.trim();
  if (!username) return;
  sending.value = true;
  try {
    const res = await fetch(`${API_URL}/api/friends/request`, {
      method: "POST",
      headers: authStore.authHeaders(),
      body: JSON.stringify({ username }),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) {
      toast.success(`Demande envoyée à ${username}`);
      newFriend.value = "";
      await loadFriends();
    } else {
      toast.error(data.error || "Impossible d'envoyer la demande");
    }
  } catch {
    toast.error("Erreur réseau");
  } finally {
    sending.value = false;
  }
};

const respond = async (requestId, accept) => {
  try {
    const res = await fetch(`${API_URL}/api/friends/respond`, {
      method: "POST",
      headers: authStore.authHeaders(),
      body: JSON.stringify({ request_id: requestId, accept }),
    });
    if (res.ok) {
      toast.success(accept ? "Ami ajouté !" : "Demande refusée");
      await loadFriends();
    } else {
      toast.error("Action impossible");
    }
  } catch {
    toast.error("Erreur réseau");
  }
};

const remove = async (friend) => {
  try {
    const res = await fetch(`${API_URL}/api/friends/${friend.user_id}`, {
      method: "DELETE",
      headers: authStore.authHeaders(),
    });
    if (res.ok) {
      friends.value = friends.value.filter((f) => f.user_id !== friend.user_id);
    } else {
      toast.error("Impossible de retirer cet ami");
    }
  } catch {
    toast.error("Erreur réseau");
  }
};

onMounted(loadFriends);
</script>

<style scoped>
.friends-panel { display: flex; flex-direction: column; gap: 16px; }

.add-row { display: flex; gap: 10px; }
.friend-input {
  flex: 1; padding: 9px 12px;
  background: #0f0f23; border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9; border-radius: 7px; font-size: 0.9rem; outline: none;
  transition: border-color 0.15s;
}
.friend-input:focus { border-color: #f97316; }
.friend-input::placeholder { color: #475569; }
.friend-input:disabled { opacity: 0.5; }
.btn-add-friend {
  background: #f97316; color: white; border: none;
  padding: 9px 16px; border-radius: 7px; cursor: pointer;
  font-weight: 700; white-space: nowrap; transition: opacity 0.15s;
}
.btn-add-friend:hover { opacity: 0.85; }
.btn-add-friend:disabled { background: #334155; color: #64748b; cursor: not-allowed; }

.friend-block h4 {
  font-size: 0.78rem; color: #94a3b8; font-weight: 700;
  margin: 0 0 8px; text-transform: uppercase; letter-spacing: 0.05em;
}
.friends-empty { color: #64748b; font-style: italic; font-size: 0.88rem; padding: 6px 0; }

.friend-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 9px 12px; border-radius: 9px; margin-bottom: 6px;
  background: #16213e; border: 1px solid rgba(255,255,255,0.07);
}
.friend-row.request { border-color: rgba(249,115,22,0.3); }
.friend-id { display: flex; align-items: center; gap: 9px; color: #f1f5f9; font-size: 0.9rem; font-weight: 600; }
.friend-id small { color: #64748b; font-weight: 500; }
.friend-dot {
  width: 26px; height: 26px; border-radius: 50%;
  background: linear-gradient(135deg, #f97316, #ea580c); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.8rem; font-weight: 700; flex-shrink: 0;
}
.friend-actions { display: flex; gap: 6px; }
.btn-accept, .btn-decline, .btn-remove {
  border: none; border-radius: 6px; cursor: pointer;
  font-weight: 700; font-size: 0.8rem; padding: 5px 10px; transition: opacity 0.15s;
}
.btn-accept { background: #22c55e; color: #fff; }
.btn-decline { background: #475569; color: #e2e8f0; }
.btn-remove { background: transparent; color: #64748b; font-size: 0.9rem; padding: 5px 8px; }
.btn-remove:hover { color: #ef4444; }
.btn-accept:hover, .btn-decline:hover { opacity: 0.85; }
</style>
