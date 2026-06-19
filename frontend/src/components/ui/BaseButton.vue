<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    :class="[
      'btn',
      `btn--${variant}`,
      `btn--${size}`,
      { 'btn--full': full, 'btn--pill': pill, 'btn--loading': loading },
    ]"
  >
    <span v-if="loading" class="btn__spinner" aria-hidden="true" />
    <slot name="prefix" />
    <slot />
    <slot name="suffix" />
  </button>
</template>

<script setup>
defineProps({
  /** Visual style */
  variant:  { type: String, default: 'primary' }, // primary | secondary | ghost | danger | link | blue
  /** Size preset */
  size:     { type: String, default: 'md' },       // sm | md | lg | xl
  /** Native button type */
  type:     { type: String, default: 'button' },
  /** Shows spinner and disables interaction */
  loading:  { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  /** Stretch to parent width */
  full:     { type: Boolean, default: false },
  /** Full border-radius (stadium shape) */
  pill:     { type: Boolean, default: false },
})
</script>

<style scoped>
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  font-family: inherit;
  font-weight: 700;
  line-height: 1;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
  text-decoration: none;
  transition:
    transform var(--ease),
    box-shadow var(--ease),
    background var(--ease),
    border-color var(--ease),
    color var(--ease),
    opacity var(--ease);
}

/* ── Sizes ── */
.btn--sm  { font-size: 0.78rem;  padding: 6px 12px; }
.btn--md  { font-size: 0.875rem; padding: 9px 18px; }
.btn--lg  { font-size: 1rem;     padding: 12px 24px; }
.btn--xl  { font-size: 1.15rem;  padding: 16px 48px; }

/* ── Modifiers ── */
.btn--full { width: 100%; }
.btn--pill { border-radius: var(--radius-full); }

/* ── Variants ── */
.btn--primary {
  background: linear-gradient(135deg, var(--orange), var(--orange-2));
  color: #fff;
  box-shadow: var(--shadow-glow);
}
.btn--primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(249,115,22,0.45);
}
.btn--primary:active:not(:disabled) { transform: translateY(0); }

.btn--secondary {
  background: var(--navy-4);
  color: #cbd5e1;
  border-color: var(--border);
}
.btn--secondary:hover:not(:disabled) { border-color: var(--orange); color: var(--orange); }

.btn--ghost {
  background: transparent;
  color: var(--text-dim);
  border-color: var(--border);
}
.btn--ghost:hover:not(:disabled) { background: rgba(255,255,255,0.04); color: #f1f5f9; }

.btn--danger {
  background: var(--error-dim);
  color: var(--error);
  border-color: var(--error-border);
}
.btn--danger:hover:not(:disabled) { background: rgba(239,68,68,0.25); }

.btn--blue {
  background: rgba(59,130,246,0.15);
  color: var(--info);
  border-color: var(--info-border);
}
.btn--blue:hover:not(:disabled) { background: rgba(59,130,246,0.25); }

.btn--link {
  background: transparent;
  border-color: transparent;
  color: var(--orange);
  padding-left: 0;
  padding-right: 0;
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 3px;
}
.btn--link:hover:not(:disabled) { color: #fb923c; }

/* ── Disabled (all variants) ── */
.btn:disabled {
  background: #1e293b;
  border-color: transparent;
  color: #475569;
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}

/* ── Loading ── */
.btn--loading { cursor: wait; }

/* ── Spinner ── */
.btn__spinner {
  display: inline-block;
  width: 13px;
  height: 13px;
  border: 2px solid rgba(255,255,255,0.25);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: btn-spin 0.65s linear infinite;
  flex-shrink: 0;
}

@keyframes btn-spin { to { transform: rotate(360deg); } }
</style>
