<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <div class="mb-8">
      <h1 class="text-3xl font-extrabold text-gray-900 tracking-tight mb-2">我的活動</h1>
      <p class="text-sm text-gray-500">
        這裡列出你在這個瀏覽器建立過或填寫過的活動（存在本機，不會同步到其他裝置）。
      </p>
    </div>

    <div v-if="events.length === 0" class="rounded-2xl border border-gray-100 bg-white p-10 text-center shadow-sm shadow-gray-100">
      <p class="text-sm text-gray-400 mb-4">目前還沒有任何紀錄。</p>
      <router-link
        to="/"
        class="inline-flex px-6 py-2.5 rounded-lg bg-gradient-to-br from-brand-500 to-brand-600 text-white text-sm font-semibold shadow-sm shadow-brand-200 transition hover:shadow-md hover:-translate-y-px"
      >
        建立新活動
      </router-link>
    </div>

    <ul v-else class="space-y-3">
      <li
        v-for="ev in events"
        :key="ev.id"
        class="flex items-center justify-between gap-4 rounded-2xl border border-gray-100 bg-white p-4 sm:p-5 shadow-sm shadow-gray-100"
      >
        <router-link :to="`/event/${ev.id}`" class="min-w-0 flex-1 group">
          <div class="flex items-center gap-2 mb-1">
            <span
              class="shrink-0 text-[11px] font-bold px-2 py-0.5 rounded-full"
              :class="ev.role === 'organizer' ? 'bg-brand-100 text-brand-700' : 'bg-gray-100 text-gray-500'"
            >
              {{ ev.role === 'organizer' ? '發起人' : '參與者' }}
            </span>
            <span class="text-xs text-gray-400">{{ formatDate(ev.visitedAt) }}</span>
          </div>
          <p class="truncate font-semibold text-gray-800 group-hover:text-brand-600 transition">
            {{ ev.title || '(未命名活動)' }}
          </p>
        </router-link>
        <button
          type="button"
          class="shrink-0 text-gray-300 hover:text-red-500 transition p-1"
          title="從清單移除"
          @click="remove(ev.id)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" class="h-4 w-4">
            <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          </svg>
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { DateTime } from 'luxon'
import { getMyEvents, removeMyEvent } from '../utils/myEvents'

const events = ref(getMyEvents())

function remove(id) {
  removeMyEvent(id)
  events.value = getMyEvents()
}

function formatDate(ts) {
  return DateTime.fromMillis(ts).toFormat('yyyy/LL/dd')
}
</script>
