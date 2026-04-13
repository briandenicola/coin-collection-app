import { ref, readonly } from 'vue'

export type DialogType = 'confirm' | 'alert'
export type DialogVariant = 'default' | 'danger'

interface DialogState {
  visible: boolean
  type: DialogType
  variant: DialogVariant
  title: string
  message: string
  confirmLabel: string
  cancelLabel: string
  resolve: ((value: boolean) => void) | null
}

const state = ref<DialogState>({
  visible: false,
  type: 'confirm',
  variant: 'default',
  title: '',
  message: '',
  confirmLabel: 'OK',
  cancelLabel: 'Cancel',
  resolve: null,
})

function showConfirm(
  message: string,
  options?: { title?: string; confirmLabel?: string; cancelLabel?: string; variant?: DialogVariant },
): Promise<boolean> {
  return new Promise((resolve) => {
    state.value = {
      visible: true,
      type: 'confirm',
      variant: options?.variant ?? 'default',
      title: options?.title ?? 'Confirm',
      message,
      confirmLabel: options?.confirmLabel ?? 'Confirm',
      cancelLabel: options?.cancelLabel ?? 'Cancel',
      resolve,
    }
  })
}

function showAlert(
  message: string,
  options?: { title?: string; confirmLabel?: string; variant?: DialogVariant },
): Promise<boolean> {
  return new Promise((resolve) => {
    state.value = {
      visible: true,
      type: 'alert',
      variant: options?.variant ?? 'default',
      title: options?.title ?? '',
      message,
      confirmLabel: options?.confirmLabel ?? 'OK',
      cancelLabel: 'Cancel',
      resolve,
    }
  })
}

function handleConfirm() {
  state.value.resolve?.(true)
  state.value.visible = false
  state.value.resolve = null
}

function handleCancel() {
  state.value.resolve?.(false)
  state.value.visible = false
  state.value.resolve = null
}

export function useDialog() {
  return {
    state: readonly(state),
    showConfirm,
    showAlert,
    handleConfirm,
    handleCancel,
  }
}
