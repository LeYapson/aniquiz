<template>
  <div
    aria-hidden="true"
    :class="['skeleton', `skeleton--${variant}`]"
    :style="rootStyle"
  >
    <!-- text: renders N stacked lines, last one shorter -->
    <template v-if="variant === 'text'">
      <span
        v-for="i in lines"
        :key="i"
        class="skeleton__line"
        :style="{ width: i === lines && lines > 1 ? '58%' : '100%' }"
      />
    </template>

    <!-- card: preset card shape with header + body lines -->
    <template v-else-if="variant === 'card'">
      <span class="skeleton__card-head" />
      <span class="skeleton__line" style="width:80%; margin-top:12px" />
      <span class="skeleton__line" style="width:60%; margin-top:8px" />
      <span class="skeleton__line" style="width:70%; margin-top:8px" />
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: { type: String, default: 'rect' }, // rect | text | circle | card
  width:   { type: String, default: '' },
  height:  { type: String, default: '' },
  lines:   { type: Number, default: 3 }, // used when variant === 'text'
})

const rootStyle = computed(() => ({
  width:  props.width  || undefined,
  height: props.variant === 'rect' && props.height ? props.height : undefined,
}))
</script>

<style scoped>
/* ── Shimmer base ── */
@keyframes shimmer {
  from { background-position: -200% center; }
  to   { background-position:  200% center; }
}

.skeleton {
  --sk-bg: linear-gradient(
    90deg,
    var(--navy-3) 25%,
    var(--navy-4) 50%,
    var(--navy-3) 75%
  );
  background: var(--sk-bg);
  background-size: 200% 100%;
  animation: shimmer 1.6s ease-in-out infinite;
  border-radius: var(--radius-sm);
}

/* ── Variants ── */
.skeleton--rect   { display: block; height: 40px; }
.skeleton--circle { display: block; border-radius: 50%; width: 40px; height: 40px; }

.skeleton--text, .skeleton--card {
  display: flex;
  flex-direction: column;
  gap: 0;
  background: none;
  animation: none;
}

/* Lines inside text / card variants get the shimmer */
.skeleton__line, .skeleton__card-head {
  display: block;
  background: var(--sk-bg);
  background-size: 200% 100%;
  animation: shimmer 1.6s ease-in-out infinite;
  border-radius: var(--radius-sm);
}
.skeleton__line      { height: 14px; }
.skeleton__card-head { height: 18px; width: 40%; }

/* Stagger shimmer for visual variety */
.skeleton__line:nth-child(2) { animation-delay: 0.1s; }
.skeleton__line:nth-child(3) { animation-delay: 0.2s; }
.skeleton__line:nth-child(4) { animation-delay: 0.3s; }
</style>
