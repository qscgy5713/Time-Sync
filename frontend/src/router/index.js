import { createRouter, createWebHistory } from 'vue-router'
import CreateEvent from '../views/CreateEvent.vue'
import EventPage from '../views/EventPage.vue'
import MyEvents from '../views/MyEvents.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'create-event', component: CreateEvent },
    { path: '/event/:id', name: 'event', component: EventPage, props: true },
    { path: '/my-events', name: 'my-events', component: MyEvents },
  ],
})
