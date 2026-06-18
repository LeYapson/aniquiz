import { reactive } from 'vue';

const stored = JSON.parse(localStorage.getItem('auth')) || {};

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
    return this.user !== null && this.token !== null;
  },

  authHeaders: () => ({
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authStore.token}`,
  }),
});
