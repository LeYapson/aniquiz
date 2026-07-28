<template>
  <div class="auth-container">
    <div class="auth-box">
      <h2>{{ title }}</h2>

      <!-- Mode mot de passe oublié -->
      <form v-if="isForgotPassword" @submit.prevent="handleForgot" novalidate>
        <BaseInput
          v-model="forgotForm.username"
          label="Pseudo"
          placeholder="Votre pseudo"
          autocomplete="username"
          :error="forgotErrors.username"
          id="forgot-username"
          required
        />
        <BaseInput
          v-model="forgotForm.email"
          type="email"
          label="Adresse Email"
          placeholder="exemple@mail.com"
          autocomplete="email"
          :error="forgotErrors.email"
          id="forgot-email"
          required
        />

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
          Envoyer le lien
        </BaseButton>
      </form>

      <!-- Mode connexion / inscription -->
      <form v-else @submit.prevent="handleSubmit" novalidate>
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

        <div v-if="isLogin" class="forgot-link">
          <BaseButton variant="link" type="button" @click="switchToForgot">
            Mot de passe oublié ?
          </BaseButton>
        </div>

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
          {{ toggleLabel }}
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { BaseButton, BaseInput } from './ui/index.js'
import { authStore } from '../authStore'
import { API_URL } from '../config'

const isLogin        = ref(true)
const isForgotPassword = ref(false)
const loading        = ref(false)
const message        = reactive({ text: '', type: '' })
const fieldErrors    = reactive({ email: '', identifier: '', password: '' })
const forgotErrors   = reactive({ username: '', email: '' })

const form       = reactive({ email: '', identifier: '', password: '' })
const forgotForm = reactive({ username: '', email: '' })

const title = computed(() => {
  if (isForgotPassword.value) return 'Mot de passe oublié'
  return isLogin.value ? 'Connexion à AniQuiz' : 'Créer un compte'
})

const toggleLabel = computed(() => {
  if (isForgotPassword.value) return 'Retour à la connexion'
  return isLogin.value ? "Pas encore de compte ? S'inscrire" : 'Déjà un compte ? Se connecter'
})

const clearErrors = () => {
  message.text           = ''
  fieldErrors.email      = ''
  fieldErrors.identifier = ''
  fieldErrors.password   = ''
  forgotErrors.username  = ''
  forgotErrors.email     = ''
}

const toggleMode = () => {
  if (isForgotPassword.value) {
    isForgotPassword.value = false
  } else {
    isLogin.value = !isLogin.value
  }
  clearErrors()
}

const switchToForgot = () => {
  isForgotPassword.value = true
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

const validateForgot = () => {
  let ok = true
  if (!forgotForm.username) {
    forgotErrors.username = 'Pseudo requis'
    ok = false
  }
  if (!forgotForm.email) {
    forgotErrors.email = 'Email requis'
    ok = false
  }
  return ok
}

const handleForgot = async () => {
  clearErrors()
  if (!validateForgot()) return

  loading.value = true
  try {
    const response = await fetch(`${API_URL}/api/auth/forgot-password`, {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ username: forgotForm.username, email: forgotForm.email }),
    })
    const data = await response.json()
    message.text = data.message || 'Email envoyé.'
    message.type = 'success'
    forgotForm.username = ''
    forgotForm.email    = ''
  } catch {
    message.text = 'Une erreur est survenue, réessaie plus tard.'
    message.type = 'error'
  } finally {
    loading.value = false
  }
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

.forgot-link {
  text-align: right;
  margin-top: -6px;
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
