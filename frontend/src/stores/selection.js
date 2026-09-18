import { defineStore } from 'pinia'

// Holds the drag-to-select state (5.3 拖拉選取實作) shared by TimeGrid instances.
export const useSelectionStore = defineStore('selection', {
  state: () => ({
    selected: new Set(),
    isMouseDown: false,
    dragMode: 'add',
  }),
  getters: {
    isSelected: (state) => (key) => state.selected.has(key),
    selectedArray: (state) => Array.from(state.selected),
    count: (state) => state.selected.size,
  },
  actions: {
    reset(initialKeys = []) {
      this.selected = new Set(initialKeys)
      this.isMouseDown = false
    },
    startDrag(key) {
      this.dragMode = this.selected.has(key) ? 'remove' : 'add'
      this.isMouseDown = true
      this.applyDrag(key)
    },
    dragOver(key) {
      if (!this.isMouseDown) return
      this.applyDrag(key)
    },
    applyDrag(key) {
      if (this.dragMode === 'add') {
        this.selected.add(key)
      } else {
        this.selected.delete(key)
      }
      this.selected = new Set(this.selected)
    },
    endDrag() {
      this.isMouseDown = false
    },
  },
})
