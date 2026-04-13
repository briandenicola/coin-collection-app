<template>
  <Teleport to="body">
    <Transition name="dialog-fade">
      <div v-if="dialog.state.value.visible" class="dialog-overlay" @click.self="onCancel">
        <div class="dialog card" role="alertdialog" :aria-label="dialog.state.value.title || 'Dialog'">
          <h3 v-if="dialog.state.value.title" class="dialog-title">{{ dialog.state.value.title }}</h3>
          <p class="dialog-message">{{ dialog.state.value.message }}</p>
          <div class="dialog-actions">
            <button
              v-if="dialog.state.value.type === 'confirm'"
              class="btn btn-secondary"
              @click="onCancel"
            >
              {{ dialog.state.value.cancelLabel }}
            </button>
            <button
              class="btn"
              :class="dialog.state.value.variant === 'danger' ? 'btn-danger' : 'btn-primary'"
              @click="dialog.handleConfirm()"
            >
              {{ dialog.state.value.confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useDialog } from '@/composables/useDialog'

const dialog = useDialog()

function onCancel() {
  dialog.handleCancel()
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.dialog {
  max-width: 400px;
  width: 90%;
  padding: 1.75rem;
  animation: dialog-enter 0.15s ease-out;
}

@keyframes dialog-enter {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.dialog-title {
  margin: 0 0 0.5rem 0;
  font-size: 1.05rem;
  color: var(--text-primary);
}

.dialog-message {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
}

.dialog-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.dialog-fade-enter-active {
  transition: opacity 0.15s ease-out;
}

.dialog-fade-leave-active {
  transition: opacity 0.1s ease-in;
}

.dialog-fade-enter-from,
.dialog-fade-leave-to {
  opacity: 0;
}
</style>
