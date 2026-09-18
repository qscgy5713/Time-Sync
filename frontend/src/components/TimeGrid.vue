<template>
  <div class="overflow-x-auto rounded-xl border border-gray-100 bg-white p-3 shadow-sm">
    <table class="border-separate border-spacing-1.5 select-none">
      <thead>
        <tr>
          <th class="w-14"></th>
          <th
            v-for="day in days"
            :key="day.value"
            class="px-1 pb-2 text-xs font-semibold text-gray-500 text-center whitespace-nowrap"
          >
            {{ day.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="slot in hours" :key="slot.value">
          <td class="pr-2 text-xs font-medium text-gray-400 text-right align-middle whitespace-nowrap tabular-nums">
            {{ slot.label }}
          </td>
          <td v-for="day in days" :key="day.value + '-' + slot.value" class="p-0">
            <div
              v-if="keyFor(day.value, slot.value)"
              :data-key="keyFor(day.value, slot.value)"
              class="w-14 h-7 rounded-lg border flex items-center justify-center text-xs font-semibold transition-all duration-100 ease-out"
              :class="[cellClasses(keyFor(day.value, slot.value)), mode === 'select' ? 'cursor-pointer active:scale-95 touch-none' : '']"
              @mousedown.prevent="onMouseDown(keyFor(day.value, slot.value))"
              @mouseenter="onMouseEnter(keyFor(day.value, slot.value))"
              @touchstart.prevent="onTouchStart(keyFor(day.value, slot.value))"
            >
              {{ cellLabel ? cellLabel(keyFor(day.value, slot.value)) : '' }}
            </div>
            <div v-else class="w-14 h-7"></div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useSelectionStore } from '../stores/selection'

const props = defineProps({
  days: { type: Array, required: true },
  hours: { type: Array, required: true },
  keyFor: { type: Function, required: true },
  mode: { type: String, default: 'select' },
  cellClass: { type: Function, default: null },
  cellLabel: { type: Function, default: null },
})

const selection = useSelectionStore()

function onMouseDown(key) {
  if (props.mode !== 'select' || !key) return
  selection.startDrag(key)
}
function onMouseEnter(key) {
  if (props.mode !== 'select' || !key) return
  selection.dragOver(key)
}
function onTouchStart(key) {
  if (props.mode !== 'select' || !key) return
  selection.startDrag(key)
}
function onTouchMove(e) {
  if (props.mode !== 'select' || !selection.isMouseDown) return
  const touch = e.touches[0]
  if (!touch) return
  const cell = document.elementFromPoint(touch.clientX, touch.clientY)?.closest('[data-key]')
  if (cell) {
    e.preventDefault()
    selection.dragOver(cell.dataset.key)
  }
}
function handleDragEnd() {
  selection.endDrag()
}
onMounted(() => {
  window.addEventListener('mouseup', handleDragEnd)
  window.addEventListener('touchmove', onTouchMove, { passive: false })
  window.addEventListener('touchend', handleDragEnd)
  window.addEventListener('touchcancel', handleDragEnd)
})
onUnmounted(() => {
  window.removeEventListener('mouseup', handleDragEnd)
  window.removeEventListener('touchmove', onTouchMove)
  window.removeEventListener('touchend', handleDragEnd)
  window.removeEventListener('touchcancel', handleDragEnd)
})

function cellClasses(key) {
  if (props.mode === 'display') {
    return props.cellClass ? props.cellClass(key) : 'bg-white border-gray-200 text-gray-400'
  }
  return selection.isSelected(key)
    ? 'bg-gradient-to-br from-brand-500 to-brand-600 border-brand-600 text-white shadow-sm shadow-brand-200'
    : 'bg-gray-50 border-gray-200 hover:bg-brand-50 hover:border-brand-200 text-gray-400'
}
</script>
