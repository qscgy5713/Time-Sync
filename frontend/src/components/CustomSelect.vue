<template>
  <div ref="rootEl" class="relative">
    <button
      type="button"
      class="w-full flex items-center justify-between gap-2 rounded-lg border border-gray-200 bg-gray-50/50 px-3.5 py-2.5 text-sm text-gray-900 text-left transition focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-400 focus:border-brand-400"
      :class="open ? 'ring-2 ring-brand-400 border-brand-400 bg-white' : ''"
      @click="open = !open"
    >
      <span class="truncate">{{ selectedLabel }}</span>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        class="h-4 w-4 shrink-0 text-gray-400 transition-transform"
        :class="open ? 'rotate-180' : ''"
      >
        <path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
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
      <ul
        v-if="open"
        class="absolute z-20 mt-1.5 max-h-64 w-full overflow-auto rounded-lg border border-gray-100 bg-white py-1 shadow-lg shadow-gray-200/60 focus:outline-none"
      >
        <li
          v-for="opt in options"
          :key="opt.value"
          class="flex items-center justify-between gap-2 px-3.5 py-2 text-sm cursor-pointer transition-colors"
          :class="opt.value === modelValue ? 'bg-brand-50 text-brand-700 font-medium' : 'text-gray-700 hover:bg-gray-50'"
          @click="select(opt.value)"
        >
          <span class="truncate">{{ opt.label }}</span>
          <svg
            v-if="opt.value === modelValue"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            class="h-4 w-4 shrink-0 text-brand-600"
          >
            <path d="M5 13l4 4L19 7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </li>
      </ul>
    </Transition>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useDismissable } from '../composables/useDismissable'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, required: true }, // [{ value, label }]
  placeholder: { type: String, default: '請選擇' },
})
const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const rootEl = ref(null)

const selectedLabel = computed(() => {
  const found = props.options.find((o) => o.value === props.modelValue)
  return found ? found.label : props.placeholder
})

function select(value) {
  emit('update:modelValue', value)
  open.value = false
}

useDismissable(rootEl, () => (open.value = false))
</script>
