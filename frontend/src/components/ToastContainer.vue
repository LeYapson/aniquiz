<template>
  <Teleport to="body">
    <div class="toast-viewport" aria-live="polite" aria-atomic="false" aria-label="Notifications">
      <TransitionGroup name="toast" tag="div" class="toast-list">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="['toast', `toast--${toast.type}`]"
          role="status"
        >
          <!-- XP toast: custom layout -->
          <template v-if="toast.type === 'xp'">
            <span class="toast__icon" aria-hidden="true">⭐</span>
            <div class="toast__body">
              <strong class="toast__title">+{{ toast.xpGained }} XP</strong>
              <p v-if="toast.levelUp" class="toast__sub toast__sub--gold">
                Niveau {{ toast.newLevel }} atteint ! 🎉
              </p>
              <p v-else class="toast__sub">
                Total : {{ toast.newXP }} XP · Niv. {{ toast.newLevel }}
              </p>
            </div>
          </template>

          <!-- Standard toasts -->
          <template v-else>
            <span class="toast__icon" aria-hidden="true">{{ ICONS[toast.type] }}</span>
            <div class="toast__body">
              <strong v-if="toast.title" class="toast__title">{{ toast.title }}</strong>
              <p class="toast__sub" :class="{ 'toast__sub--solo': !toast.title }">
                {{ toast.message }}
              </p>
            </div>
          </template>

          <button
            class="toast__close"
            aria-label="Fermer la notification"
            @click="dismiss(toast.id)"
          >×</button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { useToast } from '../composables/useToast'

const { toasts, dismiss } = useToast()

const ICONS = {
  success: '✓',
  error:   '✕',
  info:    'ℹ',
  warning: '⚠',
}
</script>

<style scoped>
/* ── Viewport (fixed anchor) ── */
.toast-viewport {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 9000;
  pointer-events: none;
}

.toast-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}

/* ── Toast card ── */
.toast {
  pointer-events: all;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 280px;
  max-width: 380px;
  padding: 14px 16px;
  border-radius: var(--radius-lg);
  background: var(--navy-3);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-lg);
  border-left: 4px solid;
}

/* Left-border accent per type */
.toast--success { border-left-color: var(--success); }
.toast--error   { border-left-color: var(--error);   }
.toast--warning { border-left-color: var(--warning); }
.toast--info    { border-left-color: var(--info);    }
.toast--xp      { border-left-color: #ffd700;        }

/* ── Icon ── */
.toast__icon {
  font-size: 1.4rem;
  flex-shrink: 0;
  line-height: 1;
}
.toast--success .toast__icon { color: var(--success); font-size: 1rem; font-weight: 900; }
.toast--error   .toast__icon { color: var(--error);   font-size: 1rem; font-weight: 900; }
.toast--warning .toast__icon { color: var(--warning); font-size: 1rem; }
.toast--info    .toast__icon { color: var(--info);    font-size: 1rem; }

/* ── Body ── */
.toast__body { flex: 1; min-width: 0; }
.toast__title {
  display: block;
  font-size: 0.9rem;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1.2;
}
.toast__sub {
  margin: 2px 0 0;
  font-size: 0.82rem;
  color: var(--text-dim);
  line-height: 1.35;
}
.toast__sub--solo { margin: 0; font-size: 0.875rem; color: #cbd5e1; }
.toast__sub--gold { color: #ffd700; font-weight: 700; margin: 2px 0 0; }

/* ── Close button ── */
.toast__close {
  align-self: flex-start;
  background: none;
  border: none;
  color: #475569;
  font-size: 1.15rem;
  cursor: pointer;
  padding: 0 2px;
  line-height: 1;
  transition: color var(--ease);
  flex-shrink: 0;
}
.toast__close:hover { color: #f1f5f9; }

/* ── Transitions ── */
.toast-enter-active { transition: opacity 0.25s ease, transform 0.25s ease; }
.toast-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.toast-move         { transition: transform 0.25s ease; }
.toast-enter-from   { opacity: 0; transform: translateX(24px); }
.toast-leave-to     { opacity: 0; transform: translateX(24px); }

/* ── Responsive: mobile moves to bottom center, above mobile nav ── */
@media (max-width: 768px) {
  .toast-viewport {
    right: 10px;
    left: 10px;
    bottom: 70px; /* clears the 56px mobile tab bar */
  }
  .toast-list { align-items: stretch; }
  .toast { min-width: unset; max-width: unset; }
}
</style>
