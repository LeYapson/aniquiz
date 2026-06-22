import { reactive, readonly } from 'vue';
import { authStore } from '../authStore';
import { API_URL } from '../config';

// État partagé (singleton) des notifications : demandes d'ami reçues +
// invitations à rejoindre un salon. Alimenté par un sondage périodique.
const state = reactive({ friendRequests: [], roomInvites: [] });
let timer = null;

const authFetch = (url, opts = {}) =>
  fetch(url, { ...opts, headers: { ...authStore.authHeaders(), ...(opts.headers || {}) } });

async function refresh() {
  if (!authStore.user) {
    state.friendRequests = [];
    state.roomInvites = [];
    return;
  }
  try {
    const [fr, inv] = await Promise.all([
      authFetch(`${API_URL}/api/friends/requests`),
      authFetch(`${API_URL}/api/invites`),
    ]);
    if (fr.ok) state.friendRequests = await fr.json();
    if (inv.ok) state.roomInvites = await inv.json();
  } catch {
    /* silencieux : on réessaiera au prochain tick */
  }
}

function start() {
  if (timer) return;
  refresh();
  timer = setInterval(refresh, 15000);
}

function stop() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

async function respondFriend(requestId, accept) {
  try {
    await authFetch(`${API_URL}/api/friends/respond`, {
      method: 'POST',
      body: JSON.stringify({ request_id: requestId, accept }),
    });
  } catch { /* ignore */ }
  state.friendRequests = state.friendRequests.filter((r) => r.request_id !== requestId);
}

async function removeInvite(inviteId) {
  try {
    await authFetch(`${API_URL}/api/invites/${inviteId}`, { method: 'DELETE' });
  } catch { /* ignore */ }
  state.roomInvites = state.roomInvites.filter((i) => i.id !== inviteId);
}

export function useNotifications() {
  return { state: readonly(state), refresh, start, stop, respondFriend, removeInvite };
}
