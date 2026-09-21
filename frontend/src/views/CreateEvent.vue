<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <div class="mb-8">
      <h1 class="text-3xl font-extrabold text-gray-900 tracking-tight mb-2">建立新活動</h1>
      <p class="text-sm text-gray-500">免登入、免註冊，建立候選時段後把連結分享給參與者即可。</p>
    </div>

    <form class="space-y-8 rounded-2xl border border-gray-100 bg-white p-6 sm:p-8 shadow-sm shadow-gray-100" @submit.prevent="submit">
      <div class="space-y-5">
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">活動名稱</label>
          <input
            v-model="form.title"
            type="text"
            maxlength="100"
            required
            placeholder="例如：Q3 開發進度同步會議"
            class="w-full rounded-lg border border-gray-200 bg-gray-50/50 px-3.5 py-2.5 text-sm text-gray-900 placeholder:text-gray-400 transition focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-400 focus:border-brand-400"
          />
        </div>

        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">補充說明（選填）</label>
          <textarea
            v-model="form.description"
            rows="2"
            placeholder="例如：請圈選下週有空的時段"
            class="w-full rounded-lg border border-gray-200 bg-gray-50/50 px-3.5 py-2.5 text-sm text-gray-900 placeholder:text-gray-400 transition focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-400 focus:border-brand-400"
          ></textarea>
        </div>

        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">你的時區</label>
          <CustomSelect v-model="form.timezone" :options="timezoneChoices" />
        </div>
      </div>

      <div class="border-t border-gray-100 pt-6 grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">開始日期</label>
          <DatePicker v-model="form.startDate" />
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">結束日期</label>
          <DatePicker v-model="form.endDate" :min="form.startDate" />
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">每日開始時間</label>
          <CustomSelect v-model="form.startHour" :options="startHourOptions" />
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">每日結束時間</label>
          <CustomSelect v-model="form.endHour" :options="endHourOptions" />
        </div>
      </div>

      <div class="border-t border-gray-100 pt-6">
        <div class="flex items-baseline justify-between mb-3">
          <label class="block text-sm font-semibold text-gray-700">圈選候選時段</label>
          <span class="text-xs text-gray-400">已選 <span class="font-semibold text-brand-600">{{ selection.count }}</span> 個時段 · 拖曳滑鼠可連續選取</span>
        </div>
        <div v-if="gridError" class="text-sm text-red-600 mb-2">{{ gridError }}</div>
        <TimeGrid
          v-else
          :days="days"
          :hours="hours"
          :key-for="keyFor"
          mode="select"
        />
      </div>

      <div v-if="errorMessage" class="rounded-lg bg-red-50 px-4 py-2.5 text-sm text-red-600">{{ errorMessage }}</div>

      <button
        type="submit"
        :disabled="submitting"
        class="w-full sm:w-auto px-8 py-2.5 rounded-lg bg-gradient-to-br from-brand-500 to-brand-600 text-white text-sm font-semibold shadow-sm shadow-brand-200 transition hover:shadow-md hover:shadow-brand-200 hover:-translate-y-px active:translate-y-0 disabled:opacity-50 disabled:pointer-events-none"
      >
        {{ submitting ? '建立中…' : '建立活動' }}
      </button>
    </form>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { DateTime } from 'luxon'
import TimeGrid from '../components/TimeGrid.vue'
import CustomSelect from '../components/CustomSelect.vue'
import DatePicker from '../components/DatePicker.vue'
import { useSelectionStore } from '../stores/selection'
import { api } from '../api'
import { TIMEZONE_OPTIONS } from '../utils/timezone'
import { SLOT_MINUTES, minutesToHHmm } from '../utils/time'
import { recordMyEvent } from '../utils/myEvents'

const router = useRouter()
const selection = useSelectionStore()

const detectedZone = Intl.DateTimeFormat().resolvedOptions().timeZone
const timezoneChoices = TIMEZONE_OPTIONS.some((tz) => tz.value === detectedZone)
  ? TIMEZONE_OPTIONS
  : [
      { value: detectedZone, label: `你的裝置時區 (${detectedZone})` },
      ...TIMEZONE_OPTIONS,
    ]

const today = DateTime.now()
const form = reactive({
  title: '',
  description: '',
  timezone: 'Asia/Taipei',
  startDate: today.toISODate(),
  endDate: today.plus({ days: 6 }).toISODate(),
  startHour: 9,
  endHour: 18,
})

function hourLabel(h) {
  return `${String(h).padStart(2, '0')}:00`
}
const startHourOptions = Array.from({ length: 24 }, (_, h) => ({ value: h, label: hourLabel(h) }))
const endHourOptions = Array.from({ length: 24 }, (_, h) => ({ value: h + 1, label: hourLabel(h + 1) }))

const errorMessage = ref('')
const submitting = ref(false)

const gridError = computed(() => {
  if (!form.startDate || !form.endDate) return '請選擇日期區間'
  if (form.startDate > form.endDate) return '結束日期不可早於開始日期'
  if (DateTime.fromISO(form.endDate).diff(DateTime.fromISO(form.startDate), 'days').days > 60) {
    return '日期區間請勿超過 60 天'
  }
  if (form.startHour >= form.endHour) return '結束時間必須晚於開始時間'
  return ''
})

const days = computed(() => {
  if (gridError.value) return []
  const start = DateTime.fromISO(form.startDate)
  const end = DateTime.fromISO(form.endDate)
  const list = []
  let cur = start
  while (cur <= end) {
    list.push({ value: cur.toISODate(), label: cur.toFormat('LL/dd (ccc)') })
    cur = cur.plus({ days: 1 })
  }
  return list
})

const hours = computed(() => {
  if (gridError.value) return []
  const list = []
  for (let m = form.startHour * 60; m < form.endHour * 60; m += SLOT_MINUTES) {
    const label = minutesToHHmm(m)
    list.push({ value: label, label })
  }
  return list
})

function keyFor(dayValue, slotValue) {
  return `${dayValue}T${slotValue}`
}

watch(
  () => [form.startDate, form.endDate, form.startHour, form.endHour],
  () => selection.reset(),
)

onMounted(() => selection.reset())

async function submit() {
  errorMessage.value = ''
  if (gridError.value) return
  if (selection.count === 0) {
    errorMessage.value = '請至少圈選一個候選時段'
    return
  }

  const starts = selection.selectedArray.map((key) =>
    DateTime.fromFormat(key, "yyyy-LL-dd'T'HH:mm", { zone: form.timezone }),
  )
  const invalidCount = starts.filter((start) => !start.isValid).length
  if (invalidCount > 0) {
    errorMessage.value = `有 ${invalidCount} 個時段因當地夏令時間調整而不存在，請重新圈選這些時段`
    return
  }

  const options = starts
    .map((start) => ({
      start_datetime: start.toUTC().toISO(),
      end_datetime: start.plus({ minutes: SLOT_MINUTES }).toUTC().toISO(),
    }))
    .sort((a, b) => (a.start_datetime < b.start_datetime ? -1 : 1))

  submitting.value = true
  try {
    const resp = await api.createEvent({
      title: form.title,
      description: form.description,
      timezone: form.timezone,
      options,
    })
    recordMyEvent(resp.id, form.title, 'organizer')
    router.push(`/event/${resp.id}`)
  } catch (err) {
    errorMessage.value = err.message || '建立活動失敗，請稍後再試'
  } finally {
    submitting.value = false
  }
}
</script>
