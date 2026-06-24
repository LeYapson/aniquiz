<template>
  <div class="notif" ref="root">
    <button
      class="notif-btn"
      :class="{ 'has-items': count > 0 }"
      @click="open = !open"
      :aria-label="`Notifications (${count})`"
      :aria-expanded="open"
    >
      🔔
      <span v-if="count > 0" class="notif-badge">{{ count > 9 ? '9+' : count }}</span>
    </button>

    <div v-if="open" class="notif-panel" role="menu">
      <p v-if="count === 0" class="notif-empty">Aucune notification</p>

      <template v-if="state.friendRequests.length">
        <h4>Demandes d'ami</h4>
        <div v-for="r in state.friendRequests" :key="`fr-${r.request_id}`" class="notif-row">
          <span class="notif-text">👤 <strong>{{ r.username }}</strong> <small>niv. {{ r.level }}</small></span>
          <span class="notif-actions">
            <button class="btn-ok" @click="respondFriend(r.request_id, true)" title="Accepter">✓</button>
            <button class="btn-no" @click="respondFriend(r.request_id, false)" title="Refuser">✕</button>
          </span>
        </div>
      </template>

      <template v-if="visibleInvites.length">
        <h4>Invitations à jouer</h4>
        <div v-for="inv in visibleInvites" :key="`inv-${inv.id}`" class="notif-row">
          <span class="notif-text">🎮 <strong>{{ inv.from_username }}</strong></span>
          <span class="notif-actions">
            <button class="btn-join" @click="join(inv)">Rejoindre</button>
            <button class="btn-no" @click="removeInvite(inv.id)" title="Ignorer">✕</button>
          </span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useNotifications } from '../composables/useNotifications';

// inGame : on masque les invitations de salon en cours de partie (rejoindre
// un autre salon depuis une partie n'a pas de sens et complique la connexion).
const props = defineProps({ inGame: { type: Boolean, default: false } });
const emit = defineEmits(['join']);

const { state, start, stop, respondFriend, removeInvite } = useNotifications();

const open = ref(false);
const root = ref(null);

const visibleInvites = computed(() => (props.inGame ? [] : state.roomInvites));
const count = computed(() => state.friendRequests.length + visibleInvites.value.length);

const join = (inv) => {
  emit('join', inv);
  removeInvite(inv.id);
  open.value = false;
};

const onDocClick = (e) => {
  if (root.value && !root.value.contains(e.target)) open.value = false;
};

onMounted(() => {
  start();
  document.addEventListener('click', onDocClick);
});
onUnmounted(() => {
  stop();
  document.removeEventListener('click', onDocClick);
});
</script>

<style scoped>
.notif { position: relative; }
.notif-btn {
  position: relative;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 6px;
  border-radius: 8px;
  line-height: 1;
  transition: background 0.15s;
}
.notif-btn:hover { background: rgba(255, 255, 255, 0.06); }
.notif-btn.has-items { filter: drop-shadow(0 0 4px rgba(249, 115, 22, 0.6)); }
.notif-badge {
  position: absolute;
  top: 0;
  right: 0;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #ef4444;
  color: #fff;
  font-size: 0.62rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notif-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 280px;
  max-height: 380px;
  overflow-y: auto;
  background: #16213e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  padding: 10px;
  z-index: 300;
}
.notif-empty { color: #64748b; font-style: italic; font-size: 0.85rem; text-align: center; padding: 14px 0; margin: 0; }
.notif-panel h4 {
  font-size: 0.7rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 8px 4px 6px;
}
.notif-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 8px;
  border-radius: 8px;
}
.notif-row:hover { background: rgba(255, 255, 255, 0.04); }
.notif-text { color: #f1f5f9; font-size: 0.85rem; }
.notif-text small { color: #64748b; }
.notif-actions { display: flex; gap: 5px; flex-shrink: 0; }
.notif-actions button {
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 700;
  font-size: 0.78rem;
  padding: 4px 8px;
}
.btn-ok { background: #22c55e; color: #fff; }
.btn-no { background: #475569; color: #e2e8f0; }
.btn-join { background: #f97316; color: #fff; }
.notif-actions button:hover { opacity: 0.85; }
</style>
