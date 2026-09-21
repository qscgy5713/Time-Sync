<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <div v-if="loading" class="flex items-center gap-2 text-sm text-gray-400">
      <span class="h-3.5 w-3.5 rounded-full border-2 border-brand-300 border-t-brand-600 animate-spin"></span>
      載入中…
    </div>
    <div v-else-if="loadError" class="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{{ loadError }}</div>

    <div v-else>
      <div class="flex items-start justify-between gap-4 mb-1.5">
        <h1 class="text-2xl sm:text-3xl font-extrabold text-gray-900 tracking-tight">{{ event.title }}</h1>
        <button
          class="shrink-0 inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-gray-600 shadow-sm transition hover:border-brand-300 hover:text-brand-600"
          @click="copyLink"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" class="h-3.5 w-3.5">
            <rect x="9" y="9" width="11" height="11" rx="2" stroke="currentColor" stroke-width="1.8" />
            <path d="M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
          </svg>
          {{ copied ? '已複製連結 ✓' : '複製分享連結' }}
        </button>
      </div>
      <p v-if="event.description" class="text-sm text-gray-500 mb-3 whitespace-pre-line">{{ event.description }}</p>
      <p class="text-xs text-gray-400 mb-6">
        發起人時區：{{ timezoneLabel(event.timezone) }} · 你目前顯示的時間為本機時區 {{ timezoneLabel(localZone) }}
      </p>

      <div class="inline-flex p-1 mb-6 rounded-xl bg-gray-100/80 gap-1">
        <button
          class="px-4 py-1.5 text-sm font-semibold rounded-lg transition"
          :class="viewMode === 'fill' ? 'bg-white text-brand-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'"
          @click="viewMode = 'fill'"
        >
          填寫我的空檔
        </button>
        <button
          class="px-4 py-1.5 text-sm font-semibold rounded-lg transition"
          :class="viewMode === 'heatmap' ? 'bg-white text-brand-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'"
          @click="viewMode = 'heatmap'"
        >
          查看熱區圖
          <span class="ml-1 inline-flex items-center justify-center rounded-full bg-brand-100 text-brand-700 text-[11px] font-bold px-1.5 py-0.5">{{ event.participants.length }}</span>
        </button>
      </div>

      <div v-if="viewMode === 'fill'" class="rounded-2xl border border-gray-100 bg-white p-6 sm:p-8 shadow-sm shadow-gray-100">
        <div v-if="myParticipantId != null" class="mb-5 flex items-center justify-between gap-3 rounded-lg bg-brand-50 px-4 py-2.5 text-xs text-brand-700">
          <span>你先前已經填寫過，這裡會更新你原本的紀錄。</span>
          <button type="button" class="shrink-0 font-semibold underline decoration-dotted hover:text-brand-800" @click="forgetIdentity">
            不是你？改填新的
          </button>
        </div>

        <div class="mb-5">
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">你的名稱</label>
          <input
            v-model="participantName"
            type="text"
            maxlength="50"
            placeholder="輸入顯示名稱"
            class="w-full sm:w-64 rounded-lg border border-gray-200 bg-gray-50/50 px-3.5 py-2.5 text-sm text-gray-900 placeholder:text-gray-400 transition focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-400 focus:border-brand-400"
          />
        </div>

        <div class="mb-3 text-xs text-gray-400">
          拖曳滑鼠圈選你有空的時段（已選 <span class="font-semibold text-brand-600">{{ selection.count }}</span> 個）
        </div>
        <TimeGrid :days="days" :hours="hours" :key-for="keyFor" mode="select" />

        <div v-if="submitError" class="rounded-lg bg-red-50 px-4 py-2.5 text-sm text-red-600 mt-4">{{ submitError }}</div>

        <button
          class="mt-5 px-8 py-2.5 rounded-lg bg-gradient-to-br from-brand-500 to-brand-600 text-white text-sm font-semibold shadow-sm shadow-brand-200 transition hover:shadow-md hover:shadow-brand-200 hover:-translate-y-px active:translate-y-0 disabled:opacity-50 disabled:pointer-events-none"
          :disabled="submitting"
          @click="submitAvailability"
        >
          {{ submitting ? '送出中…' : myParticipantId != null ? '更新我的時間' : '送出我的時間' }}
        </button>
      </div>

      <div v-else class="rounded-2xl border border-gray-100 bg-white p-6 sm:p-8 shadow-sm shadow-gray-100">
        <div class="mb-4 text-xs text-gray-500 flex flex-wrap gap-4">
          <span class="inline-flex items-center gap-1.5"><i class="w-3 h-3 rounded-full bg-green-600 inline-block"></i>全員有空</span>
          <span class="inline-flex items-center gap-1.5"><i class="w-3 h-3 rounded-full bg-green-300 inline-block"></i>過半數有空</span>
          <span class="inline-flex items-center gap-1.5"><i class="w-3 h-3 rounded-full bg-yellow-100 border border-yellow-200 inline-block"></i>少數有空</span>
          <span class="inline-flex items-center gap-1.5"><i class="w-3 h-3 rounded-full bg-gray-100 border border-gray-200 inline-block"></i>無人有空</span>
        </div>
        <TimeGrid :days="days" :hours="hours" :key-for="keyFor" mode="display" :cell-class="heatmapClass" :cell-label="heatmapLabel" />

        <div class="mt-6 pt-6 border-t border-gray-100" v-if="event.participants.length">
          <h2 class="text-sm font-semibold text-gray-700 mb-3">已填寫名單</h2>
          <ul class="flex flex-wrap gap-2">
            <li
              v-for="p in event.participants"
              :key="p.id"
              class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 border border-gray-200 pl-1 pr-3 py-1 text-xs text-gray-600"
            >
              <span class="flex h-5 w-5 items-center justify-center rounded-full bg-brand-100 text-brand-700 text-[10px] font-bold">
                {{ p.name.slice(0, 1).toUpperCase() }}
              </span>
              {{ p.name }} · {{ p.available_option_ids.length }} 個時段
            </li>
          </ul>
        </div>
        <div v-else class="mt-6 pt-6 border-t border-gray-100 text-sm text-gray-400">目前還沒有人填寫時間。</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { DateTime } from 'luxon'
