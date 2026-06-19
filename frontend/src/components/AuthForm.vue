<template>
  <div class="auth-container">
    <div class="auth-box">
      <h2>{{ isLogin ? 'Connexion à AniQuiz' : 'Créer un compte' }}</h2>

      <form @submit.prevent="handleSubmit" novalidate>
        <BaseInput
          v-if="!isLogin"
          v-model="form.email"
          type="email"
          label="Adresse Email"
          placeholder="exemple@mail.com"
          autocomplete="email"
          :error="fieldErrors.email"
          id="auth-email"
          required
        />

        <BaseInput
          v-model="form.identifier"
          :label="isLogin ? 'Pseudo ou Email' : 'Pseudo'"
          placeholder="Votre pseudo"
          autocomplete="username"
          :error="fieldErrors.identifier"
          id="auth-identifier"
          required
        />

        <BaseInput
          v-model="form.password"
          type="password"
          label="Mot de passe"
          placeholder="••••••••"
          autocomplete="current-password"
          :error="fieldErrors.password"
          id="auth-password"
          required
        />

        <!-- Inline alert for server-level feedback -->
        <div
          v-if="message.text"
          :class="['auth-alert', `auth-alert--${message.type}`]"
          role="alert"
          aria-live="polite"
        >
          {{ message.text }}
        </div>

        <BaseButton
          type="submit"
          variant="primary"
          size="lg"
          full
          :loading="loading"
          style="margin-top: 6px;"
        >
          {{ isLogin ? 'Se connecter' : "S'inscrire" }}
        </BaseButton>
      </form>

      <div class="auth-toggle">
        <BaseButton variant="link" @click="toggleMode">
          {{ isLogin ? "Pas encore de compte ? S'inscrire" : 'Déjà un compte ? Se connecter' }}
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { BaseButton, BaseInput } from './ui/index.js'
import { authStore } from '../authStore'
import { API_URL } from '../config'

const isLogin  = ref(true)
const loading  = ref(false)
const message  = reactive({ text: '', type: '' })
const fieldErrors = reactive({ email: '', identifier: '', password: '' })

const form = reactive({ email: '', identifier: '', password: '' })

const clearErrors = () => {
  message.text        = ''
  fieldErrors.email      = ''
  fieldErrors.identifier = ''
  fieldErrors.password   = ''
}

const toggleMode = () => {
  isLogin.value = !isLogin.value
  clearErrors()
}

const validate = () => {
  let ok = true
  if (!isLogin.value && !form.email) {
    fieldErrors.email = 'Email requis'
    ok = false
  }
  if (!form.identifier) {
    fieldErrors.identifier = isLogin.value ? 'Pseudo ou email requis' : 'Pseudo requis'
    ok = false
  }
  if (!form.password) {
    fieldErrors.password = 'Mot de passe requis'
    ok = false
  }
  return ok
}

const handleSubmit = async () => {
  clearErrors()
  if (!validate()) return

  loading.value = true
  const url      = isLogin.value
    ? `${API_URL}/api/auth/login`
    : `${API_URL}/api/auth/register`

  const bodyData = isLogin.value
    ? { identifier: form.identifier, password: form.password }
    : { username: form.identifier, email: form.email, password: form.password }

  try {
    const response = await fetch(url, {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify(bodyData),
    })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error || 'Une erreur est survenue')

    if (isLogin.value) {
      authStore.setUser(data.user, data.token)
    } else {
      message.text = 'Inscription réussie ! Vous pouvez vous connecter.'
      message.type = 'success'
      isLogin.value = true
    }
  } catch (err) {
    message.text = err.message
    message.type = 'error'
  } finally {
    loading.value = false
  }
}
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
  background: var(--navy-3);
  border: 1px solid var(--border);
  padding: 36px 32px;
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 400px;
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
  gap: 0;
}

.auth-box h2 {
  font-size: 1.4rem;
  font-weight: 700;
  color: #f1f5f9;
  text-align: center;
  margin-bottom: 24px;
}

form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* Inline server alert */
.auth-alert {
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  line-height: 1.4;
}
.auth-alert--error {
  background: var(--error-dim);
  color: #fca5a5;
  border: 1px solid var(--error-border);
}
.auth-alert--success {
  background: var(--success-dim);
  color: #86efac;
  border: 1px solid var(--success-border);
}

.auth-toggle {
  text-align: center;
  margin-top: 18px;
}

@media (max-width: 480px) {
  .auth-box { padding: 24px 18px; }
}
</style>
