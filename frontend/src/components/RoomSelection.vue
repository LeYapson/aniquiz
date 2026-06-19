<template>
  <div class="room-selection">

    <!-- ── Home view: single Play button ── -->
    <div v-if="!showMultiPanel" class="play-home">
      <BaseButton variant="primary" size="xl" pill @click="openModal">
        Jouer
      </BaseButton>
    </div>

    <!-- ── Multi-player panel ── -->
    <div v-else class="multi-panel">
      <div class="multi-header">
        <h2>Salons disponibles</h2>
        <BaseButton variant="ghost" size="sm" @click="closeMulti">← Retour</BaseButton>
      </div>

      <div class="room-actions">
        <BaseButton
          :variant="isCreating ? 'ghost' : 'secondary'"
          size="sm"
          @click="isCreating = !isCreating"
        >
          {{ isCreating ? '← Voir les salons' : 'Créer un salon' }}
        </BaseButton>
      </div>

      <hr />

      <!-- Create room form -->
      <div v-if="isCreating" class="form-container">
        <h3>Créer un salon</h3>

        <BaseInput
          v-model="config.room_id"
          label="Nom du salon"
          placeholder="Ex : Salon de Théau"
          autocomplete="off"
        />

        <label class="checkbox-label">
          <input v-model="config.is_private" type="checkbox" class="checkbox-native" />
          Salon privé (avec mot de passe)
        </label>

        <BaseInput
          v-if="config.is_private"
          v-model="config.password"
          type="password"
          label="Mot de passe"
          placeholder="Mot de passe requis"
          autocomplete="new-password"
        />

        <p class="settings-hint">
          Les paramètres de jeu (rounds, durée, filtres) sont configurables dans le lobby.
        </p>

        <BaseButton
          variant="primary"
          full
          :disabled="!config.room_id.trim()"
          :loading="creating"
          @click="submitCreate"
        >
          Créer et Rejoindre
        </BaseButton>
      </div>

      <!-- Room list -->
      <div v-else class="rooms-list-container">
        <div class="rooms-list-header">
          <BaseButton
            variant="blue"
            size="sm"
            :loading="loadingRooms"
            @click="fetchRooms"
          >
            Actualiser
          </BaseButton>
        </div>

        <!-- Loading skeletons -->
        <div v-if="loadingRooms" class="rooms-grid">
          <BaseCard v-for="i in 3" :key="i" accent padding="md" class="skeleton-card">
            <SkeletonLoader variant="card" />
          </BaseCard>
        </div>

        <!-- Empty state -->
        <EmptyState
          v-else-if="rooms.length === 0"
          icon="🎮"
          title="Aucun salon actif"
          description="Personne n'a encore créé de salon. Soyez le premier !"
          action-label="Créer un salon"
          action-variant="primary"
          @action="isCreating = true"
        />

        <!-- Room cards -->
        <div v-else class="rooms-grid">
          <BaseCard
            v-for="room in rooms"
            :key="room.id"
            accent
            hoverable
            padding="md"
            class="room-card"
          >
            <div class="room-info">
              <div class="room-name-row">
                <strong>{{ room.id }}</strong>
                <BaseBadge :variant="room.is_private ? 'private' : 'public'">
                  {{ room.is_private ? 'Privé' : 'Public' }}
                </BaseBadge>
              </div>
              <span class="room-meta">{{ room.players_count }} joueur{{ room.players_count !== 1 ? 's' : '' }}</span>
              <span class="room-meta">{{ room.max_rounds }} rounds</span>
            </div>

            <!-- Password prompt for private rooms -->
            <div
              v-if="selectedRoomId === room.id && room.is_private"
              class="password-prompt"
            >
              <BaseInput
                v-model="inputPassword"
                type="password"
                placeholder="Code d'accès"
                autocomplete="off"
                @keydown.enter="submitJoin(room)"
              />
              <BaseButton variant="primary" size="sm" @click="submitJoin(room)">
                Valider
              </BaseButton>
            </div>

            <BaseButton
              v-else
              :variant="room.state !== 'LOBBY' ? 'blue' : 'primary'"
              size="sm"
              full
              @click="handleRoomClick(room)"
            >
              {{ room.state === 'LOBBY' ? 'Rejoindre' : 'Regarder' }}
            </BaseButton>
          </BaseCard>
        </div>
      </div>
    </div>

    <!-- ── Mode selection modal ── -->
    <BaseModal
      :open="showModeModal"
      title="Choisis ton mode"
      size="xl"
      @update:open="showModeModal = $event"
    >
      <div class="mode-cards">
        <button
          v-for="mode in modes"
          :key="mode.key"
          class="mode-card-btn"
          :class="{ 'mode-card-btn--soon': !mode.available }"
          :style="{ backgroundImage: `url(${mode.img}), ${mode.gradient}` }"
          :disabled="!mode.available"
          @click="selectMode(mode)"
        >
          <span class="mode-card-overlay" aria-hidden="true" />
          <span class="mode-card-content">
            <span class="mode-card-title">{{ mode.title }}</span>
            <span class="mode-card-desc">{{ mode.desc }}</span>
          </span>
          <BaseBadge v-if="!mode.available" variant="warning" size="sm" class="mode-card-soon">
            Bientôt
          </BaseBadge>
        </button>
      </div>
    </BaseModal>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { BaseButton, BaseInput, BaseBadge, BaseCard, BaseModal, SkeletonLoader, EmptyState } from './ui/index.js'