import TimeGrid from '../components/TimeGrid.vue'
import { useSelectionStore } from '../stores/selection'
import { api } from '../api'
import { timezoneLabel } from '../utils/timezone'
import { SLOT_MINUTES, minutesToHHmm } from '../utils/time'
import { getStoredParticipant, storeParticipant, clearStoredParticipant } from '../utils/identity'
import { recordMyEvent } from '../utils/myEvents'

const props = defineProps({ id: { type: String, required: true } })

const selection = useSelectionStore()
const localZone = Intl.DateTimeFormat().resolvedOptions().timeZone

const event = ref(null)
const loading = ref(true)
const loadError = ref('')
const viewMode = ref('fill')
const participantName = ref('')
const submitting = ref(false)
const submitError = ref('')
const copied = ref(false)
const myParticipantId = ref(null)
const myEditToken = ref(null)

// dayValueTHH:mm -> option id
let optionByKey = new Map()
const days = ref([])
const hours = ref([])

function buildGrid() {
  const localTimes = event.value.options.map((opt) => ({
    id: opt.id,
    local: DateTime.fromISO(opt.start_datetime, { zone: 'utc' }).setZone(localZone),
  }))

  const dayValues = [...new Set(localTimes.map((t) => t.local.toISODate()))].sort()
  const totalMinutesList = localTimes.map((t) => t.local.hour * 60 + t.local.minute)
  const minMinutes = totalMinutesList.reduce((a, b) => Math.min(a, b))
  const maxMinutes = totalMinutesList.reduce((a, b) => Math.max(a, b))

  days.value = dayValues.map((v) => ({
    value: v,
    label: DateTime.fromISO(v).toFormat('LL/dd (ccc)'),
  }))
  hours.value = []
  for (let m = minMinutes; m <= maxMinutes; m += SLOT_MINUTES) {
    const label = minutesToHHmm(m)
    hours.value.push({ value: label, label })
  }

  optionByKey = new Map()
  for (const t of localTimes) {
    const label = minutesToHHmm(t.local.hour * 60 + t.local.minute)
    const key = `${t.local.toISODate()}T${label}`
    // Same local wall-clock time can arise from two different UTC instants
    // across a DST fall-back transition; keep the earliest (chronologically
    // first, since options are sorted by start_datetime) instead of silently
    // overwriting it.
    if (!optionByKey.has(key)) {
      optionByKey.set(key, t.id)
    }
  }
}

