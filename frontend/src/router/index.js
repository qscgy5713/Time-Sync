import { createRouter, createWebHistory } from 'vue-router'
import CreateEvent from '../views/CreateEvent.vue'
import EventPage from '../views/EventPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'create-event', component: CreateEvent },
    { path: '/event/:id', name: 'event', component: EventPage, props: true },
  ],
})
