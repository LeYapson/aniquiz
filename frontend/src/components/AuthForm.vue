<template>
  <div class="auth-container">
    <div class="auth-box">
      <h2>{{ isLogin ? 'Connexion à AniQuiz' : 'Créer un compte' }}</h2>
      
      <form @submit.prevent="handleSubmit">
        <div v-if="!isLogin" class="form-group">
          <label>Adresse Email</label>
          <input type="email" v-model="form.email" required placeholder="exemple@mail.com" />
        </div>

        <div class="form-group">
          <label>{{ isLogin ? 'Pseudo ou Email' : 'Pseudo' }}</label>
          <input type="text" v-model="form.identifier" required placeholder="Votre pseudo" />
        </div>

        <div class="form-group">
          <label>Mot de passe</label>
          <input type="password" v-model="form.password" required placeholder="••••••••" />
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
  
  const url = isLogin.value ? 'http://localhost:8080/api/auth/login' : 'http://localhost:8080/api/auth/register';
  
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
  min-height: 80vh;
}
.auth-box {
  background: #2a2a2a;
  padding: 30px;
  border-radius: 8px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.3);
  color: white;
}
.form-group {
  margin-bottom: 15px;
}
label {
  display: block;
  margin-bottom: 5px;
  font-size: 0.9rem;
  color: #ccc;
}
input {
  width: 100%;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #444;
  background: #1a1a1a;
  color: white;
  box-sizing: border-box;
}
.btn-submit {
  width: 100%;
  padding: 12px;
  background: #e91e63;
  border: none;
  color: white;
  font-weight: bold;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 10px;
}
.btn-submit:hover {
  background: #c2185b;
}
.toggle-mode {
  text-align: center;
  margin-top: 15px;
}
.toggle-mode button {
  background: none;
  border: none;
  color: #ff9800;
  cursor: pointer;
  text-decoration: underline;
}
.message {
  padding: 10px;
  border-radius: 4px;
  font-size: 0.9rem;
  margin-top: 10px;
}
.error { background: #d32f2f; color: white; }
.success { background: #388e3c; color: white; }
</style>