import { reactive } from 'vue';

const stored = JSON.parse(localStorage.getItem('auth')) || {};

export const authStore = reactive({
  user: stored.user || null,
  token: stored.token || null,

  setUser(userData, token) {
    this.user = userData;
    this.token = token;
    localStorage.setItem('auth', JSON.stringify({ user: userData, token }));
  },

  logout() {
    this.user = null;
    this.token = null;
    localStorage.removeItem('auth');
    localStorage.removeItem('user');
  },

  get isAuthenticated() {
    return this.user !== null && this.token !== null;
  },

  authHeaders() {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${this.token}`,
    };
  },
});