function keyFor(dayValue, slotValue) {
  const optionId = optionByKey.get(`${dayValue}T${slotValue}`)
  return optionId != null ? String(optionId) : null
}

const voteCountByOption = computed(() => {
  const counts = new Map()
  if (!event.value) return counts
  for (const p of event.value.participants) {
    for (const optionId of p.available_option_ids) {
      counts.set(optionId, (counts.get(optionId) || 0) + 1)
    }
  }
  return counts
})

function votesFor(optionId) {
  return voteCountByOption.value.get(Number(optionId)) || 0
}

function heatmapClass(key) {
  const n = event.value.participants.length
  if (n === 0) return 'bg-gray-100 text-gray-400 border-gray-200'
  const x = votesFor(key)
  if (x === n) return 'bg-green-600 text-white border-green-700'
  if (x > n / 2) return 'bg-green-300 text-gray-800 border-green-400'
  if (x === 0) return 'bg-gray-100 text-gray-400 border-gray-200'
  return 'bg-yellow-100 text-gray-600 border-yellow-200'
}

function heatmapLabel(key) {
  return String(votesFor(key))
}

function applyStoredIdentity() {
  const stored = getStoredParticipant(props.id)
  const mine = stored && event.value.participants.find((p) => p.id === stored.id)
  if (mine) {
    myParticipantId.value = mine.id
    myEditToken.value = stored.editToken
    participantName.value = mine.name
    selection.reset(mine.available_option_ids.map(String))
  } else {
    myParticipantId.value = null
    myEditToken.value = null
    selection.reset()
  }
}

function forgetIdentity() {
  clearStoredParticipant(props.id)
  myParticipantId.value = null
  myEditToken.value = null
  participantName.value = ''
  selection.reset()
}

async function loadEvent() {
  loading.value = true
  loadError.value = ''
  try {
    event.value = await api.getEvent(props.id)
    buildGrid()
    applyStoredIdentity()
  } catch (err) {
    loadError.value = err.message || '找不到這個活動'
    selection.reset()
  } finally {
    loading.value = false
  }
}

async function submitAvailability() {
  submitError.value = ''
  if (!participantName.value.trim()) {
    submitError.value = '請輸入你的名稱'
    return
  }
  if (selection.count === 0) {
    submitError.value = '請至少圈選一個有空的時段'
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: participantName.value.trim(),
      available_option_ids: selection.selectedArray.map(Number),
    }
    if (myParticipantId.value != null) {
      await api.updateParticipant(props.id, myParticipantId.value, {
        ...payload,
        edit_token: myEditToken.value,
      })
    } else {
      const resp = await api.addParticipant(props.id, payload)
      storeParticipant(props.id, resp.id, resp.edit_token)
    }
    recordMyEvent(props.id, event.value.title, 'participant')
    await loadEvent()
    viewMode.value = 'heatmap'
  } catch (err) {
    submitError.value = err.message || '送出失敗，請稍後再試'
  } finally {
    submitting.value = false
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(window.location.href)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    // clipboard API unavailable; silently ignore
  }
}

watch(() => props.id, loadEvent)
onMounted(loadEvent)
</script>
