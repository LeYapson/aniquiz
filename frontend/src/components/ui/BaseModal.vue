<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="open"
        class="modal-overlay"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="title ? `${dialogId}-title` : undefined"
        @click.self="closeOnOverlay && close()"
      >
        <div
          ref="dialogEl"
          class="modal-panel"
          :class="`modal-panel--${size}`"
          tabindex="-1"
        >
          <!-- Header -->
          <div v-if="title || showClose" class="modal-header">
            <h2 v-if="title" :id="`${dialogId}-title`" class="modal-title">
              {{ title }}
            </h2>
            <button
              v-if="showClose"
              class="modal-close"
              aria-label="Fermer"
              @click="close()"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="modal-body">
            <slot />
          </div>

          <!-- Footer (optional) -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, nextTick, onUnmounted, getCurrentInstance } from 'vue'

const props = defineProps({
  open:           { type: Boolean, default: false },
  title:          { type: String,  default: '' },
  size:           { type: String,  default: 'md' }, // sm | md | lg | xl | full
  closeOnOverlay: { type: Boolean, default: true },
  showClose:      { type: Boolean, default: true },
})

const emit = defineEmits(['update:open', 'close'])

const dialogEl   = ref(null)
const dialogId   = `modal-${getCurrentInstance()?.uid ?? Math.random().toString(36).slice(2)}`
let  prevFocus   = null

const close = () => {
  emit('update:open', false)
  emit('close')
}

/* ── Focus trap ── */
const FOCUSABLE = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(', ')

const trapFocus = (e) => {
  if (e.key === 'Escape') { close(); return }
  if (e.key !== 'Tab' || !dialogEl.value) return

  const focusable = [...dialogEl.value.querySelectorAll(FOCUSABLE)]
  if (!focusable.length) { e.preventDefault(); return }

  const first = focusable[0]
  const last  = focusable[focusable.length - 1]

  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault(); last.focus()
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault(); first.focus()
  }
}

watch(() => props.open, async (isOpen) => {
  if (isOpen) {
    prevFocus = document.activeElement
    await nextTick()
    dialogEl.value?.focus()
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', trapFocus)
  } else {
    document.body.style.overflow = ''
    document.removeEventListener('keydown', trapFocus)
    // Return focus to the trigger element
    prevFocus?.focus?.()
    prevFocus = null
  }
})

onUnmounted(() => {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', trapFocus)
})
</script>

<style scoped>
/* ── Overlay ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(8, 8, 20, 0.82);
  backdrop-filter: blur(4px);
}

/* ── Panel ── */
.modal-panel {
  position: relative;
  width: 100%;
  background: #12122a;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: var(--radius-xl);
  display: flex;
  flex-direction: column;
  max-height: calc(100dvh - 48px);
  overflow: hidden;
  outline: none;
}

/* Sizes */
.modal-panel--sm   { max-width: 400px; }
.modal-panel--md   { max-width: 560px; }
.modal-panel--lg   { max-width: 720px; }
.modal-panel--xl   { max-width: 960px; }
.modal-panel--full { max-width: 100%; height: calc(100dvh - 48px); }

/* ── Header ── */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 24px 0;
  flex-shrink: 0;
}
.modal-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: #f1f5f9;
  margin: 0;
  flex: 1;
}
.modal-close {
  width: 34px;
  height: 34px;
  padding: 0;
  flex-shrink: 0;
  border-radius: 50%;
  border: none;
  background: rgba(255,255,255,0.06);
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--ease), color var(--ease);
}
.modal-close:hover { background: rgba(255,255,255,0.14); color: #f1f5f9; }

/* ── Body ── */
.modal-body {
  padding: 20px 24px 24px;
  overflow-y: auto;
  flex: 1;
}

/* ── Footer ── */
.modal-footer {
  padding: 0 24px 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  flex-shrink: 0;
  border-top: 1px solid var(--border);
  padding-top: 16px;
}

/* ── Transitions ── */
.modal-enter-active { transition: opacity 0.18s ease, transform 0.2s ease; }
.modal-leave-active { transition: opacity 0.15s ease; }
.modal-enter-from   { opacity: 0; }
.modal-leave-to     { opacity: 0; }

.modal-enter-active .modal-panel { transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1); }
.modal-enter-from   .modal-panel { transform: scale(0.94) translateY(8px); }
</style>
