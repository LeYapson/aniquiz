<template>
  <div class="empty" role="status" :aria-label="title || 'Aucun résultat'">
    <div v-if="icon" class="empty__icon" aria-hidden="true">{{ icon }}</div>
    <h3 v-if="title"       class="empty__title">{{ title }}</h3>
    <p  v-if="description" class="empty__desc">{{ description }}</p>
    <BaseButton
      v-if="actionLabel"
      :variant="actionVariant"
      class="empty__action"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </BaseButton>
  </div>
</template>

<script setup>
import BaseButton from './BaseButton.vue'

defineProps({
  icon:          { type: String,  default: '' },
  title:         { type: String,  default: '' },
  description:   { type: String,  default: '' },
  actionLabel:   { type: String,  default: '' },
  actionVariant: { type: String,  default: 'secondary' },
})

defineEmits(['action'])
</script>

<style scoped>
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 48px 24px;
  text-align: center;
  color: var(--text-dim);
}

.empty__icon {
  font-size: 2.5rem;
  line-height: 1;
  margin-bottom: 4px;
  opacity: 0.7;
}

.empty__title {
  font-size: 1rem;
  font-weight: 700;
  color: #cbd5e1;
  margin: 0;
}

.empty__desc {
  font-size: 0.875rem;
  color: var(--text-dim);
  max-width: 320px;
  line-height: 1.55;
  margin: 0;
}

.empty__action {
  margin-top: 6px;
}
</style>
