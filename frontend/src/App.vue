<template>
  <div class="min-h-screen bg-gradient-to-b from-brand-50 via-white to-white">
    <header class="border-b border-gray-100 bg-white/80 backdrop-blur supports-[backdrop-filter]:bg-white/60 sticky top-0 z-10">
      <div class="max-w-3xl mx-auto px-6 py-4">
        <router-link to="/" class="inline-flex items-center gap-2 group">
          <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm shadow-brand-200 transition-transform group-hover:scale-105">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" class="h-4.5 w-4.5">
              <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8" />
              <path d="M12 7v5l3.5 2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </span>
          <span class="font-bold text-gray-900 tracking-tight text-lg">Time-Sync</span>
        </router-link>
      </div>
    </header>
    <router-view />
    <footer class="max-w-3xl mx-auto px-6 py-10 text-center text-xs text-gray-400">
      免登入．免註冊．時間對齊就這麼簡單
    </footer>
  </div>
</template>
