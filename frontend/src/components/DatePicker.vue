<template>
  <div ref="rootEl" class="relative">
    <button
      type="button"
      class="w-full flex items-center justify-between gap-2 rounded-lg border border-gray-200 bg-gray-50/50 px-3.5 py-2.5 text-sm text-left transition focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-400 focus:border-brand-400"
      :class="[open ? 'ring-2 ring-brand-400 border-brand-400 bg-white' : '', selectedDate ? 'text-gray-900' : 'text-gray-400']"
      @click="toggle"
    >
      <span class="truncate">{{ displayLabel }}</span>
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" class="h-4 w-4 shrink-0 text-gray-400">
        <rect x="3" y="5" width="18" height="16" rx="2" stroke="currentColor" stroke-width="1.8" />
        <path d="M3 9h18M8 3v4M16 3v4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </svg>
    </button>

    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="open"
        class="absolute z-20 mt-1.5 w-72 rounded-xl border border-gray-100 bg-white p-3 shadow-lg shadow-gray-200/60"
      >
        <div class="flex items-center justify-between mb-2 px-1">
          <button type="button" class="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition" @click="viewMonth = viewMonth.minus({ months: 1 })">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" class="h-4 w-4">
              <path d="M15 18l-6-6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
          <span class="text-sm font-semibold text-gray-800">{{ viewMonth.toFormat('yyyy 年 LL 月') }}</span>
          <button type="button" class="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition" @click="viewMonth = viewMonth.plus({ months: 1 })">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" class="h-4 w-4">
              <path d="M9 18l6-6-6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </div>

        <div class="grid grid-cols-7 mb-1">
          <span v-for="d in weekDays" :key="d" class="h-7 flex items-center justify-center text-[11px] font-medium text-gray-400">
            {{ d }}
          </span>
        </div>

        <div class="grid grid-cols-7 gap-y-1">
          <button
            v-for="cell in calendarCells"
            :key="cell.date.toISODate()"
            type="button"
            :disabled="isDisabled(cell.date)"
            class="h-8 w-8 mx-auto rounded-lg text-xs flex items-center justify-center transition"
            :class="cellClass(cell)"
            @click="selectDay(cell.date)"
          >
            {{ cell.date.day }}
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { DateTime } from 'luxon'
import { useDismissable } from '../composables/useDismissable'

const props = defineProps({
  modelValue: { type: String, default: '' },
  min: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

const open = ref(false)
const rootEl = ref(null)
const today = DateTime.now().startOf('day')
const viewMonth = ref((props.modelValue ? DateTime.fromISO(props.modelValue) : today).startOf('month'))

const selectedDate = computed(() => (props.modelValue ? DateTime.fromISO(props.modelValue) : null))
const displayLabel = computed(() => (selectedDate.value ? selectedDate.value.toFormat('yyyy/LL/dd (ccc)') : '請選擇日期'))

const calendarCells = computed(() => {
  const startOfMonth = viewMonth.value
  const startWeekday = startOfMonth.weekday % 7
  const daysInMonth = startOfMonth.daysInMonth
  const totalCells = Math.ceil((startWeekday + daysInMonth) / 7) * 7
  const cells = []
  for (let i = 0; i < totalCells; i++) {
    const date = startOfMonth.plus({ days: i - startWeekday })
    cells.push({ date, inMonth: date.month === startOfMonth.month })
  }
  return cells
})

function isDisabled(date) {
  return Boolean(props.min && date < DateTime.fromISO(props.min).startOf('day'))
}

function cellClass(cell) {
  if (selectedDate.value && cell.date.hasSame(selectedDate.value, 'day')) {
    return 'bg-gradient-to-br from-brand-500 to-brand-600 text-white font-semibold shadow-sm shadow-brand-200'
  }
  if (isDisabled(cell.date)) {
    return 'text-gray-200 cursor-not-allowed'
  }
  const base = cell.inMonth ? 'text-gray-700 hover:bg-gray-100' : 'text-gray-300 hover:bg-gray-50'
  const isToday = cell.date.hasSame(today, 'day')
  return isToday ? `${base} ring-1 ring-inset ring-brand-300 font-semibold` : base
}

function toggle() {
  if (selectedDate.value) viewMonth.value = selectedDate.value.startOf('month')
  open.value = !open.value
}

function selectDay(date) {
  if (isDisabled(date)) return
  emit('update:modelValue', date.toISODate())
  open.value = false
}

useDismissable(rootEl, () => (open.value = false))

watch(
  () => props.modelValue,
  (v) => {
    if (v) viewMonth.value = DateTime.fromISO(v).startOf('month')
  },
)
</script>