import { useToast } from '../composables/useToast'
import { authStore } from '../authStore'
import { API_URL } from '../config'

const router = useRouter()
const toast  = useToast()

const emit = defineEmits(['room-created', 'room-joined', 'panel-changed'])

/* ── State ── */
const rooms         = ref([])
const isCreating    = ref(false)
const selectedRoomId = ref(null)
const inputPassword = ref('')
const loadingRooms  = ref(false)
const creating      = ref(false)
const showModeModal = ref(false)
const showMultiPanel = ref(false)

const config = ref({ room_id: '', is_private: false, password: '' })

const modes = [
  { key: 'solo',    title: 'Solo',      desc: "Entraîne-toi à ton rythme",     img: '/kora-solo.png',    gradient: 'linear-gradient(160deg, #f97316, #b45309)', available: true  },
  { key: 'speedrun',title: 'Speed Run', desc: '5 min · Enchaîne les animes',   img: '/kora-speedrun.png',gradient: 'linear-gradient(160deg, #ef4444, #7f1d1d)', available: true  },
  { key: 'multi',   title: 'Multi',     desc: "Affronte d'autres joueurs",     img: '/kora-multi.png',   gradient: 'linear-gradient(160deg, #3b82f6, #1e3a8a)', available: true  },
  { key: 'classe',  title: 'Classé',    desc: 'Grimpe les divisions',          img: '/kora-ranked.png',  gradient: 'linear-gradient(160deg, #a855f7, #581c87)', available: false },
]

/* ── Modal ── */
const openModal  = () => { showModeModal.value = true }

const selectMode = (mode) => {
  if (!mode.available) return
  showModeModal.value = false
  if (mode.key === 'solo') {
    startSolo()
  } else if (mode.key === 'speedrun') {
    router.push('/speedrun')
  } else if (mode.key === 'multi') {
    showMultiPanel.value = true
    emit('panel-changed', true)
    fetchRooms()
  }
}

const closeMulti = () => {
  showMultiPanel.value = false
  isCreating.value     = false
  emit('panel-changed', false)
}

/* ── API ── */
const fetchRooms = async () => {
  loadingRooms.value = true
  try {
    const res = await fetch(`${API_URL}/rooms`)
    if (res.ok) {
      rooms.value = await res.json() || []
    } else {
      toast.error('Impossible de charger les salons.', { title: 'Erreur réseau' })
    }
  } catch {
    toast.error('Connexion au serveur échouée.', { title: 'Erreur réseau' })
  } finally {
    loadingRooms.value = false
  }
}

const submitCreate = async () => {
  creating.value = true
  try {
    const res = await fetch(`${API_URL}/rooms`, {
      method:  'POST',
      headers: authStore.authHeaders(),
      body:    JSON.stringify(config.value),
    })
    if (res.ok) {
      const data = await res.json()
      emit('room-created', {
        room_id:    data.room_id,
        creator_id: data.creator_id,
        password:   config.value.password,
        isCreator:  true,
      })
    } else {
      const err = await res.json()
      toast.error(err.error || 'Création du salon impossible.', { title: 'Erreur' })
    }
  } catch {
    toast.error('Connexion au serveur échouée.', { title: 'Erreur réseau' })
  } finally {
    creating.value = false
  }
}

