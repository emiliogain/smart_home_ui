import { create } from 'zustand'

interface UndoStoreState {
  visible: boolean
  message: string
  undoAction: (() => void) | null
  showUndo: (message: string, undoAction?: () => void) => void
  dismiss: () => void
}

let dismissTimeoutId: ReturnType<typeof setTimeout> | null = null

export const useUndoStore = create<UndoStoreState>((set, get) => ({
  visible: false,
  message: '',
  undoAction: null,
  showUndo: (message, undoAction) => {
    if (dismissTimeoutId) {
      clearTimeout(dismissTimeoutId)
      dismissTimeoutId = null
    }
    set({
      visible: true,
      message,
      undoAction: undoAction ?? null,
    })
    dismissTimeoutId = setTimeout(() => {
      get().dismiss()
    }, 5000)
  },
  dismiss: () => {
    if (dismissTimeoutId) {
      clearTimeout(dismissTimeoutId)
      dismissTimeoutId = null
    }
    set({ visible: false, message: '', undoAction: null })
  },
}))

export function showUndo(message: string, undoAction?: () => void): void {
  useUndoStore.getState().showUndo(message, undoAction)
}
