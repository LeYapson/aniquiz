<template>
  <div class="auth-container">
    <div class="auth-box">
      <h2>{{ isLogin ? 'Connexion à AniQuiz' : 'Créer un compte' }}</h2>
      
      <form @submit.prevent="handleSubmit">
        <div v-if="!isLogin" class="form-group">
          <label for="auth-email">Adresse Email</label>
          <input id="auth-email" type="email" v-model="form.email" required placeholder="exemple@mail.com" autocomplete="email" />
        </div>

        <div class="form-group">
          <label for="auth-identifier">{{ isLogin ? 'Pseudo ou Email' : 'Pseudo' }}</label>
          <input id="auth-identifier" type="text" v-model="form.identifier" required placeholder="Votre pseudo" autocomplete="username" />
        </div>

        <div class="form-group">
          <label for="auth-password">Mot de passe</label>
          <input id="auth-password" type="password" v-model="form.password" required placeholder="••••••••" autocomplete="current-password" />
        </div>

        <p v-if="message.text" :class="['message', message.type]">{{ message.text }}</p>

        <button type="submit" class="btn-submit">
          {{ isLogin ? 'Se connecter' : "S'inscrire" }}
        </button>
      </form>

      <div class="toggle-mode">
        <button @click="toggleMode">
          {{ isLogin ? "Pas encore de compte ? S'inscrire" : 'Déjà un compte ? Se connecter' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { authStore } from '../authStore';
import { API_URL } from '../config';

const isLogin = ref(true);
const message = reactive({ text: '', type: '' });

const form = reactive({
  email: '',
  identifier: '',
  password: ''
});

const toggleMode = () => {
  isLogin.value = !isLogin.value;
  message.text = '';
};

const handleSubmit = async () => {
  message.text = '';
  
  const url = isLogin.value ? `${API_URL}/api/auth/login` : `${API_URL}/api/auth/register`;
  
  // Préparation du corps de la requête selon le mode
  const bodyData = isLogin.value 
    ? { identifier: form.identifier, password: form.password }
    : { username: form.identifier, email: form.email, password: form.password };

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(bodyData)
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || 'Une erreur est survenue');
    }

    if (isLogin.value) {
      authStore.setUser(data.user, data.token);
      message.text = 'Connexion réussie !';
      message.type = 'success';
    } else {
      message.text = 'Inscription réussie ! Vous pouvez vous connecter.';
      message.type = 'success';
      isLogin.value = true; // Bascule sur le formulaire de connexion
    }
  } catch (err) {
    message.text = err.message;
    message.type = 'error';
  }
};
</script>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 56px);
  padding: 24px;
}
.auth-box {
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  padding: 36px 32px;
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
}
.auth-box h2 {
  font-size: 1.4rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 24px;
  text-align: center;
}
.form-group { margin-bottom: 16px; }
label {
  display: block;
  margin-bottom: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
input {
  width: 100%;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid rgba(255,255,255,0.1);
  background: #0f0f23;
  color: #f1f5f9;
  font-size: 0.95rem;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}
input:focus { border-color: #f97316; }
input::placeholder { color: #475569; }
.btn-submit {
  width: 100%;
  padding: 12px;
  background: linear-gradient(135deg, #f97316, #ea580c);
  border: none;
  color: white;
  font-weight: 700;
  font-size: 1rem;
  border-radius: 8px;
  cursor: pointer;
  margin-top: 6px;
  box-shadow: 0 4px 14px rgba(249,115,22,0.3);
  transition: transform 0.15s, box-shadow 0.15s;
}
.btn-submit:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(249,115,22,0.4);
}
.toggle-mode { text-align: center; margin-top: 18px; }
.toggle-mode button {
  background: none;
  border: none;
  color: #f97316;
  cursor: pointer;
  font-size: 0.88rem;
  text-decoration: underline;
}
.toggle-mode button:hover { color: #fb923c; }
.message {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 0.88rem;
  margin-top: 10px;
}
.error { background: rgba(239,68,68,0.15); color: #fca5a5; border: 1px solid rgba(239,68,68,0.3); }
.success { background: rgba(34,197,94,0.12); color: #86efac; border: 1px solid rgba(34,197,94,0.3); }
</style>