const handleRoomClick = (room) => {
  if (room.is_private) {
    selectedRoomId.value = room.id
    inputPassword.value  = ''
  } else {
    submitJoin(room)
  }
}

const submitJoin = (room) => {
  emit('room-joined', {
    room_id:  room.id,
    password: room.is_private ? inputPassword.value : '',
  })
}

const startSolo = async () => {
  const soloId = `solo-${authStore.user?.username ?? 'player'}-${Date.now()}`
  try {
    const res = await fetch(`${API_URL}/rooms`, {
      method:  'POST',
      headers: authStore.authHeaders(),
      body:    JSON.stringify({ room_id: soloId, is_private: true, is_solo: true, password: '' }),
    })
    if (res.ok) {
      const data = await res.json()
      emit('room-created', { room_id: data.room_id, creator_id: data.creator_id, password: '', isCreator: true })
    } else {
      toast.error('Impossible de créer une partie solo.', { title: 'Erreur' })
    }
  } catch {
    toast.error('Connexion au serveur échouée.', { title: 'Erreur réseau' })
  }
}

onMounted(fetchRooms)
</script>

<style scoped>
.room-selection {
  max-width: 860px;
  margin: 0 auto;
  padding: 40px 24px;
  color: var(--text);
}
.room-selection h2 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 20px;
  color: #f1f5f9;
}

hr { border: none; border-top: 1px solid var(--border); margin: 16px 0; }

/* ── Home play button ── */
.play-home {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

/* ── Multi header ── */
.multi-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 0;
}
.multi-header h2 { margin-bottom: 0; }

.room-actions { margin-top: 14px; }

/* ── Create form ── */
.form-container {
  background: var(--navy-3);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.form-container h3 { color: #f1f5f9; margin: 0 0 2px; }

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.875rem;
  color: var(--text-dim);
  cursor: pointer;
}
.checkbox-native {
  width: 16px;
  height: 16px;
  accent-color: var(--orange);
  cursor: pointer;
  flex-shrink: 0;
}

.settings-hint {
  color: #64748b;
  font-size: 0.8rem;
  font-style: italic;
  margin: -4px 0 0;
}

/* ── Room list ── */
.rooms-list-header { margin-bottom: 14px; }

.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 14px;
  margin-top: 4px;
}

.skeleton-card { min-height: 110px; }

/* ── Room card contents ── */
.room-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}
.room-name-row strong { color: #f1f5f9; font-size: 0.95rem; }

.room-meta {
  display: block;
  font-size: 0.8rem;
  color: var(--text-dim);
  margin-bottom: 2px;
}

.room-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* ── Password prompt ── */
.password-prompt {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}
.password-prompt > :first-child { flex: 1; }

/* ── Mode cards inside BaseModal body ── */
.mode-cards {
  display: flex;
  gap: 14px;
  width: 100%;
}
.mode-card-btn {
  position: relative;
  flex: 1;
  height: 320px;
  padding: 0;
  border: none;
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  display: flex;
  align-items: flex-end;
  background-color: var(--navy-3);
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}
.mode-card-btn:hover:not(:disabled) {
  transform: translateY(-6px);
  box-shadow: 0 14px 34px rgba(0,0,0,0.55);
}
.mode-card-btn:disabled { cursor: not-allowed; filter: grayscale(1) brightness(0.7); }
.mode-card-btn:focus-visible { outline: 2px solid var(--orange); outline-offset: 2px; }

.mode-card-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.85) 0%, rgba(0,0,0,0.2) 45%, transparent 100%);
  pointer-events: none;
}
.mode-card-content {
  position: relative;
  z-index: 1;
  width: 100%;
  padding: 18px 20px;
  text-align: left;
}
.mode-card-title { display: block; font-size: 1.35rem; font-weight: 800; color: #fff; }
.mode-card-desc  { display: block; margin-top: 3px; font-size: 0.83rem; font-weight: 500; color: rgba(255,255,255,0.82); }

.mode-card-soon {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 2;
}

/* ── Responsive ── */
@media (max-width: 700px) {
  .mode-cards { flex-wrap: wrap; }
  .mode-card-btn { flex: 1 1 calc(50% - 7px); height: 180px; }
}
@media (max-width: 480px) {
  .room-selection { padding: 20px 14px; }
  .mode-card-btn { flex: 1 1 100%; height: 140px; }
}
</style>
