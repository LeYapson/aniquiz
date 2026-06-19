<template>
  <div
    class="field"
    :class="{
      'field--error':    !!error,
      'field--disabled': disabled,
    }"
  >
    <label v-if="label" :for="uid" class="field-label">{{ label }}</label>

    <div class="field-control">
      <span v-if="$slots.prefix" class="field-adorn field-adorn--left" aria-hidden="true">
        <slot name="prefix" />
      </span>

      <input
        :id="uid"
        v-bind="$attrs"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :autocomplete="autocomplete"
        :aria-invalid="error ? 'true' : undefined"
        :aria-describedby="
          error   ? `${uid}-err`  :
          hint    ? `${uid}-hint` : undefined
        "
        :class="[
          'field-input',
          { 'field-input--prefix': $slots.prefix, 'field-input--suffix': $slots.suffix },
        ]"
        @input="$emit('update:modelValue', $event.target.value)"
        @change="$emit('change', $event.target.value)"
      />

      <span v-if="$slots.suffix" class="field-adorn field-adorn--right" aria-hidden="true">
        <slot name="suffix" />
      </span>
    </div>

    <p v-if="error" :id="`${uid}-err`"  class="field-msg field-msg--error" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="`${uid}-hint`" class="field-msg field-msg--hint">{{ hint }}</p>
  </div>
</template>

<script setup>
import { computed, getCurrentInstance } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps({
  modelValue:   { type: [String, Number], default: '' },
  type:         { type: String, default: 'text' },
  label:        { type: String, default: '' },
  placeholder:  { type: String, default: '' },
  hint:         { type: String, default: '' },
  error:        { type: String, default: '' },
  disabled:     { type: Boolean, default: false },
  /** Explicit id; auto-generated from instance uid if omitted */
  id:           { type: String, default: '' },
  autocomplete: { type: String, default: 'off' },
})

defineEmits(['update:modelValue', 'change'])

const instanceUid = getCurrentInstance()?.uid ?? Math.random().toString(36).slice(2)
const uid = computed(() => props.id || `field-${instanceUid}`)
</script>

<style scoped>
/* ── Wrapper ── */
.field { display: flex; flex-direction: column; gap: 6px; }

/* ── Label ── */
.field-label {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  line-height: 1;
}
.field--error .field-label { color: var(--error); }

/* ── Control (wrapper for adornments) ── */
.field-control { position: relative; display: flex; align-items: center; }

/* ── Input ── */
.field-input {
  width: 100%;
  padding: 10px 14px;
  background: var(--navy);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: var(--radius-sm);
  color: #f1f5f9;
  font-size: 0.95rem;
  font-family: inherit;
  line-height: 1.5;
  outline: none;
  transition: border-color var(--ease), box-shadow var(--ease);
  box-sizing: border-box;
}
.field-input::placeholder { color: #475569; }
.field-input:focus         { border-color: var(--orange); box-shadow: 0 0 0 3px rgba(249,115,22,0.15); }
.field-input:disabled      { opacity: 0.5; cursor: not-allowed; }

/* Adornment offsets */
.field-input--prefix { padding-left: 38px; }
.field-input--suffix { padding-right: 38px; }

/* ── Adornments ── */
.field-adorn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-dim);
  font-size: 0.9rem;
  pointer-events: none;
  display: flex;
  align-items: center;
}
.field-adorn--left  { left: 12px; }
.field-adorn--right { right: 12px; }

/* ── Error state ── */
.field--error .field-input { border-color: var(--error-border); }
.field--error .field-input:focus { box-shadow: 0 0 0 3px var(--error-dim); }

/* ── Messages ── */
.field-msg {
  font-size: 0.8rem;
  line-height: 1.4;
  margin: 0;
}
.field-msg--error { color: var(--error); }
.field-msg--hint  { color: var(--text-dim); }
</style>
