<template>
  <div class="auth-container">
    <div class="auth-box">
      <h2>Nouveau mot de passe</h2>

      <template v-if="success">
        <div class="auth-alert auth-alert--success" role="alert">
          Mot de passe mis à jour ! Tu peux maintenant te connecter.
        </div>
        <div class="auth-toggle">
          <BaseButton variant="link" @click="$router.push('/')">
            Retour à la connexion
          </BaseButton>
        </div>
      </template>

      <template v-else-if="!token">
        <div class="auth-alert auth-alert--error" role="alert">
          Lien invalide ou manquant. Refais une demande de réinitialisation.
        </div>
        <div class="auth-toggle">
          <BaseButton variant="link" @click="$router.push('/')">
            Retour
          </BaseButton>
        </div>
      </template>

      <template v-else>
        <form @submit.prevent="handleSubmit" novalidate>
          <BaseInput
            v-model="form.password"
            type="password"
            label="Nouveau mot de passe"
            placeholder="••••••••"
            autocomplete="new-password"
            :error="fieldErrors.password"
            id="reset-password"
            required
          />
          <BaseInput
            v-model="form.confirm"
            type="password"
            label="Confirmer le mot de passe"
            placeholder="••••••••"
            autocomplete="new-password"
            :error="fieldErrors.confirm"
            id="reset-confirm"
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
            Enregistrer le mot de passe
          </BaseButton>
        </form>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { BaseButton, BaseInput } from '../components/ui/index.js'
import { API_URL } from '../config'

const route   = useRoute()
const router  = useRouter()
const token   = ref('')
const loading = ref(false)
const success = ref(false)
const message = reactive({ text: '', type: '' })
const fieldErrors = reactive({ password: '', confirm: '' })
const form = reactive({ password: '', confirm: '' })

onMounted(() => {
  token.value = route.query.token || ''
})

const validate = () => {
  let ok = true
  fieldErrors.password = ''
  fieldErrors.confirm  = ''
  if (form.password.length < 8) {
    fieldErrors.password = 'Minimum 8 caractères'
    ok = false
  }
  if (form.password !== form.confirm) {
    fieldErrors.confirm = 'Les mots de passe ne correspondent pas'
    ok = false
  }
  return ok
}

const handleSubmit = async () => {
  message.text = ''
  if (!validate()) return

  loading.value = true
  try {
    const response = await fetch(`${API_URL}/api/auth/reset-password`, {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ token: token.value, password: form.password }),
    })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error || 'Une erreur est survenue')
    success.value = true
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
