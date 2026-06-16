<template>
  <div class="chat-panel">
    <div class="chat-header">💬 Chat</div>

    <div class="chat-messages" ref="messagesEl">
      <div v-if="messages.length === 0" class="chat-empty">Aucun message pour l'instant…</div>
      <div
        v-for="(msg, i) in messages"
        :key="i"
        class="chat-msg"
        :class="{ 'chat-msg--own': msg.username === ownUsername, 'chat-msg--system': msg.system }"
      >
        <span v-if="!msg.system" class="chat-username">{{ msg.username }}</span>
        <span class="chat-text">{{ msg.message }}</span>
      </div>
    </div>

    <form class="chat-form" @submit.prevent="send">
      <input
        v-model="draft"
        type="text"
        placeholder="Envoyer un message…"
        maxlength="200"
        :disabled="!connected"
        class="chat-input"
      />
      <button type="submit" :disabled="!connected || !draft.trim()" class="chat-send">➤</button>
    </form>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from "vue";

const props = defineProps({
  messages: { type: Array, default: () => [] },
  ownUsername: { type: String, default: "" },
  connected: { type: Boolean, default: false },
});

const emit = defineEmits(["send"]);

const draft = ref("");
const messagesEl = ref(null);

const send = () => {
  const text = draft.value.trim();
  if (!text) return;
  emit("send", text);
  draft.value = "";
};

watch(
  () => props.messages.length,
  async () => {
    await nextTick();
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight;
    }
  }
);
</script>

<style scoped>
.chat-panel {
  display: flex;
  flex-direction: column;
  background: #1a1a2e;
  border: 1px solid #2d2d4e;
  border-radius: 10px;
  overflow: hidden;
  height: 100%;
  min-height: 260px;
}

.chat-header {
  padding: 8px 14px;
  background: #16213e;
  font-size: 0.85rem;
  font-weight: bold;
  color: #f97316;
  border-bottom: 1px solid #2d2d4e;
  flex-shrink: 0;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  scrollbar-width: thin;
  scrollbar-color: #2d2d4e transparent;
}

.chat-empty {
  color: #555;
  font-size: 0.8rem;
  text-align: center;
  margin-top: 12px;
}

.chat-msg {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.chat-username {
  font-size: 0.72rem;
  font-weight: bold;
  color: #f97316;
}

.chat-msg--own .chat-username {
  color: #60a5fa;
}

.chat-msg--system .chat-text {
  color: #888;
  font-style: italic;
  font-size: 0.8rem;
}

.chat-text {
  font-size: 0.88rem;
  color: #e2e8f0;
  word-break: break-word;
}

.chat-form {
  display: flex;
  gap: 6px;
  padding: 8px 10px;
  border-top: 1px solid #2d2d4e;
  flex-shrink: 0;
}

.chat-input {
  flex: 1;
  background: #0f0f23;
  border: 1px solid #2d2d4e;
  border-radius: 6px;
  padding: 6px 10px;
  color: #e2e8f0;
  font-size: 0.85rem;
  outline: none;
}

.chat-input:focus {
  border-color: #f97316;
}

.chat-input::placeholder {
  color: #444;
}

.chat-send {
  background: #f97316;
  border: none;
  border-radius: 6px;
  padding: 6px 12px;
  color: white;
  font-size: 0.9rem;
  cursor: pointer;
  flex-shrink: 0;
}

.chat-send:disabled {
  background: #333;
  cursor: not-allowed;
}
</style>
