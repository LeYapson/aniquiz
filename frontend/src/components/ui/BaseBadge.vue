<template>
  <span :class="['badge', `badge--${variant}`, size !== 'md' && `badge--${size}`]">
    <slot />
  </span>
</template>

<script setup>
defineProps({
  /**
   * Color variant:
   * success | error | warning | info | orange | default
   * Semantic aliases: public (→ success) | private (→ warning) | soon (→ warning)
   */
  variant: { type: String, default: 'default' },
  size:    { type: String, default: 'md' }, // sm | md
})
</script>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: var(--radius-full);
  white-space: nowrap;
  line-height: 1;
}

/* Sizes */
.badge--sm { font-size: 0.65rem; padding: 2px 7px; }
.badge--md { font-size: 0.7rem;  padding: 3px 10px; }

/* Variants */
.badge--default {
  background: rgba(255,255,255,0.07);
  color: var(--text-dim);
}
.badge--success, .badge--public {
  background: var(--success-dim);
  color: var(--success);
  border: 1px solid var(--success-border);
}
.badge--error {
  background: var(--error-dim);
  color: var(--error);
  border: 1px solid var(--error-border);
}
.badge--warning, .badge--private, .badge--soon {
  background: var(--warning-dim);
  color: var(--warning);
  border: 1px solid var(--warning-border);
}
.badge--info {
  background: var(--info-dim);
  color: var(--info);
  border: 1px solid var(--info-border);
}
.badge--orange {
  background: rgba(249,115,22,0.15);
  color: #fb923c;
  border: 1px solid rgba(249,115,22,0.3);
}
.badge--blue {
  background: var(--info-dim);
  color: var(--info);
  border: 1px solid var(--info-border);
}
</style>
