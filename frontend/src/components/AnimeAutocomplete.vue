<template>
  <div class="autocomplete-wrapper">
    <label :for="inputId" class="sr-only">Nom de l'anime</label>
    <div class="autocomplete-field">
      <input
        :id="inputId"
        ref="inputRef"
        v-model="inputValue"
        @input="onInput"
        @keydown="onKeydown"
        @blur="onBlur"
        @focus="showDropdown = suggestions.length > 0"
        :placeholder="placeholder"
        autocomplete="off"
        spellcheck="false"
        aria-autocomplete="list"
        :aria-expanded="showDropdown && suggestions.length > 0"
        aria-controls="autocomplete-list"
      />
      <button
        v-if="showSubmit"
        @click="$emit('submit')"
        class="btn-submit"
        aria-label="Envoyer ma réponse"
      >
        Envoyer ma réponse
      </button>
    </div>

    <ul
      v-if="showDropdown && suggestions.length > 0"
      id="autocomplete-list"
      class="autocomplete-dropdown"
      role="listbox"
    >
      <li
        v-for="(item, i) in suggestions"
        :key="item"
        :class="{ highlighted: i === activeIndex }"
        @mousedown.prevent="select(item)"
        role="option"
        :aria-selected="i === activeIndex"
      >
        {{ item }}
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

const props = defineProps({
  modelValue: { type: String, default: '' },
  dictionary: { type: Array, default: () => [] },
  inputId: { type: String, default: 'anime-guess' },
  placeholder: { type: String, default: "Nom de l'anime..." },
  // When false, the built-in "Envoyer" button is hidden so the parent can
  // provide its own action buttons (e.g. the speed-run Valider / Skip layout).
  showSubmit: { type: Boolean, default: true },
});

const emit = defineEmits(['update:modelValue', 'submit']);

const inputRef = ref(null);
const inputValue = ref(props.modelValue ?? '');
const showDropdown = ref(false);
const activeIndex = ref(-1);

watch(() => props.modelValue, (v) => { inputValue.value = v ?? ''; });
watch(inputValue, (v) => emit('update:modelValue', v));

const suggestions = computed(() => {
  const q = inputValue.value.trim().toLowerCase();
  if (q.length < 2) return [];
  return props.dictionary
    .filter(name => name.toLowerCase().includes(q))
    .slice(0, 8);
});

const onInput = () => {
  activeIndex.value = -1;
  showDropdown.value = true;
};

const onKeydown = (e) => {
  const list = suggestions.value;
  if (!showDropdown.value || list.length === 0) {
    if (e.key === 'Enter') emit('submit');
    return;
  }
  if (e.key === 'ArrowDown') {
    e.preventDefault();
    activeIndex.value = Math.min(activeIndex.value + 1, list.length - 1);
  } else if (e.key === 'ArrowUp') {
    e.preventDefault();
    activeIndex.value = Math.max(activeIndex.value - 1, -1);
  } else if (e.key === 'Enter') {
    if (activeIndex.value >= 0) {
      select(list[activeIndex.value]);
    } else {
      showDropdown.value = false;
      emit('submit');
    }
  } else if (e.key === 'Escape') {
    showDropdown.value = false;
    activeIndex.value = -1;
  }
};

const onBlur = () => {
  setTimeout(() => { showDropdown.value = false; }, 150);
};

const select = (name) => {
  inputValue.value = name;
  showDropdown.value = false;
  activeIndex.value = -1;
  emit('update:modelValue', name);
  emit('submit');
};

defineExpose({ focus: () => inputRef.value?.focus() });
</script>

<style scoped>
.autocomplete-wrapper {
  position: relative;
  width: 100%;
}
.autocomplete-field {
  display: flex;
  gap: 8px;
}
.autocomplete-field input {
  flex: 1;
  padding: 10px 14px;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.12);
  color: #f1f5f9;
  border-radius: 8px;
  font-size: 0.95rem;
  outline: none;
  transition: border-color 0.15s;
}
.autocomplete-field input:focus { border-color: #f97316; }
.autocomplete-field input::placeholder { color: #475569; }

.btn-submit {
  padding: 10px 18px;
  background: #f97316;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 700;
  font-size: 0.88rem;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.15s;
  flex-shrink: 0;
}
.btn-submit:hover { opacity: 0.85; }

.autocomplete-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: #1e2a45;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  list-style: none;
  padding: 4px 0;
  margin: 0;
  z-index: 200;
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  max-height: 280px;
  overflow-y: auto;
}
.autocomplete-dropdown li {
  padding: 9px 14px;
  cursor: pointer;
  color: #e2e8f0;
  font-size: 0.9rem;
  transition: background 0.1s;
}
.autocomplete-dropdown li:hover,
.autocomplete-dropdown li.highlighted {
  background: rgba(249,115,22,0.15);
  color: #f97316;
}
</style>
