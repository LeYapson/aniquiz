<template>
  <span
    :class="['spinner', `spinner--${size}`, `spinner--${color}`]"
    role="status"
    :aria-label="label"
  >
    <span class="sr-only">{{ label }}</span>
  </span>
</template>

<script setup>
defineProps({
  size:  { type: String, default: 'md' },     // sm | md | lg
  color: { type: String, default: 'orange' }, // orange | white | blue | muted
  label: { type: String, default: 'Chargement…' },
})
</script>

<style scoped>
.spinner {
  display: inline-block;
  border-style: solid;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  flex-shrink: 0;
}

/* Sizes */
.spinner--sm { width: 14px; height: 14px; border-width: 2px; }
.spinner--md { width: 22px; height: 22px; border-width: 2.5px; }
.spinner--lg { width: 36px; height: 36px; border-width: 3px; }

/* Colors — track / active */
.spinner--orange { border-color: rgba(249,115,22,0.2); border-top-color: var(--orange); }
.spinner--white  { border-color: rgba(255,255,255,0.2); border-top-color: #fff; }
.spinner--blue   { border-color: rgba(59,130,246,0.2); border-top-color: var(--blue); }
.spinner--muted  { border-color: rgba(255,255,255,0.06); border-top-color: #475569; }

@keyframes spin { to { transform: rotate(360deg); } }

.sr-only {
  position: absolute;
  width: 1px; height: 1px;
  padding: 0; margin: -1px;
  overflow: hidden;
  clip: rect(0,0,0,0);
  white-space: nowrap;
  border: 0;
}
</style>
