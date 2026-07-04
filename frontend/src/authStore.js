import { reactive } from 'vue';

// Décode l'échéance (claim `exp`, en secondes) d'un JWT sans vérifier la
// signature — la validation forte reste côté serveur. Sert uniquement à éviter
// d'afficher une session déjà morte.
function tokenExpired(token) {
  if (!token) return true;
  try {
    const payload = token.split('.')[1];
    const claims = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
    if (typeof claims.exp !== 'number') return false; // pas d'exp lisible → on laisse le serveur trancher
    // Marge de 5 s pour absorber une horloge locale légèrement décalée.
    return Date.now() >= claims.exp * 1000 - 5000;
  } catch {
    return false; // token illisible : on ne force pas la déconnexion, le 401 s'en chargera
  }
}

const stored = JSON.parse(localStorage.getItem('auth')) || {};

// Session expirée pendant l'absence de l'utilisateur : on repart proprement
// déconnecté plutôt que d'afficher une session fantôme (le token est mort côté
// serveur, tous les appels API renverraient 401).
let expiredOnLoad = false;
if (stored.token && tokenExpired(stored.token)) {
  expiredOnLoad = true;
  localStorage.removeItem('auth');
  localStorage.removeItem('user');
  stored.user = null;
  stored.token = null;
}

// consumeSessionExpired : renvoie true une seule fois si la session stockée était
// expirée au chargement (permet à l'app d'afficher « session expirée »).
export function consumeSessionExpired() {
  const v = expiredOnLoad;
  expiredOnLoad = false;
  return v;
}

export const authStore = reactive({
  user: stored.user || null,
  token: stored.token || null,

  // Méthodes en arrow + référence par nom (authStore) plutôt que `this` :
  // elles restent fiables même passées par référence à un @event (ex. @logout="authStore.logout").
  setUser: (userData, token) => {
    authStore.user = userData;
    authStore.token = token;
    localStorage.setItem('auth', JSON.stringify({ user: userData, token }));
  },

  logout: () => {
    authStore.user = null;
    authStore.token = null;
    localStorage.removeItem('auth');
    localStorage.removeItem('user');
  },

  get isAuthenticated() {
    return this.user !== null && this.token !== null && !tokenExpired(this.token);
  },

  // Vérification ponctuelle (ex. au montage de l'app ou avant un appel sensible).
  isTokenExpired: () => tokenExpired(authStore.token),

  authHeaders: () => ({
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authStore.token}`,
  }),
});

// apiFetch : fetch authentifié. Injecte le header Authorization et, sur 401
// (token expiré/invalide/révoqué), déconnecte proprement pour ne pas laisser
// une session fantôme. La Response est renvoyée telle quelle pour que l'appelant
// gère ses autres cas. Ne pas utiliser pour un upload multipart (le header
// Content-Type: application/json casserait le FormData) — utiliser fetch direct.
export async function apiFetch(url, options = {}) {
  const res = await fetch(url, {
    ...options,
    headers: { ...authStore.authHeaders(), ...(options.headers || {}) },
  });
  if (res.status === 401 && authStore.token) {
    authStore.logout();
  }
  return res;
}