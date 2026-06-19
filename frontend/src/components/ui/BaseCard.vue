<template>
  <div
    :class="[
      'card',
      `card--pad-${padding}`,
      {
        'card--hover':  hoverable,
        'card--accent': accent,
        'card--flat':   flat,
      },
    ]"
  >
    <slot />
  </div>
</template>

<script setup>
defineProps({
  /** Show hover lift + border glow */
  hoverable: { type: Boolean, default: false },
  /** Left orange accent border */
  accent:    { type: Boolean, default: false },
  /** Remove background (transparent) */
  flat:      { type: Boolean, default: false },
  padding:   { type: String, default: 'md' }, // sm | md | lg
})
</script>

<style scoped>
.card {
  background: var(--navy-3);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
}

/* Padding */
.card--pad-sm { padding: 12px 14px; }
.card--pad-md { padding: 16px 18px; }
.card--pad-lg { padding: 24px 28px; }

/* Hover effect */
.card--hover {
  transition: border-color var(--ease), transform var(--ease);
  cursor: default;
}
.card--hover:hover {
  border-color: rgba(249,115,22,0.25);
  transform: translateY(-2px);
}

/* Orange accent on left */
.card--accent {
  border-left: 3px solid var(--orange);
  border-radius: var(--radius-md);
}
.card--accent:hover { border-left-color: var(--orange-2); }

/* No background */
.card--flat { background: transparent; }
</style>
